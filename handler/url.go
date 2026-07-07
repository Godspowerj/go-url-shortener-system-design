package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/godspowerjonah/url-shortener/shortener"
	"github.com/godspowerjonah/url-shortener/store"
)

// URLHandler maps incoming HTTP requests to our database store operations.
type URLHandler struct {
	store store.Store
}

func NewURLHandler(s store.Store) *URLHandler {
	return &URLHandler{store: s}
}

// Shorten creates a shortened URL from a JSON payload and saves it.
func (h *URLHandler) Shorten(response http.ResponseWriter, request *http.Request) {
	var requestBody struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil || requestBody.URL == "" {
		http.Error(response, `{"error": "please provide a valid url field"}`, http.StatusBadRequest)
		return
	}

	shortCode := shortener.GenerateCode(6)
	newEntry := store.URL{
		ShortCode:   shortCode,
		OriginalURL: requestBody.URL,
		CreatedAt:   time.Now(),
	}

	// Save to our decoupled database store
	if err := h.store.Save(shortCode, newEntry); err != nil {
		http.Error(response, `{"error": "failed to save URL"}`, http.StatusInternalServerError)
		return
	}

	responseData := map[string]string{
		"short_code": shortCode,
		"short_url":  "http://localhost:8080/" + shortCode,
		"original":   requestBody.URL,
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(responseData)
}

// Redirect extracts the short code and performs a 302 redirect to the original URL.
func (h *URLHandler) Redirect(response http.ResponseWriter, request *http.Request) {
	shortCode := chi.URLParam(request, "code")

	foundURL, exists := h.store.Get(shortCode)
	if !exists {
		http.Error(response, `{"error": "short URL not found"}`, http.StatusNotFound)
		return
	}

	http.Redirect(response, request, foundURL.OriginalURL, http.StatusFound)
}

// ListURLs lists all saved links in JSON format.
func (h *URLHandler) ListURLs(response http.ResponseWriter, request *http.Request) {
	allURLs := h.store.GetAll()
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(allURLs)
}
