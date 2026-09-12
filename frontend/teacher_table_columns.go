package frontend

import "strings"

type TeacherTableColumnID string

const (
	TeacherTableColumnDocsStatus  TeacherTableColumnID = "docsStatus"
	TeacherTableColumnConnections TeacherTableColumnID = "connectedTo"
)

type TeacherTableColumnDef struct {
	ID             TeacherTableColumnID
	Label          string
	HeaderLabel    string
	DefaultVisible bool
}

// TeacherTableOptionalColumns defines toggleable columns shown after Status on the teachers list.
// Add entries here to expose new optional columns in the toolbar dropdown.
var TeacherTableOptionalColumns = []TeacherTableColumnDef{
	{
		ID:             TeacherTableColumnDocsStatus,
		Label:          "Doc status",
		HeaderLabel:    "Docs Status",
		DefaultVisible: true,
	},
	{
		ID:             TeacherTableColumnConnections,
		Label:          "Connected to",
		HeaderLabel:    "Connections",
		DefaultVisible: true,
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
		return "None"
	}

	selected := make([]string, 0, len(TeacherTableOptionalColumns))
	for _, col := range TeacherTableOptionalColumns {
		if visible[col.ID] {
			selected = append(selected, col.Label)
		}
	}

	switch len(selected) {
	case 0:
		return "None"
	case len(TeacherTableOptionalColumns):
		return "All"
	default:
		return strings.Join(selected, ", ")
	}
}

func teacherTableColumnsDefaultSummary() string {
	return teacherTableColumnsSummary(teacherTableColumnsDefaultVisibility())
}
