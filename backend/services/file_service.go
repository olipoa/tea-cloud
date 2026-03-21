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

// RenameFile renames a file or folder within the share root.
// newName must be a simple filename with no path separators.
func (s *FileService) RenameFile(relPath, newName string) (*FileInfo, error) {
	// Reject path separators in newName
	if strings.ContainsAny(newName, "/\\") {
		return nil, fmt.Errorf("invalid name: must not contain path separators")
	}
	newName = filepath.Base(newName)
	if newName == "" || newName == "." || newName == ".." {
		return nil, fmt.Errorf("invalid name")
	}

	abs, err := s.safePath(relPath)
	if err != nil {
		return nil, err
	}
	newAbs := filepath.Join(filepath.Dir(abs), newName)
	// Verify new path is still inside root
	if _, err := s.safePath(filepath.Join(filepath.Dir(relPath), newName)); err != nil {
		return nil, err
	}
	if err := os.Rename(abs, newAbs); err != nil {
		return nil, err
	}
	rel, _ := filepath.Rel(s.Root, newAbs)
	rel = filepath.ToSlash(rel)
	info, err := os.Stat(newAbs)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(newName))
	return &FileInfo{
		Name:    newName,
		Path:    rel,
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime().UnixMilli(),
		MIME:    mimeByExt(ext),
		Ext:     ext,
	}, nil
}

// CopyPath copies a file or directory (src) into the destination directory (destDir).
// If a name conflict exists, a "(1)" suffix is appended.
func (s *FileService) CopyPath(srcRel, destDirRel string) (*FileInfo, error) {
	srcAbs, err := s.safePath(srcRel)
	if err != nil {
		return nil, err
	}
	destDirAbs, err := s.safePath(destDirRel)
	if err != nil {
		return nil, err
	}
	srcInfo, err := os.Stat(srcAbs)
	if err != nil {
		return nil, err
	}
	// Resolve conflict name
	baseName := srcInfo.Name()
	destAbs := filepath.Join(destDirAbs, baseName)
	if _, statErr := os.Stat(destAbs); statErr == nil {
		ext := filepath.Ext(baseName)
		stem := strings.TrimSuffix(baseName, ext)
		destAbs = filepath.Join(destDirAbs, stem+" (1)"+ext)
	}
	if srcInfo.IsDir() {
		if err := copyDir(srcAbs, destAbs); err != nil {
			return nil, err
		}
	} else {
		if err := copyFile(srcAbs, destAbs); err != nil {
			return nil, err
		}
	}
	rel, _ := filepath.Rel(s.Root, destAbs)
	rel = filepath.ToSlash(rel)
	newInfo, err := os.Stat(destAbs)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(destAbs)
	ext := strings.ToLower(filepath.Ext(name))
	return &FileInfo{
		Name:    name,
		Path:    rel,
		Size:    newInfo.Size(),
		IsDir:   newInfo.IsDir(),
		ModTime: newInfo.ModTime().UnixMilli(),
		MIME:    mimeByExt(ext),
		Ext:     ext,
	}, nil
}

// MovePath moves src into the destination directory.
func (s *FileService) MovePath(srcRel, destDirRel string) (*FileInfo, error) {
	srcAbs, err := s.safePath(srcRel)
	if err != nil {
		return nil, err
	}
	destDirAbs, err := s.safePath(destDirRel)
	if err != nil {
		return nil, err
	}
	// Prevent moving a directory into itself or its subdirectory
	srcClean := filepath.Clean(srcAbs) + string(os.PathSeparator)
	destClean := filepath.Clean(destDirAbs)
	if strings.HasPrefix(destClean+string(os.PathSeparator), srcClean) || filepath.Clean(srcAbs) == destClean {
		return nil, fmt.Errorf("cannot move a directory into itself or its subdirectory")
	}

	baseName := filepath.Base(srcAbs)
	destAbs := filepath.Join(destDirAbs, baseName)

	if err := os.Rename(srcAbs, destAbs); err != nil {
		// Cross-device: copy then delete
		srcInfo, statErr := os.Stat(srcAbs)
		if statErr != nil {
			return nil, err
		}
		if srcInfo.IsDir() {
			if copyErr := copyDir(srcAbs, destAbs); copyErr != nil {
				return nil, copyErr
			}
		} else {
			if copyErr := copyFile(srcAbs, destAbs); copyErr != nil {
				return nil, copyErr
			}
		}
		os.RemoveAll(srcAbs)
	}

	rel, _ := filepath.Rel(s.Root, destAbs)
	rel = filepath.ToSlash(rel)
	newInfo, err := os.Stat(destAbs)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(destAbs)
	ext := strings.ToLower(filepath.Ext(name))
	return &FileInfo{
		Name:    name,
		Path:    rel,
		Size:    newInfo.Size(),
		IsDir:   newInfo.IsDir(),
		ModTime: newInfo.ModTime().UnixMilli(),
		MIME:    mimeByExt(ext),
		Ext:     ext,
	}, nil
}

// SearchFiles recursively searches for files whose names contain keyword (case-insensitive).
// Results are capped at 200 items.
func (s *FileService) SearchFiles(dirRel, keyword string) ([]FileInfo, error) {
	dirAbs, err := s.safePath(dirRel)
	if err != nil {
		return nil, err
	}
	keyword = strings.ToLower(keyword)
	var results []FileInfo
	_ = filepath.Walk(dirAbs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if len(results) >= 200 {
			return filepath.SkipAll
		}
		if filepath.Clean(path) == filepath.Clean(dirAbs) {
			return nil // skip the root itself
		}
		if strings.Contains(strings.ToLower(info.Name()), keyword) {
			rel, _ := filepath.Rel(s.Root, path)
			rel = filepath.ToSlash(rel)
			ext := strings.ToLower(filepath.Ext(info.Name()))
			results = append(results, FileInfo{
				Name:    info.Name(),
				Path:    rel,
				Size:    info.Size(),
				IsDir:   info.IsDir(),
				ModTime: info.ModTime().UnixMilli(),
				MIME:    mimeByExt(ext),
				Ext:     ext,
			})
		}
		return nil
	})
	return results, nil
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// copyDir recursively copies a directory tree from src to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}

// Ensure time import used (mimeByExt references aren't time-dependent but other funcs use it)
var _ = time.Now

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
