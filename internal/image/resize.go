package image

import (
	stdimg "image"

	"golang.org/x/image/draw"
)

// Resize returns a new image scaled to (w,h) using Catmull-Rom resampling.
func Resize(src stdimg.Image, w, h int) stdimg.Image {
	if w <= 0 || h <= 0 {
		return src
	}
	dst := stdimg.NewRGBA(stdimg.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}
