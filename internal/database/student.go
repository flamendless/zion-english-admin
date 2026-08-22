package database

import (
	"context"

	"zion-english/internal/database/queries"
)

type Student = queries.TblStudent

func GetAllStudents(db Service) ([]Student, error) {
	rows, err := db.GetQueries().GetAllStudents(context.Background())
	if err != nil {
		return nil, err
	}
	students := make([]Student, len(rows))
	for i, row := range rows {
		students[i] = Student{
			ID:             row.ID,
			Name:           row.Name,
			Currency:       row.Currency,
			Contact:        row.Contact,
			RatePerClass:   row.RatePerClass,
			ParentName:     row.ParentName,
			AssignedColor:  row.AssignedColor,
			Status:         row.Status,
			InactiveReason: row.InactiveReason,
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		}
	}
	return students, nil
}
