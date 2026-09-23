package frontend

import "strings"

type TeacherTableColumnID string

const (
	TeacherTableColumnDocsStatus    TeacherTableColumnID = "docsStatus"
	TeacherTableColumnResumeStatus  TeacherTableColumnID = "resumeStatus"
	TeacherTableColumnConnections   TeacherTableColumnID = "connectedTo"
)

const (
	TeacherTableColumnsSummaryNone = "None"
	TeacherTableColumnsSummaryAll  = "All"
)

type TeacherTableColumnDef struct {
	ID             TeacherTableColumnID
	Label          string
	HeaderLabel    string
	DefaultVisible bool
}

var TeacherTableOptionalColumns = []TeacherTableColumnDef{
	{
		ID:             TeacherTableColumnDocsStatus,
		Label:          "Doc status",
		HeaderLabel:    "Docs Status",
		DefaultVisible: false,
	},
	{
		ID:             TeacherTableColumnResumeStatus,
		Label:          "Resume",
		HeaderLabel:    "Resume",
		DefaultVisible: false,
	},
	{
		ID:             TeacherTableColumnConnections,
		Label:          "Connected to",
		HeaderLabel:    "Connections",
		DefaultVisible: false,
	},
}

func teacherTableColumnByID(id TeacherTableColumnID) TeacherTableColumnDef {
	for _, col := range TeacherTableOptionalColumns {
		if col.ID == id {
			return col
		}
	}
	return TeacherTableColumnDef{ID: id, Label: string(id), HeaderLabel: string(id)}
}

func teacherTableColumnsDefaultVisibility() map[TeacherTableColumnID]bool {
	visible := make(map[TeacherTableColumnID]bool, len(TeacherTableOptionalColumns))
	for _, col := range TeacherTableOptionalColumns {
		visible[col.ID] = col.DefaultVisible
	}
	return visible
}

func teacherTableColumnsSummary(visible map[TeacherTableColumnID]bool) string {
	if len(TeacherTableOptionalColumns) == 0 {
		return TeacherTableColumnsSummaryNone
	}

	selected := make([]string, 0, len(TeacherTableOptionalColumns))
	for _, col := range TeacherTableOptionalColumns {
		if visible[col.ID] {
			selected = append(selected, col.Label)
		}
	}

	switch len(selected) {
	case 0:
		return TeacherTableColumnsSummaryNone
	case len(TeacherTableOptionalColumns):
		return TeacherTableColumnsSummaryAll
	default:
		return strings.Join(selected, ", ")
	}
}

func teacherTableColumnsDefaultSummary() string {
	return teacherTableColumnsSummary(teacherTableColumnsDefaultVisibility())
}
