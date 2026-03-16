package constants

import "regexp"

var (
	ReSafeName = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	ReLower    = regexp.MustCompile(`[a-z]`)
	ReUpper    = regexp.MustCompile(`[A-Z]`)
	ReDigit    = regexp.MustCompile(`\d`)
	ReSpecial  = regexp.MustCompile(`[!@#$%^&*]`)
	ReLength   = regexp.MustCompile(`^[A-Za-z\d!@#$%^&*]{8,32}$`)
)
