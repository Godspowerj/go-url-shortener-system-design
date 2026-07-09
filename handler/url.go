package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/godspowerjonah/url-shortener/service"
)

// URLHandler maps incoming HTTP requests to our service layer operations.
type URLHandler struct {
	service service.URLService
}

// NewURLHandler creates a new handler instance injected with our service.
func NewURLHandler(s service.URLService) *URLHandler {
	return &URLHandler{service: s}
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

	newEntry, err := h.service.ShortenURL(requestBody.URL)
	if err != nil {
		http.Error(response, `{"error": "failed to save URL"}`, http.StatusInternalServerError)
		return
	}

	responseData := map[string]string{
		"short_code": newEntry.ShortCode,
		"short_url":  "http://localhost:8080/" + newEntry.ShortCode,
		"original":   newEntry.OriginalURL,
	}

	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(responseData)
}

// Redirect extracts the short code and performs a 302 redirect to the original URL.
func (h *URLHandler) Redirect(response http.ResponseWriter, request *http.Request) {
	shortCode := chi.URLParam(request, "code")

	foundURL, err := h.service.ResolveURL(shortCode)
	if err != nil {
		http.Error(response, `{"error": "short URL not found"}`, http.StatusNotFound)
		return
	}

	http.Redirect(response, request, foundURL.OriginalURL, http.StatusFound)
}

// ListURLs lists all saved links in JSON format.
func (h *URLHandler) ListURLs(response http.ResponseWriter, request *http.Request) {
	allURLs, err := h.service.ListURLs()
	if err != nil {
		http.Error(response, `{"error": "failed to list URLs"}`, http.StatusInternalServerError)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	json.NewEncoder(response).Encode(allURLs)
}
