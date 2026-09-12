package constants

type SeriesScope string

const (
	SeriesScopeSingle SeriesScope = "single"
	SeriesScopeFuture SeriesScope = "future"
)

const MaxRepeatScheduleDates = 52

func ParseSeriesScope(value string) SeriesScope {
	if SeriesScope(value) == SeriesScopeFuture {
		return SeriesScopeFuture
	}
	return SeriesScopeSingle
}
