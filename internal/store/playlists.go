package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (s *Store) CreatePlaylist(name string) (int64, error) {
	res, err := s.DB.Exec(
		`INSERT INTO playlists(name, created_at) VALUES(?, ?)`,
		name, time.Now().Unix(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) RenamePlaylist(id int64, name string) error {
	_, err := s.DB.Exec(`UPDATE playlists SET name = ? WHERE id = ?`, name, id)
	return err
}

func (s *Store) DeletePlaylist(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM playlists WHERE id = ?`, id)
	return err
}

func (s *Store) ListPlaylists() ([]Playlist, error) {
	rows, err := s.DB.Query(`SELECT id, name, created_at FROM playlists ORDER BY name COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Playlist
	for rows.Next() {
		var p Playlist
		var ts int64
		if err := rows.Scan(&p.ID, &p.Name, &ts); err != nil {
			return nil, err
		}
		p.CreatedAt = time.Unix(ts, 0)
		out = append(out, p)
	}
	return out, rows.Err()
}

// FindPlaylistByName returns the playlist (if any) matching name exactly.
func (s *Store) FindPlaylistByName(name string) (*Playlist, error) {
	var p Playlist
	var ts int64
	err := s.DB.QueryRow(
		`SELECT id, name, created_at FROM playlists WHERE name = ?`, name,
	).Scan(&p.ID, &p.Name, &ts)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.CreatedAt = time.Unix(ts, 0)
	return &p, nil
}

func (s *Store) AddTrackToPlaylist(playlistID, trackID int64) error {
	var maxPos sql.NullInt64
	if err := s.DB.QueryRow(
		`SELECT MAX(position) FROM playlist_tracks WHERE playlist_id = ?`,
		playlistID,
	).Scan(&maxPos); err != nil {
		return err
	}
	pos := int64(0)
	if maxPos.Valid {
		pos = maxPos.Int64 + 1
	}
	_, err := s.DB.Exec(
		`INSERT INTO playlist_tracks(playlist_id, position, track_id) VALUES(?,?,?)`,
		playlistID, pos, trackID,
	)
	return err
}

func (s *Store) ListPlaylistTracks(playlistID int64) ([]Track, error) {
	rows, err := s.DB.Query(
		`SELECT t.id, t.video_id, t.title, t.artist, t.duration_sec, t.thumbnail_url
		 FROM playlist_tracks pt JOIN tracks t ON t.id = pt.track_id
		 WHERE pt.playlist_id = ? ORDER BY pt.position ASC`,
		playlistID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Track
	for rows.Next() {
		var t Track
		var dur int64
		var thumb sql.NullString
		if err := rows.Scan(&t.ID, &t.VideoID, &t.Title, &t.Artist, &dur, &thumb); err != nil {
			return nil, err
		}
		t.Duration = time.Duration(dur) * time.Second
		if thumb.Valid {
			t.ThumbnailURL = thumb.String
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// MovePlaylistTrack swaps two adjacent positions (used for up/down reorder).
func (s *Store) MovePlaylistTrack(playlistID int64, fromPos, toPos int64) error {
	if fromPos == toPos {
		return nil
	}
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Use a sentinel position that can't collide.
	const tmp = -1
	if _, err := tx.Exec(
		`UPDATE playlist_tracks SET position = ? WHERE playlist_id = ? AND position = ?`,
		tmp, playlistID, fromPos,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`UPDATE playlist_tracks SET position = ? WHERE playlist_id = ? AND position = ?`,
		fromPos, playlistID, toPos,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`UPDATE playlist_tracks SET position = ? WHERE playlist_id = ? AND position = ?`,
		toPos, playlistID, tmp,
	); err != nil {
		return err
	}
	return tx.Commit()
}

// QueueReplace clears the queue and refills it from the supplied track IDs in order.
func (s *Store) QueueReplace(trackIDs []int64) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM queue`); err != nil {
		return err
	}
	for i, id := range trackIDs {
		if _, err := tx.Exec(`INSERT INTO queue(position, track_id) VALUES(?,?)`, int64(i), id); err != nil {
			return fmt.Errorf("insert queue row: %w", err)
		}
	}
	return tx.Commit()
}
