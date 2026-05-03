package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Playback struct {
	Volume      int `toml:"volume"`
	SeekStepSec int `toml:"seek_step_sec"`
}

type Cache struct {
	MaxSizeMB int    `toml:"max_size_mb"`
	Dir       string `toml:"dir"`
}

type UI struct {
	Theme string `toml:"theme"`
	Mouse bool   `toml:"mouse"`
}

type Search struct {
	Results int `toml:"results"`
}

type Paths struct {
	Mpv   string `toml:"mpv"`
	YtDlp string `toml:"yt_dlp"`
}

type Config struct {
	Playback Playback `toml:"playback"`
	Cache    Cache    `toml:"cache"`
	UI       UI       `toml:"ui"`
	Search   Search   `toml:"search"`
	Paths    Paths    `toml:"paths"`
}

func Default() Config {
	return Config{
		Playback: Playback{Volume: 70, SeekStepSec: 5},
		Cache:    Cache{MaxSizeMB: 2048, Dir: ""},
		UI:       UI{Theme: "mocha", Mouse: true},
		Search:   Search{Results: 20},
		Paths:    Paths{},
	}
}

// Path returns the config file path, creating parent dirs.
func Path() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	dir := filepath.Join(base, "hymn")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir config: %w", err)
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load reads the config file, writing the default first if it doesn't exist.
func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}
	cfg := Default()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := Save(cfg); err != nil {
			return cfg, err
		}
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return cfg, fmt.Errorf("decode config: %w", err)
	}
	return cfg, nil
}

func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := toml.NewEncoder(f)
	enc.Indent = ""
	return enc.Encode(cfg)
}
