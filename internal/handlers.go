package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"android-go-server-test/internal/cbridge"
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

// SSE, no lib needed, unlike websockets
func StreamHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	send := func() {
		data, err := json.Marshal(LiveMetrics())
		if err != nil {
			return
		}
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	send()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			send()
		}
	}
}

func ChecksumHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]any{
		"checksum": cbridge.FNV1a(body),
		"backend":  cbridge.Backend,
		"bytes":    len(body),
	})
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}

const dashboardHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>android-homelab: live stats</title>
<style>
  body { background: #0b0f14; color: #d6e2ea; font-family: ui-monospace, monospace; padding: 2rem; }
  h1 { font-size: 1rem; color: #7fd08a; margin-bottom: 1.5rem; }
  .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 1rem; max-width: 900px; }
  .card { background: #131a22; border: 1px solid #223; border-radius: 8px; padding: 1rem; }
  .label { font-size: 0.75rem; color: #7a8a99; text-transform: uppercase; letter-spacing: 0.05em; }
  .value { font-size: 1.6rem; margin-top: 0.35rem; color: #e8f0f5; }
  #status { font-size: 0.8rem; color: #7a8a99; margin-bottom: 1rem; }
</style>
</head>
<body>
<h1>android-homelab // live stats</h1>
<div id="status">connecting...</div>
<div class="grid" id="grid"></div>
<script>
const grid = document.getElementById("grid");
const status = document.getElementById("status");
const order = ["uptime", "metrics_source", "load_avg_1m", "mem_used_percent", "mem_available_mb", "mem_total_mb", "go_heap_alloc_mb", "goroutines", "timestamp"];

function addCard(key, value) {
  if (key === "mem_used_percent" && typeof value === "number") value = value.toFixed(1) + "%";
  const card = document.createElement("div");
  card.className = "card";
  card.innerHTML = '<div class="label">' + key.replace(/_/g, " ") + '</div><div class="value">' + value + '</div>';
  grid.appendChild(card);
}

function render(data) {
  grid.innerHTML = "";
  for (const key of order) {
    if (!(key in data)) continue;
    addCard(key, data[key]);
  }
  if (data.helper && typeof data.helper === "object") {
    for (const key of Object.keys(data.helper)) {
      addCard("helper: " + key, data.helper[key]);
    }
  }
}

const source = new EventSource("/stream");
source.onopen = () => { status.textContent = "connected"; };
source.onerror = () => { status.textContent = "disconnected, retrying..."; };
source.onmessage = (event) => render(JSON.parse(event.data));
</script>
</body>
</html>`
