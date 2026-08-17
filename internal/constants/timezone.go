package constants

import "time"

const (
	DateLayout      = "2006-01-02"
	TimezoneNamePHT = "Asia/Manila"
)

var LocationPHT = mustLoadLocation(TimezoneNamePHT)

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic("constants: load timezone " + name + ": " + err.Error())
	}
	return loc
}
