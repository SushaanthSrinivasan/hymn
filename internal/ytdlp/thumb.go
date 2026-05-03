package ytdlp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

var thumbHTTP = &http.Client{Timeout: 10 * time.Second}

// FetchThumbnail downloads the raw bytes at url. Decoding is the caller's job.
func FetchThumbnail(ctx context.Context, url string) ([]byte, error) {
	if url == "" {
		return nil, fmt.Errorf("empty thumbnail url")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := thumbHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("thumbnail http %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
}
