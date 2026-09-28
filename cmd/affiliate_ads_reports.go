package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"zion-english/frontend"
	"zion-english/internal/constants"
	"zion-english/internal/utils"
)

const affiliateAdsReportTopN = 10

func handleAffiliateAdsReports(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()

	summaryRow, err := dbRO.GetQueries().GetAffiliateClickSummary(ctx)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load affiliate summary: %v", err), http.StatusInternalServerError)
		return
	}

	adRows, err := dbRO.GetQueries().ListAdsClickReport(ctx)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load ad report: %v", err), http.StatusInternalServerError)
		return
	}

	var publishedAds int64
	adItems := make([]frontend.AffiliateAdsReportAdItem, 0, len(adRows))
	for _, row := range adRows {
		status := constants.AdStatus(row.Status)
		if status == constants.AdStatusPublished {
			publishedAds++
		}
		adItems = append(adItems, frontend.AffiliateAdsReportAdItem{
			ID:               strconv.FormatInt(row.ID, 10),
			Name:             row.Name,
			Status:           status,
			StatusLabel:      frontend.AdStatusLabel(status),
			StatusTone:       frontend.AdStatusPillTone(status),
			PlacementLabel:   frontend.AdPlacementLabel(constants.AdPlacement(row.Placement)),
			ProductCount:     row.ProductCount,
			LinkedClickTotal: sqlAggregateInt64(row.LinkedClickTotal),
		})
	}

	topProducts, err := loadTopAffiliateProductsByClicks(ctx, affiliateAdsReportTopN)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	topAds := topAffiliateAdsReportAds(adItems, affiliateAdsReportTopN)

	topTeachers, err := loadTopTeachersByAffiliateClickEvents(ctx, affiliateAdsReportTopN)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	topAdZones, err := loadTopAdZoneClickCounts(ctx, affiliateAdsReportTopN)
	if err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	attributedAdClicks, err := dbRO.GetQueries().CountAffiliateLinkClickEventsWithAd(ctx)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to load attributed clicks: %v", err), http.StatusInternalServerError)
		return
	}

	totalClicks := sqlAggregateInt64(summaryRow.TotalClicks)
	productsWithoutClicks := summaryRow.ProductCount - summaryRow.ProductsWithClicks
	if productsWithoutClicks < 0 {
		productsWithoutClicks = 0
	}

	data := frontend.AffiliateAdsReportsPageData{
		Summary: frontend.AffiliateAdsReportSummary{
			ProductCount:          summaryRow.ProductCount,
			TotalClicks:           totalClicks,
			ProductsWithClicks:    summaryRow.ProductsWithClicks,
			ProductsWithoutClicks: productsWithoutClicks,
			PublishedAdsCount:     publishedAds,
			AttributedAdClicks:    attributedAdClicks,
		},
		TopProducts: topProducts,
		TopAds:      topAds,
		TopTeachers: topTeachers,
		TopAdZones:  topAdZones,
		Charts: frontend.AffiliateAdsReportsChartData{
			TopProducts:    chartPointsFromProducts(topProducts),
			TopAds:         chartPointsFromAds(topAds),
			TopTeachers:    chartPointsFromTeachers(topTeachers),
			ClicksByAdZone: chartPointsFromAdZones(topAdZones),
			Summary: frontend.AffiliateAdsReportsChartSummary{
				ProductsWithClicks:    summaryRow.ProductsWithClicks,
				ProductsWithoutClicks: productsWithoutClicks,
			},
		},
	}

	writeHTML(w)
	if err := frontend.AffiliateAdsReportsPage(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadTopAffiliateProductsByClicks(ctx context.Context, limit int) ([]frontend.AffiliateAdsReportProductItem, error) {
	allRows, err := dbRO.GetQueries().GetAllAffiliateProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load affiliate products: %w", err)
	}
	sort := utils.SortParams{By: "click_count", Order: utils.SortOrderDesc}
	sortAffiliateRows(allRows, sort)

	items := make([]frontend.AffiliateAdsReportProductItem, 0, limit)
	for _, row := range allRows {
		if row.ClickCount <= 0 {
			break
		}
		shop := row.ShopBrandName
		if shop == "" {
			shop = row.Brand
		}
		items = append(items, mapAffiliateAdsReportProduct(row.ID, row.Name, shop, row.ThumbnailUrl, row.ClickCount))
		if len(items) >= limit {
			break
		}
	}
	return items, nil
}

func topAffiliateAdsReportAds(ads []frontend.AffiliateAdsReportAdItem, limit int) []frontend.AffiliateAdsReportAdItem {
	sorted := make([]frontend.AffiliateAdsReportAdItem, len(ads))
	copy(sorted, ads)
	utils.SortSlice(sorted, utils.SortOrderDesc, func(a, b frontend.AffiliateAdsReportAdItem) int {
		return utils.CompareInt64(a.LinkedClickTotal, b.LinkedClickTotal)
	})
	out := make([]frontend.AffiliateAdsReportAdItem, 0, limit)
	for _, item := range sorted {
		if item.LinkedClickTotal <= 0 {
			break
		}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func chartPointsFromProducts(items []frontend.AffiliateAdsReportProductItem) []frontend.AffiliateAdsReportChartPoint {
	out := make([]frontend.AffiliateAdsReportChartPoint, 0, len(items))
	for _, item := range items {
		label := truncateAffiliateChartLabel(item.Name, 36)
		if item.ShopName != "" {
			label = truncateAffiliateChartLabel(item.Name+" · "+item.ShopName, 36)
		}
		out = append(out, frontend.AffiliateAdsReportChartPoint{
			Label:  label,
			Clicks: item.ClickCount,
		})
	}
	return out
}

func loadTopTeachersByAffiliateClickEvents(ctx context.Context, limit int) ([]frontend.AffiliateAdsReportTeacherItem, error) {
	rows, err := dbRO.GetQueries().ListTopTeachersByAffiliateClickEvents(ctx, int64(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to load teacher click events: %w", err)
	}
	items := make([]frontend.AffiliateAdsReportTeacherItem, 0, len(rows))
	for _, row := range rows {
		if !row.TeacherID.Valid || row.TeacherID.Int64 <= 0 {
			continue
		}
		name := utils.ComposePersonName(row.FirstName, row.MiddleName, row.LastName)
		items = append(items, frontend.AffiliateAdsReportTeacherItem{
			TeacherID:   row.TeacherID.Int64,
			TeacherName: name,
			Avatar: buildTeacherListAvatarProps(
				row.TeacherID.Int64,
				row.FirstName,
				row.MiddleName,
				row.LastName,
				row.AssignedColor,
				row.ProfilePicture,
			),
			ClickCount: row.ClickCount,
		})
	}
	return items, nil
}

func loadTopAdZoneClickCounts(ctx context.Context, limit int) ([]frontend.AffiliateAdsReportAdZoneItem, error) {
	rows, err := dbRO.GetQueries().ListAdZoneClickCounts(ctx, int64(limit))
	if err != nil {
		return nil, fmt.Errorf("failed to load ad zone click counts: %w", err)
	}
	items := make([]frontend.AffiliateAdsReportAdZoneItem, 0, len(rows))
	for _, row := range rows {
		if !row.AdZone.Valid || row.AdZone.String == "" {
			continue
		}
		zone := constants.AdZone(row.AdZone.String)
		items = append(items, frontend.AffiliateAdsReportAdZoneItem{
			AdName:     row.AdName,
			Zone:       zone,
			ZoneLabel:  frontend.AdZoneLabel(zone),
			ClickCount: row.ClickCount,
		})
	}
	return items, nil
}

func chartPointsFromTeachers(items []frontend.AffiliateAdsReportTeacherItem) []frontend.AffiliateAdsReportChartPoint {
	out := make([]frontend.AffiliateAdsReportChartPoint, 0, len(items))
	for _, item := range items {
		out = append(out, frontend.AffiliateAdsReportChartPoint{
			Label:  truncateAffiliateChartLabel(item.TeacherName, 36),
			Clicks: item.ClickCount,
		})
	}
	return out
}

func chartPointsFromAdZones(items []frontend.AffiliateAdsReportAdZoneItem) []frontend.AffiliateAdsReportChartPoint {
	out := make([]frontend.AffiliateAdsReportChartPoint, 0, len(items))
	for _, item := range items {
		label := truncateAffiliateChartLabel(item.AdName+" - "+item.ZoneLabel, 40)
		out = append(out, frontend.AffiliateAdsReportChartPoint{
			Label:  label,
			Clicks: item.ClickCount,
		})
	}
	return out
}

func chartPointsFromAds(items []frontend.AffiliateAdsReportAdItem) []frontend.AffiliateAdsReportChartPoint {
	out := make([]frontend.AffiliateAdsReportChartPoint, 0, len(items))
	for _, item := range items {
		out = append(out, frontend.AffiliateAdsReportChartPoint{
			Label:  truncateAffiliateChartLabel(item.Name, 36),
			Clicks: item.LinkedClickTotal,
		})
	}
	return out
}

func truncateAffiliateChartLabel(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func mapAffiliateAdsReportProduct(id int64, name, shop, thumb string, clicks int64) frontend.AffiliateAdsReportProductItem {
	return frontend.AffiliateAdsReportProductItem{
		ID:           strconv.FormatInt(id, 10),
		Name:         name,
		ShopName:     shop,
		ThumbnailURL: thumb,
		ClickCount:   clicks,
	}
}

func sqlAggregateInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case float64:
		return int64(n)
	default:
		return 0
	}
}
