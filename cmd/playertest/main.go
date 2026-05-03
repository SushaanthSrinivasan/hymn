// Throwaway smoke test for the player package.
// Usage: go run ./cmd/playertest [yt-url]
// Plays for ~15 seconds then exits. Verifies ad-free audio reaches speakers.
package main

import (
	"fmt"
	"os"
	"time"

	"hymn/internal/player"
)

func main() {
	url := "https://www.youtube.com/watch?v=jfKfPfyJRdk" // lofi girl 24/7
	if len(os.Args) > 1 {
		url = os.Args[1]
	}

	cl, err := player.Spawn("")
	if err != nil {
		fmt.Fprintln(os.Stderr, "spawn:", err)
		os.Exit(1)
	}
	defer cl.Close()

	go func() {
		for ev := range cl.Events() {
			switch ev.Kind {
			case player.EventTimePos:
				fmt.Printf("\rtime-pos: %.1fs   ", ev.Float)
			case player.EventDuration:
				fmt.Printf("\nduration: %.0fs\n", ev.Float)
			case player.EventMediaTitle:
				fmt.Printf("\ntitle: %s\n", ev.String)
			case player.EventFileLoaded:
				fmt.Println("file-loaded")
			case player.EventEOF:
				fmt.Println("eof")
			}
		}
	}()

	if err := cl.Loadfile(url); err != nil {
		fmt.Fprintln(os.Stderr, "loadfile:", err)
		os.Exit(1)
	}

	fmt.Println("playing for 15s...")
	time.Sleep(15 * time.Second)
	fmt.Println("\ndone")
}
