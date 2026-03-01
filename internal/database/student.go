package database

import (
	"context"

	"zion-english/internal/database/queries"
)

type Student = queries.TblStudent

func GetAllStudents(db Service) ([]Student, error) {
	return db.GetQueries().GetAllStudents(context.Background())
}
