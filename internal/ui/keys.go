package ui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit       key.Binding
	TogglePlay key.Binding
	Next       key.Binding
	Prev       key.Binding
	SeekFwd    key.Binding
	SeekBack   key.Binding
	VolUp      key.Binding
	VolDown    key.Binding
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Search     key.Binding
	Playlists  key.Binding
	AddTo      key.Binding
	Delete     key.Binding
	Cancel     key.Binding
}

func DefaultKeys() KeyMap {
	return KeyMap{
		Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		TogglePlay: key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "play/pause")),
		Next:       key.NewBinding(key.WithKeys("n", "ctrl+right"), key.WithHelp("n", "next")),
		Prev:       key.NewBinding(key.WithKeys("p", "ctrl+left"), key.WithHelp("p", "prev")),
		SeekFwd:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→", "seek +5s")),
		SeekBack:   key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←", "seek -5s")),
		VolUp:      key.NewBinding(key.WithKeys("+", "="), key.WithHelp("+", "vol up")),
		VolDown:    key.NewBinding(key.WithKeys("-", "_"), key.WithHelp("-", "vol down")),
		Up:         key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑", "up")),
		Down:       key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓", "down")),
		Enter:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("⏎", "select")),
		Search:     key.NewBinding(key.WithKeys("/", "ctrl+f"), key.WithHelp("/", "search")),
		Playlists:  key.NewBinding(key.WithKeys("P"), key.WithHelp("P", "playlists")),
		AddTo:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add to playlist")),
		Delete:     key.NewBinding(key.WithKeys("x", "delete"), key.WithHelp("x", "remove")),
		Cancel:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}
}
