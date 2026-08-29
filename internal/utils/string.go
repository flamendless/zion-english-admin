package utils

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
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

func NullIfEmptyString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// ProfileNameEditable reports whether a teacher may set this name field on their profile.
// It checks the raw stored database value only; display placeholders such as "-" are never used.
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
