package ytdlp

import (
	"fmt"
	"os/exec"
)

type Client struct {
	binary string
}

func New(binary string) (*Client, error) {
	if binary == "" {
		p, err := exec.LookPath("yt-dlp")
		if err != nil {
			return nil, fmt.Errorf("yt-dlp not found in PATH: %w", err)
		}
		binary = p
	}
	return &Client{binary: binary}, nil
}
