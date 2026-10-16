package constants

const DefaultRateDurationMinutes = 25

type RateDurationMinutes int

const (
	RateDuration25 RateDurationMinutes = 25
	RateDuration50 RateDurationMinutes = 50
)

func ValidRateDurationMinutes(n int) bool {
	return n == int(RateDuration25) || n == int(RateDuration50)
}

type RateDurationOption struct {
	Value int
	Label string
}

func RateDurationOptions() []RateDurationOption {
	return []RateDurationOption{
		{Value: int(RateDuration25), Label: "25 minutes"},
		{Value: int(RateDuration50), Label: "50 minutes"},
	}
}
