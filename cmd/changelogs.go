package cmd

import (
	"errors"
	"net/http"
	"os"
	"zion-english/frontend"
	"zion-english/internal/changelog"
	"zion-english/internal/logs"
	"zion-english/internal/version"

	"go.uber.org/zap"
)

func handleChangelogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := changelog.ResolvePath()
	doc, err := changelog.Read(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logs.Log().Error("changelogs file missing", zap.String("path", path), zap.Error(err))
			data := changelog.BuildChangelogsView(nil, version.Get())
			if renderErr := frontend.Changelogs(data).Render(r.Context(), w); renderErr != nil {
				logs.Log().Error("failed to render changelogs page", zap.Error(renderErr))
			}
			return
		}
		logs.Log().Error("failed to read changelogs", zap.String("path", path), zap.Error(err))
		HttpError(w, "Failed to read changelogs", http.StatusInternalServerError)
		return
	}

	data := changelog.BuildChangelogsView(doc, version.Get())
	if err := frontend.Changelogs(data).Render(r.Context(), w); err != nil {
		logs.Log().Error("failed to render changelogs page", zap.Error(err))
	}
}
