package image

import "github.com/BourgeoisBear/rasterm"

type Caps int

const (
	CapsNone Caps = iota
	CapsSixel
	CapsITerm2
	CapsKitty
)

func Detect() Caps {
	switch {
	case rasterm.IsKittyCapable():
		return CapsKitty
	case rasterm.IsItermCapable():
		return CapsITerm2
	}
	if ok, _ := rasterm.IsSixelCapable(); ok {
		return CapsSixel
	}
	return CapsNone
}
