package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/database/queries"
	"zion-english/internal/metatags"
)

func handleMetaPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "meta", "/edit"); ok {
		handleMetaEdit(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "meta", "/delete"); ok {
		handleMetaDelete(w, r, id)
		return
	}
	HttpError(w, MsgNotFound, http.StatusNotFound)
}

func handleMeta(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleMetaList(w, r)
	case http.MethodPost:
		handleMetaCreate(w, r)
	default:
		HttpError(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func handleMetaList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := dbRO.GetQueries().GetAllMetaTags(ctx)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load meta tags: %v", err), http.StatusInternalServerError)
		return
	}

	items := make([]frontend.MetaTagListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, frontend.MetaTagListItem{
			ID:          strconv.FormatInt(row.ID, 10),
			Name:        row.Name,
			Content:     row.Content,
			Value:       row.Value,
			SortOrder:   row.SortOrder,
			ContentDisp: frontend.MetaTagPreview(row.Content),
			ValueDisp:   frontend.MetaTagPreview(row.Value),
		})
	}

	if err := frontend.MetaTagsPage(frontend.MetaTagsListData{Tags: items}).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleMetaCreate(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, "/meta")
		return
	}

	req := parseMetaTagRequest(r)
	if err := metatags.ValidateRequest(req); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/meta")
		return
	}
	req = metatags.NormalizeRequest(req)

	existing, err := dbRO.GetQueries().GetMetaTagByName(ctx, req.Name)
	if err == nil && existing.ID > 0 {
		setErrorFlash(w, metatags.ErrDuplicateName.Error())
		HttpRedirect(w, r, "/meta")
		return
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		setErrorFlash(w, fmt.Sprintf("Failed to check meta tag: %v", err))
		HttpRedirect(w, r, "/meta")
		return
	}

	id, err := dbRW.GetQueries().InsertMetaTag(ctx, queries.InsertMetaTagParams{
		Name:      req.Name,
		Content:   req.Content,
		Value:     req.Value,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			setErrorFlash(w, metatags.ErrDuplicateName.Error())
		} else {
			setErrorFlash(w, fmt.Sprintf("Failed to add meta tag: %v", err))
		}
		HttpRedirect(w, r, "/meta")
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "meta", fmt.Sprintf("Added meta tag #%d: %s", id, req.Name))
	setSuccessFlash(w, "Meta tag added successfully")
	HttpRedirect(w, r, "/meta")
}

func handleMetaEdit(w http.ResponseWriter, r *http.Request, metaID int64) {
	switch r.Method {
	case http.MethodGet:
		ctx := r.Context()
		row, err := dbRO.GetQueries().GetMetaTagByID(ctx, metaID)
		if err != nil {
			setErrorFlash(w, "Meta tag not found")
			HttpRedirect(w, r, "/meta")
			return
		}

		if err := frontend.MetaTagEditPage(frontend.MetaTagFormData{
			ID:        strconv.FormatInt(metaID, 10),
			Name:      row.Name,
			Content:   row.Content,
			Value:     row.Value,
			SortOrder: row.SortOrder,
			IsEdit:    true,
		}).Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		handleMetaUpdate(w, r, metaID)
	default:
		HttpError(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func handleMetaUpdate(w http.ResponseWriter, r *http.Request, metaID int64) {
	ctx := r.Context()
	editPath := fmt.Sprintf("/meta/%d/edit", metaID)

	existing, err := dbRO.GetQueries().GetMetaTagByID(ctx, metaID)
	if err != nil {
		setErrorFlash(w, "Meta tag not found")
		HttpRedirect(w, r, "/meta")
		return
	}

	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, editPath)
		return
	}

	req := parseMetaTagRequest(r)
	if err := metatags.ValidateRequest(req); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, editPath)
		return
	}
	req = metatags.NormalizeRequest(req)

	if req.Name != existing.Name {
		other, err := dbRO.GetQueries().GetMetaTagByName(ctx, req.Name)
		if err == nil && other.ID != metaID {
			setErrorFlash(w, metatags.ErrDuplicateName.Error())
			HttpRedirect(w, r, editPath)
			return
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			setErrorFlash(w, fmt.Sprintf("Failed to check meta tag: %v", err))
			HttpRedirect(w, r, editPath)
			return
		}
	}

	if err := dbRW.GetQueries().UpdateMetaTag(ctx, queries.UpdateMetaTagParams{
		Name:      req.Name,
		Content:   req.Content,
		Value:     req.Value,
		SortOrder: req.SortOrder,
		ID:        metaID,
	}); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			setErrorFlash(w, metatags.ErrDuplicateName.Error())
		} else {
			setErrorFlash(w, fmt.Sprintf("Failed to update meta tag: %v", err))
		}
		HttpRedirect(w, r, editPath)
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "meta", fmt.Sprintf("Updated meta tag #%d: %s", metaID, req.Name))
	setSuccessFlash(w, "Meta tag updated successfully")
	HttpRedirect(w, r, "/meta")
}

func handleMetaDelete(w http.ResponseWriter, r *http.Request, metaID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	row, err := dbRO.GetQueries().GetMetaTagByID(ctx, metaID)
	if err != nil {
		HttpError(w, "Meta tag not found", http.StatusNotFound)
		return
	}

	if err := dbRW.GetQueries().DeleteMetaTag(ctx, metaID); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to delete meta tag: %v", err))
		HttpRedirect(w, r, "/meta")
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "meta", fmt.Sprintf("Deleted meta tag #%d: %s", metaID, row.Name))
	setSuccessFlash(w, "Meta tag deleted successfully")
	HttpRedirect(w, r, "/meta")
}

func parseMetaTagRequest(r *http.Request) metatags.Request {
	sortOrder := int64(0)
	if raw := strings.TrimSpace(r.FormValue("sort_order")); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n >= 0 {
			sortOrder = n
		}
	}
	return metatags.Request{
		Name:      r.FormValue("name"),
		Content:   r.FormValue("content"),
		Value:     r.FormValue("value"),
		SortOrder: sortOrder,
	}
}
