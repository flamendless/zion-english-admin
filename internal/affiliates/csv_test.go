package affiliates

import (
	"strings"
	"testing"
)

func TestParseProductLinksCSV_sample(t *testing.T) {
	raw := `Item Id,Item Name,Price,Sales,Shop Name,Commission Rate,Commission,Product Link,Offer Link
45410921075,Test Product,75,10K+,Mirra Beauty,19.5%,₱14.63,https://shopee.ph/product/1807467467/45410921075,https://s.shopee.ph/5q8TMoBv7b
`
	rows, err := ParseProductLinksCSV(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].ItemID != "45410921075" || rows[0].ShopName != "Mirra Beauty" {
		t.Fatalf("unexpected row: %+v", rows[0])
	}
}
