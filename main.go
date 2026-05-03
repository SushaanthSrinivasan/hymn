package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"

	"hymn/internal/cache"
	"hymn/internal/config"
	"hymn/internal/player"
	"hymn/internal/store"
	"hymn/internal/ui"
	"hymn/internal/ytdlp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(1)
	}
	if err := checkDeps(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dbPath, err := defaultDBPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config dir:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir:", err)
		os.Exit(1)
	}
	db, err := store.Open(dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "store:", err)
		os.Exit(1)
	}
	defer db.Close()

	cacheDir := cfg.Cache.Dir
	if cacheDir == "" {
		cacheDir, err = cache.DefaultDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "cache:", err)
			os.Exit(1)
		}
	} else if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "cache mkdir:", err)
		os.Exit(1)
	}
	capBytes := int64(cfg.Cache.MaxSizeMB) * 1024 * 1024
	if err := cache.Evict(db, capBytes); err != nil {
		fmt.Fprintln(os.Stderr, "cache evict:", err)
	}

	pl, err := player.Spawn(cfg.Paths.Mpv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mpv:", err)
		os.Exit(1)
	}
	defer pl.Close()

	yt, err := ytdlp.New(cfg.Paths.YtDlp)
	if err != nil {
		fmt.Fprintln(os.Stderr, "yt-dlp:", err)
		os.Exit(1)
	}

	m := ui.NewModel(db, pl, yt, cacheDir)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "tea:", err)
		os.Exit(1)
	}
}

func checkDeps(cfg config.Config) error {
	mpvBin := cfg.Paths.Mpv
	if mpvBin == "" {
		if _, err := exec.LookPath("mpv"); err != nil {
			return fmt.Errorf("mpv not found in PATH.\nInstall hint: %s", installHint("mpv"))
		}
	} else if _, err := os.Stat(mpvBin); err != nil {
		return fmt.Errorf("mpv path %q from config is not accessible: %w", mpvBin, err)
	}
	ytBin := cfg.Paths.YtDlp
	if ytBin == "" {
		if _, err := exec.LookPath("yt-dlp"); err != nil {
			return fmt.Errorf("yt-dlp not found in PATH.\nInstall hint: %s", installHint("yt-dlp"))
		}
	} else if _, err := os.Stat(ytBin); err != nil {
		return fmt.Errorf("yt-dlp path %q from config is not accessible: %w", ytBin, err)
	}
	return nil
}

func installHint(name string) string {
	switch runtime.GOOS {
	case "windows":
		if name == "mpv" {
			return "winget install shinchiro.mpv"
		}
		return "winget install yt-dlp.yt-dlp"
	case "darwin":
		return "brew install " + name
	default:
		return "apt/dnf/pacman install " + name
	}
}

func defaultDBPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "hymn", "hymn.db"), nil
}
