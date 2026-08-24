package meetings

import "fmt"

func interfaceToString(v any) string {
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

func stringToNullInterface(value string) any {
	if value == "" {
		return nil
	}
	return value
}
