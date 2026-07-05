package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/godspowerjonah/url-shortener/handler"
	"github.com/godspowerjonah/url-shortener/store"
)

func main() {
	// Create the router (traffic controller for incoming requests)
	router := chi.NewRouter()

	// Middleware: logs every request and recovers from panics
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Create the SQLite store — data is saved to urls.db in the current directory
	sqliteStore := store.NewSQLiteStore("urls.db")
	urlHandler := handler.NewURLHandler(sqliteStore)

	// Register routes: path + HTTP method → handler function
	router.Post("/shorten", urlHandler.Shorten)  // POST /shorten  → Shorten()
	router.Get("/urls", urlHandler.ListURLs)     // GET  /urls     → ListURLs()
	router.Get("/{code}", urlHandler.Redirect)   // GET  /{code}   → Redirect()

	fmt.Println("🚀 URL Shortener running on http://localhost:8080")
	fmt.Println("   POST /shorten     → shorten a URL")
	fmt.Println("   GET  /{code}      → redirect to original URL")
	fmt.Println("   GET  /urls        → list all URLs")

	// Start the server — this blocks forever, waiting for requests
	http.ListenAndServe(":8080", router)
}