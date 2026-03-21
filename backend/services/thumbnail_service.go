package services

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var videoExts = map[string]bool{
	".mp4": true, ".mkv": true, ".webm": true,
	".avi": true, ".mov": true, ".flv": true,
	".wmv": true, ".m4v": true, ".ts": true,
}

// ThumbnailService generates and caches video thumbnails using ffmpeg.
type ThumbnailService struct {
	fileSvc  *FileService
	cacheDir string
}

func NewThumbnailService(fileSvc *FileService) *ThumbnailService {
	cacheDir := filepath.Join(os.TempDir(), "tea-cloud-thumbs")
	os.MkdirAll(cacheDir, 0755)
	svc := &ThumbnailService{fileSvc: fileSvc, cacheDir: cacheDir}
	go svc.cleanOldCache()
	return svc
}

// IsVideoExt returns true if the extension belongs to a known video format.
func (s *ThumbnailService) IsVideoExt(ext string) bool {
	return videoExts[strings.ToLower(ext)]
}

// Generate extracts a JPEG thumbnail for a video file.
// Returns the JPEG bytes, or an error if ffmpeg is unavailable or extraction fails.
func (s *ThumbnailService) Generate(relPath string) ([]byte, error) {
	absPath, err := s.fileSvc.safePath(relPath)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(relPath))
	if !videoExts[ext] {
		return nil, fmt.Errorf("not a supported video file")
	}

	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found in PATH")
	}

	stat, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	// Cache key: file path + mtime
	key := fmt.Sprintf("%s:%d", relPath, stat.ModTime().UnixNano())
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
	cacheFile := filepath.Join(s.cacheDir, hash+".jpg")

	if data, err := os.ReadFile(cacheFile); err == nil {
		return data, nil
	}

	// Try to grab a frame at 5 seconds (works for most videos)
	args := []string{"-ss", "5", "-i", absPath, "-vframes", "1", "-f", "image2", "-q:v", "4", "-an", "-y", cacheFile}
	if err := exec.Command(ffmpeg, args...).Run(); err != nil {
		// Fallback: grab first frame (for short videos < 5s)
		args2 := []string{"-i", absPath, "-vframes", "1", "-f", "image2", "-q:v", "4", "-an", "-y", cacheFile}
		if err2 := exec.Command(ffmpeg, args2...).Run(); err2 != nil {
			return nil, fmt.Errorf("ffmpeg thumbnail extraction failed: %v", err2)
		}
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *ThumbnailService) cleanOldCache() {
	entries, err := os.ReadDir(s.cacheDir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(filepath.Join(s.cacheDir, e.Name())); err == nil {
				log.Printf("[INFO] cleaned old thumbnail cache: %s", e.Name())
			}
		}
	}
}
