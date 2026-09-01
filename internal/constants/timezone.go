package constants

import "time"

const (
	MonthLayout           = "2006-01"
	DateLayout            = "2006-01-02"
	DateTimeLayout        = "2006-01-02 15:04"
	DateTimeSecondsLayout = "2006-01-02 15:04:05"
	TimeHMLayout          = "15:04"
	TimeHMSecondsLayout   = "15:04:05"
	TimezoneNamePHT       = "Asia/Manila"
)

var LocationPHT = mustLoadLocation(TimezoneNamePHT)

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic("constants: load timezone " + name + ": " + err.Error())
	}
	return loc
}
