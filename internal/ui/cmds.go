package ui

import (
	"context"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"hymn/internal/cache"
	"hymn/internal/player"
	"hymn/internal/store"
	"hymn/internal/ytdlp"
)

// waitForPlayerEvent pumps one event from the player into the tea Update loop.
// Update must re-issue this command after handling each PlayerEventMsg.
func waitForPlayerEvent(p player.Client) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-p.Events()
		if !ok {
			return nil
		}
		return PlayerEventMsg{E: ev}
	}
}

func loadQueueCmd(db *store.Store) tea.Cmd {
	return func() tea.Msg {
		q, err := db.QueueList()
		if err != nil {
			return errMsg{Err: err}
		}
		return QueueChangedMsg{Tracks: q}
	}
}

func enqueueAndPlayCmd(db *store.Store, p player.Client, cacheDir string, r ytdlp.Result) tea.Cmd {
	return func() tea.Msg {
		t := store.Track{
			VideoID:      r.VideoID,
			Title:        r.Title,
			Artist:       r.Artist,
			Duration:     r.Duration,
			ThumbnailURL: r.ThumbnailURL,
		}
		id, err := db.UpsertTrack(t)
		if err != nil {
			return errMsg{Err: err}
		}
		pos, err := db.QueueAppend(id)
		if err != nil {
			return errMsg{Err: err}
		}
		t.ID = id
		if err := loadTrack(db, p, cacheDir, t, r.URL); err != nil {
			return errMsg{Err: err}
		}
		queue, err := db.QueueList()
		if err != nil {
			return errMsg{Err: err}
		}
		_ = db.HistoryAdd(id)
		return playStartedMsg{Track: t, Index: int(pos), Queue: queue}
	}
}

func fetchThumbnailCmd(videoID, url string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b, err := ytdlp.FetchThumbnail(ctx, url)
		return ThumbnailMsg{VideoID: videoID, Bytes: b, Err: err}
	}
}

func playQueueIndexCmd(db *store.Store, p player.Client, cacheDir string, idx int) tea.Cmd {
	return func() tea.Msg {
		queue, err := db.QueueList()
		if err != nil {
			return errMsg{Err: err}
		}
		if idx < 0 || idx >= len(queue) {
			return nil
		}
		t := queue[idx]
		if err := loadTrack(db, p, cacheDir, t, ""); err != nil {
			return errMsg{Err: err}
		}
		_ = db.HistoryAdd(t.ID)
		return playStartedMsg{Track: t, Index: idx, Queue: queue}
	}
}

// loadTrack picks cache-hit (local file, instant start) over cache-miss (stream
// from YT and stream-record to disk for next time). networkURL may be empty —
// in that case it's reconstructed from VideoID.
func loadTrack(db *store.Store, p player.Client, cacheDir string, t store.Track, networkURL string) error {
	cached, _ := db.CacheGet(t.VideoID)
	if cached != nil {
		if _, err := os.Stat(cached.Path); err == nil {
			if err := p.Loadfile(cached.Path); err != nil {
				return err
			}
			_ = p.Pause(false)
			_ = db.CacheTouch(t.VideoID)
			return nil
		}
	}
	if cacheDir != "" {
		_ = p.SetProperty("stream-record", cache.PathFor(cacheDir, t.VideoID))
	}
	if networkURL == "" {
		networkURL = "https://www.youtube.com/watch?v=" + t.VideoID
	}
	if err := p.Loadfile(networkURL); err != nil {
		return err
	}
	_ = p.Pause(false)
	return nil
}

// loadPlaylistCmd appends or replaces the queue with a playlist's tracks. If
// replace is set and the playlist has at least one track, the first track is
// also loaded into the player.
func loadPlaylistCmd(db *store.Store, p player.Client, cacheDir string, playlistID int64, replace bool) tea.Cmd {
	return func() tea.Msg {
		tracks, err := db.ListPlaylistTracks(playlistID)
		if err != nil {
			return errMsg{Err: err}
		}
		if len(tracks) == 0 {
			return nil
		}
		ids := make([]int64, len(tracks))
		for i, t := range tracks {
			ids[i] = t.ID
		}
		if replace {
			if err := db.QueueReplace(ids); err != nil {
				return errMsg{Err: err}
			}
			queue, err := db.QueueList()
			if err != nil {
				return errMsg{Err: err}
			}
			t := queue[0]
			if err := loadTrack(db, p, cacheDir, t, ""); err != nil {
				return errMsg{Err: err}
			}
			_ = db.HistoryAdd(t.ID)
			return playStartedMsg{Track: t, Index: 0, Queue: queue}
		}
		// append: insert each track at the end
		for _, id := range ids {
			if _, err := db.QueueAppend(id); err != nil {
				return errMsg{Err: err}
			}
		}
		queue, err := db.QueueList()
		if err != nil {
			return errMsg{Err: err}
		}
		return QueueChangedMsg{Tracks: queue}
	}
}

// recordCacheCmd is fired on EOF to register the freshly recorded file in the cache index.
func recordCacheCmd(db *store.Store, cacheDir, videoID string) tea.Cmd {
	return func() tea.Msg {
		if cacheDir == "" || videoID == "" {
			return nil
		}
		path := cache.PathFor(cacheDir, videoID)
		st, err := os.Stat(path)
		if err != nil {
			return nil
		}
		_ = db.CacheUpsert(store.CacheEntry{
			VideoID:    videoID,
			Path:       path,
			SizeBytes:  st.Size(),
			LastUsedAt: time.Now(),
		})
		return nil
	}
}
