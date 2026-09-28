package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/ads"
	"zion-english/internal/featureflags"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

const adAffiliateSearchPageSize = 20

type adsListPageState struct {
	Form           frontend.AdFormData
	OpenCreateModal bool
	OpenEditModal   bool
}

func handleAdsPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "ads", "/view"); ok {
		handleAdView(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "ads", "/edit"); ok {
		handleAdEdit(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "ads", "/delete"); ok {
		handleAdDeletePath(w, r, id)
		return
	}
	HttpError(w, MsgNotFound, http.StatusNotFound)
}

func handleAds(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleAdsList(w, r, adsListPageState{})
	case http.MethodPost:
		handleAdCreate(w, r)
	default:
		HttpError(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func handleAdsList(w http.ResponseWriter, r *http.Request, state adsListPageState) {
	ctx := r.Context()
	sort := parseListSort(r, frontend.ListSortKindAds)
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))

	rows, err := dbRO.GetQueries().GetAllAds(ctx)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load ads: %v", err), http.StatusInternalServerError)
		return
	}

	productCounts := make(map[int64]int64)
	for _, row := range rows {
		n, err := dbRO.GetQueries().CountAdAffiliateProducts(ctx, row.ID)
		if err != nil {
			logs.Log().Error("count ad affiliate products", zap.Error(err), zap.Int64("ad_id", row.ID))
			continue
		}
		productCounts[row.ID] = n
	}

	rows = filterAdsRows(rows, query, statusFilter)
	sortAdsRows(rows, productCounts, sort)

	items := make([]frontend.AdListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapAdListItem(row, productCounts[row.ID]))
	}

	data := frontend.AdsPageData{
		Items:           items,
		Query:           query,
		StatusFilter:    statusFilter,
		SortBy:          sort.By,
		SortOrder:       string(sort.Order),
		FilterPath:      utils.URL("/ads"),
		Form:            state.Form,
		OpenCreateModal: state.OpenCreateModal,
		OpenEditModal:   state.OpenEditModal,
	}

	if err := frontend.AdsPage(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func mapAdListItem(row queries.TblAd, productCount int64) frontend.AdListItem {
	status := constants.AdStatus(row.Status)
	return frontend.AdListItem{
		ID:              strconv.FormatInt(row.ID, 10),
		Name:            row.Name,
		Status:          status,
		StatusLabel:     frontend.AdStatusLabel(status),
		StatusTone:      frontend.AdStatusPillTone(status),
		Placement:       constants.AdPlacement(row.Placement),
		PlacementLabel:  frontend.AdPlacementLabel(constants.AdPlacement(row.Placement)),
		AdType:          constants.AdType(row.AdType),
		TypeLabel:       frontend.AdTypeLabel(constants.AdType(row.AdType)),
		ProductCount:    productCount,
		RandomizeSummary: frontend.AdRandomizeSummary(
			constants.AdRandomizeKind(row.RandomizeKind),
			constants.AdTimerInterval(row.TimerInterval),
		),
		IsDeleted: status == constants.AdStatusDeleted,
	}
}

func handleAdCreate(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, "/ads")
		return
	}
	req := parseAdRequest(r)
	req = ads.NormalizeRequest(req)
	if err := ads.ValidateRequest(req); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/ads")
		return
	}

	id, err := saveAd(ctx, 0, req)
	if err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/ads")
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "ads", fmt.Sprintf("Created ad #%d: %s", id, req.Name))
	setSuccessFlash(w, "Ad saved")
	HttpRedirect(w, r, "/ads")
}

func handleAdEdit(w http.ResponseWriter, r *http.Request, adID int64) {
	switch r.Method {
	case http.MethodGet:
		ctx := r.Context()
		form, err := adFormFromID(ctx, adID)
		if err != nil {
			if r.Header.Get(headerHXRequest) == "true" {
				HttpError(w, "Ad not found", http.StatusNotFound)
				return
			}
			setErrorFlash(w, "Ad not found")
			HttpRedirect(w, r, "/ads")
			return
		}
		if r.Header.Get(headerHXRequest) == "true" {
			writeHTML(w)
			if err := frontend.AdEditModal(form, true).Render(ctx, w); err != nil {
				HttpError(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		handleAdsList(w, r, adsListPageState{Form: form, OpenEditModal: true})
	case http.MethodPost:
		handleAdUpdate(w, r, adID)
	default:
		HttpError(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func handleAdUpdate(w http.ResponseWriter, r *http.Request, adID int64) {
	ctx := r.Context()
	row, err := dbRO.GetQueries().GetAdByID(ctx, adID)
	if err != nil {
		setErrorFlash(w, "Ad not found")
		HttpRedirect(w, r, "/ads")
		return
	}
	if row.Status == string(constants.AdStatusDeleted) {
		setErrorFlash(w, "Deleted ads cannot be edited")
		HttpRedirect(w, r, "/ads")
		return
	}

	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, fmt.Sprintf("/ads/%d/edit", adID))
		return
	}
	req := parseAdRequest(r)
	req = ads.NormalizeRequest(req)
	if err := ads.ValidateRequest(req); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, fmt.Sprintf("/ads/%d/edit", adID))
		return
	}

	if _, err := saveAd(ctx, adID, req); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, fmt.Sprintf("/ads/%d/edit", adID))
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "ads", fmt.Sprintf("Updated ad #%d: %s", adID, req.Name))
	setSuccessFlash(w, "Ad updated")
	HttpRedirect(w, r, "/ads")
}

func handleAdView(w http.ResponseWriter, r *http.Request, adID int64) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx := r.Context()
	view, err := adViewFromID(ctx, adID)
	if err != nil {
		if r.Header.Get(headerHXRequest) == "true" {
			HttpError(w, "Ad not found", http.StatusNotFound)
			return
		}
		setErrorFlash(w, "Ad not found")
		HttpRedirect(w, r, "/ads")
		return
	}
	if r.Header.Get(headerHXRequest) == "true" {
		writeHTML(w)
		if err := frontend.AdViewModal(view).Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	HttpRedirect(w, r, "/ads")
}

func handleAdDeletePath(w http.ResponseWriter, r *http.Request, adID int64) {
	switch r.Method {
	case http.MethodGet:
		ctx := r.Context()
		row, err := dbRO.GetQueries().GetAdByID(ctx, adID)
		if err != nil {
			HttpError(w, "Ad not found", http.StatusNotFound)
			return
		}
		if r.Header.Get(headerHXRequest) == "true" {
			writeHTML(w)
			if err := frontend.AdDeleteModal(strconv.FormatInt(adID, 10), row.Name).Render(ctx, w); err != nil {
				HttpError(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		HttpRedirect(w, r, "/ads")
	case http.MethodPost:
		handleAdSoftDelete(w, r, adID)
	default:
		HttpError(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func handleAdSoftDelete(w http.ResponseWriter, r *http.Request, adID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	ctx := r.Context()
	row, err := dbRO.GetQueries().GetAdByID(ctx, adID)
	if err != nil {
		HttpError(w, "Ad not found", http.StatusNotFound)
		return
	}
	if row.Status == string(constants.AdStatusDeleted) {
		setErrorFlash(w, "Ad is already deleted")
		HttpRedirect(w, r, "/ads")
		return
	}
	if err := dbRW.GetQueries().SoftDeleteAd(ctx, adID); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to delete ad: %v", err))
		HttpRedirect(w, r, "/ads")
		return
	}
	insertAuditLogAs(ctx, auth.GetUser(ctx), "ads", fmt.Sprintf("Deleted ad #%d: %s", adID, row.Name))
	setSuccessFlash(w, "Ad deleted")
	HttpRedirect(w, r, "/ads")
}

func handleAdAffiliateSearch(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx := r.Context()
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	formID := strings.TrimSpace(r.URL.Query().Get("form_id"))
	if formID == "" {
		formID = "adCreateForm"
	}
	page := 1
	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			page = n
		}
	}
	selected := r.URL.Query()["affiliate_product_ids"]
	selectedSet := make(map[string]struct{}, len(selected))
	for _, id := range selected {
		selectedSet[strings.TrimSpace(id)] = struct{}{}
	}

	countParams := queries.CountAffiliateProductsForAdsParams{
		Column1: q,
		Column2: sql.NullString{String: q, Valid: q != ""},
		Column3: sql.NullString{String: q, Valid: q != ""},
	}
	total, err := dbRO.GetQueries().CountAffiliateProductsForAds(ctx, countParams)
	if err != nil {
		HttpError(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}
	totalPages := int((total + adAffiliateSearchPageSize - 1) / adAffiliateSearchPageSize)
	if totalPages < 1 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}
	offset := int64((page - 1) * adAffiliateSearchPageSize)

	rows, err := dbRO.GetQueries().SearchAffiliateProductsForAds(ctx, queries.SearchAffiliateProductsForAdsParams{
		Column1: q,
		Column2: sql.NullString{String: q, Valid: q != ""},
		Column3: sql.NullString{String: q, Valid: q != ""},
		Limit:   adAffiliateSearchPageSize,
		Offset:  offset,
	})
	if err != nil {
		HttpError(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	items := make([]frontend.AdAffiliateSearchItem, 0, len(rows))
	for _, row := range rows {
		idStr := strconv.FormatInt(row.ID, 10)
		_, checked := selectedSet[idStr]
		items = append(items, frontend.AdAffiliateSearchItem{
			ID:           idStr,
			Name:         row.Name,
			ShopName:     row.ShopName,
			PriceDisplay: row.PriceDisplay,
			ThumbnailURL: row.ThumbnailUrl,
			Checked:      checked,
		})
	}

	panel := frontend.AdAffiliateSearchPanelData{
		Items:         items,
		Query:         q,
		Page:          page,
		TotalPages:    totalPages,
		Total:         total,
		SelectedCount: len(selectedSet),
		FormID:        formID,
	}

	writeHTML(w)
	if err := frontend.AdAffiliateSearchPanel(panel).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

type adAffiliateProductIDsResponse struct {
	IDs []string `json:"ids"`
}

func handleAdAffiliateProductIDs(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx := r.Context()
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	rows, err := dbRO.GetQueries().ListAffiliateProductIDsForAds(ctx, queries.ListAffiliateProductIDsForAdsParams{
		Column1: q,
		Column2: sql.NullString{String: q, Valid: q != ""},
		Column3: sql.NullString{String: q, Valid: q != ""},
	})
	if err != nil {
		HttpError(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}
	ids := make([]string, 0, len(rows))
	for _, id := range rows {
		ids = append(ids, strconv.FormatInt(id, 10))
	}
	writeJSON(w)
	if err := json.NewEncoder(w).Encode(adAffiliateProductIDsResponse{IDs: ids}); err != nil {
		logs.Log().Error("encode affiliate product ids", zap.Error(err))
	}
}

func saveAd(ctx context.Context, adID int64, req ads.Request) (int64, error) {
	req = ads.NormalizeRequest(req)
	for _, pid := range req.AffiliateProductIDs {
		if _, err := dbRO.GetQueries().GetAffiliateProductByID(ctx, pid); err != nil {
			return 0, fmt.Errorf("affiliate product #%d not found", pid)
		}
	}

	if adID == 0 {
		id, err := dbRW.GetQueries().InsertAd(ctx, queries.InsertAdParams{
			Name:           req.Name,
			Placement:      req.Placement,
			AdType:         req.AdType,
			Status:         req.Status,
			RandomizeKind:  req.RandomizeKind,
			TimerInterval:  req.TimerInterval,
			SortOrder:      req.SortOrder,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to save ad: %w", err)
		}
		adID = id
	} else {
		if err := dbRW.GetQueries().UpdateAd(ctx, queries.UpdateAdParams{
			Name:          req.Name,
			Placement:     req.Placement,
			AdType:        req.AdType,
			Status:        req.Status,
			RandomizeKind: req.RandomizeKind,
			TimerInterval: req.TimerInterval,
			SortOrder:     req.SortOrder,
			ID:            adID,
		}); err != nil {
			return 0, fmt.Errorf("failed to update ad: %w", err)
		}
	}

	if err := dbRW.GetQueries().DeleteAdAffiliateProducts(ctx, adID); err != nil {
		return 0, fmt.Errorf("failed to update ad products: %w", err)
	}
	for i, pid := range req.AffiliateProductIDs {
		if err := dbRW.GetQueries().InsertAdAffiliateProduct(ctx, queries.InsertAdAffiliateProductParams{
			AdID:                adID,
			AffiliateProductID:  pid,
			SortOrder:           int64(i),
		}); err != nil {
			return 0, fmt.Errorf("failed to link affiliate product: %w", err)
		}
	}
	return adID, nil
}

func parseAdRequest(r *http.Request) ads.Request {
	sortOrder := int64(0)
	if raw := strings.TrimSpace(r.FormValue("sort_order")); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
			sortOrder = n
		}
	}
	var productIDs []int64
	for _, raw := range r.Form["affiliate_product_ids"] {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
			productIDs = append(productIDs, id)
		}
	}
	return ads.Request{
		Name:                r.FormValue("name"),
		Placement:           r.FormValue("placement"),
		AdType:              r.FormValue("ad_type"),
		Status:              r.FormValue("status"),
		RandomizeKind:       r.FormValue("randomize_kind"),
		TimerInterval:       r.FormValue("timer_interval"),
		SortOrder:           sortOrder,
		AffiliateProductIDs: productIDs,
	}
}

func adFormFromID(ctx context.Context, adID int64) (frontend.AdFormData, error) {
	row, err := dbRO.GetQueries().GetAdByID(ctx, adID)
	if err != nil {
		return frontend.AdFormData{}, err
	}
	ids, err := dbRO.GetQueries().GetAdAffiliateProductIDs(ctx, adID)
	if err != nil {
		return frontend.AdFormData{}, err
	}
	selected := make([]frontend.AdSelectedProduct, 0, len(ids))
	for _, pid := range ids {
		prod, err := dbRO.GetQueries().GetAffiliateProductByID(ctx, pid)
		if err != nil {
			continue
		}
		shop := prod.ShopBrandName
		if shop == "" {
			shop = prod.Brand
		}
		selected = append(selected, frontend.AdSelectedProduct{
			ID:           strconv.FormatInt(prod.ID, 10),
			Name:         prod.Name,
			ShopName:     shop,
			PriceDisplay: prod.PriceDisplay,
		})
	}
	return frontend.AdFormData{
		ID:              strconv.FormatInt(row.ID, 10),
		Name:            row.Name,
		Placement:       constants.AdPlacement(row.Placement),
		AdType:          constants.AdType(row.AdType),
		Status:          constants.AdStatus(row.Status),
		RandomizeKind:   constants.AdRandomizeKind(row.RandomizeKind),
		TimerInterval:   constants.AdTimerInterval(row.TimerInterval),
		SortOrder:       row.SortOrder,
		SelectedProducts: selected,
		IsEdit:          true,
		IsDeleted:       row.Status == string(constants.AdStatusDeleted),
	}, nil
}

func adViewFromID(ctx context.Context, adID int64) (frontend.AdViewData, error) {
	row, err := dbRO.GetQueries().GetAdByID(ctx, adID)
	if err != nil {
		return frontend.AdViewData{}, err
	}
	ids, err := dbRO.GetQueries().GetAdAffiliateProductIDs(ctx, adID)
	if err != nil {
		return frontend.AdViewData{}, err
	}
	cards := make([]frontend.AffiliateProductCardData, 0, len(ids))
	for _, pid := range ids {
		prod, err := dbRO.GetQueries().GetAffiliateProductByID(ctx, pid)
		if err != nil {
			continue
		}
		shop := prod.ShopBrandName
		if shop == "" {
			shop = prod.Brand
		}
		cards = append(cards, mapAffiliateProductCard(
			prod.ID,
			prod.Name,
			shop,
			prod.PriceDisplay,
			prod.Sales,
			prod.ThumbnailUrl,
			prod.AffiliateUrl,
		))
	}
	status := constants.AdStatus(row.Status)
	return frontend.AdViewData{
		Name:             row.Name,
		StatusLabel:      frontend.AdStatusLabel(status),
		StatusTone:       frontend.AdStatusPillTone(status),
		PlacementLabel:   frontend.AdPlacementLabel(constants.AdPlacement(row.Placement)),
		TypeLabel:        frontend.AdTypeLabel(constants.AdType(row.AdType)),
		RandomizeSummary: frontend.AdRandomizeSummary(
			constants.AdRandomizeKind(row.RandomizeKind),
			constants.AdTimerInterval(row.TimerInterval),
		),
		ProductCards: cards,
	}, nil
}

func handleAdChromePartial(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	ctx := r.Context()
	if !featureflags.AdsDisplayEnabled(ctx, dbRO) || ads.SuppressAdsForRequest(ctx, r, dbRO.GetQueries()) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	ctx = ads.LoadRequestContext(ctx, w, r, dbRO.GetQueries(), true)
	slots := ads.GetResolvedSlots(ctx)
	if len(slots.Top) == 0 && len(slots.Left) == 0 && len(slots.Right) == 0 &&
		len(slots.Bottom) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeHTML(w)
	if err := frontend.AdChromeRefreshFragment().Render(ctx, w); err != nil {
		logs.Log().Error("render ad chrome partial", zap.Error(err))
	}
}
