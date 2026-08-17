package constants

func ValidCurrency(currency string) bool {
	switch currency {
	case "KRW", "CAD", "YEN", "PHP":
		return true
	default:
		return false
	}
}

func ValidStudentStatus(status string) bool {
	return status == "active" || status == "inactive"
}

func ValidClassStatus(status string) bool {
	return status == "conducted" || status == "cancelled" || status == "rescheduled"
}

func ValidSex(sex string) bool {
	return sex == "" || sex == "M" || sex == "F"
}

func ValidPassword(p string) bool {
	return ReLength.MatchString(p) &&
		ReLower.MatchString(p) &&
		ReUpper.MatchString(p) &&
		ReDigit.MatchString(p) &&
		ReSpecial.MatchString(p)
}
