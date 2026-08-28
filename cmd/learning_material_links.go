package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
	"zion-english/internal/learningmaterials"
)

func parseLearningMaterialIDs(r *http.Request) []int64 {
	seen := make(map[int64]struct{})
	var ids []int64
	for _, raw := range r.Form["learning_material_ids"] {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func validateLearningMaterialIDs(ctx context.Context, user auth.User, ids []int64) error {
	for _, id := range ids {
		material, err := dbRO.GetQueries().GetLearningMaterialByID(ctx, id)
		if err != nil {
			return fmt.Errorf("learning material %d not found", id)
		}
		if material.Status == learningmaterials.StatusDeleted || material.DeletedAt.Valid {
			return fmt.Errorf("learning material %d is not available", id)
		}
		if !learningmaterials.CanView(user, material.OwnerID, material.Status, material.Access) {
			return fmt.Errorf("you do not have access to learning material %d", id)
		}
	}
	return nil
}

func replaceClassRecordLearningMaterials(ctx context.Context, q *queries.Queries, classRecordID int64, materialIDs []int64) error {
	if err := q.DeleteClassRecordLearningMaterialLinks(ctx, classRecordID); err != nil {
		return err
	}
	for _, materialID := range materialIDs {
		if err := q.InsertClassRecordLearningMaterialLink(ctx, queries.InsertClassRecordLearningMaterialLinkParams{
			ClassRecordID: classRecordID,
			MaterialID:    materialID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func replaceScheduledClassLearningMaterials(ctx context.Context, q *queries.Queries, scheduleID int64, materialIDs []int64) error {
	if err := q.DeleteScheduledClassLearningMaterialLinks(ctx, scheduleID); err != nil {
		return err
	}
	for _, materialID := range materialIDs {
		if err := q.InsertScheduledClassLearningMaterialLink(ctx, queries.InsertScheduledClassLearningMaterialLinkParams{
			ScheduledClassID: scheduleID,
			MaterialID:       materialID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func copyScheduledClassLearningMaterials(ctx context.Context, q *queries.Queries, scheduleID, classRecordID int64) error {
	rows, err := q.GetLearningMaterialsByScheduledClassID(ctx, scheduleID)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return replaceClassRecordLearningMaterials(ctx, q, classRecordID, ids)
}

func loadClassRecordLearningMaterialLinks(ctx context.Context, classRecordID int64) ([]frontend.ClassLearningMaterialLink, error) {
	rows, err := dbRO.GetQueries().GetLearningMaterialsByClassRecordID(ctx, classRecordID)
	if err != nil {
		return nil, err
	}
	return mapLearningMaterialLinkRows(rows), nil
}

func loadScheduledClassLearningMaterialLinks(ctx context.Context, scheduleID int64) ([]frontend.ClassLearningMaterialLink, error) {
	rows, err := dbRO.GetQueries().GetLearningMaterialsByScheduledClassID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	return mapScheduledLearningMaterialLinkRows(rows), nil
}

func mapLearningMaterialLinkRows(rows []queries.GetLearningMaterialsByClassRecordIDRow) []frontend.ClassLearningMaterialLink {
	items := make([]frontend.ClassLearningMaterialLink, 0, len(rows))
	for _, row := range rows {
		items = append(items, frontend.ClassLearningMaterialLink{
			ID:    strconv.FormatInt(row.ID, 10),
			Title: row.Title,
			URL:   row.Url,
		})
	}
	return items
}

func mapScheduledLearningMaterialLinkRows(rows []queries.GetLearningMaterialsByScheduledClassIDRow) []frontend.ClassLearningMaterialLink {
	items := make([]frontend.ClassLearningMaterialLink, 0, len(rows))
	for _, row := range rows {
		items = append(items, frontend.ClassLearningMaterialLink{
			ID:    strconv.FormatInt(row.ID, 10),
			Title: row.Title,
			URL:   row.Url,
		})
	}
	return items
}

func searchLearningMaterialLinks(ctx context.Context, user auth.User, q string) ([]frontend.ClassLearningMaterialLink, error) {
	isSuperuser := int64(0)
	if user.Role == auth.RoleSuperuser || user.Role == auth.RoleAdmin {
		isSuperuser = 1
	}
	rows, err := dbRO.GetQueries().SearchLearningMaterialsByTitle(ctx, queries.SearchLearningMaterialsByTitleParams{
		Column1: q,
		Column2: sql.NullString{String: q, Valid: q != ""},
		Column3: isSuperuser,
		OwnerID: user.ID,
	})
	if err != nil {
		return nil, err
	}
	items := make([]frontend.ClassLearningMaterialLink, 0, len(rows))
	for _, row := range rows {
		items = append(items, frontend.ClassLearningMaterialLink{
			ID:    strconv.FormatInt(row.ID, 10),
			Title: row.Title,
			URL:   row.Url,
		})
	}
	return items, nil
}

func handleSearchLearningMaterials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	w.Header().Set("Content-Type", "text/html")
	if q == "" {
		if err := frontend.LearningMaterialSearchResults(nil).Render(r.Context(), w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	items, err := searchLearningMaterialLinks(r.Context(), auth.GetUser(r.Context()), q)
	if err != nil {
		HttpError(w, "Failed to search learning materials", http.StatusInternalServerError)
		return
	}

	if err := frontend.LearningMaterialSearchResults(items).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func saveClassRecordLearningMaterials(ctx context.Context, user auth.User, classRecordID int64, materialIDs []int64) error {
	if err := validateLearningMaterialIDs(ctx, user, materialIDs); err != nil {
		return err
	}
	return replaceClassRecordLearningMaterials(ctx, dbRW.GetQueries(), classRecordID, materialIDs)
}

func saveScheduledClassLearningMaterials(ctx context.Context, user auth.User, scheduleID int64, materialIDs []int64) error {
	if err := validateLearningMaterialIDs(ctx, user, materialIDs); err != nil {
		return err
	}
	return replaceScheduledClassLearningMaterials(ctx, dbRW.GetQueries(), scheduleID, materialIDs)
}
