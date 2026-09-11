package cmd

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
	"zion-english/internal/trainingmaterials"
	"zion-english/internal/utils"
)

func handleTrainingMaterials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	page := utils.ParsePageQuery(r)

	var rows []queries.TblTrainingMaterial
	var err error
	if auth.HasAdminAccess(user.Role) {
		page.Total, err = dbRO.GetQueries().CountTrainingMaterialsForAdmin(ctx)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to count training materials: %v", err), http.StatusInternalServerError)
			return
		}
		rows, err = dbRO.GetQueries().GetTrainingMaterialsPagedForAdmin(ctx, queries.GetTrainingMaterialsPagedForAdminParams{
			Limit:  int64(page.Size),
			Offset: int64(page.Offset()),
		})
	} else {
		page.Total, err = dbRO.GetQueries().CountTrainingMaterialsForTeacher(ctx)
		if err != nil {
			HttpError(w, fmt.Sprintf("Failed to count training materials: %v", err), http.StatusInternalServerError)
			return
		}
		rows, err = dbRO.GetQueries().GetTrainingMaterialsPagedForTeacher(ctx, queries.GetTrainingMaterialsPagedForTeacherParams{
			Limit:  int64(page.Size),
			Offset: int64(page.Offset()),
		})
	}
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load training materials: %v", err), http.StatusInternalServerError)
		return
	}

	items := buildTrainingMaterialListItems(rows, user)
	filterPath := utils.URL("/training-materials")
	data := frontend.TrainingMaterialsData{
		Materials:      items,
		CanCreate:      trainingmaterials.CanManage(user),
		PageNumber:     page.Number,
		PageTotalPages: page.TotalPages(),
		PageTotal:      page.Total,
		PrevURL:        utils.BuildPageURLAt(filterPath, page.Number-1, page.Size, nil),
		NextURL:        utils.BuildPageURLAt(filterPath, page.Number+1, page.Size, nil),
		HasPrev:        page.HasPrev(),
		HasNext:        page.HasNext(),
	}

	if err := frontend.TrainingMaterials(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTrainingMaterialsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "training-materials", "/view"); ok {
		handleTrainingMaterialView(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "training-materials", "/edit"); ok {
		handleTrainingMaterialEdit(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "training-materials", "/delete"); ok {
		handleTrainingMaterialDelete(w, r, id)
		return
	}
	HttpError(w, "Not found", http.StatusNotFound)
}

func handleTrainingMaterialCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !trainingmaterials.CanManage(user) {
		HttpError(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, "/training-materials")
		return
	}

	req := parseTrainingMaterialRequest(r)
	if err := trainingmaterials.ValidateRequest(req); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/training-materials")
		return
	}

	id, err := dbRW.GetQueries().InsertTrainingMaterial(ctx, queries.InsertTrainingMaterialParams{
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Url:         strings.TrimSpace(req.URL),
		Status:      req.Status,
		CreatedByID: user.ID,
	})
	if err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to create training material: %v", err))
		HttpRedirect(w, r, "/training-materials")
		return
	}

	insertAuditLogAs(ctx, user, "training-materials", fmt.Sprintf("Created training material #%d", id))
	setSuccessFlash(w, "Training material created successfully")
	HttpRedirect(w, r, "/training-materials")
}

func handleTrainingMaterialView(w http.ResponseWriter, r *http.Request, materialID int64) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	material, err := dbRO.GetQueries().GetTrainingMaterialByID(ctx, materialID)
	if err != nil {
		HttpError(w, "Training material not found", http.StatusNotFound)
		return
	}
	if !trainingmaterials.CanView(user, material) {
		HttpError(w, "Forbidden", http.StatusForbidden)
		return
	}

	data := frontend.TrainingMaterialViewData{
		ID:          strconv.FormatInt(materialID, 10),
		Title:       material.Title,
		Description: material.Description,
		URL:         material.Url,
		Status:      material.Status,
		CreatedAt:   material.CreatedAt,
		UpdatedAt:   material.UpdatedAt,
		DeletedAt:   material.DeletedAt.String,
		CanEdit:     trainingmaterials.CanManage(user) && material.Status != string(trainingmaterials.StatusDeleted),
		CanDelete:   trainingmaterials.CanManage(user) && material.Status != string(trainingmaterials.StatusDeleted),
	}

	if err := frontend.TrainingMaterialViewModal(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTrainingMaterialEdit(w http.ResponseWriter, r *http.Request, materialID int64) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !trainingmaterials.CanManage(user) {
		if r.Method == http.MethodGet {
			HttpError(w, "Forbidden", http.StatusForbidden)
		} else {
			setErrorFlash(w, "You do not have permission to edit training materials")
			HttpRedirect(w, r, "/training-materials")
		}
		return
	}

	material, err := dbRO.GetQueries().GetTrainingMaterialByID(ctx, materialID)
	if err != nil {
		if r.Method == http.MethodGet {
			HttpError(w, "Training material not found", http.StatusNotFound)
		} else {
			setErrorFlash(w, "Training material not found")
			HttpRedirect(w, r, "/training-materials")
		}
		return
	}
	if material.Status == string(trainingmaterials.StatusDeleted) {
		if r.Method == http.MethodGet {
			HttpError(w, "Training material not found", http.StatusNotFound)
		} else {
			setErrorFlash(w, "Deleted training materials cannot be edited")
			HttpRedirect(w, r, "/training-materials")
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		data := frontend.TrainingMaterialFormData{
			ID:          strconv.FormatInt(materialID, 10),
			Title:       material.Title,
			Description: material.Description,
			URL:         material.Url,
			Status:      material.Status,
			IsEdit:      true,
		}
		if err := frontend.TrainingMaterialFormModal(data).Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			setErrorFlash(w, "Invalid form data")
			HttpRedirect(w, r, "/training-materials")
			return
		}
		req := parseTrainingMaterialRequest(r)
		if err := trainingmaterials.ValidateRequest(req); err != nil {
			setErrorFlash(w, err.Error())
			HttpRedirect(w, r, "/training-materials")
			return
		}
		if err := dbRW.GetQueries().UpdateTrainingMaterial(ctx, queries.UpdateTrainingMaterialParams{
			Title:       strings.TrimSpace(req.Title),
			Description: strings.TrimSpace(req.Description),
			Url:         strings.TrimSpace(req.URL),
			Status:      req.Status,
			ID:          materialID,
		}); err != nil {
			setErrorFlash(w, fmt.Sprintf("Failed to update training material: %v", err))
			HttpRedirect(w, r, "/training-materials")
			return
		}
		insertAuditLogAs(ctx, user, "training-materials", fmt.Sprintf("Updated training material #%d", materialID))
		setSuccessFlash(w, "Training material updated successfully")
		HttpRedirect(w, r, "/training-materials")
	default:
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTrainingMaterialDelete(w http.ResponseWriter, r *http.Request, materialID int64) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !trainingmaterials.CanManage(user) {
		HttpError(w, "Forbidden", http.StatusForbidden)
		return
	}

	material, err := dbRO.GetQueries().GetTrainingMaterialByID(ctx, materialID)
	if err != nil {
		HttpError(w, "Training material not found", http.StatusNotFound)
		return
	}
	if material.Status == string(trainingmaterials.StatusDeleted) {
		setErrorFlash(w, "Training material is already deleted")
		HttpRedirect(w, r, "/training-materials")
		return
	}

	if err := dbRW.GetQueries().DeleteTrainingMaterial(ctx, materialID); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to delete training material: %v", err))
		HttpRedirect(w, r, "/training-materials")
		return
	}

	insertAuditLogAs(ctx, user, "training-materials", fmt.Sprintf("Deleted training material #%d", materialID))
	setSuccessFlash(w, "Training material deleted successfully")
	HttpRedirect(w, r, "/training-materials")
}

func parseTrainingMaterialRequest(r *http.Request) trainingmaterials.Request {
	status := r.FormValue("status")
	if status == "" {
		status = string(trainingmaterials.StatusDraft)
	}
	return trainingmaterials.Request{
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		URL:         r.FormValue("url"),
		Status:      status,
	}
}

func buildTrainingMaterialListItems(rows []queries.TblTrainingMaterial, user auth.User) []frontend.TrainingMaterialListItem {
	items := make([]frontend.TrainingMaterialListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, frontend.TrainingMaterialListItem{
			ID:                strconv.FormatInt(row.ID, 10),
			Title:             row.Title,
			Description:       row.Description,
			URL:               row.Url,
			Status:            row.Status,
			CreatedAt:         row.CreatedAt,
			UpdatedAt:         row.UpdatedAt,
			PublicationStatus: trainingMaterialPublicationStatus(row.Status),
			PublicationTone:   trainingMaterialPublicationTone(row.Status),
			IsDeleted:         row.Status == string(trainingmaterials.StatusDeleted),
			CanEdit:           trainingmaterials.CanManage(user) && row.Status != string(trainingmaterials.StatusDeleted),
			CanDelete:         trainingmaterials.CanManage(user) && row.Status != string(trainingmaterials.StatusDeleted),
		})
	}
	return items
}

func trainingMaterialPublicationStatus(status string) string {
	switch trainingmaterials.Status(status) {
	case trainingmaterials.StatusPublished:
		return "Published"
	case trainingmaterials.StatusDeleted:
		return "Deleted"
	default:
		return "Draft"
	}
}

func trainingMaterialPublicationTone(status string) frontend.PillTone {
	switch trainingmaterials.Status(status) {
	case trainingmaterials.StatusPublished:
		return frontend.PillToneSuccess
	case trainingmaterials.StatusDeleted:
		return frontend.PillToneError
	default:
		return frontend.PillToneNeutral
	}
}
