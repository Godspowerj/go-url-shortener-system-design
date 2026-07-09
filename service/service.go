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
	store store.Store
}

// NewURLService creates a new instance of our business logic service.
func NewURLService(s store.Store) URLService {
	return &urlService{store: s}
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

	// Increment the click count in the database
	if err := s.store.Update(foundURL.Id, foundURL.ClickCount+1); err != nil {
		log.Println("failed to increment click count:", err)
	}

	// Update the field in our returned struct as well, so it reflects the click count update
	foundURL.ClickCount++

	return foundURL, nil
}

// ListURLs lists all saved links from the database.
func (s *urlService) ListURLs() ([]store.URL, error) {
	allURLs := s.store.GetAll()
	return allURLs, nil
}
