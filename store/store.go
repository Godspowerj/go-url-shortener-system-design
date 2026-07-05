package store

import "time"

// URL represents a single shortened URL entry stored in the database.
type URL struct {
	Id          int       `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}

// Store is the interface that any storage backend must implement.
// This allows us to swap between MemoryStore and SQLiteStore easily.
type Store interface {
	Save(shortCode string, url URL) error
	Get(shortCode string) (URL, bool)
	GetAll() []URL
}
