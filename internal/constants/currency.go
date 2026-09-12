package constants

var CurrencyCodes = []string{"KRW", "CAD", "YEN", "PHP"}

func CurrencySymbol(code string) string {
	switch code {
	case "KRW":
		return "₩"
	case "CAD":
		return "C$"
	case "YEN":
		return "¥"
	case "PHP":
		return "₱"
	default:
		return ""
	}
}
