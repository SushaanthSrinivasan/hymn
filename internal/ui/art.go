package ui

import (
	"github.com/charmbracelet/lipgloss"

	hymnimg "hymn/internal/image"
	"hymn/internal/theme"
)

type artModel struct {
	bounds   Bounds
	theme    theme.Theme
	renderer *hymnimg.Renderer

	rawBytes []byte // last fetched (kept so we can re-render on resize)
	videoID  string // owner of rawBytes — drop stale thumbs
	rendered string // cached render fitted to current inner box
	cachedW  int
	cachedH  int
}

func newArtModel(t theme.Theme, r *hymnimg.Renderer) artModel {
	return artModel{theme: t, renderer: r}
}

func (a *artModel) SetBounds(b Bounds) {
	old := a.bounds
	a.bounds = b
	if old.W != b.W || old.H != b.H {
		a.rerender()
	}
}

func (a *artModel) SetThumb(videoID string, data []byte) {
	a.videoID = videoID
	a.rawBytes = data
	a.rerender()
}

// Clear drops the current thumbnail (used on track change before new fetch).
func (a *artModel) Clear() {
	a.videoID = ""
	a.rawBytes = nil
	a.rendered = ""
}

func (a *artModel) rerender() {
	innerW := a.bounds.W - 2
	innerH := a.bounds.H - 2
	if innerW < 4 || innerH < 2 || a.renderer == nil || len(a.rawBytes) == 0 {
		a.rendered = ""
		return
	}
	s, err := a.renderer.Render(a.rawBytes, innerW, innerH)
	if err != nil {
		a.rendered = ""
		return
	}
	a.rendered = s
	a.cachedW = innerW
	a.cachedH = innerH
}

func (a artModel) View() string {
	innerW := a.bounds.W - 2
	innerH := a.bounds.H - 2
	if innerW < 4 || innerH < 2 {
		return ""
	}
	if a.rendered == "" {
		placeholder := lipgloss.NewStyle().
			Foreground(a.theme.Muted).
			Width(innerW).
			Height(innerH).
			Align(lipgloss.Center, lipgloss.Center).
			Render("♪")
		return a.theme.Border.Render(placeholder)
	}
	body := hymnimg.Pad(a.rendered, innerH)
	return a.theme.Border.Render(body)
}
