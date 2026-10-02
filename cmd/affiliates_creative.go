package cmd

import (
	"net/http"
	"strings"

	"zion-english/internal/affiliates"
	"zion-english/internal/constants"
	"zion-english/internal/logs"

	"go.uber.org/zap"
)

func handleAffiliateCreativeImageByKey(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	programID := strings.TrimSpace(r.URL.Query().Get("program_id"))
	adID := strings.TrimSpace(r.URL.Query().Get("ad_id"))
	serveImpactCreative(w, r, programID, adID)
}

func handleAffiliateCreativeImageByID(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	id, ok := extractPathID(r, "affiliate-creative", "")
	if !ok {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}

	ctx := r.Context()
	row, err := dbRO.GetQueries().GetAffiliateProductByID(ctx, id)
	if err != nil {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}
	if constants.AffiliateProvider(row.Provider) != constants.AffiliateProviderImpact {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}
	serveImpactCreative(w, r, row.ProgramID, row.ItemID)
}

func serveImpactCreative(w http.ResponseWriter, r *http.Request, programID, adID string) {
	body, contentType, err := affiliates.FetchImpactCreative(r.Context(), programID, adID)
	if err != nil {
		logs.Log().Warn("impact creative image",
			zap.Error(err),
			zap.String("program_id", programID),
			zap.String("ad_id", adID),
		)
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=900")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
