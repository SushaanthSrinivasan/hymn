package store

import (
	"database/sql"
	"fmt"
)

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL);`,

	`CREATE TABLE IF NOT EXISTS tracks (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		video_id      TEXT NOT NULL UNIQUE,
		title         TEXT NOT NULL,
		artist        TEXT NOT NULL,
		duration_sec  INTEGER NOT NULL,
		thumbnail_url TEXT
	);`,

	`CREATE TABLE IF NOT EXISTS queue (
		position INTEGER PRIMARY KEY,
		track_id INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE
	);`,

	`CREATE TABLE IF NOT EXISTS history (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		track_id  INTEGER NOT NULL REFERENCES tracks(id),
		played_at INTEGER NOT NULL
	);`,

	`CREATE TABLE IF NOT EXISTS playlists (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		name       TEXT NOT NULL UNIQUE,
		created_at INTEGER NOT NULL
	);`,

	`CREATE TABLE IF NOT EXISTS playlist_tracks (
		playlist_id INTEGER NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
		position    INTEGER NOT NULL,
		track_id    INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
		PRIMARY KEY (playlist_id, position)
	);`,

	`CREATE INDEX IF NOT EXISTS idx_playlist_tracks_pl ON playlist_tracks(playlist_id);`,

	`CREATE TABLE IF NOT EXISTS cached_tracks (
		video_id     TEXT PRIMARY KEY,
		path         TEXT NOT NULL,
		size_bytes   INTEGER NOT NULL,
		last_used_at INTEGER NOT NULL
	);`,

	`CREATE TABLE IF NOT EXISTS settings (
		key   TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);`,
}

func migrate(db *sql.DB) error {
	var current int
	row := db.QueryRow(`SELECT COALESCE(MAX(version),0) FROM schema_version`)
	if err := row.Scan(&current); err != nil {
		// schema_version table doesn't exist yet — apply migration 0 first.
		if _, err := db.Exec(migrations[0]); err != nil {
			return fmt.Errorf("create schema_version: %w", err)
		}
		current = 0
	}
	for i := current; i < len(migrations); i++ {
		if i == 0 && current == 0 {
			// already created above on the cold path; skip if no rows.
			var n int
			if err := db.QueryRow(`SELECT COUNT(*) FROM schema_version`).Scan(&n); err == nil && n == 0 {
				if _, err := db.Exec(`INSERT INTO schema_version(version) VALUES (0)`); err != nil {
					return err
				}
			}
		}
		if _, err := db.Exec(migrations[i]); err != nil {
			return fmt.Errorf("migration %d: %w", i, err)
		}
		if i > 0 {
			if _, err := db.Exec(`UPDATE schema_version SET version = ?`, i); err != nil {
				return err
			}
		}
	}
	return nil
}
