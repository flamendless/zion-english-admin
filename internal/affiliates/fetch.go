package affiliates

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"zion-english/internal/logs"

	"go.uber.org/zap"
)

const (
	fetchMaxBodyBytes = 512 * 1024
	fetchTimeout      = 15 * time.Second
	fetchUserAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	// Shopee serves product og:image and SSR meta only to link-preview crawlers.
	fetchProductPageUserAgent = "facebookexternalhit/1.1"
	fetchMaxAttempts          = 3
)

type PreviewResult struct {
	AffiliateURL  string
	ProductURL    string
	ShopID        string
	ItemID        string
	Name          string
	Brand         string
	PriceDisplay  string
	ThumbnailURL  string
}

var fetchClient = &http.Client{
	Timeout: fetchTimeout,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 8 {
			return fmt.Errorf("too many redirects")
		}
		if _, err := validateShopeeURL(req.URL.String()); err != nil {
			return err
		}
		return nil
	},
}

func FetchProductPreview(ctx context.Context, rawAffiliateURL string) (PreviewResult, error) {
	start := time.Now()
	rawAffiliateURL = strings.TrimSpace(rawAffiliateURL)
	if rawAffiliateURL == "" {
		logAffiliatePreview("validate", rawAffiliateURL, "", start, ErrAffiliateURLRequired)
		return PreviewResult{}, ErrAffiliateURLRequired
	}
	if !isValidHTTPSURL(rawAffiliateURL) {
		logAffiliatePreview("validate", rawAffiliateURL, "", start, ErrInvalidAffiliateURL)
		return PreviewResult{}, ErrInvalidAffiliateURL
	}

	productURL, err := resolveToProductURL(ctx, rawAffiliateURL)
	if err != nil {
		logAffiliatePreview("resolve", rawAffiliateURL, "", start, err)
		return PreviewResult{}, err
	}

	ids, ok := ParseProductIDsFromURL(productURL)
	if !ok {
		logAffiliatePreview("parse_ids", rawAffiliateURL, productURL, start, ErrNotProductPage)
		return PreviewResult{}, ErrNotProductPage
	}

	pageHTML, err := fetchPageHTML(ctx, productURL)
	if err != nil {
		logAffiliatePreview("fetch_html", rawAffiliateURL, productURL, start, ErrFetchBlocked,
			zap.String("fetch_error", err.Error()),
			zap.String("shop_id", ids.ShopID),
			zap.String("item_id", ids.ItemID),
		)
		return PreviewResult{}, ErrFetchBlocked
	}

	name := extractTitle(pageHTML)
	thumb := extractMetaImage(pageHTML)
	brand, price := extractBrandAndPrice(pageHTML)

	if name == "" && thumb == "" && price == "" {
		logAffiliatePreview("parse_preview", rawAffiliateURL, productURL, start, ErrPreviewUnavailable,
			zap.String("shop_id", ids.ShopID),
			zap.String("item_id", ids.ItemID),
			zap.Int("html_bytes", len(pageHTML)),
		)
		return PreviewResult{}, ErrPreviewUnavailable
	}

	result := PreviewResult{
		AffiliateURL: rawAffiliateURL,
		ProductURL:   productURL,
		ShopID:       ids.ShopID,
		ItemID:       ids.ItemID,
		Name:         name,
		Brand:        brand,
		PriceDisplay: price,
		ThumbnailURL: thumb,
	}
	logAffiliatePreview("ok", rawAffiliateURL, productURL, start, nil,
		zap.String("shop_id", result.ShopID),
		zap.String("item_id", result.ItemID),
		zap.String("name", result.Name),
		zap.String("brand", result.Brand),
		zap.String("price_display", result.PriceDisplay),
		zap.Bool("has_thumbnail", result.ThumbnailURL != ""),
		zap.Int("html_bytes", len(pageHTML)),
	)
	return result, nil
}

func logAffiliatePreview(stage, affiliateURL, productURL string, start time.Time, err error, extra ...zap.Field) {
	fields := []zap.Field{
		zap.String("stage", stage),
		zap.String("affiliate_url", affiliateURL),
		zap.Duration("duration", time.Since(start)),
	}
	if productURL != "" {
		fields = append(fields, zap.String("product_url", productURL))
	}
	fields = append(fields, extra...)
	if err != nil {
		fields = append(fields, zap.Error(err))
		logs.Log().Warn("affiliate fetch", fields...)
		return
	}
	logs.Log().Info("affiliate fetch", fields...)
}

func resolveToProductURL(ctx context.Context, raw string) (string, error) {
	if ids, ok := ParseProductIDsFromURL(raw); ok {
		_, err := validateShopeeURL(raw)
		if err != nil {
			return "", err
		}
		_ = ids
		return raw, nil
	}

	finalURL, err := followRedirects(ctx, raw)
	if err != nil {
		return "", err
	}
	if _, ok := ParseProductIDsFromURL(finalURL); !ok {
		logs.Log().Warn("affiliate fetch redirect target is not a product page",
			zap.String("affiliate_url", raw),
			zap.String("resolved_url", finalURL),
		)
		return "", ErrNotProductPage
	}
	return finalURL, nil
}

func followRedirects(ctx context.Context, raw string) (string, error) {
	if _, err := validateShopeeURL(raw); err != nil {
		return "", ErrInvalidAffiliateURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", fetchUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := fetchClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	final := resp.Request.URL.String()
	if _, err := validateShopeeURL(final); err != nil {
		return "", ErrInvalidAffiliateURL
	}
	return final, nil
}

func fetchPageHTML(ctx context.Context, pageURL string) (string, error) {
	if _, err := validateShopeeURL(pageURL); err != nil {
		return "", err
	}

	var lastErr error
	for attempt := 1; attempt <= fetchMaxAttempts; attempt++ {
		html, err := fetchPageHTMLOnce(ctx, pageURL)
		if err == nil {
			return html, nil
		}
		lastErr = err
		if !isRetryableFetchErr(err) || attempt == fetchMaxAttempts {
			break
		}
		backoff := time.Duration(attempt) * 250 * time.Millisecond
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(backoff):
		}
	}
	return "", lastErr
}

func fetchPageHTMLOnce(ctx context.Context, pageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", fetchProductPageUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://shopee.ph/")

	resp, err := fetchClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("status %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, fetchMaxBodyBytes))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func isRetryableFetchErr(err error) bool {
	if err == nil {
		return false
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "temporary failure")
}

func validateShopeeURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrInvalidAffiliateURL
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, ErrInvalidAffiliateURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, ErrInvalidAffiliateURL
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return nil, ErrInvalidAffiliateURL
	}
	if err := validateFetchHost(host); err != nil {
		return nil, err
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return nil, ErrInvalidAffiliateURL
		}
	}
	return parsed, nil
}

func validateFetchHost(host string) error {
	allowed := []string{
		"shopee.ph",
		"shope.ee",
		"s.shopee.ph",
		"shopee.com.ph",
	}
	for _, suffix := range allowed {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return nil
		}
	}
	return ErrInvalidAffiliateURL
}
