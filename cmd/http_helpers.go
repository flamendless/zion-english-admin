package cmd

import (
	"net/http"
	"zion-english/internal/utils"
)

const (
	headerContentType = "Content-Type"
	headerHXRedirect  = "HX-Redirect"
	headerHXTrigger   = "HX-Trigger"
	headerHXRequest   = "HX-Request"
	mimeHTML          = "text/html"
	mimeJSON          = "application/json"
)

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		HttpError(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func writeHTML(w http.ResponseWriter) {
	w.Header().Set(headerContentType, mimeHTML)
}

func writeJSON(w http.ResponseWriter) {
	w.Header().Set(headerContentType, mimeJSON)
}

func setHXRedirect(w http.ResponseWriter, path string) {
	w.Header().Set(headerHXRedirect, utils.URL(path))
}

const hxTriggerBackgroundJobsRefresh = "backgroundJobsRefresh"

func triggerBackgroundJobsRefresh(w http.ResponseWriter) {
	w.Header().Set(headerHXTrigger, hxTriggerBackgroundJobsRefresh)
}
