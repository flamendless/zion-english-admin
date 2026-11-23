package cmd

import (
	"net/http"

	"zion-english/frontend"
)

func handleAccesses(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	data := frontend.AccessesData{
		Matrix: frontend.NavAccessMatrixRows(),
	}
	if err := frontend.Accesses(data).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}
