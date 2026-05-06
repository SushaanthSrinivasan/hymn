package theme

import (
	"image/color"
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

// RGBA parses a #rrggbb lipgloss.Color into an image/color.RGBA. Falls back
// to opaque black on parse failure (e.g. named colors or empty values).
func RGBA(c lipgloss.Color) color.RGBA {
	s := string(c)
	if len(s) == 7 && s[0] == '#' {
		r, e1 := strconv.ParseUint(s[1:3], 16, 8)
		g, e2 := strconv.ParseUint(s[3:5], 16, 8)
		b, e3 := strconv.ParseUint(s[5:7], 16, 8)
		if e1 == nil && e2 == nil && e3 == nil {
			return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
		}
	}
	return color.RGBA{A: 255}
}

// Catppuccin Mocha palette — single theme for v0.1.
type Theme struct {
	Base    lipgloss.Color
	Surface lipgloss.Color
	Text    lipgloss.Color
	Subtext lipgloss.Color
	Muted   lipgloss.Color
	Accent  lipgloss.Color
	Pink    lipgloss.Color
	Red     lipgloss.Color
	Green   lipgloss.Color
	Yellow  lipgloss.Color

	Title    lipgloss.Style
	Subtle   lipgloss.Style
	AccentS  lipgloss.Style
	Border   lipgloss.Style
	Selected lipgloss.Style
	Bar      lipgloss.Style
	BarFill  lipgloss.Style
}

func Mocha() Theme {
	t := Theme{
		Base:    lipgloss.Color("#1e1e2e"),
		Surface: lipgloss.Color("#313244"),
		Text:    lipgloss.Color("#cdd6f4"),
		Subtext: lipgloss.Color("#a6adc8"),
		Muted:   lipgloss.Color("#6c7086"),
		Accent:  lipgloss.Color("#cba6f7"),
		Pink:    lipgloss.Color("#f5c2e7"),
		Red:     lipgloss.Color("#f38ba8"),
		Green:   lipgloss.Color("#a6e3a1"),
		Yellow:  lipgloss.Color("#f9e2af"),
	}
	t.Title = lipgloss.NewStyle().Foreground(t.Accent).Bold(true)
	t.Subtle = lipgloss.NewStyle().Foreground(t.Muted)
	t.AccentS = lipgloss.NewStyle().Foreground(t.Accent)
	t.Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Surface)
	t.Selected = lipgloss.NewStyle().
		Foreground(t.Base).
		Background(t.Accent).
		Bold(true)
	t.Bar = lipgloss.NewStyle().Foreground(t.Surface)
	t.BarFill = lipgloss.NewStyle().Foreground(t.Accent)
	return t
}
