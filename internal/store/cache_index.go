package store

import (
	"database/sql"
	"errors"
	"time"
)

func (s *Store) CacheUpsert(e CacheEntry) error {
	_, err := s.DB.Exec(
		`INSERT INTO cached_tracks(video_id,path,size_bytes,last_used_at) VALUES(?,?,?,?)
		 ON CONFLICT(video_id) DO UPDATE SET path=excluded.path, size_bytes=excluded.size_bytes, last_used_at=excluded.last_used_at`,
		e.VideoID, e.Path, e.SizeBytes, e.LastUsedAt.Unix(),
	)
	return err
}

func (s *Store) CacheTouch(videoID string) error {
	_, err := s.DB.Exec(
		`UPDATE cached_tracks SET last_used_at = ? WHERE video_id = ?`,
		time.Now().Unix(), videoID,
	)
	return err
}

func (s *Store) CacheGet(videoID string) (*CacheEntry, error) {
	var e CacheEntry
	var lastUsed int64
	err := s.DB.QueryRow(
		`SELECT video_id, path, size_bytes, last_used_at FROM cached_tracks WHERE video_id = ?`,
		videoID,
	).Scan(&e.VideoID, &e.Path, &e.SizeBytes, &lastUsed)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	e.LastUsedAt = time.Unix(lastUsed, 0)
	return &e, nil
}

func (s *Store) CacheList() ([]CacheEntry, error) {
	rows, err := s.DB.Query(
		`SELECT video_id, path, size_bytes, last_used_at FROM cached_tracks`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CacheEntry
	for rows.Next() {
		var e CacheEntry
		var lastUsed int64
		if err := rows.Scan(&e.VideoID, &e.Path, &e.SizeBytes, &lastUsed); err != nil {
			return nil, err
		}
		e.LastUsedAt = time.Unix(lastUsed, 0)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) CacheDelete(videoID string) error {
	_, err := s.DB.Exec(`DELETE FROM cached_tracks WHERE video_id = ?`, videoID)
	return err
}
