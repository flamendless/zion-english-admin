package affiliates

import (
	"net/url"
	"strings"
)

const MaxNameLen = 256

type SaveRequest struct {
	AffiliateURL   string
	ProductURL     string
	ShopID         string
	ItemID         string
	Name           string
	Brand          string
	PriceDisplay   string
	ThumbnailURL   string
	SortOrder      int64
	Sales          string
	ShopName       string
	CommissionRate string
	Commission     string
}

func NormalizeSaveRequest(req SaveRequest) SaveRequest {
	return SaveRequest{
		AffiliateURL: strings.TrimSpace(req.AffiliateURL),
		ProductURL:   strings.TrimSpace(req.ProductURL),
		ShopID:       strings.TrimSpace(req.ShopID),
		ItemID:       strings.TrimSpace(req.ItemID),
		Name:         strings.TrimSpace(req.Name),
		Brand:        strings.TrimSpace(req.Brand),
		PriceDisplay: strings.TrimSpace(req.PriceDisplay),
		ThumbnailURL:   strings.TrimSpace(req.ThumbnailURL),
		SortOrder:      req.SortOrder,
		Sales:          strings.TrimSpace(req.Sales),
		ShopName:       strings.TrimSpace(req.ShopName),
		CommissionRate: strings.TrimSpace(req.CommissionRate),
		Commission:     strings.TrimSpace(req.Commission),
	}
}

func ValidateSave(req SaveRequest) error {
	req = NormalizeSaveRequest(req)
	if req.AffiliateURL == "" {
		return ErrAffiliateURLRequired
	}
	if !isValidHTTPSURL(req.AffiliateURL) {
		return ErrInvalidAffiliateURL
	}
	if req.Name == "" {
		return ErrNameRequired
	}
	if len(req.Name) > MaxNameLen {
		return ErrNameTooLong
	}
	if req.ProductURL != "" && !isValidHTTPSURL(req.ProductURL) {
		return ErrInvalidProductURL
	}
	if req.ThumbnailURL != "" && !isValidHTTPSURL(req.ThumbnailURL) {
		return ErrInvalidThumbnailURL
	}
	return nil
}

func isValidHTTPSURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func IsValidHTTPURL(raw string) bool {
	return isValidHTTPSURL(strings.TrimSpace(raw))
}
