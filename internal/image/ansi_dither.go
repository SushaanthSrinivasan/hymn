package image

import (
	stdimg "image"
	"image/color"
	"strings"
)

// RenderSextantDither is RenderSextant with Floyd-Steinberg error diffusion
// applied to the binarization step. Instead of thresholding each sub-pixel
// independently against its 6-pixel cell average, we build a single luminance
// field for the whole image, dither it to 1-bit using error diffusion, and
// use that as the on/off mask. Cell colors are still computed from the
// original (non-dithered) source.
//
// This trades sharper local contrast for smoother gradients — looks better
// on photos / album art with subtle tonal transitions, can over-stipple on
// sharp logos.
func RenderSextantDither(src stdimg.Image, w, h int, bg color.Color) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	gridW, gridH := 2*w, 3*h
	img := FitForGrid(src, gridW, gridH, 3, 4, bg)

	// Build luminance buffer.
	lum := make([]float64, gridW*gridH)
	for y := 0; y < gridH; y++ {
		for x := 0; x < gridW; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			lum[y*gridW+x] = 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
		}
	}

	// Floyd-Steinberg: threshold each pixel against the global midpoint,
	// diffuse error to right (7/16), bottom-left (3/16), bottom (5/16),
	// bottom-right (1/16). Mask stores 0 (dark) or 1 (lit).
	mask := make([]bool, gridW*gridH)
	const mid = 32768.0 // 65535/2 — RGBA() returns 16-bit values
	for y := 0; y < gridH; y++ {
		for x := 0; x < gridW; x++ {
			i := y*gridW + x
			old := lum[i]
			var newVal float64
			if old >= mid {
				newVal = 65535
				mask[i] = true
			} else {
				newVal = 0
				mask[i] = false
			}
			err := old - newVal
			if x+1 < gridW {
				lum[i+1] += err * 7 / 16
			}
			if y+1 < gridH {
				if x > 0 {
					lum[i+gridW-1] += err * 3 / 16
				}
				lum[i+gridW] += err * 5 / 16
				if x+1 < gridW {
					lum[i+gridW+1] += err * 1 / 16
				}
			}
		}
	}

	var sb strings.Builder
	sb.Grow(w * h * 32)

	for cy := 0; cy < h; cy++ {
		for cx := 0; cx < w; cx++ {
			var pattern uint8
			var fgR, fgG, fgB, fgN uint64
			var bgR, bgG, bgB, bgN uint64
			for i := 0; i < 6; i++ {
				row := i / 2
				col := i % 2
				x := 2*cx + col
				y := 3*cy + row
				r, g, b, _ := img.At(x, y).RGBA()
				if mask[y*gridW+x] {
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
			var fg, bgc [3]uint8
			if fgN > 0 {
				fg[0] = uint8((fgR / fgN) >> 8)
				fg[1] = uint8((fgG / fgN) >> 8)
				fg[2] = uint8((fgB / fgN) >> 8)
			}
			if bgN > 0 {
				bgc[0] = uint8((bgR / bgN) >> 8)
				bgc[1] = uint8((bgG / bgN) >> 8)
				bgc[2] = uint8((bgB / bgN) >> 8)
			} else {
				bgc = fg
			}
			writeCell(&sb, fg, bgc, sextantRune(pattern))
		}
		sb.WriteString("\x1b[0m\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}
