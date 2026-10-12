package cmd

import (
	"bytes"
	"context"
	"net/http"
	"zion-english/internal/logs"
	"zion-english/internal/utils"

	"github.com/a-h/templ"
	"go.uber.org/zap"
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

// renderHTMLComponent renders to a buffer first so a failed templ render does not commit 200 then HttpError(500).
func renderHTMLComponent(ctx context.Context, w http.ResponseWriter, logTag string, component templ.Component) {
	var buf bytes.Buffer
	if err := component.Render(ctx, &buf); err != nil {
		logs.Log().Error(logTag, zap.Error(err))
		HttpError(w, MsgSomethingWrong, http.StatusInternalServerError)
		return
	}
	writeHTML(w)
	if _, err := w.Write(buf.Bytes()); err != nil {
		logs.Log().Error(logTag, zap.Error(err))
	}
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
