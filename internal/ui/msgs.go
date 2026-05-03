package ui

import (
	"hymn/internal/player"
	"hymn/internal/store"
	"hymn/internal/ytdlp"
)

// Player → UI
type PlayerEventMsg struct{ E player.Event }

// Async results
type SearchResultsMsg struct {
	Query   string
	Results []ytdlp.Result
	Err     error
}
type ThumbnailMsg struct {
	VideoID string
	Bytes   []byte
	Err     error
}

// State changes (re-read from store)
type QueueChangedMsg struct{ Tracks []store.Track }
type PlaylistsChangedMsg struct{}

// Commands from UI → app
type EnqueueAndPlayMsg struct{ R ytdlp.Result }
type PlayIndexMsg struct{ Index int }
type SeekMsg struct{ DeltaSec float64 }
type VolumeMsg struct{ Delta int }

type OpenAddToPlaylistMsg struct{ TrackID int64 }
type CreatePlaylistMsg struct{ Name string }
type LoadPlaylistMsg struct {
	ID      int64
	Replace bool
}

type initDoneMsg struct{}
type errMsg struct{ Err error }

type playStartedMsg struct {
	Track store.Track
	Index int
	Queue []store.Track
}
