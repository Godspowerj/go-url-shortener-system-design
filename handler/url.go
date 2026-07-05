package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/godspowerjonah/url-shortener/shortener"
	"github.com/godspowerjonah/url-shortener/store"
)

// URLHandler holds the store (database) and handles all URL-related HTTP requests.
type URLHandler struct {
	store *store.MemoryStore
}

// NewURLHandler creates a new URLHandler wired to the given store.
func NewURLHandler(memoryStore *store.MemoryStore) *URLHandler {
	return &URLHandler{store: memoryStore}
}

// Shorten handles POST /shorten.
// Reads a URL from the request body, generates a short code, saves it, and responds with the result.
func (h *URLHandler) Shorten(response http.ResponseWriter, request *http.Request) {
	// Define the shape of the expected request body
	var requestBody struct {
		URL string `json:"url"`
	}

	// Decode the JSON body into requestBody
	decodeError := json.NewDecoder(request.Body).Decode(&requestBody)
	if decodeError != nil || requestBody.URL == "" {
		http.Error(response, `{"error": "please provide a valid url field"}`, http.StatusBadRequest)
		return
	}

	// Generate a random 6-character short code (e.g. "xK92pQ")
	shortCode := shortener.GenerateCode(6)

	// Build the URL entry and save it to the store
	newEntry := store.URL{
		ShortCode:   shortCode,
		OriginalURL: requestBody.URL,
		CreatedAt:   time.Now(),
	}
	h.store.Save(shortCode, newEntry)

	// Build and send the JSON response
	responseData := map[string]string{
		"short_code": shortCode,
		"short_url":  "http://localhost:8080/" + shortCode,
		"original":   requestBody.URL,
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(responseData)
}

// Redirect handles GET /{code}.
// Looks up the short code and redirects the user to the original URL.
func (h *URLHandler) Redirect(response http.ResponseWriter, request *http.Request) {
	// Extract the short code from the URL path (e.g. /xK92pQ → "xK92pQ")
	shortCode := chi.URLParam(request, "code")

	// Look it up in the store
	foundURL, exists := h.store.Get(shortCode)
	if !exists {
		http.Error(response, `{"error": "short URL not found"}`, http.StatusNotFound)
		return
	}

	// Redirect the user to the original URL (302 temporary redirect)
	http.Redirect(response, request, foundURL.OriginalURL, http.StatusFound)
}

// ListURLs handles GET /urls.
// Returns all stored URLs as a JSON array.
func (h *URLHandler) ListURLs(response http.ResponseWriter, request *http.Request) {
	allURLs := h.store.GetAll()
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(allURLs)
}
