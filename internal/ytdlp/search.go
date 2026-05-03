package ytdlp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

type rawThumb struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type rawEntry struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	URL        string     `json:"url"`
	Duration   *float64   `json:"duration"`
	Uploader   string     `json:"uploader"`
	Channel    string     `json:"channel"`
	LiveStatus string     `json:"live_status"`
	IsLive     *bool      `json:"is_live"`
	Thumbnails []rawThumb `json:"thumbnails"`
	Thumbnail  string     `json:"thumbnail"`
}

type rawSearch struct {
	Entries []rawEntry `json:"entries"`
}

// Search runs yt-dlp -J --flat-playlist "ytsearch<n>:<query>" and returns
// the parsed results. Livestreams are filtered out (v0.1 scope).
func (c *Client) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	if limit <= 0 {
		limit = 20
	}
	target := "ytsearch" + strconv.Itoa(limit) + ":" + query
	cmd := exec.CommandContext(ctx, c.binary, "-J", "--flat-playlist", "--no-warnings", target)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("yt-dlp: %s", string(ee.Stderr))
		}
		return nil, err
	}
	var raw rawSearch
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("decode yt-dlp output: %w", err)
	}
	results := make([]Result, 0, len(raw.Entries))
	for _, e := range raw.Entries {
		if isLive(e) {
			continue
		}
		if e.ID == "" {
			continue
		}
		r := Result{
			VideoID:      e.ID,
			URL:          fallback(e.URL, "https://www.youtube.com/watch?v="+e.ID),
			Title:        e.Title,
			Artist:       fallback(e.Channel, e.Uploader),
			ThumbnailURL: pickThumb(e),
		}
		if e.Duration != nil {
			r.Duration = time.Duration(*e.Duration * float64(time.Second))
		}
		results = append(results, r)
	}
	return results, nil
}

func isLive(e rawEntry) bool {
	if e.IsLive != nil && *e.IsLive {
		return true
	}
	switch e.LiveStatus {
	case "is_live", "is_upcoming":
		return true
	}
	return false
}

func fallback(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func pickThumb(e rawEntry) string {
	if len(e.Thumbnails) > 0 {
		// Highest resolution wins; thumbnails are usually ascending.
		best := e.Thumbnails[0]
		for _, t := range e.Thumbnails[1:] {
			if t.Width*t.Height > best.Width*best.Height {
				best = t
			}
		}
		if best.URL != "" {
			return best.URL
		}
	}
	return e.Thumbnail
}
