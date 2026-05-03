package cache

import (
	"os"
	"sort"

	"hymn/internal/store"
)

// Evict trims the cache to fit under capBytes by deleting the least-recently-used
// files until total size drops below the cap. Called once at startup.
func Evict(db *store.Store, capBytes int64) error {
	entries, err := db.CacheList()
	if err != nil {
		return err
	}
	var total int64
	for _, e := range entries {
		total += e.SizeBytes
	}
	if total <= capBytes {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastUsedAt.Before(entries[j].LastUsedAt)
	})
	for _, e := range entries {
		if total <= capBytes {
			break
		}
		_ = os.Remove(e.Path)
		_ = db.CacheDelete(e.VideoID)
		total -= e.SizeBytes
	}
	return nil
}
