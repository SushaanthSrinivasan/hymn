package image

import (
	stdimg "image"
	"strconv"
	"strings"
)

// RenderHalfBlock renders an image as ANSI half-block characters, two pixels
// per cell (top = foreground, bottom = background, glyph = "▀"). The image is
// resized to w cells wide × 2h pixels tall first.
func RenderHalfBlock(src stdimg.Image, w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	img := Resize(src, w, 2*h)
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var sb strings.Builder
	sb.Grow(width * height * 24)

	for cy := 0; cy < height; cy += 2 {
		for cx := 0; cx < width; cx++ {
			r1, g1, b1, _ := img.At(bounds.Min.X+cx, bounds.Min.Y+cy).RGBA()
			r2, g2, b2 := uint32(0), uint32(0), uint32(0)
			if cy+1 < height {
				r2, g2, b2, _ = img.At(bounds.Min.X+cx, bounds.Min.Y+cy+1).RGBA()
			}
			sb.WriteString("\x1b[38;2;")
			writeColor(&sb, r1, g1, b1)
			sb.WriteString("m\x1b[48;2;")
			writeColor(&sb, r2, g2, b2)
			sb.WriteString("m▀")
		}
		sb.WriteString("\x1b[0m\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

func writeColor(sb *strings.Builder, r, g, b uint32) {
	sb.WriteString(strconv.Itoa(int(r >> 8)))
	sb.WriteByte(';')
	sb.WriteString(strconv.Itoa(int(g >> 8)))
	sb.WriteByte(';')
	sb.WriteString(strconv.Itoa(int(b >> 8)))
}
