package utils

import (
	"database/sql"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"zion-english/internal/constants"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}
	return string(b)
}

func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func InterfaceToString(v any) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprint(v)
	}
}

func NullIfEmptyString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func NullInt64(id int64) sql.NullInt64 {
	if id <= 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: id, Valid: true}
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ProfileNameEditable reports whether a teacher may set this name field on their profile.
// It checks the raw stored database value only; display placeholders such as constants.DisplayNotSet are never used.
func ProfileNameEditable(stored string) bool {
	return stored == ""
}

func ComposePersonName(first, middle, last string) string {
	parts := make([]string, 0, 3)
	for _, part := range []string{strings.TrimSpace(first), strings.TrimSpace(middle), strings.TrimSpace(last)} {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, " ")
}

// TeacherCompactDisplayName shortens a full person name for list cells: first word plus last-name initial.
func TeacherCompactDisplayName(fullName string) string {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" || fullName == constants.DisplayNotSet {
		return fullName
	}
	parts := strings.Fields(fullName)
	if len(parts) == 1 {
		return parts[0]
	}
	lastRunes := []rune(parts[len(parts)-1])
	if len(lastRunes) == 0 {
		return parts[0]
	}
	initial := string(unicode.ToUpper(lastRunes[0]))
	return parts[0] + " " + initial
}

func SanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = constants.ReSafeName.ReplaceAllString(name, "_")
	if len(name) > 50 {
		name = name[:50]
	}
	if name == "" {
		name = "file"
	}
	return name
}

func SanitizeSheetName(name string) string {
	replacer := strings.NewReplacer(
		"\\", "",
		"/", "",
		"?", "",
		"*", "",
		"[", "",
		"]", "",
		":", "",
	)
	name = strings.TrimSpace(replacer.Replace(name))
	if len(name) > 31 {
		name = name[:31]
	}
	return name
}
