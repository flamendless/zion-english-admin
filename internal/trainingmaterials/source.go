package trainingmaterials

type SourceType string

const (
	SourceYouTube SourceType = "youtube"
)

func ValidSourceType(v string) bool {
	return v == string(SourceYouTube)
}

func AllowedSourceTypes() []SourceType {
	return []SourceType{SourceYouTube}
}
