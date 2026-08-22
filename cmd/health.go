package cmd

import (
	"encoding/json"
	"net/http"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := dbRO.Health()
	w.Header().Set("Content-Type", "application/json")
	if stats["status"] != "up" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	if r.Method == http.MethodHead {
		return
	}
	_ = json.NewEncoder(w).Encode(stats)
}
