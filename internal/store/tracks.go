package store

import (
	"database/sql"
	"errors"
	"time"
)

// UpsertTrack inserts a track keyed on video_id, or returns the existing row's
// id if one already exists. It does not overwrite existing metadata.
func (s *Store) UpsertTrack(t Track) (int64, error) {
	var id int64
	err := s.DB.QueryRow(`SELECT id FROM tracks WHERE video_id = ?`, t.VideoID).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	res, err := s.DB.Exec(
		`INSERT INTO tracks(video_id,title,artist,duration_sec,thumbnail_url) VALUES (?,?,?,?,?)`,
		t.VideoID, t.Title, t.Artist, int64(t.Duration.Seconds()), nullStr(t.ThumbnailURL),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetTrack(id int64) (Track, error) {
	var t Track
	var dur int64
	var thumb sql.NullString
	err := s.DB.QueryRow(
		`SELECT id,video_id,title,artist,duration_sec,thumbnail_url FROM tracks WHERE id = ?`, id,
	).Scan(&t.ID, &t.VideoID, &t.Title, &t.Artist, &dur, &thumb)
	if err != nil {
		return Track{}, err
	}
	t.Duration = time.Duration(dur) * time.Second
	if thumb.Valid {
		t.ThumbnailURL = thumb.String
	}
	return t, nil
}

// QueueAppend adds a track to the end of the queue and returns its position.
func (s *Store) QueueAppend(trackID int64) (int64, error) {
	var maxPos sql.NullInt64
	if err := s.DB.QueryRow(`SELECT MAX(position) FROM queue`).Scan(&maxPos); err != nil {
		return 0, err
	}
	pos := int64(0)
	if maxPos.Valid {
		pos = maxPos.Int64 + 1
	}
	_, err := s.DB.Exec(`INSERT INTO queue(position,track_id) VALUES(?,?)`, pos, trackID)
	if err != nil {
		return 0, err
	}
	return pos, nil
}

// QueueList returns the queue in position order, joined to tracks.
func (s *Store) QueueList() ([]Track, error) {
	rows, err := s.DB.Query(
		`SELECT t.id, t.video_id, t.title, t.artist, t.duration_sec, t.thumbnail_url
		 FROM queue q JOIN tracks t ON t.id = q.track_id ORDER BY q.position ASC`,
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

func (s *Store) QueueClear() error {
	_, err := s.DB.Exec(`DELETE FROM queue`)
	return err
}

// QueueRemoveAt removes the row at the given queue position and shifts later rows down.
func (s *Store) QueueRemoveAt(pos int64) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM queue WHERE position = ?`, pos); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE queue SET position = position - 1 WHERE position > ?`, pos); err != nil {
		return err
	}
	return tx.Commit()
}

// HistoryAdd records a play event.
func (s *Store) HistoryAdd(trackID int64) error {
	_, err := s.DB.Exec(
		`INSERT INTO history(track_id, played_at) VALUES(?,?)`,
		trackID, time.Now().Unix(),
	)
	return err
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
