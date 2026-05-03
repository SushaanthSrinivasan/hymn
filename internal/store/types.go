package store

import "time"

type Track struct {
	ID           int64
	VideoID      string
	Title        string
	Artist       string
	Duration     time.Duration
	ThumbnailURL string
}

type CacheEntry struct {
	VideoID    string
	Path       string
	SizeBytes  int64
	LastUsedAt time.Time
}

type Playlist struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}
