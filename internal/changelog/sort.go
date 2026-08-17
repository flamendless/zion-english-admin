package changelog

import (
	"sort"
	"strings"
	"time"

	"zion-english/internal/utils"
)

func sortVersions(versions []Version) {
	sort.Slice(versions, func(i, j int) bool {
		ti, tiOK := parseVersionDate(versions[i].Date)
		tj, tjOK := parseVersionDate(versions[j].Date)
		if tiOK && tjOK {
			if ti.Equal(tj) {
				return strings.Compare(versions[i].Version, versions[j].Version) > 0
			}
			return ti.After(tj)
		}
		if tiOK != tjOK {
			return tiOK
		}
		return strings.Compare(versions[i].Version, versions[j].Version) > 0
	})
}

func parseVersionDate(value string) (time.Time, bool) {
	t, err := utils.ParseDatePHT(value)
	if err != nil || t == nil {
		return time.Time{}, false
	}
	return *t, true
}
