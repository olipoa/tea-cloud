package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

type Config struct {
	Port      int
	ShareDir  string
	NodeName  string
	MaxUpload int64 // bytes
}

var defaultShareDir = getDefaultShareDir()

func getDefaultShareDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "TeaCloud")
	}
	return filepath.Join(home, "tea-cloud")
}

func Load() *Config {
	port := 8080
	shareDir := defaultShareDir
	nodeName, _ := os.Hostname()

	if v := os.Getenv("TEA_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 65535 {
			port = n
		}
	}
	if v := os.Getenv("TEA_SHARE_DIR"); v != "" {
		shareDir = v
	}
	if v := os.Getenv("TEA_NODE_NAME"); v != "" {
		nodeName = v
	}

	// Ensure share directory exists
	if err := os.MkdirAll(shareDir, 0755); err != nil {
		shareDir = "."
	}

	abs, err := filepath.Abs(shareDir)
	if err == nil {
		shareDir = abs
	}

	return &Config{
		Port:      port,
		ShareDir:  shareDir,
		NodeName:  nodeName,
		MaxUpload: 10 * 1024 * 1024 * 1024, // 10 GB
	}
}
