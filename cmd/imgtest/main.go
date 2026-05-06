// Throwaway smoke test for the image pipeline.
// Usage: go run ./cmd/imgtest [width] [height]
// Fetches a known YT thumbnail and prints the half-block render.
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	hymnimg "hymn/internal/image"
	"hymn/internal/ytdlp"
)

const thumbURL = "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg"

func main() {
	w, h := 40, 20
	if len(os.Args) >= 3 {
		w, _ = strconv.Atoi(os.Args[1])
		h, _ = strconv.Atoi(os.Args[2])
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	b, err := ytdlp.FetchThumbnail(ctx, thumbURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetch:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "fetched %d bytes; caps=%d; rendering %dx%d cells\n",
		len(b), hymnimg.Detect(), w, h)
	r := hymnimg.NewRenderer()
	s, err := r.Render(b, w, h)
	if err != nil {
		fmt.Fprintln(os.Stderr, "render:", err)
		os.Exit(1)
	}
	fmt.Print(s)
	fmt.Println()
}
