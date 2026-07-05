package store

import (
	"sync"
	"time"
)

// URL represents a single shortened URL entry stored in the database.
type URL struct {
	Id          int       `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}

// MemoryStore is an in-memory database that holds all URL entries in a map.
// sync.RWMutex makes it safe when multiple requests come in at the same time.
type MemoryStore struct {
	mutex sync.RWMutex
	urls  map[string]URL // key = short code, value = URL entry
}

// NewMemoryStore creates and returns a new, empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		urls: make(map[string]URL),
	}
}

// Save stores a URL entry in the map under the given short code.
func (memoryStore *MemoryStore) Save(shortCode string, url URL) {
	memoryStore.mutex.Lock()
	defer memoryStore.mutex.Unlock()

	memoryStore.urls[shortCode] = url
}

// Get looks up a URL by its short code.
// Returns the URL entry and a boolean indicating whether it was found.
func (memoryStore *MemoryStore) Get(shortCode string) (URL, bool) {
	memoryStore.mutex.RLock()
	defer memoryStore.mutex.RUnlock()

	foundURL, exists := memoryStore.urls[shortCode]
	return foundURL, exists
}

// GetAll returns all stored URL entries as a slice.
func (memoryStore *MemoryStore) GetAll() []URL {
	memoryStore.mutex.RLock()
	defer memoryStore.mutex.RUnlock()

	allURLs := make([]URL, 0, len(memoryStore.urls))
	for _, url := range memoryStore.urls {
		allURLs = append(allURLs, url)
	}
	return allURLs
}
