package frontend

import "zion-english/internal/constants"

var ThumbnailOrientationFilterOptions = []StatusOption{
	{Value: "", Label: "All orientations"},
	{Value: string(constants.ThumbnailOrientationLandscape), Label: constants.ThumbnailOrientationLabel(constants.ThumbnailOrientationLandscape)},
	{Value: string(constants.ThumbnailOrientationPortrait), Label: constants.ThumbnailOrientationLabel(constants.ThumbnailOrientationPortrait)},
	{Value: string(constants.ThumbnailOrientationSquare), Label: constants.ThumbnailOrientationLabel(constants.ThumbnailOrientationSquare)},
}

var ThumbnailOrientationEditOptions = []StatusOption{
	{Value: "", Label: "Not set"},
	{Value: string(constants.ThumbnailOrientationLandscape), Label: constants.ThumbnailOrientationLabel(constants.ThumbnailOrientationLandscape)},
	{Value: string(constants.ThumbnailOrientationPortrait), Label: constants.ThumbnailOrientationLabel(constants.ThumbnailOrientationPortrait)},
	{Value: string(constants.ThumbnailOrientationSquare), Label: constants.ThumbnailOrientationLabel(constants.ThumbnailOrientationSquare)},
}
