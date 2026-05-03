package ui

import (
	"strings"

	"hymn/internal/store"
	"hymn/internal/theme"
)

type nowPlayingModel struct {
	bounds Bounds
	theme  theme.Theme
	track  *store.Track
	pos    float64
	dur    float64
	paused bool
	volume int
}

func newNowPlaying(t theme.Theme) nowPlayingModel {
	return nowPlayingModel{theme: t, volume: 70}
}

// Transport buttons are rendered at fixed offsets so click hitboxes are
// deterministic. Inside the bordered box, content origin is (1, 1).
// Button row sits on content line 3 (after title/artist/blank).
const (
	btnPrevX = 0
	btnPrevW = 4 // "[<<]"
	btnPlayX = 5
	btnPlayW = 4 // "[||]" / "[ >]"
	btnNextX = 10
	btnNextW = 4 // "[>>]"
	btnRow   = 3
)

type Button int

const (
	BtnNone Button = iota
	BtnPrev
	BtnPlay
	BtnNext
)

func (n nowPlayingModel) HitButton(mx, my int) Button {
	// translate to inside-box coords
	ix := mx - (n.bounds.X + 1)
	iy := my - (n.bounds.Y + 1)
	if iy != btnRow {
		return BtnNone
	}
	switch {
	case ix >= btnPrevX && ix < btnPrevX+btnPrevW:
		return BtnPrev
	case ix >= btnPlayX && ix < btnPlayX+btnPlayW:
		return BtnPlay
	case ix >= btnNextX && ix < btnNextX+btnNextW:
		return BtnNext
	}
	return BtnNone
}

func (n nowPlayingModel) View() string {
	innerW := n.bounds.W - 2
	innerH := n.bounds.H - 2
	if innerW < 20 || innerH < 4 {
		return ""
	}
	title := "(nothing playing)"
	artist := ""
	if n.track != nil {
		title = n.track.Title
		artist = n.track.Artist
	}
	titleLine := n.theme.Title.Render(truncate(title, innerW))
	artistLine := n.theme.Subtle.Render(truncate(artist, innerW))

	playBtn := "[ >]"
	if !n.paused && n.track != nil {
		playBtn = "[||]"
	}
	transport := n.theme.AccentS.Render("[<<]") + " " +
		n.theme.AccentS.Render(playBtn) + " " +
		n.theme.AccentS.Render("[>>]")

	timeStr := formatDur(n.pos) + " / " + formatDur(n.dur)
	volStr := "vol " + itoa(n.volume)
	// transport visual width is 14 (fixed: btnNextX+btnNextW)
	barW := innerW - 14 - 1 - len(timeStr) - 2 - len(volStr) - 1
	if barW < 8 {
		barW = 8
	}
	bar := renderBar(n.theme, n.pos, n.dur, barW)
	bottom := transport + " " + bar + "  " + n.theme.Subtle.Render(timeStr) + " " + n.theme.Subtle.Render(volStr)

	body := strings.Join([]string{titleLine, artistLine, "", bottom}, "\n")
	return n.theme.Border.Width(innerW).Height(innerH).Render(body)
}

func renderBar(t theme.Theme, pos, dur float64, w int) string {
	if w < 4 {
		return ""
	}
	frac := 0.0
	if dur > 0 {
		frac = pos / dur
		if frac < 0 {
			frac = 0
		}
		if frac > 1 {
			frac = 1
		}
	}
	filled := int(float64(w) * frac)
	if filled > w {
		filled = w
	}
	rest := w - filled
	return t.BarFill.Render(strings.Repeat("━", filled)) + t.Bar.Render(strings.Repeat("━", rest))
}

