package services

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileInfo represents metadata about a file or directory.
type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`    // relative to share root
	Size    int64  `json:"size"`
	IsDir   bool   `json:"isDir"`
	ModTime int64  `json:"modTime"` // Unix timestamp (ms)
	MIME    string `json:"mime"`
	Ext     string `json:"ext"`
}

// FileService handles all file operations within the share directory.
type FileService struct {
	Root string
}

func NewFileService(root string) *FileService {
	return &FileService{Root: root}
}

// safePath resolves a relative user-provided path to an absolute path,
// rejecting any path that would escape the share root (path traversal prevention).
func (s *FileService) safePath(rel string) (string, error) {
	// Clean the path and make it relative
	rel = filepath.FromSlash(rel)
	rel = filepath.Clean(rel)

	// Build candidate absolute path
	abs := filepath.Join(s.Root, rel)

	// Ensure the resolved path is still inside Root
	rootClean := filepath.Clean(s.Root) + string(os.PathSeparator)
	absClean := filepath.Clean(abs)

	if absClean != filepath.Clean(s.Root) && !strings.HasPrefix(absClean+string(os.PathSeparator), rootClean) {
		return "", fmt.Errorf("forbidden: path escapes share root")
	}
	return absClean, nil
}

// ListDir lists the contents of a directory relative to the share root.
func (s *FileService) ListDir(relPath string) ([]FileInfo, error) {
	abs, err := s.safePath(relPath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}

	result := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		// Build relative path from root
		fullAbs := filepath.Join(abs, e.Name())
		rel, _ := filepath.Rel(s.Root, fullAbs)
		rel = filepath.ToSlash(rel)

		mimeType := ""
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if !e.IsDir() {
			mimeType = mimeByExt(ext)
		}

		result = append(result, FileInfo{
			Name:    e.Name(),
			Path:    rel,
			Size:    info.Size(),
			IsDir:   e.IsDir(),
			ModTime: info.ModTime().UnixMilli(),
			MIME:    mimeType,
			Ext:     ext,
		})
	}
	return result, nil
}

// GetFileInfo returns metadata for a single file.
func (s *FileService) GetFileInfo(relPath string) (*FileInfo, error) {
	abs, err := s.safePath(relPath)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(info.Name()))
	rel, _ := filepath.Rel(s.Root, abs)
	rel = filepath.ToSlash(rel)
	return &FileInfo{
		Name:    info.Name(),
		Path:    rel,
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime().UnixMilli(),
		MIME:    mimeByExt(ext),
		Ext:     ext,
	}, nil
}

// OpenFile opens a file for reading; caller is responsible for closing it.
func (s *FileService) OpenFile(relPath string) (*os.File, *FileInfo, error) {
	info, err := s.GetFileInfo(relPath)
	if err != nil {
		return nil, nil, err
	}
	if info.IsDir {
		return nil, nil, fmt.Errorf("cannot open a directory")
	}
	abs, _ := s.safePath(relPath)
	f, err := os.Open(abs)
	if err != nil {
		return nil, nil, err
	}
	return f, info, nil
}

// SaveFile saves uploaded data to a file inside relDir with the given filename.
func (s *FileService) SaveFile(relDir, filename string, src io.Reader) (*FileInfo, error) {
	dirAbs, err := s.safePath(relDir)
	if err != nil {
		return nil, err
	}
	// Sanitize filename: no path separators allowed
	filename = filepath.Base(filename)
	if filename == "" || filename == "." {
		return nil, fmt.Errorf("invalid filename")
	}

	destAbs := filepath.Join(dirAbs, filename)
	// Re-verify the destination is inside root
	rootClean := filepath.Clean(s.Root) + string(os.PathSeparator)
	if destClean := filepath.Clean(destAbs); destClean != filepath.Clean(s.Root) && !strings.HasPrefix(destClean+string(os.PathSeparator), rootClean) {
		return nil, fmt.Errorf("forbidden: destination escapes share root")
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(dirAbs, 0755); err != nil {
		return nil, err
	}

	f, err := os.Create(destAbs)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := io.Copy(f, src); err != nil {
		return nil, err
	}

	stat, _ := os.Stat(destAbs)
	rel, _ := filepath.Rel(s.Root, destAbs)
	rel = filepath.ToSlash(rel)
	ext := strings.ToLower(filepath.Ext(filename))
	return &FileInfo{
		Name:    filename,
		Path:    rel,
		Size:    stat.Size(),
		IsDir:   false,
		ModTime: time.Now().UnixMilli(),
		MIME:    mimeByExt(ext),
		Ext:     ext,
	}, nil
}

// DeletePath deletes a file or directory (recursively) at the given relative path.
func (s *FileService) DeletePath(relPath string) error {
	abs, err := s.safePath(relPath)
	if err != nil {
		return err
	}
	// Do not allow deleting the root itself
	if filepath.Clean(abs) == filepath.Clean(s.Root) {
		return fmt.Errorf("cannot delete share root")
	}
	return os.RemoveAll(abs)
}

// MkDir creates a directory at the given relative path.
func (s *FileService) MkDir(relPath string) error {
	abs, err := s.safePath(relPath)
	if err != nil {
		return err
	}
	return os.MkdirAll(abs, 0755)
}

// AbsPath returns the validated absolute path for use with http.ServeContent.
func (s *FileService) AbsPath(relPath string) (string, error) {
	return s.safePath(relPath)
}

// mimeByExt returns a MIME type for common extensions, falling back to
// the standard library's mime package.
func mimeByExt(ext string) string {
	known := map[string]string{
		".mp4":  "video/mp4",
		".mkv":  "video/x-matroska",
		".webm": "video/webm",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".mp3":  "audio/mpeg",
		".flac": "audio/flac",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".aac":  "audio/aac",
		".m4a":  "audio/mp4",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".bmp":  "image/bmp",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".md":   "text/markdown",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
		".7z":   "application/x-7z-compressed",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
	}
	if m, ok := known[ext]; ok {
		return m
	}
	if m := mime.TypeByExtension(ext); m != "" {
		return m
	}
	return "application/octet-stream"
}
