package affiliates

import (
	"encoding/csv"
	"io"
	"strings"

	"zion-english/internal/constants"
)

type ImpactCSVRow struct {
	AdID           string
	State          string
	ProgramID      string
	Name           string
	AdType         string
	TrackingLink   string
	CreativeWidth  string
	CreativeHeight string
	Orientation    constants.ThumbnailOrientation
}

var impactRequiredHeaders = []string{
	"adid",
	"state",
	"programid",
	"name",
	"adtype",
	"trackinglink",
}

func ParseImpactAdsCSV(r io.Reader) ([]ImpactCSVRow, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return nil, ErrCSVInvalid
	}
	col := mapImpactCSVColumns(header)
	for _, required := range impactRequiredHeaders {
		if col[required] < 0 {
			return nil, ErrCSVInvalid
		}
	}

	var rows []ImpactCSVRow
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
		creativeWidth := fieldAt(record, col["thirdpartyservableadcreativewidth"])
		creativeHeight := fieldAt(record, col["thirdpartyservableadcreativeheight"])
		row := ImpactCSVRow{
			AdID:           fieldAt(record, col["adid"]),
			State:          fieldAt(record, col["state"]),
			ProgramID:      fieldAt(record, col["programid"]),
			Name:           fieldAt(record, col["name"]),
			AdType:         fieldAt(record, col["adtype"]),
			TrackingLink:   fieldAt(record, col["trackinglink"]),
			CreativeWidth:  creativeWidth,
			CreativeHeight: creativeHeight,
			Orientation:    OrientationFromCreativeDimensions(creativeWidth, creativeHeight),
		}
		if row.AdID == "" || row.Name == "" || row.TrackingLink == "" {
			return nil, ErrCSVInvalid
		}
		if !IsValidHTTPURL(row.TrackingLink) {
			return nil, ErrCSVInvalid
		}
		rows = append(rows, row)
		if len(rows) > maxCSVRows {
			return nil, ErrCSVTooManyRows
		}
	}
	if len(rows) == 0 {
		return nil, ErrCSVInvalid
	}
	return rows, nil
}

func mapImpactCSVColumns(header []string) map[string]int {
	out := make(map[string]int)
	for i, h := range header {
		key := normalizeImpactCSVHeader(h)
		if key == "" {
			continue
		}
		out[key] = i
	}
	for _, required := range impactRequiredHeaders {
		if _, ok := out[required]; !ok {
			out[required] = -1
		}
	}
	for _, optional := range []string{
		"thirdpartyservableadcreativewidth",
		"thirdpartyservableadcreativeheight",
	} {
		if _, ok := out[optional]; !ok {
			out[optional] = -1
		}
	}
	return out
}

func normalizeImpactCSVHeader(h string) string {
	h = strings.TrimSpace(h)
	h = strings.TrimPrefix(h, "\ufeff")
	return strings.ToLower(h)
}

func fieldAt(record []string, idx int) string {
	if idx < 0 || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}
