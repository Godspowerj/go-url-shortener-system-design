package store

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite" // registers the sqlite driver
)

// SQLiteStore is a persistent store backed by a SQLite database file.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens (or creates) a SQLite database at the given file path,
// creates the urls table if it doesn't exist, and returns the store.
func NewSQLiteStore(dbPath string) *SQLiteStore {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open sqlite database: %v", err)
	}

	// Create the urls table if it doesn't already exist
	createTable := `
	CREATE TABLE IF NOT EXISTS urls (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		short_code  TEXT NOT NULL UNIQUE,
		original_url TEXT NOT NULL,
		created_at  DATETIME NOT NULL
	);`

	if _, err = db.Exec(createTable); err != nil {
		log.Fatalf("failed to create urls table: %v", err)
	}

	return &SQLiteStore{db: db}
}

// Save inserts a new URL entry into the database.
func (s *SQLiteStore) Save(shortCode string, url URL) error {
	_, err := s.db.Exec(
		`INSERT INTO urls (short_code, original_url, created_at) VALUES (?, ?, ?)`,
		shortCode, url.OriginalURL, url.CreatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

// Get looks up a URL by its short code.
// Returns the URL entry and true if found, or an empty URL and false if not found.
func (s *SQLiteStore) Get(shortCode string) (URL, bool) {
	row := s.db.QueryRow(
		`SELECT id, short_code, original_url, created_at FROM urls WHERE short_code = ?`,
		shortCode,
	)

	var u URL
	var createdAtStr string
	if err := row.Scan(&u.Id, &u.ShortCode, &u.OriginalURL, &createdAtStr); err != nil {
		return URL{}, false
	}

	u.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	return u, true
}

// GetAll returns all stored URL entries.
func (s *SQLiteStore) GetAll() []URL {
	rows, err := s.db.Query(
		`SELECT id, short_code, original_url, created_at FROM urls ORDER BY id DESC`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var urls []URL
	for rows.Next() {
		var u URL
		var createdAtStr string
		if err := rows.Scan(&u.Id, &u.ShortCode, &u.OriginalURL, &createdAtStr); err != nil {
			continue
		}
		u.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		urls = append(urls, u)
	}
	return urls
}
