package internal

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"sync"
)

// in-memory on purpose - contrast against guestbook.go's sqlite persistence
var pastes = struct {
	mu    sync.RWMutex
	items map[string]string
}{items: make(map[string]string)}

func newPasteID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func PastebinCreateHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil || len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	id := newPasteID()
	pastes.mu.Lock()
	pastes.items[id] = string(body)
	pastes.mu.Unlock()

	writeJSON(w, map[string]string{"id": id, "url": "/pastebin/" + id})
}

func PastebinGetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	pastes.mu.RLock()
	content, ok := pastes.items[id]
	pastes.mu.RUnlock()

	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(content))
}
