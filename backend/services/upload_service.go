package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// UploadSession tracks state for a single resumable upload.
type UploadSession struct {
	ID          string
	Filename    string
	SaveDir     string
	TotalSize   int64
	TotalChunks int
	TempDir     string
	CreatedAt   time.Time
}

// UploadService manages resumable chunked file uploads.
type UploadService struct {
	fileSvc  *FileService
	sessions map[string]*UploadSession
	mu       sync.Mutex
}

func NewUploadService(fileSvc *FileService) *UploadService {
	svc := &UploadService{
		fileSvc:  fileSvc,
		sessions: make(map[string]*UploadSession),
	}
	go svc.cleanLoop()
	return svc
}

// Init creates a new upload session and returns it.
func (s *UploadService) Init(filename, saveDir string, totalSize int64, totalChunks int) (*UploadSession, error) {
	if totalChunks <= 0 {
		return nil, fmt.Errorf("totalChunks must be > 0")
	}

	// Validate save directory
	if _, err := s.fileSvc.safePath(saveDir); err != nil {
		return nil, err
	}

	// Sanitize filename
	filename = filepath.Base(filename)
	if filename == "" || filename == "." {
		return nil, fmt.Errorf("invalid filename")
	}

	// Generate cryptographically random upload ID
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate upload ID: %v", err)
	}
	id := hex.EncodeToString(b)

	// Create isolated temp directory for this upload's chunks
	tempDir := filepath.Join(os.TempDir(), "tea-cloud-uploads", id)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}

	session := &UploadSession{
		ID:          id,
		Filename:    filename,
		SaveDir:     saveDir,
		TotalSize:   totalSize,
		TotalChunks: totalChunks,
		TempDir:     tempDir,
		CreatedAt:   time.Now(),
	}

	s.mu.Lock()
	s.sessions[id] = session
	s.mu.Unlock()

	return session, nil
}

// GetStatus returns the session and a list of already-uploaded chunk indices.
func (s *UploadService) GetStatus(id string) (*UploadSession, []int, error) {
	s.mu.Lock()
	session, ok := s.sessions[id]
	s.mu.Unlock()

	if !ok {
		return nil, nil, fmt.Errorf("upload session not found")
	}

	chunks, err := s.listChunks(session.TempDir)
	if err != nil {
		return session, []int{}, nil
	}
	return session, chunks, nil
}

// SaveChunk writes a chunk to the session's temp directory.
func (s *UploadService) SaveChunk(id string, index int, data io.Reader) error {
	s.mu.Lock()
	session, ok := s.sessions[id]
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("upload session not found")
	}
	if index < 0 || index >= session.TotalChunks {
		return fmt.Errorf("chunk index out of range: %d (total: %d)", index, session.TotalChunks)
	}

	chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk-%d", index))
	f, err := os.Create(chunkPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, data)
	return err
}

// Complete assembles all chunks into the final file and removes the session.
func (s *UploadService) Complete(id string) (*FileInfo, error) {
	s.mu.Lock()
	session, ok := s.sessions[id]
	s.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("upload session not found")
	}

	chunks, err := s.listChunks(session.TempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunks: %v", err)
	}
	if len(chunks) != session.TotalChunks {
		return nil, fmt.Errorf("incomplete upload: have %d chunks, need %d", len(chunks), session.TotalChunks)
	}

	sort.Ints(chunks)

	// Resolve and validate destination directory
	dirAbs, err := s.fileSvc.safePath(session.SaveDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dirAbs, 0755); err != nil {
		return nil, err
	}

	destPath := filepath.Join(dirAbs, session.Filename)

	// Extra path traversal check on final destination
	rootClean := filepath.Clean(s.fileSvc.Root) + string(os.PathSeparator)
	destClean := filepath.Clean(destPath)
	if destClean != filepath.Clean(s.fileSvc.Root) &&
		!strings.HasPrefix(destClean+string(os.PathSeparator), rootClean) {
		return nil, fmt.Errorf("forbidden: destination path escapes share root")
	}

	dest, err := os.Create(destPath)
	if err != nil {
		return nil, err
	}
	defer dest.Close()

	for _, idx := range chunks {
		chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("chunk-%d", idx))
		chunk, err := os.Open(chunkPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open chunk %d: %v", idx, err)
		}
		_, copyErr := io.Copy(dest, chunk)
		chunk.Close()
		if copyErr != nil {
			return nil, fmt.Errorf("failed to write chunk %d: %v", idx, copyErr)
		}
	}

	// Clean up temp dir and session
	os.RemoveAll(session.TempDir)
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()

	stat, err := os.Stat(destPath)
	if err != nil {
		return nil, err
	}
	rel, _ := filepath.Rel(s.fileSvc.Root, destPath)
	rel = filepath.ToSlash(rel)
	ext := strings.ToLower(filepath.Ext(session.Filename))

	return &FileInfo{
		Name:    session.Filename,
		Path:    rel,
		Size:    stat.Size(),
		IsDir:   false,
		ModTime: time.Now().UnixMilli(),
		MIME:    mimeByExt(ext),
		Ext:     ext,
	}, nil
}

// listChunks returns the indices of chunks already written to tempDir.
func (s *UploadService) listChunks(tempDir string) ([]int, error) {
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []int{}, nil
		}
		return nil, err
	}
	var chunks []int
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "chunk-") {
			idx, err := strconv.Atoi(strings.TrimPrefix(name, "chunk-"))
			if err == nil {
				chunks = append(chunks, idx)
			}
		}
	}
	return chunks, nil
}

func (s *UploadService) cleanLoop() {
	s.cleanSessions() // run once at startup to clean stale temp dirs
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		s.cleanSessions()
	}
}

func (s *UploadService) cleanSessions() {
	cutoff := time.Now().Add(-24 * time.Hour)

	// Collect expired sessions under lock, then release before doing disk I/O
	// to avoid blocking concurrent Init/SaveChunk requests.
	type expiredEntry struct {
		id      string
		tempDir string
	}
	s.mu.Lock()
	var toClean []expiredEntry
	for id, sess := range s.sessions {
		if sess.CreatedAt.Before(cutoff) {
			toClean = append(toClean, expiredEntry{id, sess.TempDir})
			delete(s.sessions, id)
		}
	}
	s.mu.Unlock()

	for _, entry := range toClean {
		os.RemoveAll(entry.tempDir)
		log.Printf("[INFO] cleaned expired upload session: %s", entry.id)
	}
}
