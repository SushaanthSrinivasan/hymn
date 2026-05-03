package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"hymn/internal/store"
	"hymn/internal/theme"
)

type queueModel struct {
	bounds  Bounds
	theme   theme.Theme
	tracks  []store.Track
	cursor  int
	playing int // index in tracks of currently playing track, -1 if none
}

func newQueueModel(t theme.Theme) queueModel {
	return queueModel{theme: t, playing: -1}
}

func (q *queueModel) SetTracks(tracks []store.Track) {
	q.tracks = tracks
	if q.cursor >= len(tracks) {
		q.cursor = len(tracks) - 1
	}
	if q.cursor < 0 {
		q.cursor = 0
	}
}

func (q *queueModel) MoveCursor(delta int) {
	if len(q.tracks) == 0 {
		q.cursor = 0
		return
	}
	q.cursor += delta
	if q.cursor < 0 {
		q.cursor = 0
	}
	if q.cursor >= len(q.tracks) {
		q.cursor = len(q.tracks) - 1
	}
}

func (q queueModel) Cursor() int { return q.cursor }

func (q queueModel) View() string {
	innerW := q.bounds.W - 4
	innerH := q.bounds.H - 2
	if innerW < 10 || innerH < 1 {
		return ""
	}
	header := q.theme.Title.Render("Queue")
	if len(q.tracks) == 0 {
		body := q.theme.Subtle.Render("(empty — press / to search)")
		content := lipgloss.JoinVertical(lipgloss.Left, header, "", body)
		return q.theme.Border.Width(q.bounds.W - 2).Height(q.bounds.H - 2).Render(content)
	}
	rows := make([]string, 0, len(q.tracks))
	rows = append(rows, header, "")
	for i, t := range q.tracks {
		marker := "  "
		if i == q.playing {
			marker = q.theme.AccentS.Render("▶ ")
		}
		title := t.Title
		dur := formatDur(t.Duration.Seconds())
		// title may need truncation
		room := innerW - len(marker) - len(dur) - 1
		if room < 4 {
			room = 4
		}
		title = truncate(title, room)
		line := marker + title + strings.Repeat(" ", max(1, innerW-len(marker)-room-len(dur))) + q.theme.Subtle.Render(dur)
		if i == q.cursor {
			line = q.theme.Selected.Render(padRight(line, innerW))
		}
		rows = append(rows, line)
		if len(rows) >= innerH {
			break
		}
	}
	body := strings.Join(rows, "\n")
	return q.theme.Border.Width(q.bounds.W - 2).Height(q.bounds.H - 2).Render(body)
}

func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	if w <= 1 {
		return string(r[:w])
	}
	return string(r[:w-1]) + "…"
}

func padRight(s string, w int) string {
	// crude; we don't strip ANSI, so this is best-effort.
	return s
}

func formatDur(sec float64) string {
	if sec <= 0 {
		return "--:--"
	}
	total := int(sec)
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return spad(h) + ":" + spad(m) + ":" + spad(s)
	}
	return spad(m) + ":" + spad(s)
}

func spad(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := "0123456789"
	out := ""
	for n > 0 {
		out = string(digits[n%10]) + out
		n /= 10
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
