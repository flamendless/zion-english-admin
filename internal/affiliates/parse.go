package affiliates

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
)

var (
	productPathRe      = regexp.MustCompile(`/product/(\d+)/(\d+)`)
	productSlugRe      = regexp.MustCompile(`-i\.(\d+)\.(\d+)(?:\?|$|/)`)
	productNamedPathRe = regexp.MustCompile(`(?i)shopee\.ph/[^/?#]+/(\d+)/(\d+)`)
	pricePHPRe    = regexp.MustCompile(`₱\s*[\d,]+(?:\.\d+)?`)
	titleTagRe    = regexp.MustCompile(`(?is)<title[^>]*>([^<]+)</title>`)
	ldJSONRe      = regexp.MustCompile(`(?is)<script[^>]+type=["']application/ld\+json["'][^>]*>([\s\S]*?)</script>`)
	metaTagRe          = regexp.MustCompile(`<meta[^>]+>`)
	attrRe             = regexp.MustCompile(`(\w+)\s*=\s*["']([^"']*)["']`)
	productImagesArrayRe = regexp.MustCompile(`"images"\s*:\s*\[\s*"(ph-[^"]+)"`)
	shopeeCDNFileRe      = regexp.MustCompile(`https?://down-ph\.img\.susercontent\.com/file/([^\s"'<>]+)`)
	heroPaintBlockRe     = regexp.MustCompile(`(?is)<(?:picture|source|img)[^>]*elementtiming\s*=\s*["']shopee:heroComponentPaint["'][^>]*>`)
	heroAttrURLRe        = regexp.MustCompile(`(?is)(?:src|srcset)\s*=\s*["']([^"']+)["']`)
)

type ProductIDs struct {
	ShopID string
	ItemID string
}

func ParseProductIDsFromURL(pageURL string) (ProductIDs, bool) {
	if m := productPathRe.FindStringSubmatch(pageURL); len(m) == 3 {
		return ProductIDs{ShopID: m[1], ItemID: m[2]}, true
	}
	if m := productSlugRe.FindStringSubmatch(pageURL); len(m) == 3 {
		return ProductIDs{ShopID: m[1], ItemID: m[2]}, true
	}
	if m := productNamedPathRe.FindStringSubmatch(pageURL); len(m) == 3 {
		return ProductIDs{ShopID: m[1], ItemID: m[2]}, true
	}
	return ProductIDs{}, false
}

func extractTitle(pageHTML string) string {
	for _, property := range []string{"og:title", "twitter:title"} {
		if t := extractMetaContent(pageHTML, property); t != "" {
			return cleanShopeeTitle(t)
		}
	}
	if m := titleTagRe.FindStringSubmatch(pageHTML); len(m) == 2 {
		return cleanShopeeTitle(html.UnescapeString(m[1]))
	}
	return ""
}

func extractMetaImage(pageHTML string) string {
	if img := extractHeroProductImage(pageHTML); img != "" {
		return normalizeShopeeImageURL(img)
	}
	for _, property := range []string{
		"og:image",
		"og:image:secure_url",
		"twitter:image",
	} {
		if img := extractMetaContent(pageHTML, property); img != "" {
			return normalizeShopeeImageURL(img)
		}
	}
	if imageID := extractShopeeProductImageID(pageHTML); imageID != "" {
		return ShopeeCDNURL(imageID)
	}
	return ""
}

func extractHeroProductImage(pageHTML string) string {
	tags := heroPaintBlockRe.FindAllString(pageHTML, -1)
	for _, tag := range tags {
		for _, m := range heroAttrURLRe.FindAllStringSubmatch(tag, -1) {
			if len(m) < 2 {
				continue
			}
			if url := firstURLFromSrcset(m[1]); url != "" {
				return url
			}
		}
	}
	return ""
}

func firstURLFromSrcset(raw string) string {
	raw = strings.TrimSpace(html.UnescapeString(raw))
	if raw == "" {
		return ""
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		url := strings.Fields(part)[0]
		if url != "" {
			return url
		}
	}
	return raw
}

func normalizeShopeeImageURL(raw string) string {
	raw = firstURLFromSrcset(raw)
	if raw == "" {
		return ""
	}
	if m := shopeeCDNFileRe.FindStringSubmatch(raw); len(m) == 2 {
		return ShopeeCDNURL(trimShopeeImageFileToken(m[1]))
	}
	if strings.HasPrefix(raw, "ph-") {
		return ShopeeCDNURL(trimShopeeImageFileToken(raw))
	}
	if idx := strings.Index(raw, "@"); idx > 0 && strings.Contains(raw, "susercontent.com/file/") {
		return raw[:idx]
	}
	return raw
}

func trimShopeeImageFileToken(token string) string {
	token = strings.TrimSpace(token)
	if i := strings.Index(token, "@"); i > 0 {
		token = token[:i]
	}
	if i := strings.Index(token, "?"); i > 0 {
		token = token[:i]
	}
	return token
}

func extractShopeeProductImageID(pageHTML string) string {
	if m := productImagesArrayRe.FindStringSubmatch(pageHTML); len(m) == 2 {
		return m[1]
	}
	return ""
}

func ShopeeCDNURL(imageID string) string {
	imageID = strings.TrimSpace(imageID)
	if imageID == "" {
		return ""
	}
	return "https://down-ph.img.susercontent.com/file/" + imageID
}

func extractMetaContent(pageHTML, property string) string {
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

func parseHTMLAttrs(tag string) map[string]string {
	attrs := make(map[string]string)
	for _, match := range attrRe.FindAllStringSubmatch(tag, -1) {
		attrs[strings.ToLower(match[1])] = match[2]
	}
	return attrs
}

func cleanShopeeTitle(title string) string {
	title = strings.TrimSpace(html.UnescapeString(title))
	for _, suffix := range []string{
		" | Shopee Philippines",
		" - Shopee Philippines",
		" | Shopee PH",
		" - Shopee PH",
	} {
		if strings.HasSuffix(title, suffix) {
			title = strings.TrimSuffix(title, suffix)
		}
	}
	return strings.TrimSpace(title)
}

func extractBrandAndPrice(pageHTML string) (brand, price string) {
	for _, block := range extractJSONLDBlocks(pageHTML) {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(block), &payload); err != nil {
			continue
		}
		b, p := brandPriceFromLD(payload)
		if brand == "" && b != "" {
			brand = b
		}
		if price == "" && p != "" {
			price = p
		}
		if brand != "" && price != "" {
			return brand, price
		}
	}
	if price == "" {
		if m := pricePHPRe.FindString(pageHTML); m != "" {
			price = strings.TrimSpace(m)
		}
	}
	return brand, price
}

func extractJSONLDBlocks(pageHTML string) []string {
	var blocks []string
	for _, m := range ldJSONRe.FindAllStringSubmatch(pageHTML, -1) {
		if len(m) < 2 {
			continue
		}
		block := strings.TrimSpace(m[1])
		if block != "" {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

func brandPriceFromLD(node map[string]interface{}) (brand, price string) {
	if t, _ := node["@type"].(string); strings.EqualFold(t, "Product") {
		brand = ldBrand(node["brand"])
		price = ldOfferPrice(node["offers"])
		return brand, price
	}
	if graph, ok := node["@graph"].([]interface{}); ok {
		for _, item := range graph {
			if m, ok := item.(map[string]interface{}); ok {
				b, p := brandPriceFromLD(m)
				if brand == "" && b != "" {
					brand = b
				}
				if price == "" && p != "" {
					price = p
				}
			}
		}
	}
	return brand, price
}

func ldBrand(v interface{}) string {
	switch b := v.(type) {
	case string:
		return strings.TrimSpace(b)
	case map[string]interface{}:
		if name, _ := b["name"].(string); name != "" {
			return strings.TrimSpace(name)
		}
	}
	return ""
}

func ldOfferPrice(v interface{}) string {
	switch offers := v.(type) {
	case map[string]interface{}:
		return formatLDPrice(offers)
	case []interface{}:
		for _, o := range offers {
			if m, ok := o.(map[string]interface{}); ok {
				if p := formatLDPrice(m); p != "" {
					return p
				}
			}
		}
	}
	return ""
}

func formatLDPrice(offer map[string]interface{}) string {
	var price string
	switch p := offer["price"].(type) {
	case string:
		price = strings.TrimSpace(p)
	case float64:
		price = fmt.Sprintf("%.2f", p)
		price = strings.TrimRight(strings.TrimRight(price, "0"), ".")
	}
	if price == "" {
		return ""
	}
	currency, _ := offer["priceCurrency"].(string)
	currency = strings.TrimSpace(currency)
	if currency == "PHP" {
		return "₱" + price
	}
	if currency != "" {
		return currency + " " + price
	}
	return price
}
