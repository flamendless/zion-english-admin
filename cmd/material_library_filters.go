package cmd

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"zion-english/frontend"
	"zion-english/internal/utils"
)

type materialLibraryFilters struct {
	Query    string
	Status   string
	Access   string
	TagID    int64
	Progress string
}

func parseMaterialLibraryFilters(r *http.Request) materialLibraryFilters {
	tagID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("tag_id")), 10, 64)
	return materialLibraryFilters{
		Query:    strings.TrimSpace(r.URL.Query().Get("q")),
		Status:   strings.TrimSpace(r.URL.Query().Get("status")),
		Access:   strings.TrimSpace(r.URL.Query().Get("access")),
		TagID:    tagID,
		Progress: strings.TrimSpace(r.URL.Query().Get("progress")),
	}
}

func (f materialLibraryFilters) active() bool {
	return f.Query != "" || f.Status != "" || f.Access != "" || f.TagID > 0 || f.Progress != ""
}

func materialLibraryFilterParams(f materialLibraryFilters, sort utils.SortParams) map[string]string {
	params := make(map[string]string)
	if f.Query != "" {
		params["q"] = f.Query
	}
	if f.Status != "" {
		params["status"] = f.Status
	}
	if f.Access != "" {
		params["access"] = f.Access
	}
	if f.TagID > 0 {
		params["tag_id"] = strconv.FormatInt(f.TagID, 10)
	}
	if f.Progress != "" {
		params["progress"] = f.Progress
	}
	for k, v := range sort.QueryValues() {
		if v != "" {
			params[k] = v
		}
	}
	return params
}

func matchesMaterialQuery(title, description, query string) bool {
	if query == "" {
		return true
	}
	q := strings.ToLower(query)
	return strings.Contains(strings.ToLower(title), q) || strings.Contains(strings.ToLower(description), q)
}

func materialHasTag(tagIDs []int64, tagID int64) bool {
	if tagID <= 0 {
		return true
	}
	for _, id := range tagIDs {
		if id == tagID {
			return true
		}
	}
	return false
}

func filterLearningMaterialRows(rows []learningMaterialRow, tagsByMaterial map[int64][]int64, filters materialLibraryFilters) []learningMaterialRow {
	if !filters.active() {
		return rows
	}
	filtered := make([]learningMaterialRow, 0, len(rows))
	for _, row := range rows {
		if filters.Status != "" && row.Status != filters.Status {
			continue
		}
		if filters.Access != "" && row.Access != filters.Access {
			continue
		}
		if !matchesMaterialQuery(row.Title, row.Description, filters.Query) {
			continue
		}
		if !materialHasTag(tagsByMaterial[row.ID], filters.TagID) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func filterTrainingMaterialRows(
	rows []trainingMaterialRow,
	tagsByMaterial map[int64][]int64,
	progressByMaterial map[int64]frontend.TrainingMaterialProgressSummary,
	filters materialLibraryFilters,
	isTeacher bool,
) []trainingMaterialRow {
	if !filters.active() {
		return rows
	}
	filtered := make([]trainingMaterialRow, 0, len(rows))
	for _, row := range rows {
		if filters.Status != "" && row.Status != filters.Status {
			continue
		}
		if !matchesMaterialQuery(row.Title, row.Description, filters.Query) {
			continue
		}
		if !materialHasTag(tagsByMaterial[row.ID], filters.TagID) {
			continue
		}
		if isTeacher && filters.Progress != "" {
			progress := progressByMaterial[row.ID]
			switch filters.Progress {
			case "completed":
				if !progress.IsCompleted {
					continue
				}
			case "in_progress":
				if progress.IsCompleted || progress.ProgressPercent <= 0 {
					continue
				}
			case "not_started":
				if progress.IsCompleted || progress.ProgressPercent > 0 {
					continue
				}
			}
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func loadMaterialTagIDsByMaterial(ctx context.Context, materialIDs []int64, training bool) (map[int64][]int64, error) {
	tagsByMaterial := make(map[int64][]int64, len(materialIDs))
	if len(materialIDs) == 0 {
		return tagsByMaterial, nil
	}
	if training {
		tagRows, err := dbRO.GetQueries().GetTagsByTrainingMaterialIDs(ctx, materialIDs)
		if err != nil {
			return nil, err
		}
		for _, row := range tagRows {
			tagsByMaterial[row.MaterialID] = append(tagsByMaterial[row.MaterialID], row.ID)
		}
		return tagsByMaterial, nil
	}
	tagRows, err := dbRO.GetQueries().GetTagsByMaterialIDs(ctx, materialIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range tagRows {
		tagsByMaterial[row.MaterialID] = append(tagsByMaterial[row.MaterialID], row.ID)
	}
	return tagsByMaterial, nil
}

func materialTagFilterValue(tagID int64) string {
	if tagID <= 0 {
		return ""
	}
	return strconv.FormatInt(tagID, 10)
}
