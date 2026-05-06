package image

import (
	stdimg "image"
	"image/color"

	"golang.org/x/image/draw"
)

// CenterCropAspect returns a sub-image of src cropped to (aspectW : aspectH),
// centered on the source. If src already matches the aspect, the source is
// returned untouched; otherwise the wider/taller axis is trimmed evenly from
// both sides.
func CenterCropAspect(src stdimg.Image, aspectW, aspectH int) stdimg.Image {
	sw := src.Bounds().Dx()
	sh := src.Bounds().Dy()
	if sw <= 0 || sh <= 0 || aspectW <= 0 || aspectH <= 0 {
		return src
	}
	target := float64(aspectW) / float64(aspectH)
	srcA := float64(sw) / float64(sh)
	var newW, newH int
	if srcA > target {
		newH = sh
		newW = int(float64(sh)*target + 0.5)
	} else {
		newW = sw
		newH = int(float64(sw)/target + 0.5)
	}
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}
	if newW == sw && newH == sh {
		return src
	}
	offX := src.Bounds().Min.X + (sw-newW)/2
	offY := src.Bounds().Min.Y + (sh-newH)/2
	dst := stdimg.NewRGBA(stdimg.Rect(0, 0, newW, newH))
	draw.Draw(dst, dst.Bounds(), src, stdimg.Point{X: offX, Y: offY}, draw.Src)
	return dst
}

// FitForGrid resizes src to fit inside a sub-pixel grid of (gridW, gridH)
// while preserving the source's visible aspect ratio. (subVisW, subVisH) is
// the visible aspect of a single sub-pixel — half-block sub-pixels are 1:1,
// sextants are 3:4 (W:H), octants are 1:1.
//
// The returned image is exactly (gridW, gridH) pixels with the source
// centered and the leftover area filled with bg. This is what produces the
// "letterbox / pillarbox" effect that keeps album art from stretching when
// it doesn't match the panel's aspect.
func FitForGrid(src stdimg.Image, gridW, gridH, subVisW, subVisH int, bg color.Color) stdimg.Image {
	if gridW <= 0 || gridH <= 0 {
		return stdimg.NewRGBA(stdimg.Rect(0, 0, 1, 1))
	}
	canvas := stdimg.NewRGBA(stdimg.Rect(0, 0, gridW, gridH))
	if bg != nil {
		draw.Draw(canvas, canvas.Bounds(), &stdimg.Uniform{C: bg}, stdimg.Point{}, draw.Src)
	}
	sw := src.Bounds().Dx()
	sh := src.Bounds().Dy()
	if sw <= 0 || sh <= 0 {
		return canvas
	}

	// Source is assumed to have square pixels (1:1). Compute scale to fit
	// inside the panel's *visible* area while preserving aspect.
	panelVisW := float64(gridW * subVisW)
	panelVisH := float64(gridH * subVisH)
	s := panelVisW / float64(sw)
	if sy := panelVisH / float64(sh); sy < s {
		s = sy
	}
	visW := float64(sw) * s
	visH := float64(sh) * s

	// Convert visible dimensions back to sub-pixel grid units.
	nw := int(visW/float64(subVisW) + 0.5)
	nh := int(visH/float64(subVisH) + 0.5)
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	if nw > gridW {
		nw = gridW
	}
	if nh > gridH {
		nh = gridH
	}

	resized := Resize(src, nw, nh)
	offX := (gridW - nw) / 2
	offY := (gridH - nh) / 2
	draw.Draw(canvas, stdimg.Rect(offX, offY, offX+nw, offY+nh), resized, stdimg.Point{}, draw.Src)
	return canvas
}
