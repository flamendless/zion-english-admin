package affiliates

import "testing"

func TestExtractShopeeProductImageID_fromEmbeddedJSON(t *testing.T) {
	html := `{"item":{"images":["ph-11134258-820lf-mt8u8yyjzbih79"],"name":"Test"}}`
	id := extractShopeeProductImageID(html)
	if id != "ph-11134258-820lf-mt8u8yyjzbih79" {
		t.Fatalf("unexpected id: %s", id)
	}
	url := extractMetaImage(html)
	if url != "https://down-ph.img.susercontent.com/file/ph-11134258-820lf-mt8u8yyjzbih79" {
		t.Fatalf("unexpected url: %s", url)
	}
}
