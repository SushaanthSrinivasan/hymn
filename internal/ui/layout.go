package ui

type Bounds struct {
	X, Y, W, H int
}

func (b Bounds) Hit(mx, my int) bool {
	return mx >= b.X && mx < b.X+b.W && my >= b.Y && my < b.Y+b.H
}

// computeLayout splits the screen into art / queue / now-playing panes.
// Row 1 takes ~90% height; row 2 (now-playing) takes the rest (min 6 lines).
// The art panel is always visually square: artW = 2 × artH in cells, since
// terminal cells are roughly 1:2 W:H. When width-capped, artH shrinks to
// preserve squareness; the empty column below art is padded by lipgloss.
func computeLayout(width, height int) (art, queue, np Bounds) {
	if width < 40 {
		width = 40
	}
	if height < 12 {
		height = 12
	}
	npH := height / 8
	if npH < 6 {
		npH = 6
	}
	if npH > 8 {
		npH = 8
	}
	row1H := height - npH

	// Pick the largest visually-square panel that fits both width/2 and
	// row1H. Cells are ~1:2 W:H, so visual square means artW = 2 × artH.
	artW := row1H * 2
	if artW > width/2 {
		artW = width / 2
	}
	if artW < 16 {
		artW = 16
	}
	artH := artW / 2
	if artH > row1H {
		artH = row1H
	}
	art = Bounds{X: 0, Y: 0, W: artW, H: artH}
	queue = Bounds{X: artW, Y: 0, W: width - artW, H: row1H}
	np = Bounds{X: 0, Y: row1H, W: width, H: npH}
	return
}
