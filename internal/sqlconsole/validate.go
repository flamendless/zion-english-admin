package sqlconsole

import (
	"regexp"
	"strings"
	"unicode"
	"zion-english/internal/constants"
)

var forbiddenKeywords = []string{
	"INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER", "REPLACE",
	"ATTACH", "DETACH", "PRAGMA", "VACUUM", "REINDEX", "ANALYZE",
	"BEGIN", "COMMIT", "ROLLBACK", "SAVEPOINT", "EXPLAIN",
}

var startsSelectOrWith = regexp.MustCompile(`^(?i)(with\b|select\b)`)

func NormalizeQuery(sql string) string {
	return strings.TrimSpace(sql)
}

func ValidateReadOnlyQuery(sql string) (string, error) {
	q := NormalizeQuery(sql)
	if q == "" {
		return "", ErrSQLQueryEmpty
	}
	if len(q) > constants.SQLConsoleMaxQueryBytes {
		return "", ErrSQLQueryTooLong
	}

	withoutTrailing := strings.TrimSuffix(q, ";")
	if strings.Contains(withoutTrailing, ";") {
		return "", ErrSQLMultipleStatements
	}
	q = strings.TrimSpace(withoutTrailing)
	if q == "" {
		return "", ErrSQLQueryEmpty
	}

	compact := collapseWhitespace(q)
	if !startsSelectOrWith.MatchString(compact) {
		return "", ErrSQLNotSelect
	}
	if containsForbiddenKeyword(compact) {
		return "", ErrSQLForbiddenKeyword
	}
	return q, nil
}

func collapseWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	lastSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
			}
			continue
		}
		lastSpace = false
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func containsForbiddenKeyword(compact string) bool {
	upper := strings.ToUpper(compact)
	for _, kw := range forbiddenKeywords {
		if wordMatch(upper, kw) {
			return true
		}
	}
	return false
}

func wordMatch(upper string, keyword string) bool {
	idx := 0
	for {
		pos := strings.Index(upper[idx:], keyword)
		if pos < 0 {
			return false
		}
		start := idx + pos
		end := start + len(keyword)
		beforeOK := start == 0 || !isIdentByte(upper[start-1])
		afterOK := end == len(upper) || !isIdentByte(upper[end])
		if beforeOK && afterOK {
			return true
		}
		idx = start + 1
		if idx >= len(upper) {
			return false
		}
	}
}

func isIdentByte(b byte) bool {
	return b == '_' || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}
