package image

import (
	"bytes"
	stdimg "image"
	"image/color/palette"
	stddraw "image/draw"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"github.com/BourgeoisBear/rasterm"
	_ "golang.org/x/image/webp"
)

type Renderer struct {
	caps Caps
}

func NewRenderer() *Renderer {
	return &Renderer{caps: Detect()}
}

func (r *Renderer) Caps() Caps { return r.caps }

// Render decodes the supplied bytes (JPEG, PNG, or WebP) and returns a string
// suitable for direct printing into a terminal cell box of (w,h) cells.
func (r *Renderer) Render(data []byte, w, h int) (string, error) {
	if w <= 0 || h <= 0 {
		return "", nil
	}
	img, _, err := stdimg.Decode(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	switch r.caps {
	case CapsKitty:
		// Kitty: target a pixel size that fits the cell box.
		px := Resize(img, w*8, h*16)
		var buf bytes.Buffer
		if err := rasterm.KittyWriteImage(&buf, px, rasterm.KittyImgOpts{}); err != nil {
			break
		}
		return buf.String(), nil
	case CapsITerm2:
		px := Resize(img, w*8, h*16)
		var buf bytes.Buffer
		if err := rasterm.ItermWriteImage(&buf, px); err != nil {
			break
		}
		return buf.String(), nil
	case CapsSixel:
		px := Resize(img, w*8, h*16)
		pal := stdimg.NewPaletted(px.Bounds(), palette.Plan9)
		stddraw.FloydSteinberg.Draw(pal, pal.Bounds(), px, stdimg.Point{})
		var buf bytes.Buffer
		if err := rasterm.SixelWriteImage(&buf, pal); err != nil {
			break
		}
		return buf.String(), nil
	}
	// Fallback (and recovery from any of the above failing).
	return RenderHalfBlock(img, w, h), nil
}

// Pad ensures the rendered art is at least h lines tall (centered for
// graphics protocols whose output may be a single escape-blob line).
func Pad(s string, h int) string {
	lines := strings.Split(s, "\n")
	for len(lines) < h {
		lines = append(lines, "")
	}
	if len(lines) > h {
		lines = lines[:h]
	}
	return strings.Join(lines, "\n")
}
