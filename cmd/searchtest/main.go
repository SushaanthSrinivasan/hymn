// Throwaway smoke test for the ytdlp package.
// Usage: go run ./cmd/searchtest [query]
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"hymn/internal/ytdlp"
)

func main() {
	q := "lofi hip hop"
	if len(os.Args) > 1 {
		q = strings.Join(os.Args[1:], " ")
	}
	cl, err := ytdlp.New("")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	t0 := time.Now()
	res, err := cl.Search(ctx, q, 10)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("got %d results in %s\n\n", len(res), time.Since(t0).Round(time.Millisecond))
	for i, r := range res {
		fmt.Printf("%2d. [%s] %s — %s (%s)\n", i+1, r.DurationString(), r.Title, r.Artist, r.VideoID)
	}
}
