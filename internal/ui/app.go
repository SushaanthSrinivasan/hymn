package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	hymnimg "hymn/internal/image"
	"hymn/internal/player"
	"hymn/internal/store"
	"hymn/internal/theme"
	"hymn/internal/ytdlp"
)

type Model struct {
	db       *store.Store
	player   player.Client
	yt       *ytdlp.Client
	cacheDir string
	theme    theme.Theme
	keys     KeyMap

	width, height int

	art   artModel
	queue queueModel
	np    nowPlayingModel

	search     searchModel
	showSearch bool

	playlists     playlistModel
	showPlaylists bool

	playingIdx int // index into queue.tracks, -1 if none

	err error
}

func NewModel(db *store.Store, p player.Client, yt *ytdlp.Client, cacheDir string) Model {
	t := theme.Mocha()
	r := hymnimg.NewRenderer()
	return Model{
		db:         db,
		player:     p,
		yt:         yt,
		cacheDir:   cacheDir,
		theme:      t,
		keys:       DefaultKeys(),
		art:        newArtModel(t, r),
		queue:      newQueueModel(t),
		np:         newNowPlaying(t),
		search:     newSearchModel(t, yt),
		playlists:  newPlaylistModel(t, db),
		playingIdx: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitForPlayerEvent(m.player),
		loadQueueCmd(m.db),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		artB, qB, npB := computeLayout(m.width, m.height)
		m.art.SetBounds(artB)
		m.queue.bounds = qB
		m.np.bounds = npB

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.showSearch {
			var cmd tea.Cmd
			var closeIt bool
			m.search, cmd, closeIt = m.search.Update(msg)
			if closeIt {
				m.search.Close()
				m.showSearch = false
			}
			return m, cmd
		}
		if m.showPlaylists {
			var cmd tea.Cmd
			var closeIt bool
			m.playlists, cmd, closeIt = m.playlists.Update(msg)
			if closeIt {
				m.showPlaylists = false
			}
			return m, cmd
		}
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Search):
			m.search.Open()
			m.showSearch = true
			return m, nil
		case key.Matches(msg, m.keys.Playlists):
			m.playlists.OpenBrowse()
			m.showPlaylists = true
			return m, nil
		case key.Matches(msg, m.keys.AddTo):
			if m.queue.Cursor() < len(m.queue.tracks) {
				tid := m.queue.tracks[m.queue.Cursor()].ID
				m.playlists.OpenAdd(tid)
				m.showPlaylists = true
			}
			return m, nil
		case key.Matches(msg, m.keys.TogglePlay):
			cmds = append(cmds, func() tea.Msg { _ = m.player.TogglePause(); return nil })
		case key.Matches(msg, m.keys.SeekFwd):
			cmds = append(cmds, func() tea.Msg { _ = m.player.Seek(5); return nil })
		case key.Matches(msg, m.keys.SeekBack):
			cmds = append(cmds, func() tea.Msg { _ = m.player.Seek(-5); return nil })
		case key.Matches(msg, m.keys.VolUp):
			m.np.volume = clampVol(m.np.volume + 5)
			v := m.np.volume
			cmds = append(cmds, func() tea.Msg { _ = m.player.SetVolume(v); return nil })
		case key.Matches(msg, m.keys.VolDown):
			m.np.volume = clampVol(m.np.volume - 5)
			v := m.np.volume
			cmds = append(cmds, func() tea.Msg { _ = m.player.SetVolume(v); return nil })
		case key.Matches(msg, m.keys.Down):
			m.queue.MoveCursor(1)
		case key.Matches(msg, m.keys.Up):
			m.queue.MoveCursor(-1)
		case key.Matches(msg, m.keys.Enter):
			if m.queue.Cursor() < len(m.queue.tracks) {
				idx := m.queue.Cursor()
				cmds = append(cmds, playQueueIndexCmd(m.db, m.player, m.cacheDir, idx))
			}
		}

	case tea.MouseMsg:
		if m.showSearch {
			break
		}
		if cmd := m.handleMouse(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

	case PlayerEventMsg:
		switch msg.E.Kind {
		case player.EventTimePos:
			m.np.pos = msg.E.Float
		case player.EventDuration:
			m.np.dur = msg.E.Float
		case player.EventPause:
			m.np.paused = msg.E.Bool
		case player.EventVolume:
			m.np.volume = int(msg.E.Float + 0.5)
		case player.EventMediaTitle:
			if m.np.track != nil {
				m.np.track.Title = msg.E.String
			}
		case player.EventEOF:
			// register the just-recorded file in the cache index
			if m.np.track != nil {
				cmds = append(cmds, recordCacheCmd(m.db, m.cacheDir, m.np.track.VideoID))
			}
			// auto-advance: play next item if any
			next := m.playingIdx + 1
			if next >= 0 && next < len(m.queue.tracks) {
				cmds = append(cmds, playQueueIndexCmd(m.db, m.player, m.cacheDir, next))
			}
		}
		cmds = append(cmds, waitForPlayerEvent(m.player))

	case QueueChangedMsg:
		m.queue.SetTracks(msg.Tracks)
		m.queue.playing = m.playingIdx

	case EnqueueAndPlayMsg:
		cmds = append(cmds, enqueueAndPlayCmd(m.db, m.player, m.cacheDir, msg.R))

	case LoadPlaylistMsg:
		cmds = append(cmds, loadPlaylistCmd(m.db, m.player, m.cacheDir, msg.ID, msg.Replace))

	case PlaylistsChangedMsg:
		// modal handles its own reload

	case playStartedMsg:
		m.playingIdx = msg.Index
		m.queue.playing = msg.Index
		t := msg.Track
		m.np.track = &t
		m.np.pos = 0
		m.np.dur = float64(t.Duration / time.Second)
		// kick off thumbnail fetch for the new track
		m.art.Clear()
		if t.ThumbnailURL != "" {
			cmds = append(cmds, fetchThumbnailCmd(t.VideoID, t.ThumbnailURL))
		}

	case ThumbnailMsg:
		if msg.Err != nil || len(msg.Bytes) == 0 {
			break
		}
		// only apply if it still matches the currently playing track
		if m.np.track != nil && m.np.track.VideoID == msg.VideoID {
			m.art.SetThumb(msg.VideoID, msg.Bytes)
		}

	case errMsg:
		m.err = msg.Err

	default:
		// forward to whichever modal is open (textinput.Blink etc.)
		if m.showSearch {
			var cmd tea.Cmd
			var closeIt bool
			m.search, cmd, closeIt = m.search.Update(msg)
			if closeIt {
				m.search.Close()
				m.showSearch = false
			}
			return m, cmd
		}
		if m.showPlaylists {
			var cmd tea.Cmd
			var closeIt bool
			m.playlists, cmd, closeIt = m.playlists.Update(msg)
			if closeIt {
				m.showPlaylists = false
			}
			return m, cmd
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, m.art.View(), m.queue.View())
	row2 := m.np.View()
	base := lipgloss.JoinVertical(lipgloss.Left, row1, row2)
	if m.showSearch {
		base = m.search.View(m.width, m.height)
	} else if m.showPlaylists {
		base = m.playlists.View(m.width, m.height)
	}
	lines := strings.Split(base, "\n")
	if len(lines) > m.height {
		lines = lines[:m.height]
	}
	return strings.Join(lines, "\n")
}

func clampVol(v int) int {
	if v < 0 {
		return 0
	}
	if v > 130 {
		return 130
	}
	return v
}
