package player

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

func ipcPath(pid int) string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\hymn-mpv-` + strconv.Itoa(pid)
	}
	dir := os.TempDir()
	return filepath.Join(dir, "hymn-mpv-"+strconv.Itoa(pid)+".sock")
}

func mpvArgs(socket string) []string {
	return []string{
		"--idle=yes",
		"--no-video",
		"--no-terminal",
		"--input-ipc-server=" + socket,
		"--cache=yes",
		"--cache-secs=30",
		"--demuxer-max-bytes=64MiB",
		"--demuxer-max-back-bytes=32MiB",
		"--ytdl=yes",
		"--ytdl-format=bestaudio[ext=opus]/bestaudio",
		"--audio-display=no",
		"--volume=70",
		"--gapless-audio=yes",
		"--prefetch-playlist=yes",
	}
}

// spawnMPV starts mpv and returns the running cmd plus the socket path.
// Caller must wait for the IPC socket to become dialable separately.
func spawnMPV(mpvPath string) (*exec.Cmd, string, error) {
	if mpvPath == "" {
		p, err := exec.LookPath("mpv")
		if err != nil {
			return nil, "", fmt.Errorf("mpv not found in PATH: %w", err)
		}
		mpvPath = p
	}
	socket := ipcPath(os.Getpid())
	cmd := exec.Command(mpvPath, mpvArgs(socket)...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return nil, "", fmt.Errorf("start mpv: %w", err)
	}
	return cmd, socket, nil
}

// pollDial repeatedly tries to dial the IPC socket until it succeeds or
// the timeout expires. mpv takes a moment to bind the socket after spawn.
func pollDial(socket string, timeout time.Duration) (conn, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		c, err := dial(socket)
		if err == nil {
			return c, nil
		}
		lastErr = err
		time.Sleep(50 * time.Millisecond)
	}
	return nil, fmt.Errorf("ipc dial timeout: %w", lastErr)
}
