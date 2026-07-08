package internal

import (
	"encoding/json"
	"net/http"
	"time"
)

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"message": "Hello from Android!",
		"project": Project,
	})
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"status": "ok",
	})
}

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, SystemInfo())
}

func TimeHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"time": time.Now().Format(time.RFC3339),
	})
}
