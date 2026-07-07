package store

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore is the struct box that holds our active database connection pointer.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens the DB file and prepares it.
func NewSQLiteStore(dbPath string) *SQLiteStore {
	// file: prefix is required so the driver parses query parameters.
	// WAL mode and busy_timeout allow concurrent readers and prevent database lock errors under load.
	dsn := "file:" + dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
	
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("failed to open sqlite database: %v", err)
	}

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

	// Pack the connection into our SQLiteStore box and return it.
	return &SQLiteStore{db: db}
}

// Save inserts a new shortened URL mapping into the database.
func (s *SQLiteStore) Save(shortCode string, url URL) error {
	_, err := s.db.Exec(
		`INSERT INTO urls (short_code, original_url, created_at) VALUES (?, ?, ?)`,
		shortCode, url.OriginalURL, url.CreatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

// Get finds a shortened URL using the short code.
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

// GetAll returns all saved links, sorted newest first.
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
