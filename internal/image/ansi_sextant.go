package image

import (
	stdimg "image"
	"image/color"
	"strconv"
	"strings"
)

// RenderSextant renders src as ANSI sextant blocks — 6 sub-pixels per cell in
// a 2×3 grid using the Symbols for Legacy Computing Unicode block
// (U+1FB00-U+1FB3B). Roughly 3× the apparent vertical resolution of half-block
// at the cost of 2-color-per-cell quantization.
//
// Each cell encodes two colors (the average of the "lit" sub-pixels and the
// average of the "dark" ones) and one of 64 sextant glyphs. Font support:
// Cascadia Mono (Windows Terminal default) and JetBrains Mono (WezTerm
// bundled) both have these glyphs since 2021. Older fonts will render tofu.
//
// bg is the color used for letterboxing when the source aspect doesn't match
// the panel — pass color.Black or your terminal's background.
func RenderSextant(src stdimg.Image, w, h int, bg color.Color) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	img := FitForGrid(src, 2*w, 3*h, 3, 4, bg)

	var sb strings.Builder
	sb.Grow(w * h * 32)

	for cy := 0; cy < h; cy++ {
		for cx := 0; cx < w; cx++ {
			var pix [6]color.Color
			for i := 0; i < 6; i++ {
				row := i / 2
				col := i % 2
				pix[i] = img.At(2*cx+col, 3*cy+row)
			}

			// Threshold by luminance: pixels at-or-above the cell average go
			// into the foreground partition; the rest are background.
			var lums [6]float64
			var sumLum float64
			for i, p := range pix {
				r, g, b, _ := p.RGBA()
				lums[i] = 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
				sumLum += lums[i]
			}
			avgLum := sumLum / 6.0

			var pattern uint8
			var fgR, fgG, fgB, fgN uint64
			var bgR, bgG, bgB, bgN uint64
			for i, p := range pix {
				r, g, b, _ := p.RGBA()
				if lums[i] >= avgLum {
					pattern |= 1 << i
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

			var fr, fg, fb, br, bgC, bb uint8
			if fgN > 0 {
				fr = uint8((fgR / fgN) >> 8)
				fg = uint8((fgG / fgN) >> 8)
				fb = uint8((fgB / fgN) >> 8)
			}
			if bgN > 0 {
				br = uint8((bgR / bgN) >> 8)
				bgC = uint8((bgG / bgN) >> 8)
				bb = uint8((bgB / bgN) >> 8)
			} else {
				br, bgC, bb = fr, fg, fb
			}

			sb.WriteString("\x1b[38;2;")
			sb.WriteString(strconv.Itoa(int(fr)))
			sb.WriteByte(';')
			sb.WriteString(strconv.Itoa(int(fg)))
			sb.WriteByte(';')
			sb.WriteString(strconv.Itoa(int(fb)))
			sb.WriteString("m\x1b[48;2;")
			sb.WriteString(strconv.Itoa(int(br)))
			sb.WriteByte(';')
			sb.WriteString(strconv.Itoa(int(bgC)))
			sb.WriteByte(';')
			sb.WriteString(strconv.Itoa(int(bb)))
			sb.WriteString("m")
			sb.WriteRune(sextantRune(pattern))
		}
		sb.WriteString("\x1b[0m\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

// sextantRune maps a 6-bit pattern (bit 0 = top-left, bit 5 = bottom-right) to
// the corresponding Unicode glyph. Patterns 0, 21, 42, and 63 alias to
// pre-existing characters (space, ▌, ▐, █); the other 60 live at U+1FB00-1FB3B.
func sextantRune(p uint8) rune {
	switch {
	case p == 0:
		return ' '
	case p < 21:
		return rune(0x1FB00 + int(p) - 1)
	case p == 21:
		return '▌'
	case p < 42:
		return rune(0x1FB14 + int(p) - 22)
	case p == 42:
		return '▐'
	case p < 63:
		return rune(0x1FB28 + int(p) - 43)
	default:
		return '█'
	}
}
