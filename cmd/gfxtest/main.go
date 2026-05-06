// Compare ANSI block renderers side-by-side. Fetches a YouTube thumbnail
// (or any URL passed as the first arg) and prints it five times: half-block,
// quadrant, sextant, octant, and Floyd-Steinberg-dithered sextant.
//
// Usage:
//
//	go run ./cmd/gfxtest                       # default thumbnail, 30x15
//	go run ./cmd/gfxtest <width> <height>      # set panel size
//	go run ./cmd/gfxtest <url>                 # custom image URL
//	go run ./cmd/gfxtest <url> <width> <height>
package main

import (
	"bytes"
	"context"
	"fmt"
	stdimg "image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strconv"
	"time"

	hymnimg "hymn/internal/image"
	"hymn/internal/ytdlp"
)

const defaultThumb = "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg"

func main() {
	url := defaultThumb
	w, h := 30, 15
	args := os.Args[1:]
	// Argument parsing: first numeric arg starts the size pair.
	if len(args) > 0 {
		if _, err := strconv.Atoi(args[0]); err != nil {
			url = args[0]
			args = args[1:]
		}
	}
	if len(args) >= 2 {
		if v, err := strconv.Atoi(args[0]); err == nil {
			w = v
		}
		if v, err := strconv.Atoi(args[1]); err == nil {
			h = v
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	raw, err := ytdlp.FetchThumbnail(ctx, url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetch:", err)
		os.Exit(1)
	}
	img, _, err := stdimg.Decode(bytes.NewReader(raw))
	if err != nil {
		fmt.Fprintln(os.Stderr, "decode:", err)
		os.Exit(1)
	}

	// Catppuccin Mocha base for letterbox bars.
	bg := color.RGBA{R: 0x1e, G: 0x1e, B: 0x2e, A: 0xff}

	type variant struct {
		name string
		s    string
	}
	variants := []variant{
		{"HALF-BLOCK (1×2 sub-pixels per cell)", hymnimg.RenderHalfBlock(img, w, h, bg)},
		{"QUADRANT (2×2 sub-pixels per cell)", hymnimg.RenderQuadrant(img, w, h, bg)},
		{"SEXTANT (2×3 sub-pixels per cell) — default", hymnimg.RenderSextant(img, w, h, bg)},
		{"SEXTANT + FLOYD-STEINBERG DITHER", hymnimg.RenderSextantDither(img, w, h, bg)},
	}

	for _, v := range variants {
		fmt.Printf("\n\x1b[1;38;2;203;166;247m=== %s ===\x1b[0m\n\n", v.name)
		fmt.Println(v.s)
	}
	fmt.Printf("\nSource: %s | Panel: %d×%d cells\n", url, w, h)
}
