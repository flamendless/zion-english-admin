package utils

import (
	"database/sql"
	"math"
	"strings"
	"time"
)

const SensitiveChangeCooldown = 7 * 24 * time.Hour

func PersonInitials(first, middle, last, fullName string) string {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	middle = strings.TrimSpace(middle)

	if first != "" && last != "" {
		return initialsFromRunes([]rune(first), []rune(last))
	}
	if first != "" && middle != "" {
		return initialsFromRunes([]rune(first), []rune(middle))
	}
	if first != "" {
		return string([]rune(strings.ToUpper(first))[:1])
	}

	parts := strings.Fields(strings.TrimSpace(fullName))
	switch len(parts) {
	case 0:
		return "?"
	case 1:
		r := []rune(strings.ToUpper(parts[0]))
		if len(r) == 0 {
			return "?"
		}
		if len(r) == 1 {
			return string(r)
		}
		return string(r[:2])
	default:
		firstRunes := []rune(strings.ToUpper(parts[0]))
		lastRunes := []rune(strings.ToUpper(parts[len(parts)-1]))
		return initialsFromRunes(firstRunes, lastRunes)
	}
}

func initialsFromRunes(first, last []rune) string {
	if len(first) == 0 || len(last) == 0 {
		return "?"
	}
	return string(first[:1]) + string(last[:1])
}

func SensitiveChangeAllowed(lastChanged sql.NullTime, now time.Time) (allowed bool, daysRemaining int) {
	if !lastChanged.Valid {
		return true, 0
	}

	elapsed := now.Sub(lastChanged.Time)
	if elapsed >= SensitiveChangeCooldown {
		return true, 0
	}

	remaining := SensitiveChangeCooldown - elapsed
	daysRemaining = max(int(math.Ceil(remaining.Hours()/24)), 1)
	return false, daysRemaining
}
