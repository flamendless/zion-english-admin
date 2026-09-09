package frontend

import (
	"net/url"

	"zion-english/internal/utils"
)

func teacherSearchURL(hiddenID, inputID, resultsID string) string {
	query := url.Values{}
	query.Set("hiddenId", hiddenID)
	query.Set("inputId", inputID)
	query.Set("resultsId", resultsID)
	return utils.URL("/api/teachers/search") + "?" + query.Encode()
}
