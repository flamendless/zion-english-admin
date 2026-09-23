package cmd

import (
	"encoding/json"
	"net/http"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		HttpError(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	stats := dbRO.Health()
	writeJSON(w)
	if stats["status"] != "up" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	if r.Method == http.MethodHead {
		return
	}
	_ = json.NewEncoder(w).Encode(stats)
}
