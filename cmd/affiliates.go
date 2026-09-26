package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/affiliates"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/utils"

	"go.uber.org/zap"
)

const maxAffiliateCSVBytes = 2 * 1024 * 1024

func handleAffiliateLink(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	id, ok := extractPathID(r, "affiliate-link", "")
	if !ok {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}

	ctx := r.Context()
	row, err := dbRO.GetQueries().GetAffiliateProductByID(ctx, id)
	if err != nil {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}

	dest := strings.TrimSpace(row.AffiliateUrl)
	if !affiliates.IsValidHTTPURL(dest) {
		HttpError(w, MsgNotFound, http.StatusNotFound)
		return
	}

	if err := dbRW.GetQueries().IncrementAffiliateProductClickCount(ctx, id); err != nil {
		logs.Log().Error("increment affiliate product click count", zap.Error(err), zap.Int64("product_id", id))
	}

	http.Redirect(w, r, dest, http.StatusFound)
}

func handleAffiliatesPath(w http.ResponseWriter, r *http.Request) {
	if id, ok := extractPathID(r, "affiliates", "/view"); ok {
		handleAffiliateView(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "affiliates", "/edit"); ok {
		handleAffiliateEdit(w, r, id)
		return
	}
	if id, ok := extractPathID(r, "affiliates", "/delete"); ok {
		handleAffiliateDelete(w, r, id)
		return
	}
	HttpError(w, MsgNotFound, http.StatusNotFound)
}

func handleAffiliates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleAffiliatesList(w, r, affiliatesListPageState{})
	case http.MethodPost:
		handleAffiliateCreate(w, r)
	default:
		HttpError(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func handleAffiliateUpload(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)

	if err := r.ParseMultipartForm(maxAffiliateCSVBytes); err != nil {
		setErrorFlash(w, "CSV file is too large or invalid")
		HttpRedirect(w, r, "/affiliates")
		return
	}

	file, header, err := r.FormFile("csv")
	if err != nil {
		setErrorFlash(w, "Please choose a CSV file to upload")
		HttpRedirect(w, r, "/affiliates")
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	fileSize := header.Size

	data, err := io.ReadAll(io.LimitReader(file, maxAffiliateCSVBytes+1))
	if err != nil {
		logAffiliateUploadFailure(ctx, user, filename, fileSize, err)
		setErrorFlash(w, "Failed to read CSV file")
		HttpRedirect(w, r, "/affiliates")
		return
	}
	if int64(len(data)) > maxAffiliateCSVBytes {
		logAffiliateUploadFailure(ctx, user, filename, fileSize, fmt.Errorf("file exceeds size limit"))
		setErrorFlash(w, "CSV file is too large")
		HttpRedirect(w, r, "/affiliates")
		return
	}

	rows, err := affiliates.ParseProductLinksCSV(strings.NewReader(string(data)))
	if err != nil {
		logAffiliateUploadFailure(ctx, user, filename, fileSize, err)
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/affiliates")
		return
	}

	staged, thumbsFound, importLogs := buildStagedAffiliateImportItems(ctx, rows)
	if len(staged) == 0 {
		logAffiliateUploadFailure(ctx, user, filename, fileSize, fmt.Errorf("no products to review"))
		setErrorFlash(w, "No products found in CSV")
		HttpRedirect(w, r, "/affiliates")
		return
	}

	insertUploadLog(ctx, user, uploadLogEntry{
		Module:        "affiliates",
		Outcome:       constants.UploadLogOutcomeSucceeded,
		Kind:          constants.UploadLogKindAffiliateCSV,
		Summary:       fmt.Sprintf("Parsed %d products from '%s' (%d thumbnails found)", len(staged), filename, thumbsFound),
		Filename:      filename,
		FileSize:      fileSize,
		FileSizeValid: fileSize > 0,
	})

	handleAffiliatesList(w, r, affiliatesListPageState{
		ImportPreview: frontend.AffiliateImportPreviewData{
			Filename:        filename,
			FileSizeDisplay: utils.FormatFileSize(fileSize),
			FileSizeBytes:   fileSize,
			ProcessedCount:  len(staged),
			TotalCount:      len(rows),
			ThumbnailsFound: thumbsFound,
			PendingSave:     true,
			ImportLogs:      importLogs,
			StagedItems:     staged,
		},
		OpenPreviewModal: true,
	})
}

func buildStagedAffiliateImportItems(ctx context.Context, rows []affiliates.CSVRow) ([]frontend.AffiliateStagedImportItem, int, []frontend.AffiliateImportLogEntry) {
	staged := make([]frontend.AffiliateStagedImportItem, 0, len(rows))
	importLogs := make([]frontend.AffiliateImportLogEntry, 0)
	thumbsFound := 0
	for i, row := range rows {
		shopeeShopID, itemID := "", row.ItemID
		if ids, ok := affiliates.ParseProductIDsFromURL(row.ProductLink); ok {
			shopeeShopID = ids.ShopID
			if itemID == "" {
				itemID = ids.ItemID
			}
		}

		thumb, thumbErr := affiliates.FetchThumbnailFromProductURL(ctx, row.ProductLink)
		if thumbErr != nil {
			importLogs = append(importLogs, frontend.AffiliateImportLogEntry{
				ItemID:  itemID,
				Message: affiliates.UserFacingMessage(thumbErr),
			})
		} else {
			thumbsFound++
		}
		priceDisplay := affiliates.FormatPriceDisplay(row.Price)

		staged = append(staged, frontend.AffiliateStagedImportItem{
			Index: i,
			Card: frontend.AffiliateProductCardData{
				Name:          row.ItemName,
				ShopName:      row.ShopName,
				PriceDisplay:  priceDisplay,
				Sales:         row.Sales,
				ThumbnailURL:  thumb,
				AffiliateURL:  row.OfferLink,
				IncludeInSave: true,
				FormIndex:     i,
			},
			ItemID:         itemID,
			ItemName:       row.ItemName,
			Price:          row.Price,
			Sales:          row.Sales,
			ShopName:       row.ShopName,
			CommissionRate: row.CommissionRate,
			Commission:     row.Commission,
			ProductLink:    row.ProductLink,
			OfferLink:      row.OfferLink,
			ShopeeShopID:   shopeeShopID,
			ThumbnailURL:   thumb,
			PriceDisplay:   priceDisplay,
		})
	}
	return staged, thumbsFound, importLogs
}

func handleAffiliateImportSave(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)

	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, "/affiliates")
		return
	}

	filename := strings.TrimSpace(r.FormValue("import_filename"))
	fileSize, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("import_file_size")), 10, 64)
	csvRowCount, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("import_csv_row_count")))
	stagedCount, _ := strconv.Atoi(strings.TrimSpace(r.FormValue("staged_count")))
	if stagedCount <= 0 || filename == "" {
		setErrorFlash(w, "Nothing to save")
		HttpRedirect(w, r, "/affiliates")
		return
	}

	included := make(map[int]bool)
	for _, raw := range r.Form["include"] {
		idx, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || idx < 0 || idx >= stagedCount {
			continue
		}
		included[idx] = true
	}
	if len(included) == 0 {
		setErrorFlash(w, "Select at least one product to save")
		HttpRedirect(w, r, "/affiliates")
		return
	}

	var createdBy sql.NullInt64
	if user.ID > 0 {
		createdBy = sql.NullInt64{Int64: user.ID, Valid: true}
	}
	batchID, err := dbRW.GetQueries().InsertAffiliateImportBatch(ctx, queries.InsertAffiliateImportBatchParams{
		Filename:      filename,
		FileSizeBytes: fileSize,
		CsvRowCount:   int64(csvRowCount),
		CreatedBy:     createdBy,
		CreatedByName: user.Name,
	})
	if err != nil {
		setErrorFlash(w, "Failed to record import batch")
		HttpRedirect(w, r, "/affiliates")
		return
	}

	inserted := 0
	sortOrder := 0
	for i := 0; i < stagedCount; i++ {
		if !included[i] {
			continue
		}
		prefix := fmt.Sprintf("staged_%d_", i)
		offerLink := strings.TrimSpace(r.FormValue(prefix + "offer_link"))
		productLink := strings.TrimSpace(r.FormValue(prefix + "product_link"))
		itemName := strings.TrimSpace(r.FormValue(prefix + "item_name"))
		if offerLink == "" || productLink == "" || itemName == "" {
			continue
		}

		affiliatedShopID, err := database.EnsureAffiliatedProductShopID(ctx, dbRW, r.FormValue(prefix+"shop_name"))
		if err != nil {
			logs.Log().Error("ensure affiliated product shop on import save", zap.Error(err))
			continue
		}

		_, err = dbRW.GetQueries().InsertAffiliateProduct(ctx, queries.InsertAffiliateProductParams{
			AffiliateUrl:     offerLink,
			ProductUrl:       productLink,
			ShopID:           strings.TrimSpace(r.FormValue(prefix + "shopee_shop_id")),
			ItemID:           strings.TrimSpace(r.FormValue(prefix + "item_id")),
			Name:             itemName,
			Brand:            "",
			PriceDisplay:     strings.TrimSpace(r.FormValue(prefix + "price_display")),
			ThumbnailUrl:     strings.TrimSpace(r.FormValue(prefix + "thumbnail_url")),
			SortOrder:        int64(sortOrder),
			ImportBatchID:    batchID,
			Sales:            strings.TrimSpace(r.FormValue(prefix + "sales")),
			AffiliatedShopID: affiliatedShopID,
			CommissionRate:   strings.TrimSpace(r.FormValue(prefix + "commission_rate")),
			Commission:       strings.TrimSpace(r.FormValue(prefix + "commission")),
		})
		if err != nil {
			logs.Log().Error("insert affiliate product from import save", zap.Error(err))
			continue
		}
		inserted++
		sortOrder++
	}

	if inserted == 0 {
		setErrorFlash(w, "Failed to save any products")
		HttpRedirect(w, r, "/affiliates")
		return
	}

	if err := dbRW.GetQueries().UpdateAffiliateImportBatchCount(ctx, queries.UpdateAffiliateImportBatchCountParams{
		ProductCount: int64(inserted),
		ID:           batchID,
	}); err != nil {
		logs.Log().Error("update affiliate import batch count", zap.Error(err))
	}

	insertAuditLogAs(ctx, user, "affiliates", fmt.Sprintf("Saved %d products from CSV batch #%d (%s)", inserted, batchID, filename))
	setSuccessFlash(w, fmt.Sprintf("Saved %d affiliate products.", inserted))
	HttpRedirect(w, r, "/affiliates")
}

func logAffiliateUploadFailure(ctx context.Context, user auth.User, filename string, fileSize int64, err error) {
	insertUploadLog(ctx, user, uploadLogEntry{
		Module:        "affiliates",
		Outcome:       constants.UploadLogOutcomeFailed,
		Kind:          constants.UploadLogKindAffiliateCSV,
		Summary:       fmt.Sprintf("affiliate CSV upload failed for '%s': %v", filename, err),
		Filename:      filename,
		FileSize:      fileSize,
		FileSizeValid: fileSize > 0,
	})
}

type affiliatesListPageState struct {
	Form             frontend.AffiliateFormData
	OpenEditModal    bool
	ImportPreview    frontend.AffiliateImportPreviewData
	OpenPreviewModal bool
}

func handleAffiliatesList(w http.ResponseWriter, r *http.Request, state affiliatesListPageState) {
	ctx := r.Context()
	sort := parseListSort(r, frontend.ListSortKindAffiliate)
	query := strings.TrimSpace(r.URL.Query().Get("q"))

	rows, err := dbRO.GetQueries().GetAllAffiliateProducts(ctx)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load affiliates: %v", err), http.StatusInternalServerError)
		return
	}

	rows = filterAffiliateRows(rows, query)
	sortAffiliateRows(rows, sort)

	items := make([]frontend.AffiliateListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapAffiliateListItem(row))
	}

	data := frontend.AffiliatesPageData{
		Items:            items,
		Query:            query,
		SortBy:           sort.By,
		SortOrder:        string(sort.Order),
		FilterPath:       utils.URL("/affiliates"),
		Form:             state.Form,
		OpenEditModal:    state.OpenEditModal,
		ImportPreview:    state.ImportPreview,
		OpenPreviewModal: state.OpenPreviewModal,
	}

	if err := frontend.AffiliatesPage(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func mapAffiliateListItem(row queries.GetAllAffiliateProductsRow) frontend.AffiliateListItem {
	shop := row.ShopBrandName
	if shop == "" {
		shop = row.Brand
	}
	return frontend.AffiliateListItem{
		ID:            strconv.FormatInt(row.ID, 10),
		Name:          row.Name,
		ShopName:      shop,
		PriceDisplay:  row.PriceDisplay,
		ClickCount:    row.ClickCount,
		ThumbnailURL:  row.ThumbnailUrl,
		AffiliateURL:  frontend.AffiliateProductLinkHref(row.ID, row.AffiliateUrl),
		AffiliateDisp: frontend.AffiliateURLPreview(row.AffiliateUrl),
	}
}

func mapAffiliateProductCardFromBatch(row queries.GetAffiliateProductsByImportBatchIDRow) frontend.AffiliateProductCardData {
	return mapAffiliateProductCard(
		row.ID,
		row.Name,
		row.ShopBrandName,
		row.PriceDisplay,
		row.Sales,
		row.ThumbnailUrl,
		row.AffiliateUrl,
	)
}

func mapAffiliateProductCardFromByID(row queries.GetAffiliateProductByIDRow) frontend.AffiliateProductCardData {
	return mapAffiliateProductCard(
		row.ID,
		row.Name,
		row.ShopBrandName,
		row.PriceDisplay,
		row.Sales,
		row.ThumbnailUrl,
		row.AffiliateUrl,
	)
}

func mapAffiliateProductCard(productID int64, name, shopName, priceDisplay, sales, thumbnailURL, affiliateURL string) frontend.AffiliateProductCardData {
	return frontend.AffiliateProductCardData{
		ProductID:    productID,
		Name:         name,
		ShopName:     shopName,
		PriceDisplay: priceDisplay,
		Sales:        sales,
		ThumbnailURL: thumbnailURL,
		AffiliateURL: affiliateURL,
	}
}

func handleAffiliateView(w http.ResponseWriter, r *http.Request, affiliateID int64) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	row, err := dbRO.GetQueries().GetAffiliateProductByID(ctx, affiliateID)
	if err != nil {
		setErrorFlash(w, "Affiliate product not found")
		HttpRedirect(w, r, "/affiliates")
		return
	}

	card := mapAffiliateProductCardFromByID(row)
	if r.Header.Get(headerHXRequest) == "true" {
		writeHTML(w)
		if err := frontend.AffiliateProductViewModal(card).Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	HttpRedirect(w, r, "/affiliates")
}

func handleAffiliateCreate(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, "/affiliates")
		return
	}

	req := parseAffiliateSaveRequest(r)
	if err := affiliates.ValidateSave(req); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/affiliates")
		return
	}
	req = affiliates.NormalizeSaveRequest(req)

	affiliatedShopID, err := database.EnsureAffiliatedProductShopID(ctx, dbRW, req.ShopName)
	if err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to save shop: %v", err))
		HttpRedirect(w, r, "/affiliates")
		return
	}

	id, err := dbRW.GetQueries().InsertAffiliateProduct(ctx, queries.InsertAffiliateProductParams{
		AffiliateUrl:     req.AffiliateURL,
		ProductUrl:       req.ProductURL,
		ShopID:           req.ShopID,
		ItemID:           req.ItemID,
		Name:             req.Name,
		Brand:            req.Brand,
		PriceDisplay:     req.PriceDisplay,
		ThumbnailUrl:     req.ThumbnailURL,
		SortOrder:        req.SortOrder,
		ImportBatchID:    0,
		Sales:            req.Sales,
		AffiliatedShopID: affiliatedShopID,
		CommissionRate:   req.CommissionRate,
		Commission:       req.Commission,
	})
	if err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to save affiliate product: %v", err))
		HttpRedirect(w, r, "/affiliates")
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "affiliates", fmt.Sprintf("Added affiliate product #%d: %s", id, req.Name))
	setSuccessFlash(w, "Affiliate product saved")
	HttpRedirect(w, r, "/affiliates")
}

func handleAffiliateEdit(w http.ResponseWriter, r *http.Request, affiliateID int64) {
	switch r.Method {
	case http.MethodGet:
		ctx := r.Context()
		row, err := dbRO.GetQueries().GetAffiliateProductByID(ctx, affiliateID)
		if err != nil {
			if r.Header.Get(headerHXRequest) == "true" {
				HttpError(w, "Affiliate product not found", http.StatusNotFound)
				return
			}
			setErrorFlash(w, "Affiliate product not found")
			HttpRedirect(w, r, "/affiliates")
			return
		}
		form := affiliateFormFromRow(row)
		if r.Header.Get(headerHXRequest) == "true" {
			writeHTML(w)
			if err := frontend.AffiliateEditModal(form, true).Render(ctx, w); err != nil {
				HttpError(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		handleAffiliatesList(w, r, affiliatesListPageState{
			Form:          form,
			OpenEditModal: true,
		})
	case http.MethodPost:
		handleAffiliateUpdate(w, r, affiliateID)
	default:
		HttpError(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func handleAffiliateUpdate(w http.ResponseWriter, r *http.Request, affiliateID int64) {
	ctx := r.Context()
	editPath := fmt.Sprintf("/affiliates/%d/edit", affiliateID)

	_, err := dbRO.GetQueries().GetAffiliateProductByID(ctx, affiliateID)
	if err != nil {
		setErrorFlash(w, "Affiliate product not found")
		HttpRedirect(w, r, "/affiliates")
		return
	}

	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form data")
		HttpRedirect(w, r, editPath)
		return
	}

	req := parseAffiliateSaveRequest(r)
	if err := affiliates.ValidateSave(req); err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, editPath)
		return
	}
	req = affiliates.NormalizeSaveRequest(req)

	affiliatedShopID, err := database.EnsureAffiliatedProductShopID(ctx, dbRW, req.ShopName)
	if err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to save shop: %v", err))
		HttpRedirect(w, r, editPath)
		return
	}

	if err := dbRW.GetQueries().UpdateAffiliateProduct(ctx, queries.UpdateAffiliateProductParams{
		AffiliateUrl:     req.AffiliateURL,
		ProductUrl:       req.ProductURL,
		ShopID:           req.ShopID,
		ItemID:           req.ItemID,
		Name:             req.Name,
		Brand:            req.Brand,
		PriceDisplay:     req.PriceDisplay,
		ThumbnailUrl:     req.ThumbnailURL,
		SortOrder:        req.SortOrder,
		Sales:            req.Sales,
		AffiliatedShopID: affiliatedShopID,
		CommissionRate:   req.CommissionRate,
		Commission:       req.Commission,
		ID:               affiliateID,
	}); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to update affiliate product: %v", err))
		HttpRedirect(w, r, editPath)
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "affiliates", fmt.Sprintf("Updated affiliate product #%d: %s", affiliateID, req.Name))
	setSuccessFlash(w, "Affiliate product updated")
	HttpRedirect(w, r, "/affiliates")
}

func handleAffiliateDelete(w http.ResponseWriter, r *http.Request, affiliateID int64) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	row, err := dbRO.GetQueries().GetAffiliateProductByID(ctx, affiliateID)
	if err != nil {
		HttpError(w, "Affiliate product not found", http.StatusNotFound)
		return
	}

	if err := dbRW.GetQueries().DeleteAffiliateProduct(ctx, affiliateID); err != nil {
		setErrorFlash(w, fmt.Sprintf("Failed to delete affiliate product: %v", err))
		HttpRedirect(w, r, "/affiliates")
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "affiliates", fmt.Sprintf("Deleted affiliate product #%d: %s", affiliateID, row.Name))
	setSuccessFlash(w, "Affiliate product deleted")
	HttpRedirect(w, r, "/affiliates")
}

func affiliateFormFromRow(row queries.GetAffiliateProductByIDRow) frontend.AffiliateFormData {
	return frontend.AffiliateFormData{
		ID:             strconv.FormatInt(row.ID, 10),
		AffiliateURL:   row.AffiliateUrl,
		ProductURL:     row.ProductUrl,
		ShopID:         row.ShopID,
		ItemID:         row.ItemID,
		Name:           row.Name,
		Brand:          row.Brand,
		PriceDisplay:   row.PriceDisplay,
		ThumbnailURL:   row.ThumbnailUrl,
		SortOrder:      row.SortOrder,
		Sales:          row.Sales,
		ShopName:       row.ShopBrandName,
		CommissionRate: row.CommissionRate,
		Commission:     row.Commission,
		IsEdit:         true,
	}
}

func parseAffiliateSaveRequest(r *http.Request) affiliates.SaveRequest {
	sortOrder := int64(0)
	if raw := strings.TrimSpace(r.FormValue("sort_order")); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
			sortOrder = n
		}
	}
	return affiliates.SaveRequest{
		AffiliateURL:   r.FormValue("affiliate_url"),
		ProductURL:     r.FormValue("product_url"),
		ShopID:         r.FormValue("shop_id"),
		ItemID:         r.FormValue("item_id"),
		Name:           r.FormValue("name"),
		Brand:          r.FormValue("brand"),
		PriceDisplay:   r.FormValue("price_display"),
		ThumbnailURL:   r.FormValue("thumbnail_url"),
		SortOrder:      sortOrder,
		Sales:          r.FormValue("sales"),
		ShopName:       r.FormValue("shop_name"),
		CommissionRate: r.FormValue("commission_rate"),
		Commission:     r.FormValue("commission"),
	}
}
