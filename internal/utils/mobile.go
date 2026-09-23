package utils

import (
	"strings"
	"zion-english/internal/constants"
)

func ValidMobileNumber(mobile string) bool {
	mobile = strings.TrimSpace(mobile)
	if len(mobile) < 7 || len(mobile) > 32 {
		return false
	}
	if !constants.ReMobileNumber.MatchString(mobile) {
		return false
	}
	digits := 0
	for _, r := range mobile {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	return digits >= 7 && digits <= 15
}
