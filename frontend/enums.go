package frontend

type ClassRecordSource string

const (
	ClassRecordSourceRecord    ClassRecordSource = "record"
	ClassRecordSourceScheduled ClassRecordSource = "scheduled"
)

func (s ClassRecordSource) String() string {
	return string(s)
}

type ClassActionContext string

const (
	ClassActionContextSchedule ClassActionContext = "schedule"
	ClassActionContextClasses  ClassActionContext = "classes"
)

func (c ClassActionContext) String() string {
	return string(c)
}

type ListFilterKind string

const (
	ListFilterKindTeacher ListFilterKind = "teacher"
	ListFilterKindClass   ListFilterKind = "class"
	ListFilterKindStudent ListFilterKind = "student"
)

func (k ListFilterKind) String() string {
	return string(k)
}
