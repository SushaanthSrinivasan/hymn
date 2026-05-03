package player

import "io"

type conn interface {
	io.ReadWriteCloser
}

var dial func(path string) (conn, error)
