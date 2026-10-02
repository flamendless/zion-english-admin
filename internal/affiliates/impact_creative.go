package affiliates

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"zion-english/internal/utils"
)

const (
	impactCreativeMaxBytes = 3 * 1024 * 1024
	impactCreativeTimeout  = 15 * time.Second
)

var impactCreativeClient = &http.Client{Timeout: impactCreativeTimeout}

func ValidImpactCreativePart(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 24 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func ImpactCreativeImagePath(productID int64, programID, itemID string) string {
	if productID > 0 {
		return fmt.Sprintf("/affiliate-creative/%d", productID)
	}
	programID = strings.TrimSpace(programID)
	itemID = strings.TrimSpace(itemID)
	if !ValidImpactCreativePart(programID) || !ValidImpactCreativePart(itemID) {
		return ""
	}
	return fmt.Sprintf(
		"/affiliate-creative/image?program_id=%s&ad_id=%s",
		url.QueryEscape(programID),
		url.QueryEscape(itemID),
	)
}

func ImpactCreativeImageURL(productID int64, programID, itemID string) string {
	path := ImpactCreativeImagePath(productID, programID, itemID)
	if path == "" {
		return ""
	}
	return utils.URL(path)
}

func FetchImpactCreative(ctx context.Context, programID, adID string) ([]byte, string, error) {
	programID = strings.TrimSpace(programID)
	adID = strings.TrimSpace(adID)
	if !ValidImpactCreativePart(programID) || !ValidImpactCreativePart(adID) {
		return nil, "", ErrImpactCreativeInvalid
	}
	remoteURL := ImpactDisplayAdThumbnailURL(programID, adID)
	if remoteURL == "" {
		return nil, "", ErrImpactCreativeInvalid
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remoteURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", fetchUserAgent)

	resp, err := impactCreativeClient.Do(req)
	if err != nil {
		return nil, "", ErrImpactCreativeFetch
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", ErrImpactCreativeFetch
	}

	ct := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "image/") {
		return nil, "", ErrImpactCreativeFetch
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, impactCreativeMaxBytes+1))
	if err != nil {
		return nil, "", ErrImpactCreativeFetch
	}
	if len(body) > impactCreativeMaxBytes {
		return nil, "", ErrImpactCreativeFetch
	}
	return body, ct, nil
}
