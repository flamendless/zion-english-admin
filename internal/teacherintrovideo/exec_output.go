package teacherintrovideo

import "strings"

const maxExecOutputLogChars = 2048

func trimExecOutput(out []byte) string {
	s := strings.TrimSpace(string(out))
	if s == "" {
		return "(no output)"
	}
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > maxExecOutputLogChars {
		return s[:maxExecOutputLogChars] + "..."
	}
	return s
}
