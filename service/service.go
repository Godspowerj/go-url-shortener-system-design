package service

import (
	"errors"
	"log"
	"time"

	"github.com/godspowerjonah/url-shortener/shortener"
	"github.com/godspowerjonah/url-shortener/store"
)

// URLService defines the business logic contract for our URL shortener operations.
type URLService interface {
	ShortenURL(originalURL string) (store.URL, error)
	ResolveURL(shortCode string) (store.URL, error)
	ListURLs() ([]store.URL, error)
}

type urlService struct {
	store       store.Store
	clicksQueue chan int
}

// NewURLService creates a new instance of our business logic service.
func NewURLService(s store.Store) URLService {
	svc := &urlService{
		store:       s,
		clicksQueue: make(chan int, 10000),
	}
	// Start the asynchronous background worker
	go svc.worker()
	return svc
}

// ShortenURL handles validating a URL, generating a unique short code, and saving the entry.
func (s *urlService) ShortenURL(originalURL string) (store.URL, error) {
	if originalURL == "" {
		return store.URL{}, errors.New("please provide a valid url field")
	}

	shortCode := shortener.GenerateCode(6)
	newEntry := store.URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
	}

	if err := s.store.Save(shortCode, newEntry); err != nil {
		return store.URL{}, err
	}

	return newEntry, nil
}

// ResolveURL retrieves a URL entry by short code and increments its click count.
func (s *urlService) ResolveURL(shortCode string) (store.URL, error) {
	foundURL, exists := s.store.Get(shortCode)
	if !exists {
		return store.URL{}, errors.New("short URL not found")
	}

	// Push the click event to the channel asynchronously (non-blocking)
	select {
	case s.clicksQueue <- foundURL.Id:
	default:
		log.Println("warning: clicks queue is full, dropping click update for ID:", foundURL.Id)
	}
	
	return foundURL, nil
}

// ListURLs lists all saved links from the database.
func (s *urlService) ListURLs() ([]store.URL, error) {
	allURLs := s.store.GetAll()
	return allURLs, nil
}

// worker runs in the background, consuming URL IDs and executing SQLite updates sequentially.
func (s *urlService) worker() {
	for id := range s.clicksQueue {
		if err := s.store.IncrementClick(id); err != nil {
			log.Println("failed to increment click count in background:", err)
		}
	}
}
