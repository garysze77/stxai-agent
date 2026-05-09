package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func dataDir() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "stxai", "data")
}

func Open() (*Store, error) {
	dir := dataDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", filepath.Join(dir, "stxai.db"))
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id, created_at);

		CREATE TABLE IF NOT EXISTS preferences (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`)
	return err
}

func (s *Store) SaveMessage(sessionID, role, content string) error {
	_, err := s.db.Exec(
		"INSERT INTO messages (session_id, role, content) VALUES (?, ?, ?)",
		sessionID, role, content,
	)
	return err
}

func (s *Store) GetHistory(sessionID string, limit int) ([]Message, error) {
	rows, err := s.db.Query(
		"SELECT role, content, created_at FROM messages WHERE session_id = ? ORDER BY created_at DESC LIMIT ?",
		sessionID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.Role, &m.Content, &m.Timestamp); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}

	// Reverse to chronological order
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (s *Store) SetPreference(key, value string) error {
	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO preferences (key, value) VALUES (?, ?)",
		key, value,
	)
	return err
}

func (s *Store) GetPreference(key string) (string, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM preferences WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("not found: %s", key)
	}
	return value, err
}

func (s *Store) DeleteOldMessages(before time.Time) error {
	_, err := s.db.Exec("DELETE FROM messages WHERE created_at < ?", before)
	return err
}
