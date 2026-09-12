package learningmaterials

import (
	"context"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	previewMaxBodyBytes = 512 * 1024
	previewTimeout      = 10 * time.Second
	previewUserAgent    = "Mozilla/5.0 (compatible; ZionEnglishAdmin/1.0; +https://zion-english)"
)

var (
	metaTagRe = regexp.MustCompile(`<meta[^>]+>`)
	attrRe    = regexp.MustCompile(`(\w+)\s*=\s*["']([^"']*)["']`)
	linkTagRe = regexp.MustCompile(`<link[^>]+>`)
)

var previewClient = &http.Client{
	Timeout: previewTimeout,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		if _, err := validatePreviewTargetURL(req.URL.String()); err != nil {
			return err
		}
		return nil
	},
}

func ResolveThumbnailURL(ctx context.Context, rawURL string) (string, error) {
	parsedPageURL, err := validatePreviewTargetURL(rawURL)
	if err != nil {
		return "", err
	}
	pageURL := parsedPageURL.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", previewUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := previewClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("preview fetch failed: %s", resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, previewMaxBodyBytes))
	if err != nil {
		return "", err
	}

	pageHTML := string(body)
	for _, property := range []string{
		"og:image",
		"og:image:secure_url",
		"og:image:url",
		"twitter:image",
		"twitter:image:src",
	} {
		if thumb := extractMetaImage(pageHTML, property); thumb != "" {
			return resolveAssetURL(pageURL, thumb), nil
		}
	}
	if thumb := extractPreloadImage(pageHTML); thumb != "" {
		return resolveAssetURL(pageURL, thumb), nil
	}
	if thumb := extractLinkIcon(pageHTML, "apple-touch-icon"); thumb != "" {
		return resolveAssetURL(pageURL, thumb), nil
	}
	if thumb := extractLinkIcon(pageHTML, "icon"); thumb != "" {
		return resolveAssetURL(pageURL, thumb), nil
	}

	host := parsedPageURL.Hostname()
	if host != "" {
		return fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=128", host), nil
	}
	return "", nil
}

func validatePreviewTargetURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrURLRequired
	}
	if !isValidURL(raw) {
		return nil, ErrInvalidURL
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, ErrInvalidURL
	}
	if parsed.Host == "" {
		return nil, ErrInvalidURL
	}
	if err := validatePreviewHost(parsed.Hostname()); err != nil {
		return nil, err
	}
	return parsed, nil
}

func validatePreviewHost(host string) error {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return ErrInvalidURL
	}
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") {
		return ErrInvalidURL
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return ErrInvalidURL
		}
	}
	return nil
}

func extractMetaImage(pageHTML, property string) string {
	property = strings.ToLower(property)
	for _, tag := range metaTagRe.FindAllString(pageHTML, -1) {
		attrs := parseHTMLAttrs(tag)
		prop := strings.ToLower(attrs["property"])
		name := strings.ToLower(attrs["name"])
		if prop != property && name != property {
			continue
		}
		if content := strings.TrimSpace(html.UnescapeString(attrs["content"])); content != "" {
			return content
		}
	}
	return ""
}

func extractLinkIcon(pageHTML, rel string) string {
	rel = strings.ToLower(rel)
	for _, tag := range linkTagRe.FindAllString(pageHTML, -1) {
		attrs := parseHTMLAttrs(tag)
		linkRel := strings.ToLower(attrs["rel"])
		if linkRel != rel {
			continue
		}
		if href := strings.TrimSpace(html.UnescapeString(attrs["href"])); href != "" {
			return href
		}
	}
	return ""
}

func extractPreloadImage(pageHTML string) string {
	for _, tag := range linkTagRe.FindAllString(pageHTML, -1) {
		attrs := parseHTMLAttrs(tag)
		if strings.ToLower(attrs["as"]) != "image" {
			continue
		}
		if srcset := attrs["imagesrcset"]; srcset != "" {
			if thumb := firstSrcSetURL(srcset); thumb != "" {
				return thumb
			}
		}
		if href := strings.TrimSpace(html.UnescapeString(attrs["href"])); href != "" {
			return href
		}
	}
	return ""
}

func firstSrcSetURL(srcset string) string {
	srcset = strings.TrimSpace(html.UnescapeString(srcset))
	if srcset == "" {
		return ""
	}
	for _, part := range strings.Split(srcset, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Fields(part)
		if len(fields) > 0 && fields[0] != "" {
			return fields[0]
		}
	}
	return ""
}

func parseHTMLAttrs(tag string) map[string]string {
	attrs := make(map[string]string)
	for _, match := range attrRe.FindAllStringSubmatch(tag, -1) {
		attrs[strings.ToLower(match[1])] = match[2]
	}
	return attrs
}

func resolveAssetURL(pageURL, asset string) string {
	asset = strings.TrimSpace(asset)
	if asset == "" {
		return ""
	}
	parsedAsset, err := url.Parse(asset)
	if err != nil {
		return asset
	}
	if parsedAsset.IsAbs() {
		return parsedAsset.String()
	}
	base, err := url.Parse(pageURL)
	if err != nil {
		return asset
	}
	return base.ResolveReference(parsedAsset).String()
}

func NormalizeThumbnailURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !isValidURL(raw) {
		return ""
	}
	return raw
}
