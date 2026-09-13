package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func DriveURLToExportURL(driveURL string, format string) (string, error) {
	parsedURL, err := url.Parse(driveURL)
	if err != nil {
		return "", err
	}

	if parsedURL.Host != "docs.google.com" {
		return "", fmt.Errorf("not a Google Docs URL: expected docs.google.com, got %s", parsedURL.Host)
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) < 4 || pathParts[0] != "spreadsheets" || pathParts[1] != "d" {
		return "", ErrInvalidDriveSpreadsheetPath
	}

	exportURL := &url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
		Path:   fmt.Sprintf("/spreadsheets/d/%s/export", pathParts[2]),
	}

	gid := "0"
	if parsedURL.Query().Has("gid") {
		gid = parsedURL.Query().Get("gid")
	} else if parsedURL.Fragment != "" {
		if after, ok := strings.CutPrefix(parsedURL.Fragment, "gid="); ok {
			gid = after
		} else {
			fragmentParts := strings.Split(parsedURL.Fragment, "&")
			for _, part := range fragmentParts {
				if strings.HasPrefix(part, "gid=") {
					gid = strings.TrimPrefix(part, "gid=")
					break
				}
			}
		}
	}

	query := url.Values{}
	query.Set("format", format)
	query.Set("gid", gid)
	exportURL.RawQuery = query.Encode()
	return exportURL.String(), nil
}

func ValidateDriveSpreadsheetURL(url string) error {
	if url == "" {
		return ErrDriveURLRequired
	}
	if _, err := DriveURLToExportURL(url, "csv"); err != nil {
		return ErrInvalidDriveSpreadsheetURL
	}
	return nil
}
