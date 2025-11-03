package processor

import "time"

type ClassRecord struct {
	Name              string
	DurationInMinutes int
	Rate              float64
	StartTime         *time.Time
	EndTime           *time.Time
	GoogleLink        string
	Status            string
	Date              string
}

type ColumnIndices struct {
	Name      int
	Duration  int
	Rate      int
	Status    int
	StartTime int
	EndTime   int
	Link      int
}

type ColumnMapping struct {
	NameCol      string
	DurationCol  string
	RateCol      string
	StatusCol    string
	StartTimeCol string
	EndTimeCol   string
	LinkCol      string
}

func (m ColumnMapping) ToColumnIndices() ColumnIndices {
	return ColumnIndices{
		Name:      ColumnLetterToIndex(m.NameCol),
		Duration:  ColumnLetterToIndex(m.DurationCol),
		Rate:      ColumnLetterToIndex(m.RateCol),
		Status:    ColumnLetterToIndex(m.StatusCol),
		StartTime: ColumnLetterToIndex(m.StartTimeCol),
		EndTime:   ColumnLetterToIndex(m.EndTimeCol),
		Link:      ColumnLetterToIndex(m.LinkCol),
	}
}
