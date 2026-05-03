package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"hymn/internal/store"
	"hymn/internal/theme"
)

type plMode int

const (
	plBrowse  plMode = iota // listing playlists, just for navigation
	plView                  // viewing tracks of one playlist
	plAddPick               // 'a' key flow: pick playlist (or create new) to add a track to
	plAddNew                // entering name for a new playlist (then add the held track)
)

type playlistModel struct {
	theme theme.Theme
	db    *store.Store

	mode      plMode
	playlists []store.Playlist
	tracks    []store.Track
	cursor    int

	currentPL  *store.Playlist
	addTrackID int64 // valid when mode is plAddPick / plAddNew

	input textinput.Model
	err   error
}

func newPlaylistModel(t theme.Theme, db *store.Store) playlistModel {
	in := textinput.New()
	in.Placeholder = "playlist name"
	in.Prompt = "› "
	in.CharLimit = 80
	in.Width = 40
	in.PromptStyle = t.AccentS
	in.Cursor.Style = t.AccentS
	return playlistModel{theme: t, db: db, input: in}
}

func (p *playlistModel) OpenBrowse() {
	p.mode = plBrowse
	p.cursor = 0
	p.tracks = nil
	p.currentPL = nil
	p.addTrackID = 0
	p.err = nil
	p.input.Blur()
	p.reload()
}

func (p *playlistModel) OpenAdd(trackID int64) {
	p.mode = plAddPick
	p.cursor = 0
	p.tracks = nil
	p.currentPL = nil
	p.addTrackID = trackID
	p.err = nil
	p.input.Blur()
	p.reload()
}

func (p *playlistModel) reload() {
	pls, err := p.db.ListPlaylists()
	if err != nil {
		p.err = err
		return
	}
	p.playlists = pls
}

// Update handles messages while modal is open. Returns updated modal,
// optional cmd, and whether modal wants to close.
func (p playlistModel) Update(msg tea.Msg) (playlistModel, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if p.mode == plView {
				p.mode = plBrowse
				p.cursor = 0
				return p, nil, false
			}
			if p.mode == plAddNew {
				p.mode = plAddPick
				p.input.Blur()
				return p, nil, false
			}
			return p, nil, true
		case "up", "k":
			if p.input.Focused() {
				break
			}
			if p.cursor > 0 {
				p.cursor--
			}
			return p, nil, false
		case "down", "j":
			if p.input.Focused() {
				break
			}
			if p.cursor < p.maxCursor() {
				p.cursor++
			}
			return p, nil, false
		case "enter", "shift+enter":
			return p.handleEnter(msg.String() == "shift+enter")
		}
	case PlaylistsChangedMsg:
		p.reload()
		return p, nil, false
	}
	if p.input.Focused() {
		var cmd tea.Cmd
		p.input, cmd = p.input.Update(msg)
		return p, cmd, false
	}
	return p, nil, false
}

func (p playlistModel) handleEnter(replace bool) (playlistModel, tea.Cmd, bool) {
	switch p.mode {
	case plBrowse:
		if len(p.playlists) == 0 {
			return p, nil, false
		}
		pl := p.playlists[p.cursor]
		tracks, err := p.db.ListPlaylistTracks(pl.ID)
		if err != nil {
			p.err = err
			return p, nil, false
		}
		p.currentPL = &pl
		p.tracks = tracks
		p.mode = plView
		p.cursor = 0
		return p, nil, false

	case plView:
		if p.currentPL == nil {
			return p, nil, true
		}
		ids := make([]int64, len(p.tracks))
		for i, t := range p.tracks {
			ids[i] = t.ID
		}
		return p, func() tea.Msg {
			return LoadPlaylistMsg{ID: p.currentPL.ID, Replace: replace}
		}, true

	case plAddPick:
		// cursor 0 = "+ new playlist", cursor 1+ = existing playlists
		if p.cursor == 0 {
			p.mode = plAddNew
			p.input.SetValue("")
			p.input.Focus()
			return p, textinput.Blink, false
		}
		idx := p.cursor - 1
		if idx >= 0 && idx < len(p.playlists) {
			pl := p.playlists[idx]
			tid := p.addTrackID
			return p, func() tea.Msg {
				_ = p.db.AddTrackToPlaylist(pl.ID, tid)
				return PlaylistsChangedMsg{}
			}, true
		}

	case plAddNew:
		name := strings.TrimSpace(p.input.Value())
		if name == "" {
			return p, nil, false
		}
		tid := p.addTrackID
		db := p.db
		return p, func() tea.Msg {
			existing, _ := db.FindPlaylistByName(name)
			var id int64
			if existing != nil {
				id = existing.ID
			} else {
				newID, err := db.CreatePlaylist(name)
				if err != nil {
					return errMsg{Err: err}
				}
				id = newID
			}
			_ = db.AddTrackToPlaylist(id, tid)
			return PlaylistsChangedMsg{}
		}, true
	}
	return p, nil, false
}

func (p playlistModel) maxCursor() int {
	switch p.mode {
	case plBrowse:
		if len(p.playlists) == 0 {
			return 0
		}
		return len(p.playlists) - 1
	case plView:
		if len(p.tracks) == 0 {
			return 0
		}
		return len(p.tracks) - 1
	case plAddPick:
		return len(p.playlists) // 1 extra for "+ new"
	}
	return 0
}

func (p playlistModel) View(width, height int) string {
	w := width - 8
	if w > 70 {
		w = 70
	}
	if w < 30 {
		w = 30
	}
	var title, body string
	switch p.mode {
	case plBrowse:
		title = "Playlists"
		body = p.renderPlaylistList(w)
		body += "\n" + p.theme.Subtle.Render("⏎ open   esc close")
	case plView:
		name := "?"
		if p.currentPL != nil {
			name = p.currentPL.Name
		}
		title = "Playlist · " + name
		body = p.renderTrackList(w)
		body += "\n" + p.theme.Subtle.Render("⏎ append to queue   shift+⏎ replace queue   esc back")
	case plAddPick:
		title = "Add to playlist"
		body = p.renderAddPick(w)
		body += "\n" + p.theme.Subtle.Render("⏎ pick   esc cancel")
	case plAddNew:
		title = "New playlist"
		body = p.input.View() + "\n\n" + p.theme.Subtle.Render("⏎ create   esc back")
	}

	if p.err != nil {
		body += "\n" + p.theme.AccentS.Render("error: "+p.err.Error())
	}

	box := p.theme.Border.
		BorderForeground(p.theme.Accent).
		Width(w).
		Render(p.theme.Title.Render(title) + "\n\n" + body)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func (p playlistModel) renderPlaylistList(w int) string {
	if len(p.playlists) == 0 {
		return p.theme.Subtle.Render("(no playlists yet — press 'a' on a track to create one)")
	}
	var lines []string
	for i, pl := range p.playlists {
		line := pl.Name
		if i == p.cursor {
			line = p.theme.Selected.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (p playlistModel) renderTrackList(w int) string {
	if len(p.tracks) == 0 {
		return p.theme.Subtle.Render("(empty)")
	}
	var lines []string
	for i, t := range p.tracks {
		line := truncate(t.Title, w-12) + "  " + p.theme.Subtle.Render(formatDur(t.Duration.Seconds()))
		if i == p.cursor {
			line = p.theme.Selected.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (p playlistModel) renderAddPick(w int) string {
	var lines []string
	first := "+ new playlist"
	if p.cursor == 0 {
		first = p.theme.Selected.Render(first)
	}
	lines = append(lines, first)
	for i, pl := range p.playlists {
		line := pl.Name
		if p.cursor == i+1 {
			line = p.theme.Selected.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
