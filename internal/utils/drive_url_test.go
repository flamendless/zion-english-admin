package utils

import (
	"strings"
	"testing"
)

func TestDriveURLToExportURL(t *testing.T) {
	tests := []struct {
		name        string
		driveURL    string
		format      string
		expectedURL string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid URL with csv format",
			driveURL:    "https://docs.google.com/spreadsheets/d/1OAvXZCxxVKDmc0zTUfntLA_7sxMffsDvzREe_XaUd5g/edit?usp=sharing",
			format:      "csv",
			expectedURL: "https://docs.google.com/spreadsheets/d/1OAvXZCxxVKDmc0zTUfntLA_7sxMffsDvzREe_XaUd5g/export?format=csv&gid=0",
			expectError: false,
		},
		{
			name:        "valid URL with xlsx format",
			driveURL:    "https://docs.google.com/spreadsheets/d/1OAvXZCxxVKDmc0zTUfntLA_7sxMffsDvzREe_XaUd5g/edit?usp=sharing",
			format:      "xlsx",
			expectedURL: "https://docs.google.com/spreadsheets/d/1OAvXZCxxVKDmc0zTUfntLA_7sxMffsDvzREe_XaUd5g/export?format=xlsx&gid=0",
			expectError: false,
		},
		{
			name:        "valid URL with pdf format",
			driveURL:    "https://docs.google.com/spreadsheets/d/1OAvXZCxxVKDmc0zTUfntLA_7sxMffsDvzREe_XaUd5g/view",
			format:      "pdf",
			expectedURL: "https://docs.google.com/spreadsheets/d/1OAvXZCxxVKDmc0zTUfntLA_7sxMffsDvzREe_XaUd5g/export?format=pdf&gid=0",
			expectError: false,
		},
		{
			name:        "valid URL with different path endings",
			driveURL:    "https://docs.google.com/spreadsheets/d/ABC123XYZ456/edit",
			format:      "csv",
			expectedURL: "https://docs.google.com/spreadsheets/d/ABC123XYZ456/export?format=csv&gid=0",
			expectError: false,
		},
		{
			name:        "invalid URL - missing trailing segment",
			driveURL:    "https://docs.google.com/spreadsheets/d/ABC123XYZ456",
			format:      "csv",
			expectError: true,
			errorMsg:    "invalid Google Sheets URL path",
		},
		{
			name:        "valid URL with ods format",
			driveURL:    "https://docs.google.com/spreadsheets/d/1OAvXZCxxVKDmc0zTUfntLA_7sxMffsDvzREe_XaUd5g/edit",
			format:      "ods",
			expectedURL: "https://docs.google.com/spreadsheets/d/1OAvXZCxxVKDmc0zTUfntLA_7sxMffsDvzREe_XaUd5g/export?format=ods&gid=0",
			expectError: false,
		},
		{
			name:        "invalid URL - malformed",
			driveURL:    "not-a-valid-url",
			format:      "csv",
			expectError: true,
		},
		{
			name:        "invalid URL - wrong host",
			driveURL:    "https://example.com/spreadsheets/d/ABC123/edit",
			format:      "csv",
			expectError: true,
			errorMsg:    "not a Google Docs URL",
		},
		{
			name:        "invalid URL - missing document ID",
			driveURL:    "https://docs.google.com/spreadsheets/d/",
			format:      "csv",
			expectError: true,
			errorMsg:    "invalid Google Sheets URL path",
		},
		{
			name:        "invalid URL - wrong path structure",
			driveURL:    "https://docs.google.com/documents/d/ABC123/edit",
			format:      "csv",
			expectError: true,
			errorMsg:    "invalid Google Sheets URL path",
		},
		{
			name:        "invalid URL - missing spreadsheets segment",
			driveURL:    "https://docs.google.com/d/ABC123/edit",
			format:      "csv",
			expectError: true,
			errorMsg:    "invalid Google Sheets URL path",
		},
		{
			name:        "invalid URL - missing 'd' segment",
			driveURL:    "https://docs.google.com/spreadsheets/ABC123/edit",
			format:      "csv",
			expectError: true,
			errorMsg:    "invalid Google Sheets URL path",
		},
		{
			name:        "valid URL with http scheme",
			driveURL:    "http://docs.google.com/spreadsheets/d/ABC123/edit",
			format:      "csv",
			expectedURL: "http://docs.google.com/spreadsheets/d/ABC123/export?format=csv&gid=0",
			expectError: false,
		},
		{
			name:        "valid URL with complex document ID",
			driveURL:    "https://docs.google.com/spreadsheets/d/1aB2cD3eF4gH5iJ6kL7mN8oP9qR0sT1uV2wX3yZ4/edit",
			format:      "xlsx",
			expectedURL: "https://docs.google.com/spreadsheets/d/1aB2cD3eF4gH5iJ6kL7mN8oP9qR0sT1uV2wX3yZ4/export?format=xlsx&gid=0",
			expectError: false,
		},
		{
			name:        "valid URL with gid in query parameter",
			driveURL:    "https://docs.google.com/spreadsheets/d/1RZUx74fXMtFYoWsLaJvfvWB2ULXHX2S5IMecdj48p90/edit?gid=1190559729",
			format:      "csv",
			expectedURL: "https://docs.google.com/spreadsheets/d/1RZUx74fXMtFYoWsLaJvfvWB2ULXHX2S5IMecdj48p90/export?format=csv&gid=1190559729",
			expectError: false,
		},
		{
			name:        "valid URL with gid in fragment",
			driveURL:    "https://docs.google.com/spreadsheets/d/1RZUx74fXMtFYoWsLaJvfvWB2ULXHX2S5IMecdj48p90/edit#gid=1190559729",
			format:      "csv",
			expectedURL: "https://docs.google.com/spreadsheets/d/1RZUx74fXMtFYoWsLaJvfvWB2ULXHX2S5IMecdj48p90/export?format=csv&gid=1190559729",
			expectError: false,
		},
		{
			name:        "valid URL with gid in both query and fragment (query takes precedence)",
			driveURL:    "https://docs.google.com/spreadsheets/d/1RZUx74fXMtFYoWsLaJvfvWB2ULXHX2S5IMecdj48p90/edit?gid=1190559729#gid=999999",
			format:      "csv",
			expectedURL: "https://docs.google.com/spreadsheets/d/1RZUx74fXMtFYoWsLaJvfvWB2ULXHX2S5IMecdj48p90/export?format=csv&gid=1190559729",
			expectError: false,
		},
		{
			name:        "valid URL with gid in query and other parameters",
			driveURL:    "https://docs.google.com/spreadsheets/d/1RZUx74fXMtFYoWsLaJvfvWB2ULXHX2S5IMecdj48p90/edit?usp=sharing&gid=1190559729",
			format:      "xlsx",
			expectedURL: "https://docs.google.com/spreadsheets/d/1RZUx74fXMtFYoWsLaJvfvWB2ULXHX2S5IMecdj48p90/export?format=xlsx&gid=1190559729",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DriveURLToExportURL(tt.driveURL, tt.format)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errorMsg != "" {
					if !strings.Contains(err.Error(), tt.errorMsg) {
						t.Errorf("expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expectedURL {
					t.Errorf("expected URL %s, got %s", tt.expectedURL, result)
				}
			}
		})
	}
}

func BenchmarkDriveURLToExportURL(b *testing.B) {
	validURL := "https://docs.google.com/spreadsheets/d/1OAvXZCxxVKDmc0zTUfntLA_7sxMffsDvzREe_XaUd5g/edit?usp=sharing"

	b.Run("valid URL CSV", func(b *testing.B) {
		for b.Loop() {
			_, _ = DriveURLToExportURL(validURL, "csv")
		}
	})

	b.Run("valid URL XLSX", func(b *testing.B) {
		for b.Loop() {
			_, _ = DriveURLToExportURL(validURL, "xlsx")
		}
	})

	b.Run("valid URL PDF", func(b *testing.B) {
		for b.Loop() {
			_, _ = DriveURLToExportURL(validURL, "pdf")
		}
	})
}

func BenchmarkDriveURLToExportURLInvalid(b *testing.B) {
	invalidURL := "not-a-valid-url"

	b.Run("invalid URL", func(b *testing.B) {
		for b.Loop() {
			_, _ = DriveURLToExportURL(invalidURL, "csv")
		}
	})

	wrongHostURL := "https://example.com/spreadsheets/d/ABC123/edit"
	b.Run("wrong host", func(b *testing.B) {
		for b.Loop() {
			_, _ = DriveURLToExportURL(wrongHostURL, "csv")
		}
	})

	invalidPathURL := "https://docs.google.com/documents/d/ABC123/edit"
	b.Run("invalid path", func(b *testing.B) {
		for b.Loop() {
			_, _ = DriveURLToExportURL(invalidPathURL, "csv")
		}
	})
}

func BenchmarkDriveURLToExportURLComplexID(b *testing.B) {
	complexURL := "https://docs.google.com/spreadsheets/d/1aB2cD3eF4gH5iJ6kL7mN8oP9qR0sT1uV2wX3yZ4/edit"

	b.Run("complex document ID", func(b *testing.B) {
		for b.Loop() {
			_, _ = DriveURLToExportURL(complexURL, "csv")
		}
	})
}
