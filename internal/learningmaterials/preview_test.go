package learningmaterials

import (
	"testing"
)

func TestExtractMetaImage(t *testing.T) {
	html := `<html><head>
		<meta property="og:image" content="https://cdn.example.com/hero.png">
		<meta name="twitter:image" content="https://cdn.example.com/twitter.png">
	</head></html>`

	got := extractMetaImage(html, "og:image")
	if got != "https://cdn.example.com/hero.png" {
		t.Fatalf("og:image = %q, want hero.png", got)
	}

	got = extractMetaImage(html, "twitter:image")
	if got != "https://cdn.example.com/twitter.png" {
		t.Fatalf("twitter:image = %q", got)
	}
}

func TestExtractMetaImageUnescapesEntities(t *testing.T) {
	html := `<meta property="og:image" content="https://cdn.example.com/a.png?x=1&amp;y=2" />`
	got := extractMetaImage(html, "og:image")
	if got != "https://cdn.example.com/a.png?x=1&y=2" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractPreloadImage(t *testing.T) {
	html := `<link rel="preload" as="image" imageSrcSet="https://pbs.twimg.com/media/abc?format=webp&amp;name=small 680w, https://pbs.twimg.com/media/abc?format=webp&amp;name=large 2048w" />`
	got := extractPreloadImage(html)
	want := "https://pbs.twimg.com/media/abc?format=webp&name=small"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveAssetURL(t *testing.T) {
	got := resolveAssetURL("https://example.com/docs/page", "/images/thumb.jpg")
	if got != "https://example.com/images/thumb.jpg" {
		t.Fatalf("resolveAssetURL = %q", got)
	}
}

func TestValidatePreviewTargetURL(t *testing.T) {
	_, err := validatePreviewTargetURL("http://127.0.0.1/test")
	if err == nil {
		t.Fatal("expected localhost rejection")
	}

	got, err := validatePreviewTargetURL("https://example.com/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.String() != "https://example.com/path" {
		t.Fatalf("got %q", got.String())
	}
}

func TestNormalizeThumbnailURL(t *testing.T) {
	if got := NormalizeThumbnailURL(" https://img.example.com/a.png "); got != "https://img.example.com/a.png" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeThumbnailURL("not-a-url"); got != "" {
		t.Fatalf("got %q", got)
	}
}
