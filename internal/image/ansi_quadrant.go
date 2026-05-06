package image

import (
	stdimg "image"
	"image/color"
	"strconv"
	"strings"
)

// RenderQuadrant renders src as ANSI quadrant blocks — 4 sub-pixels per cell
// in a 2×2 grid (▘▝▀▖▌▞▛▗▚▐▜▄▙▟█). 2× horizontal/vertical fidelity over
// half-block. Universally supported (BMP characters).
func RenderQuadrant(src stdimg.Image, w, h int, bg color.Color) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	// Sub-pixel visible aspect: cell is 1:2 W:H, sub-pixel = 0.5 × 1 = 1:2.
	img := FitForGrid(src, 2*w, 2*h, 1, 2, bg)

	var sb strings.Builder
	sb.Grow(w * h * 30)

	for cy := 0; cy < h; cy++ {
		for cx := 0; cx < w; cx++ {
			var pix [4]color.Color
			for i := 0; i < 4; i++ {
				row := i / 2
				col := i % 2
				pix[i] = img.At(2*cx+col, 2*cy+row)
			}
			fg, bgC, pattern := partition2(pix[:])
			writeCell(&sb, fg, bgC, quadrantRune(pattern))
		}
		sb.WriteString("\x1b[0m\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

// quadrantRune maps a 4-bit pattern to a quadrant glyph.
// Bit 0=TL, 1=TR, 2=BL, 3=BR.
func quadrantRune(p uint8) rune {
	return [16]rune{
		' ', '▘', '▝', '▀',
		'▖', '▌', '▞', '▛',
		'▗', '▚', '▐', '▜',
		'▄', '▙', '▟', '█',
	}[p&0x0F]
}

// partition2 splits pixels into "fg" (lit) and "bg" (dark) groups by luminance
// threshold (cell average), returns averaged colors per group and the bit
// pattern with bit i set when pixel i is fg.
func partition2(pix []color.Color) (fg [3]uint8, bg [3]uint8, pattern uint8) {
	n := len(pix)
	lums := make([]float64, n)
	var sumLum float64
	for i, p := range pix {
		r, g, b, _ := p.RGBA()
		lums[i] = 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
		sumLum += lums[i]
	}
	avg := sumLum / float64(n)

	var fgR, fgG, fgB, fgN uint64
	var bgR, bgG, bgB, bgN uint64
	for i, p := range pix {
		r, g, b, _ := p.RGBA()
		if lums[i] >= avg {
			pattern |= 1 << uint(i)
			fgR += uint64(r)
			fgG += uint64(g)
			fgB += uint64(b)
			fgN++
		} else {
			bgR += uint64(r)
			bgG += uint64(g)
			bgB += uint64(b)
			bgN++
		}
	}
	if fgN > 0 {
		fg[0] = uint8((fgR / fgN) >> 8)
		fg[1] = uint8((fgG / fgN) >> 8)
		fg[2] = uint8((fgB / fgN) >> 8)
	}
	if bgN > 0 {
		bg[0] = uint8((bgR / bgN) >> 8)
		bg[1] = uint8((bgG / bgN) >> 8)
		bg[2] = uint8((bgB / bgN) >> 8)
	} else {
		bg = fg
	}
	return
}

func writeCell(sb *strings.Builder, fg, bg [3]uint8, glyph rune) {
	sb.WriteString("\x1b[38;2;")
	sb.WriteString(strconv.Itoa(int(fg[0])))
	sb.WriteByte(';')
	sb.WriteString(strconv.Itoa(int(fg[1])))
	sb.WriteByte(';')
	sb.WriteString(strconv.Itoa(int(fg[2])))
	sb.WriteString("m\x1b[48;2;")
	sb.WriteString(strconv.Itoa(int(bg[0])))
	sb.WriteByte(';')
	sb.WriteString(strconv.Itoa(int(bg[1])))
	sb.WriteByte(';')
	sb.WriteString(strconv.Itoa(int(bg[2])))
	sb.WriteString("m")
	sb.WriteRune(glyph)
}
