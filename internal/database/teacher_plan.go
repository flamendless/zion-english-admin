package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"
)

type TeacherPlanGrant = queries.TblTeacherPlanTransaction

func EffectiveTeacherPlanTier(ctx context.Context, q *queries.Queries, teacherID int64) (constants.TeacherPlanTier, *TeacherPlanGrant, error) {
	if teacherID <= 0 {
		return constants.TeacherPlanTierFree, nil, nil
	}
	row, err := q.GetEffectiveTeacherPlanGrant(ctx, teacherID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return constants.TeacherPlanTierFree, nil, nil
		}
		return "", nil, err
	}
	return constants.TeacherPlanTierPro, &row, nil
}

func MonthlyPlanEffectiveEndPHT(startDate string) (string, error) {
	start, err := utils.ParseDatePHT(startDate)
	if err != nil || start == nil {
		return "", err
	}
	loc := constants.LocationPHT
	t := start.In(loc)
	endOfMonth := time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 0, loc)
	return utils.DateTimeSecondsPHT(endOfMonth), nil
}

func CurrentMonthDateRangePHT() (start string, end string) {
	now := time.Now().In(constants.LocationPHT)
	startT := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, constants.LocationPHT)
	endT := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, constants.LocationPHT)
	return utils.DatePHT(startT), utils.DatePHT(endT)
}
