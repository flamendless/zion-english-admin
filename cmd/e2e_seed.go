package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"zion-english/internal/auth"
	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
	"zion-english/internal/utils"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

const (
	e2eTeacherEmail    = "e2e-teacher@zion-english.test"
	e2eTeacherPassword = "E2eTeacherPass1!"
	e2eStudentName     = "E2E Test Student"
	e2eCredentialsFile = "e2e/.auth-credentials.json"
)

type e2eCredentials struct {
	Email              string  `json:"email"`
	Password           string  `json:"password"`
	StudentName        string  `json:"studentName"`
	StudentRate        float64 `json:"studentRate"`
	Currency           string  `json:"currency"`
	SuperuserEmail         string `json:"superuserEmail"`
	SuperuserPassword      string `json:"superuserPassword"`
	OverdueScheduledDate   string `json:"overdueScheduledDate"`
}

var cmdE2ESeed = &cobra.Command{
	Use:   "e2e-seed",
	Short: "Reset E2E database and seed teacher, student, and assignment",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := database.DBPathFromEnv()
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			panic(fmt.Sprintf("failed to remove e2e database: %v", err))
		}
		if err := database.Init(dbPath); err != nil {
			panic(fmt.Sprintf("failed to initialize database: %v", err))
		}
		defer database.Close()

		dbRW := database.New(database.DB_MODE_RW)
		ctx := context.Background()
		q := dbRW.GetQueries()

		hashed, err := bcrypt.GenerateFromPassword([]byte(e2eTeacherPassword), bcrypt.DefaultCost)
		if err != nil {
			panic(fmt.Sprintf("failed to hash password: %v", err))
		}

		if err := q.InsertTeacher(ctx, queries.InsertTeacherParams{
			FirstName:      "E2E",
			MiddleName:     "",
			LastName:       "Teacher",
			Birthdate:      "1990-01-01",
			Address:        "E2E Address",
			JoiningDate:    "2020-01-01",
			MobileNumber:   "+6391700000001",
			Email:          e2eTeacherEmail,
			Certifications: sql.NullString{},
			AssignedColor:  constants.DefaultTeacherAssignedColor,
			RatePerClass:   500,
			Currency:       "PHP",
			DriveUrl:       "",
			Sex:            sql.NullString{},
			Password:       string(hashed),
			Status:         string(constants.TeacherStatusApproved),
		}); err != nil {
			panic(fmt.Sprintf("failed to insert teacher: %v", err))
		}

		teacher, err := q.GetTeacherByEmail(ctx, e2eTeacherEmail)
		if err != nil {
			panic(fmt.Sprintf("failed to load teacher: %v", err))
		}
		if err := q.InsertTeacherRole(ctx, queries.InsertTeacherRoleParams{
			TeacherID: teacher.ID,
			Role:      string(constants.TeacherRoleTeacher),
		}); err != nil {
			panic(fmt.Sprintf("failed to insert teacher role: %v", err))
		}

		studentID, err := q.InsertStudent(ctx, queries.InsertStudentParams{
			Name:           e2eStudentName,
			Currency:       "PHP",
			Contact:        sql.NullString{},
			RatePerClass:   500,
			ParentName:     sql.NullString{},
			ParentRate:     sql.NullFloat64{},
			ParentCurrency: sql.NullString{},
			AssignedColor:  constants.DefaultAssignedColor,
			Status:         string(constants.StudentStatusActive),
			InactiveReason: sql.NullString{},
			DeletedReason:  sql.NullString{},
		})
		if err != nil {
			panic(fmt.Sprintf("failed to insert student: %v", err))
		}

		if err := q.InsertTeacherStudentM2M(ctx, queries.InsertTeacherStudentM2MParams{
			TeacherID: teacher.ID,
			StudentID: studentID,
		}); err != nil {
			panic(fmt.Sprintf("failed to link teacher and student: %v", err))
		}

		todayPHT, err := time.ParseInLocation(constants.DateLayout, utils.TodayPHT(), constants.LocationPHT)
		if err != nil {
			panic(fmt.Sprintf("failed to parse today PHT: %v", err))
		}
		overdueDate := todayPHT.AddDate(0, 0, -1).Format(constants.DateLayout)
		_, err = q.InsertScheduledClass(ctx, queries.InsertScheduledClassParams{
			StudentID:       studentID,
			TeacherID:       teacher.ID,
			ScheduledDate:   overdueDate,
			StartTime:       sql.NullString{String: "09:00", Valid: true},
			DurationMinutes: 60,
			Rate:            500,
			Currency:        "PHP",
			IsTrialClass:    0,
			SeriesID:        sql.NullInt64{},
			Reason:          sql.NullString{},
			CreatedByRole:   string(auth.RoleTeacher),
		})
		if err != nil {
			panic(fmt.Sprintf("failed to insert overdue scheduled class: %v", err))
		}

		cfg := conf.Conf()
		creds := e2eCredentials{
			Email:                e2eTeacherEmail,
			Password:             e2eTeacherPassword,
			StudentName:          e2eStudentName,
			StudentRate:          500,
			Currency:             "PHP",
			SuperuserEmail:       cfg.SuperuserUsername,
			SuperuserPassword:    cfg.SuperuserPassword,
			OverdueScheduledDate: overdueDate,
		}
		if err := writeE2ECredentials(creds); err != nil {
			panic(fmt.Sprintf("failed to write credentials: %v", err))
		}
		fmt.Printf("e2e seed ok: teacher %s, student %q (id %d)\n", e2eTeacherEmail, e2eStudentName, studentID)
	},
}

func writeE2ECredentials(creds e2eCredentials) error {
	dir := filepath.Dir(e2eCredentialsFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(e2eCredentialsFile, data, 0644)
}

func init() {
	rootCmd.AddCommand(cmdE2ESeed)
}
