//go:build !windows

package player

import "net"

func init() {
	dial = func(path string) (conn, error) {
		c, err := net.Dial("unix", path)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
}
