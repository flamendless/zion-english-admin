package constants

type ScheduledClassStatus string

const (
	ScheduledClassStatusScheduled ScheduledClassStatus = "scheduled"
)

func (s ScheduledClassStatus) String() string {
	return string(s)
}
