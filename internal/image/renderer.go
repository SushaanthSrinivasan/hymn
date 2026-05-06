package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	stdimg "image"
	"image/color"
	"image/color/palette"
	stddraw "image/draw"
	"image/jpeg"
	_ "image/png"
	"strings"

	"github.com/BourgeoisBear/rasterm"
	_ "golang.org/x/image/webp"
)

type Renderer struct {
	caps Caps
	bg   color.Color // letterbox color for non-square sources
}

func NewRenderer() *Renderer {
	return &Renderer{caps: Detect(), bg: color.Black}
}

func (r *Renderer) Caps() Caps { return r.caps }

// SetBackground sets the letterbox/pillarbox fill color used when the source
// image's aspect ratio doesn't match the target panel.
func (r *Renderer) SetBackground(c color.Color) {
	if c != nil {
		r.bg = c
	}
}

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
	// For ANSI text rendering, center-crop the source to the panel's visual
	// aspect (w cells wide × 2h "half-cells" tall — cells are roughly 1:2 W:H
	// in most fonts). This makes square album covers fill exactly and trims
	// the edges of 16:9 video thumbnails so they sit like cover art instead
	// of letterboxing into a square panel. Native protocol paths handle
	// their own sizing.
	switch r.caps {
	case CapsNone, CapsHalfBlock, CapsQuadrant, CapsSextantDither:
		img = CenterCropAspect(img, w, 2*h)
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
		// rasterm.ItermWriteImage doesn't pass cell-size params, so the
		// terminal sizes the image from raw pixels and the layout collapses.
		// Build the escape directly with width=Ncells;height=Mcells so the
		// image is constrained to the panel.
		px := Resize(img, w*8, h*16)
		var jpg bytes.Buffer
		if err := jpeg.Encode(&jpg, px, &jpeg.Options{Quality: 85}); err != nil {
			break
		}
		b64 := base64.StdEncoding.EncodeToString(jpg.Bytes())
		return fmt.Sprintf(
			"\x1b]1337;File=inline=1;width=%d;height=%d;preserveAspectRatio=1:%s\x07",
			w, h, b64,
		), nil
	case CapsSixel:
		px := Resize(img, w*8, h*16)
		pal := stdimg.NewPaletted(px.Bounds(), palette.Plan9)
		stddraw.FloydSteinberg.Draw(pal, pal.Bounds(), px, stdimg.Point{})
		var buf bytes.Buffer
		if err := rasterm.SixelWriteImage(&buf, pal); err != nil {
			break
		}
		return buf.String(), nil
	case CapsHalfBlock:
		return RenderHalfBlock(img, w, h, r.bg), nil
	case CapsQuadrant:
		return RenderQuadrant(img, w, h, r.bg), nil
	case CapsSextantDither:
		return RenderSextantDither(img, w, h, r.bg), nil
	}
	// Default ANSI text path: sextants give 3× the apparent vertical
	// resolution of half-blocks. Cascadia Mono / JetBrains Mono support the
	// glyphs since 2021. Aspect ratio preserved via FitForGrid; bars fill
	// with r.bg.
	return RenderSextant(img, w, h, r.bg), nil
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
