package image

import (
	"os"
	"strings"

	"github.com/BourgeoisBear/rasterm"
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
// Auto-detection currently returns half-block (CapsNone) for terminals that
// aren't a known native-Kitty / native-iTerm2 host. Native pixel protocols
// (iTerm2 OSC 1337, Kitty graphics) don't compose reliably with bubbletea's
// frame compositor — see charmbracelet/bubbletea#163 — so we don't try to
// auto-select them in mixed environments (WezTerm, ConPTY-on-Windows, etc).
//
// HYMN_GFX overrides detection. Recognized values:
//
//   sextant (default) — 2×3 sub-pixels per cell, Symbols for Legacy Computing
//   halfblock         — 2 sub-pixels per cell (▀); fallback for fonts lacking sextants
//   quadrant          — 2×2 sub-pixels per cell (▘▝▖▗); universal font support
//   dither            — sextant + Floyd-Steinberg error diffusion (smoother gradients)
//   kitty / iterm2 / sixel — opt-in native graphics protocols (fragile; see #163)
//   off / none         — alias for sextant
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
	case "sextant", "none", "off":
		return CapsNone
	}
	if rasterm.IsKittyCapable() {
		return CapsKitty
	}
	if rasterm.IsItermCapable() {
		return CapsITerm2
	}
	return CapsNone
}
