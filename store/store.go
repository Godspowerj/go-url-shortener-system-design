package store

import "time"

// URL represents a single shortened URL entry stored in the database.
type URL struct {
	Id          int       `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	ClickCount  int       `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// Store is the interface that decouples our HTTP layer from our database.
// As long as a database struct has these 3 methods, handler.go can use it.
type Store interface {
	Save(shortCode string, url URL) error
	Get(shortCode string) (URL, bool)
	IncrementClick(id int) error
	GetAll() []URL
}
