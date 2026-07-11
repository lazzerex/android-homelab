package main

import (
	"log"
	"net/http"
	"os"

	"android-go-server-test/internal"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := internal.InitGuestbook(); err != nil {
		log.Fatalf("guestbook init failed: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", internal.IndexHandler)
	mux.HandleFunc("/health", internal.HealthHandler)
	mux.HandleFunc("/info", internal.InfoHandler)
	mux.HandleFunc("/time", internal.TimeHandler)
	mux.HandleFunc("/stream", internal.StreamHandler)
	mux.HandleFunc("/dashboard", internal.DashboardHandler)
	mux.HandleFunc("/checksum", internal.ChecksumHandler)
	mux.HandleFunc("GET /guestbook", internal.GuestbookListHandler)
	mux.HandleFunc("POST /guestbook", internal.GuestbookAddHandler)
	mux.HandleFunc("POST /pastebin", internal.PastebinCreateHandler)
	mux.HandleFunc("GET /pastebin/{id}", internal.PastebinGetHandler)

	limiter := internal.NewRateLimiter(5, 10) // 5 req/s per IP, burst 10

	log.Printf("Starting %s %s", internal.Project, internal.Version)
	log.Printf("Listening on :%s", port)

	handler := internal.LoggingMiddleware(limiter.Middleware(mux))
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
