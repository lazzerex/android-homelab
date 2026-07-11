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

	mux := http.NewServeMux()

	mux.HandleFunc("/", internal.IndexHandler)
	mux.HandleFunc("/health", internal.HealthHandler)
	mux.HandleFunc("/info", internal.InfoHandler)
	mux.HandleFunc("/time", internal.TimeHandler)
	mux.HandleFunc("/stream", internal.StreamHandler)
	mux.HandleFunc("/dashboard", internal.DashboardHandler)
	mux.HandleFunc("/checksum", internal.ChecksumHandler)

	log.Printf("Starting %s %s", internal.Project, internal.Version)
	log.Printf("Listening on :%s", port)

	log.Fatal(http.ListenAndServe(":"+port, internal.LoggingMiddleware(mux)))
}
