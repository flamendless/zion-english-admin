package processor

import "strings"

var LightStudentColors = []string{
	"FFE6E6", // light red
	"E6F3FF", // light blue
	"E6FFE6", // light green
	"FFF9E6", // light yellow
	"F0E6FF", // light purple
	"FFE6CC", // light orange
	"FFE6F0", // light pink
	"E6FFFF", // light cyan
	"F5F5F5", // light gray
	"E6F5FF", // light sky blue
	"FFF0E6", // light peach
	"F0FFE6", // light lime
	"FFE6F5", // light rose
	"E6E6FF", // light lavender
	"FFFFE6", // light cream
}

var defaultStudentColors = map[string]bool{
	"#90c020": true,
	"#b9d283": true,
}

func StudentColor(index int) string {
	return LightStudentColors[index%len(LightStudentColors)]
}

func StudentColorHex(index int) string {
	return "#" + StudentColor(index)
}

func IsDefaultStudentColor(color string) bool {
	return defaultStudentColors[strings.ToLower(color)]
}
