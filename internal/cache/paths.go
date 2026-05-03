package cache

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultDir returns <UserCacheDir>/hymn/audio. The directory is created if missing.
func DefaultDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("user cache dir: %w", err)
	}
	dir := filepath.Join(base, "hymn", "audio")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir cache: %w", err)
	}
	return dir, nil
}

// PathFor returns <dir>/<videoID>.cache. The extension is fixed: mpv detects
// audio format from content, so the suffix is just a marker for hymn's bookkeeping.
func PathFor(dir, videoID string) string {
	return filepath.Join(dir, videoID+".cache")
}
