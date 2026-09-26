package affiliates

import "testing"

func TestParseProductIDsFromURL(t *testing.T) {
	tests := []struct {
		url      string
		shop     string
		item     string
		wantOK   bool
	}{
		{
			url:    "https://shopee.ph/product/123/456",
			shop:   "123",
			item:   "456",
			wantOK: true,
		},
		{
			url:    "https://shopee.ph/Some-Name-i.99.88888",
			shop:   "99",
			item:   "88888",
			wantOK: true,
		},
		{
			url:    "https://shopee.ph/opaanlp/1807467467/45410921075",
			shop:   "1807467467",
			item:   "45410921075",
			wantOK: true,
		},
		{
			url:    "https://shope.ee/abc",
			wantOK: false,
		},
	}
	for _, tc := range tests {
		ids, ok := ParseProductIDsFromURL(tc.url)
		if ok != tc.wantOK {
			t.Fatalf("url %q: ok=%v want %v", tc.url, ok, tc.wantOK)
		}
		if !tc.wantOK {
			continue
		}
		if ids.ShopID != tc.shop || ids.ItemID != tc.item {
			t.Fatalf("url %q: ids=%+v want shop=%s item=%s", tc.url, ids, tc.shop, tc.item)
		}
	}
}

func TestCleanShopeeTitle(t *testing.T) {
	got := cleanShopeeTitle("Widget Pro | Shopee Philippines")
	if got != "Widget Pro" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractHeroProductImage(t *testing.T) {
	html := `<picture class="i9ihcI"><source srcset="https://down-ph.img.susercontent.com/file/ph-11134207-7ra0h-mdn732w2hfjvb4@resize_w450_nl.webp 1x, https://down-ph.img.susercontent.com/file/ph-11134207-7ra0h-mdn732w2hfjvb4@resize_w900_nl.webp 2x" type="image/webp" elementtiming="shopee:heroComponentPaint"><img src="https://down-ph.img.susercontent.com/file/ph-11134207-7ra0h-mdn732w2hfjvb4" elementtiming="shopee:heroComponentPaint"></picture>`
	got := extractMetaImage(html)
	want := "https://down-ph.img.susercontent.com/file/ph-11134207-7ra0h-mdn732w2hfjvb4"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeShopeeImageURL(t *testing.T) {
	raw := "https://down-ph.img.susercontent.com/file/ph-11134207-7ra0h-mdn732w2hfjvb4@resize_w450_nl.webp 1x"
	got := normalizeShopeeImageURL(raw)
	want := "https://down-ph.img.susercontent.com/file/ph-11134207-7ra0h-mdn732w2hfjvb4"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
