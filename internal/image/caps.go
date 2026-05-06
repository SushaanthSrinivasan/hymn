package image

import (
	"os"
	"strings"
)

type Caps int

const (
	CapsNone          Caps = iota // ANSI text via sextants (default)
	CapsHalfBlock                 // ANSI text via half-blocks (fallback for fonts without U+1FB00 sextants)
	CapsQuadrant                  // ANSI text via quadrants (universal font support, 2x2 sub-pixels)
	CapsSextantDither             // sextants with Floyd-Steinberg error diffusion
	CapsSixel
	CapsITerm2
	CapsKitty
)

// Detect returns the image-rendering backend.
//
// Default (HYMN_GFX unset) is always sextant. Native graphics protocols
// (iTerm2 OSC 1337, Kitty graphics, sixel) don't compose with bubbletea's
// frame compositor — see charmbracelet/bubbletea#163 — so we never auto-
// select them; users have to opt in via HYMN_GFX explicitly.
//
// Recognized HYMN_GFX values:
//
//   sextant (default)   2×3 sub-pixels per cell, Symbols for Legacy Computing
//   halfblock           2 sub-pixels per cell (▀); fallback for fonts lacking sextants
//   quadrant            2×2 sub-pixels per cell (▘▝▖▗); universal font support
//   dither              sextant + Floyd-Steinberg error diffusion
//   kitty/iterm2/sixel  opt-in native graphics protocols (fragile)
//   off/none            alias for sextant
func Detect() Caps {
	switch strings.ToLower(os.Getenv("HYMN_GFX")) {
	case "kitty":
		return CapsKitty
	case "iterm", "iterm2":
		return CapsITerm2
	case "sixel":
		return CapsSixel
	case "halfblock":
		return CapsHalfBlock
	case "quadrant":
		return CapsQuadrant
	case "dither", "sextant-dither":
		return CapsSextantDither
	}
	return CapsNone
}
