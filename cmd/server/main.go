package main

import (
	"os"
	"github.com/joho/godotenv"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/godspowerjonah/url-shortener/handler"
	"github.com/godspowerjonah/url-shortener/service"
	"github.com/godspowerjonah/url-shortener/store"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	
	// Setup the web router
	router := chi.NewRouter()

	// Add logging middleware so we can watch incoming requests in the terminal
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	redisStore := store.NewRedisStore(redisAddr)

	// Initialize SQLite (this returns the box holding the database connection)
	sqliteStore := store.NewSQLiteStore("urls.db")

	// Initialize URL service and inject database store
	urlService := service.NewURLService(sqliteStore, redisStore)

	// Pass the service to the handler so the endpoints can use it
	urlHandler := handler.NewURLHandler(urlService)

	// Define HTTP endpoints
	router.Post("/shorten", urlHandler.Shorten)
	router.Get("/urls", urlHandler.ListURLs)
	router.Get("/{code}", urlHandler.Redirect)

	fmt.Println("🚀 URL Shortener running on http://localhost:8080")
	fmt.Println("   POST /shorten     → shorten a URL")
	fmt.Println("   GET  /{code}      → redirect to original URL")
	fmt.Println("   GET  /urls        → list all URLs")

	// Start the server
	http.ListenAndServe(":8080", router)
}
