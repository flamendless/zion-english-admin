package database

import (
	"context"

	"zion-english/internal/database/queries"
)

type Teacher = queries.TblTeacher

func GetAllTeachers(db Service) ([]Teacher, error) {
	return db.GetQueries().GetAllTeachers(context.Background())
}
