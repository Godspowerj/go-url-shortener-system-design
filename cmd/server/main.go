package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/godspowerjonah/url-shortener/handler"
	"github.com/godspowerjonah/url-shortener/service"
	"github.com/godspowerjonah/url-shortener/store"
)

func main() {
	// 1. Setup the web router
	router := chi.NewRouter()

	// 2. Add logging middleware so we can watch incoming requests in the terminal
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// 3. Initialize SQLite (this returns the box holding the database connection)
	sqliteStore := store.NewSQLiteStore("urls.db")

	// 4. Initialize URL service and inject database store
	urlService := service.NewURLService(sqliteStore)

	// 5. Pass the service to the handler so the endpoints can use it
	urlHandler := handler.NewURLHandler(urlService)

	// 6. Define HTTP endpoints
	router.Post("/shorten", urlHandler.Shorten)
	router.Get("/urls", urlHandler.ListURLs)
	router.Get("/{code}", urlHandler.Redirect)

	fmt.Println("🚀 URL Shortener running on http://localhost:8080")
	fmt.Println("   POST /shorten     → shorten a URL")
	fmt.Println("   GET  /{code}      → redirect to original URL")
	fmt.Println("   GET  /urls        → list all URLs")

	// 7. Start the server
	http.ListenAndServe(":8080", router)
}
