package ui

import (
	"context"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"hymn/internal/theme"
	"hymn/internal/ytdlp"
)

type searchModel struct {
	theme   theme.Theme
	yt      *ytdlp.Client
	input   textinput.Model
	results []ytdlp.Result
	cursor  int
	loading bool
	err     error
}

func newSearchModel(t theme.Theme, yt *ytdlp.Client) searchModel {
	in := textinput.New()
	in.Placeholder = "search youtube…"
	in.Prompt = "› "
	in.CharLimit = 200
	in.Width = 40
	in.PromptStyle = t.AccentS
	in.Cursor.Style = t.AccentS
	return searchModel{theme: t, yt: yt, input: in}
}

func (s *searchModel) Open() {
	s.input.SetValue("")
	s.input.Focus()
	s.results = nil
	s.cursor = 0
	s.loading = false
	s.err = nil
}

func (s *searchModel) Close() {
	s.input.Blur()
}

// Update handles messages while the modal is open. Returns the updated
// modal, an optional command, and whether the modal wants to be closed.
func (s searchModel) Update(msg tea.Msg) (searchModel, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return s, nil, true
		case "enter":
			if s.input.Focused() {
				q := strings.TrimSpace(s.input.Value())
				if q == "" {
					return s, nil, false
				}
				s.loading = true
				s.results = nil
				s.err = nil
				return s, doSearchCmd(s.yt, q), false
			}
			// On result: emit EnqueueAndPlay
			if s.cursor >= 0 && s.cursor < len(s.results) {
				r := s.results[s.cursor]
				return s, func() tea.Msg { return EnqueueAndPlayMsg{R: r} }, true
			}
		case "tab":
			// toggle focus between input and results
			if s.input.Focused() {
				s.input.Blur()
			} else {
				s.input.Focus()
			}
			return s, nil, false
		case "up", "ctrl+p":
			if !s.input.Focused() && s.cursor > 0 {
				s.cursor--
			}
			return s, nil, false
		case "down", "ctrl+n":
			if !s.input.Focused() && s.cursor < len(s.results)-1 {
				s.cursor++
			}
			return s, nil, false
		}
	case SearchResultsMsg:
		s.loading = false
		if msg.Err != nil {
			s.err = msg.Err
			return s, nil, false
		}
		s.results = msg.Results
		s.cursor = 0
		s.input.Blur()
		return s, nil, false
	}
	if s.input.Focused() {
		var cmd tea.Cmd
		s.input, cmd = s.input.Update(msg)
		return s, cmd, false
	}
	return s, nil, false
}

func (s searchModel) View(width, height int) string {
	w := width - 8
	if w > 80 {
		w = 80
	}
	if w < 30 {
		w = 30
	}
	body := s.input.View()
	body += "\n\n"
	switch {
	case s.loading:
		body += s.theme.Subtle.Render("searching…")
	case s.err != nil:
		body += s.theme.AccentS.Render("error: " + s.err.Error())
	case len(s.results) == 0 && s.input.Value() != "":
		body += s.theme.Subtle.Render("no results")
	default:
		max := height - 8
		if max < 5 {
			max = 5
		}
		end := len(s.results)
		if end > max {
			end = max
		}
		for i := 0; i < end; i++ {
			r := s.results[i]
			line := truncate(r.Title, w-12) + "  " + s.theme.Subtle.Render(r.DurationString())
			if i == s.cursor && !s.input.Focused() {
				line = s.theme.Selected.Render(line)
			}
			body += line + "\n"
		}
		if len(s.results) > end {
			body += s.theme.Subtle.Render("…")
		}
	}

	hint := s.theme.Subtle.Render("⏎ search/select   tab focus   esc cancel")
	box := s.theme.Border.
		BorderForeground(s.theme.Accent).
		Width(w).
		Render(s.theme.Title.Render("Search") + "\n\n" + body + "\n" + hint)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func doSearchCmd(yt *ytdlp.Client, query string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		res, err := yt.Search(ctx, query, 20)
		return SearchResultsMsg{Query: query, Results: res, Err: err}
	}
}
