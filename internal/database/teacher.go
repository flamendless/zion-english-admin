package database

import (
	"context"

	"zion-english/internal/database/queries"
)

type Teacher = queries.TblTeacher

func GetAllTeachers(db Service) ([]Teacher, error) {
	rows, err := db.GetQueries().GetAllTeachers(context.Background())
	if err != nil {
		return nil, err
	}
	teachers := make([]Teacher, len(rows))
	for i, r := range rows {
		teachers[i] = Teacher{
			ID:             r.ID,
			FirstName:      r.FirstName,
			MiddleName:     r.MiddleName,
			LastName:       r.LastName,
			Birthdate:      r.Birthdate,
			Address:        r.Address,
			JoiningDate:    r.JoiningDate,
			MobileNumber:   r.MobileNumber,
			Email:          r.Email,
			Certifications: r.Certifications,
			AssignedColor:  r.AssignedColor,
			RatePerClass:   r.RatePerClass,
			Currency:       r.Currency,
			DriveUrl:       r.DriveUrl,
			Sex:            r.Sex,
			Status:         r.Status,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
		}
	}
	return teachers, nil
}
