package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleMouse routes a tea.MouseMsg to the right pane and returns commands
// for the resulting state changes. Modal-open is handled by the caller.
func (m *Model) handleMouse(msg tea.MouseMsg) tea.Cmd {
	switch msg.Action {
	case tea.MouseActionPress:
		switch msg.Button {
		case tea.MouseButtonLeft:
			return m.handleLeftClick(msg.X, msg.Y)
		case tea.MouseButtonWheelUp:
			if m.queue.bounds.Hit(msg.X, msg.Y) {
				m.queue.MoveCursor(-1)
			}
		case tea.MouseButtonWheelDown:
			if m.queue.bounds.Hit(msg.X, msg.Y) {
				m.queue.MoveCursor(1)
			}
		}
	}
	return nil
}

func (m *Model) handleLeftClick(x, y int) tea.Cmd {
	switch {
	case m.queue.bounds.Hit(x, y):
		// row index inside the queue body. Layout:
		//   border (y=0) | header (y=1) | blank (y=2) | row 0 (y=3)...
		row := y - m.queue.bounds.Y - 3
		if row >= 0 && row < len(m.queue.tracks) {
			m.queue.cursor = row
			return playQueueIndexCmd(m.db, m.player, m.cacheDir, row)
		}
	case m.np.bounds.Hit(x, y):
		btn := m.np.HitButton(x, y)
		switch btn {
		case BtnPrev:
			if m.playingIdx > 0 {
				return playQueueIndexCmd(m.db, m.player, m.cacheDir, m.playingIdx-1)
			}
		case BtnPlay:
			p := m.player
			return func() tea.Msg { _ = p.TogglePause(); return nil }
		case BtnNext:
			if m.playingIdx >= 0 && m.playingIdx+1 < len(m.queue.tracks) {
				return playQueueIndexCmd(m.db, m.player, m.cacheDir, m.playingIdx+1)
			}
		}
	}
	return nil
}

