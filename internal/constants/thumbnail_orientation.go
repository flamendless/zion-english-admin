package constants

type ThumbnailOrientation string

const (
	ThumbnailOrientationLandscape ThumbnailOrientation = "landscape"
	ThumbnailOrientationPortrait  ThumbnailOrientation = "portrait"
	ThumbnailOrientationSquare    ThumbnailOrientation = "square"
)

func ValidThumbnailOrientation(value string) bool {
	switch ThumbnailOrientation(value) {
	case ThumbnailOrientationLandscape, ThumbnailOrientationPortrait, ThumbnailOrientationSquare:
		return true
	default:
		return false
	}
}

func ThumbnailOrientationLabel(o ThumbnailOrientation) string {
	switch o {
	case ThumbnailOrientationLandscape:
		return "Landscape"
	case ThumbnailOrientationPortrait:
		return "Portrait"
	case ThumbnailOrientationSquare:
		return "Square"
	default:
		return ""
	}
}
