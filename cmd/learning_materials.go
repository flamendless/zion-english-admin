package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
	"zion-english/internal/learningmaterials"
	"zion-english/internal/utils"
)

func handleLearningMaterials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	page := utils.ParsePageQuery(r)

	var rows []queries.TblLearningMaterial
	var err error
	if user.Role == auth.RoleSuperuser {
		page.Total, err = dbRO.GetQueries().CountLearningMaterialsForSuperuser(ctx)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to count materials: %v", err), http.StatusInternalServerError)
			return
		}
		rows, err = dbRO.GetQueries().GetLearningMaterialsPagedForSuperuser(ctx, queries.GetLearningMaterialsPagedForSuperuserParams{
			Limit:  int64(page.Size),
			Offset: int64(page.Offset()),
		})
	} else {
		page.Total, err = dbRO.GetQueries().CountLearningMaterialsForUser(ctx, user.ID)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to count materials: %v", err), http.StatusInternalServerError)
			return
		}
		rows, err = dbRO.GetQueries().GetLearningMaterialsPagedForUser(ctx, queries.GetLearningMaterialsPagedForUserParams{
			OwnerID: user.ID,
			Limit:   int64(page.Size),
			Offset:  int64(page.Offset()),
		})
	}
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load materials: %v", err), http.StatusInternalServerError)
		return
	}

	items, err := buildLearningMaterialListItems(ctx, rows, user)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	existingTags, err := dbRO.GetQueries().GetAllLearningMaterialTags(ctx)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load tags: %v", err), http.StatusInternalServerError)
		return
	}

	filterPath := utils.URL("/learning-materials")
	data := frontend.LearningMaterialsData{
		Materials:      items,
		ExistingTags:   mapLearningMaterialTags(existingTags),
		CanCreate:      true,
		PageNumber:     page.Number,
		PageTotalPages: page.TotalPages(),
		PageTotal:      page.Total,
		PrevURL:        utils.BuildPageURLAt(filterPath, page.Number-1, page.Size, nil),
		NextURL:        utils.BuildPageURLAt(filterPath, page.Number+1, page.Size, nil),
		HasPrev:        page.HasPrev(),
		HasNext:        page.HasNext(),
	}

	if err := frontend.LearningMaterials(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleLearningMaterialsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "learning-materials", "/view"); ok {
		handleLearningMaterialView(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "learning-materials", "/edit"); ok {
		handleLearningMaterialEdit(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "learning-materials", "/delete"); ok {
		handleLearningMaterialDelete(w, r, id)
		return
	}
	HttpError(w, "Not found", http.StatusNotFound)
}

func handleLearningMaterialCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, "/learning-materials")
		return
	}

	req := parseLearningMaterialRequest(r)
	if err := learningmaterials.ValidateRequest(req); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/learning-materials")
		return
	}

	ownerID := user.ID
	if user.Role == auth.RoleSuperuser {
		ownerID = 0
	}

	id, err := dbRW.GetQueries().InsertLearningMaterial(ctx, queries.InsertLearningMaterialParams{
		OwnerID:     ownerID,
		Description: strings.TrimSpace(req.Description),
		Url:         strings.TrimSpace(req.URL),
		Access:      req.Access,
		Status:      req.Status,
	})
	if err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to create material: %v", err))
		HttpRedirect(w, r, "/learning-materials")
		return
	}

	if err := learningmaterials.ReplaceMaterialTags(ctx, dbRW.GetQueries(), id, req.TagLabels); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/learning-materials")
		return
	}

	insertAuditLogAs(ctx, user, "learning-materials", fmt.Sprintf("Created learning material #%d", id))
	setSuccessFlash(w, "Learning material created successfully")
	HttpRedirect(w, r, "/learning-materials")
}

func handleLearningMaterialView(w http.ResponseWriter, r *http.Request, materialID int64) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	material, tags, err := loadLearningMaterialWithTags(ctx, materialID)
	if err != nil {
		HttpError(w, "Learning material not found", http.StatusNotFound)
		return
	}
	if !learningmaterials.CanView(user, material) {
		HttpError(w, "Forbidden", http.StatusForbidden)
		return
	}

	ownerName, err := resolveLearningMaterialOwnerName(ctx, material.OwnerID)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := frontend.LearningMaterialViewData{
		ID:          strconv.FormatInt(materialID, 10),
		Description: material.Description,
		URL:         material.Url,
		Access:      material.Access,
		Status:      material.Status,
		OwnerName:   ownerName,
		CreatedAt:   material.CreatedAt,
		UpdatedAt:   material.UpdatedAt,
		DeletedAt:   material.DeletedAt.String,
		Tags:        mapLearningMaterialTags(tags),
		CanEdit:     learningmaterials.CanEdit(user, material),
		CanDelete:   learningmaterials.CanDelete(user),
	}

	if err := frontend.LearningMaterialViewModal(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleLearningMaterialEdit(w http.ResponseWriter, r *http.Request, materialID int64) {
	ctx := r.Context()
	user := auth.GetUser(ctx)

	material, err := dbRO.GetQueries().GetLearningMaterialByID(ctx, materialID)
	if err != nil {
		if r.Method == http.MethodGet {
			HttpError(w, "Learning material not found", http.StatusNotFound)
		} else {
			setErrorFlash(w, "Learning material not found")
			HttpRedirect(w, r, "/learning-materials")
		}
		return
	}
	if !learningmaterials.CanEdit(user, material) {
		if r.Method == http.MethodGet {
			HttpError(w, "Forbidden", http.StatusForbidden)
		} else {
			setErrorFlash(w, "You do not have permission to edit this material")
			HttpRedirect(w, r, "/learning-materials")
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		tags, err := dbRO.GetQueries().GetTagsByMaterialID(ctx, materialID)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to load tags: %v", err), http.StatusInternalServerError)
			return
		}
		existingTags, err := dbRO.GetQueries().GetAllLearningMaterialTags(ctx)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to load tags: %v", err), http.StatusInternalServerError)
			return
		}

		data := frontend.LearningMaterialFormData{
			ID:           strconv.FormatInt(materialID, 10),
			Description:  material.Description,
			URL:          material.Url,
			Access:       material.Access,
			Status:       material.Status,
			SelectedTags: mapLearningMaterialTags(tags),
			ExistingTags: mapLearningMaterialTags(existingTags),
			IsEdit:       true,
		}
		if err := frontend.LearningMaterialFormModal(data).Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			setErrorFlash(w, "Invalid form data")
			HttpRedirect(w, r, "/learning-materials")
			return
		}
		req := parseLearningMaterialRequest(r)
		if err := learningmaterials.ValidateRequest(req); err != nil {
			setErrorFlash(w, err.Error())
			HttpRedirect(w, r, "/learning-materials")
			return
		}
		if err := dbRW.GetQueries().UpdateLearningMaterial(ctx, queries.UpdateLearningMaterialParams{
			Description: strings.TrimSpace(req.Description),
			Url:         strings.TrimSpace(req.URL),
			Access:      req.Access,
			Status:      req.Status,
			ID:          materialID,
		}); err != nil {
			setErrorFlash(w, fmt.Sprintf("Failed to update material: %v", err))
			HttpRedirect(w, r, "/learning-materials")
			return
		}
		if err := learningmaterials.ReplaceMaterialTags(ctx, dbRW.GetQueries(), materialID, req.TagLabels); err != nil {
			setErrorFlash(w, err.Error())
			HttpRedirect(w, r, "/learning-materials")
			return
		}
		insertAuditLogAs(ctx, user, "learning-materials", fmt.Sprintf("Updated learning material #%d", materialID))
		setSuccessFlash(w, "Learning material updated successfully")
		HttpRedirect(w, r, "/learning-materials")
	default:
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleLearningMaterialDelete(w http.ResponseWriter, r *http.Request, materialID int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !learningmaterials.CanDelete(user) {
		HttpError(w, "Forbidden", http.StatusForbidden)
		return
	}

	material, err := dbRO.GetQueries().GetLearningMaterialByID(ctx, materialID)
	if err != nil {
		HttpError(w, "Learning material not found", http.StatusNotFound)
		return
	}
	if material.Status == learningmaterials.StatusDeleted {
		setErrorFlash(w, "Learning material is already deleted")
		HttpRedirect(w, r, "/learning-materials")
		return
	}

	if err := dbRW.GetQueries().DeleteLearningMaterial(ctx, materialID); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to delete material: %v", err))
		HttpRedirect(w, r, "/learning-materials")
		return
	}

	insertAuditLogAs(ctx, user, "learning-materials", fmt.Sprintf("Deleted learning material #%d", materialID))
	setSuccessFlash(w, "Learning material deleted successfully")
	HttpRedirect(w, r, "/learning-materials")
}

func parseLearningMaterialRequest(r *http.Request) learningmaterials.Request {
	status := r.FormValue("status")
	if status == "" {
		status = learningmaterials.StatusDraft
	}
	access := r.FormValue("access")
	if access == "" {
		access = learningmaterials.AccessPublic
	}
	return learningmaterials.Request{
		Description: r.FormValue("description"),
		URL:         r.FormValue("url"),
		Access:      access,
		Status:      status,
		TagLabels:   r.Form["tags"],
	}
}

func loadLearningMaterialWithTags(ctx context.Context, materialID int64) (queries.TblLearningMaterial, []queries.TblLearningMaterialTag, error) {
	material, err := dbRO.GetQueries().GetLearningMaterialByID(ctx, materialID)
	if err != nil {
		return queries.TblLearningMaterial{}, nil, err
	}
	tags, err := dbRO.GetQueries().GetTagsByMaterialID(ctx, materialID)
	if err != nil {
		return queries.TblLearningMaterial{}, nil, err
	}
	return material, tags, nil
}

func buildLearningMaterialListItems(ctx context.Context, rows []queries.TblLearningMaterial, user auth.User) ([]frontend.LearningMaterialListItem, error) {
	if len(rows) == 0 {
		return nil, nil
	}

	materialIDs := make([]int64, 0, len(rows))
	ownerIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		materialIDs = append(materialIDs, row.ID)
		if row.OwnerID > 0 {
			ownerIDs = append(ownerIDs, row.OwnerID)
		}
	}

	tagRows, err := dbRO.GetQueries().GetTagsByMaterialIDs(ctx, materialIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}
	tagsByMaterial := make(map[int64][]frontend.LearningMaterialTag)
	for _, tr := range tagRows {
		tagsByMaterial[tr.MaterialID] = append(tagsByMaterial[tr.MaterialID], frontend.LearningMaterialTag{
			ID:    strconv.FormatInt(tr.ID, 10),
			Label: tr.Label,
			Color: tr.Color,
		})
	}

	ownerNames, err := resolveLearningMaterialOwnerNames(ctx, ownerIDs)
	if err != nil {
		return nil, err
	}

	items := make([]frontend.LearningMaterialListItem, 0, len(rows))
	for _, row := range rows {
		ownerName := "Superuser"
		if row.OwnerID > 0 {
			ownerName = ownerNames[row.OwnerID]
			if ownerName == "" {
				ownerName = "Unknown"
			}
		}
		items = append(items, frontend.LearningMaterialListItem{
			ID:                strconv.FormatInt(row.ID, 10),
			Description:       row.Description,
			URL:               row.Url,
			Access:            row.Access,
			Status:            row.Status,
			OwnerName:         ownerName,
			CreatedAt:         row.CreatedAt,
			UpdatedAt:         row.UpdatedAt,
			Tags:              tagsByMaterial[row.ID],
			PublicationStatus: learningMaterialPublicationStatus(row.Status),
			PublicationTone:   learningMaterialPublicationTone(row.Status),
			AccessLabel:       learningMaterialAccessLabel(row.Access),
			AccessTone:        learningMaterialAccessTone(row.Access),
			IsDeleted:         row.Status == learningmaterials.StatusDeleted,
			CanEdit:           learningmaterials.CanEdit(user, row),
			CanDelete:         learningmaterials.CanDelete(user),
		})
	}
	return items, nil
}

func resolveLearningMaterialOwnerName(ctx context.Context, ownerID int64) (string, error) {
	if ownerID == 0 {
		return "Superuser", nil
	}
	teacher, err := dbRO.GetQueries().GetTeacherNameByID(ctx, ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "Unknown", nil
		}
		return "", fmt.Errorf("failed to load owner: %w", err)
	}
	return utils.ComposePersonName(teacher.FirstName, teacher.MiddleName, teacher.LastName), nil
}

func resolveLearningMaterialOwnerNames(ctx context.Context, ownerIDs []int64) (map[int64]string, error) {
	names := make(map[int64]string, len(ownerIDs))
	for _, id := range ownerIDs {
		if _, ok := names[id]; ok {
			continue
		}
		name, err := resolveLearningMaterialOwnerName(ctx, id)
		if err != nil {
			return nil, err
		}
		names[id] = name
	}
	return names, nil
}

func mapLearningMaterialTags(tags []queries.TblLearningMaterialTag) []frontend.LearningMaterialTag {
	out := make([]frontend.LearningMaterialTag, 0, len(tags))
	for _, tag := range tags {
		out = append(out, frontend.LearningMaterialTag{
			ID:    strconv.FormatInt(tag.ID, 10),
			Label: tag.Label,
			Color: tag.Color,
		})
	}
	return out
}

func learningMaterialPublicationStatus(status string) string {
	switch status {
	case learningmaterials.StatusPublished:
		return "Published"
	case learningmaterials.StatusDeleted:
		return "Deleted"
	default:
		return "Draft"
	}
}

func learningMaterialPublicationTone(status string) frontend.PillTone {
	switch status {
	case learningmaterials.StatusPublished:
		return frontend.PillToneSuccess
	case learningmaterials.StatusDeleted:
		return frontend.PillToneError
	default:
		return frontend.PillToneNeutral
	}
}

func learningMaterialAccessLabel(access string) string {
	if access == learningmaterials.AccessPrivate {
		return "Private"
	}
	return "Public"
}

func learningMaterialAccessTone(access string) frontend.PillTone {
	if access == learningmaterials.AccessPrivate {
		return frontend.PillToneWarning
	}
	return frontend.PillToneInfo
}
