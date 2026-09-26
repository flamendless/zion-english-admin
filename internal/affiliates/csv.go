package affiliates

import (
	"encoding/csv"
	"io"
	"strings"
)

const maxCSVRows = 500

type CSVRow struct {
	ItemID          string
	ItemName        string
	Price           string
	Sales           string
	ShopName        string
	CommissionRate  string
	Commission      string
	ProductLink     string
	OfferLink       string
}

var expectedCSVHeaders = []string{
	"item id",
	"item name",
	"price",
	"sales",
	"shop name",
	"commission rate",
	"commission",
	"product link",
	"offer link",
}

func ParseProductLinksCSV(r io.Reader) ([]CSVRow, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return nil, ErrCSVInvalid
	}
	if len(header) < len(expectedCSVHeaders) {
		return nil, ErrCSVInvalid
	}
	normalized := normalizeCSVHeaderRow(header)
	if !csvHeaderMatches(normalized) {
		return nil, ErrCSVInvalid
	}

	var rows []CSVRow
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, ErrCSVInvalid
		}
		if len(record) == 0 || csvRecordEmpty(record) {
			continue
		}
		if len(record) < len(expectedCSVHeaders) {
			return nil, ErrCSVInvalid
		}
		row := CSVRow{
			ItemID:         strings.TrimSpace(record[0]),
			ItemName:       strings.TrimSpace(record[1]),
			Price:          strings.TrimSpace(record[2]),
			Sales:          strings.TrimSpace(record[3]),
			ShopName:       strings.TrimSpace(record[4]),
			CommissionRate: strings.TrimSpace(record[5]),
			Commission:     strings.TrimSpace(record[6]),
			ProductLink:    strings.TrimSpace(record[7]),
			OfferLink:      strings.TrimSpace(record[8]),
		}
		if row.ItemName == "" || row.OfferLink == "" || row.ProductLink == "" {
			return nil, ErrCSVInvalid
		}
		if !isValidHTTPSURL(row.ProductLink) || !isValidHTTPSURL(row.OfferLink) {
			return nil, ErrCSVInvalid
		}
		rows = append(rows, row)
		if len(rows) > maxCSVRows {
			return nil, ErrCSVTooManyRows
		}
	}
	if len(rows) == 0 {
		return nil, ErrCSVNoRows
	}
	return rows, nil
}

func normalizeCSVHeaderRow(header []string) []string {
	out := make([]string, len(header))
	for i, h := range header {
		h = strings.TrimSpace(h)
		h = strings.TrimPrefix(h, "\ufeff")
		out[i] = strings.ToLower(h)
	}
	return out
}

func csvHeaderMatches(normalized []string) bool {
	for i, want := range expectedCSVHeaders {
		if i >= len(normalized) || normalized[i] != want {
			return false
		}
	}
	return true
}

func csvRecordEmpty(record []string) bool {
	for _, cell := range record {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func FormatPriceDisplay(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "₱") {
		return raw
	}
	return "₱" + raw
}
