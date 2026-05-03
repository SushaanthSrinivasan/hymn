package ui

type Bounds struct {
	X, Y, W, H int
}

func (b Bounds) Hit(mx, my int) bool {
	return mx >= b.X && mx < b.X+b.W && my >= b.Y && my < b.Y+b.H
}

// computeLayout splits the screen into art / queue / now-playing panes.
// Row 1 takes ~90% height; row 2 (now-playing) takes the rest (min 4 lines).
// Within row 1, art is roughly square, queue takes the remainder.
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

	// Art: square in cell aspect (cells are ~2:1 H:W).
	artW := row1H * 2
	if artW > width/2 {
		artW = width / 2
	}
	if artW < 16 {
		artW = 16
	}
	art = Bounds{X: 0, Y: 0, W: artW, H: row1H}
	queue = Bounds{X: artW, Y: 0, W: width - artW, H: row1H}
	np = Bounds{X: 0, Y: row1H, W: width, H: npH}
	return
}
