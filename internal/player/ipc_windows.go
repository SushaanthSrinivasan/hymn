//go:build windows

package player

import "github.com/Microsoft/go-winio"

func init() {
	dial = func(path string) (conn, error) {
		c, err := winio.DialPipe(path, nil)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
}
