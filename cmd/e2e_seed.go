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
	e2eTeacherEmail       = "e2e-teacher@zion-english.test"
	e2eTeacherPassword    = "E2eTeacherPass1!"
	e2eAdminEmail         = "e2e-admin@zion-english.test"
	e2eAdminPassword      = "E2eAdminPass1!"
	e2eStudentName        = "E2E Test Student"
	e2eSiblingStudentName = "E2E Sibling Student"
	e2eParentName         = "E2E Parent"
	e2eCredentialsFile    = "e2e/.auth-credentials.json"
)

type e2eCredentials struct {
	Email                string  `json:"email"`
	Password             string  `json:"password"`
	TeacherID            int64   `json:"teacherId"`
	StudentName          string  `json:"studentName"`
	StudentID            int64   `json:"studentId"`
	StudentRate          float64 `json:"studentRate"`
	Currency             string  `json:"currency"`
	SuperuserEmail       string  `json:"superuserEmail"`
	SuperuserPassword    string  `json:"superuserPassword"`
	OverdueScheduledDate string  `json:"overdueScheduledDate"`
	AdminEmail           string  `json:"adminEmail"`
	AdminPassword        string  `json:"adminPassword"`
	SiblingStudentName   string  `json:"siblingStudentName"`
	ParentName           string  `json:"parentName"`
}

var cmdE2ESeed = &cobra.Command{
	Use:   "e2e-seed",
	Short: "Reset E2E database and seed teacher, student, and assignment",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := database.E2EDBPath()
		fmt.Printf("e2e-seed: resetting %s (never data/zion.db)\n", dbPath)
		database.RemoveSQLiteDBFiles(dbPath)
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
			FirstName:           "E2E",
			MiddleName:          "",
			LastName:            "Teacher",
			Birthdate:           "1990-01-01",
			Address:             "E2E Address",
			JoiningDate:         "2020-01-01",
			MobileNumber:        "+6391700000001",
			Email:               e2eTeacherEmail,
			Certifications:      sql.NullString{},
			AssignedColor:       constants.DefaultTeacherAssignedColor,
			RatePerClass:        500,
			RateDurationMinutes: int64(constants.DefaultRateDurationMinutes),
			Currency:            "PHP",
			DriveUrl:            "",
			Sex:                 sql.NullString{},
			Password:            string(hashed),
			Status:              string(constants.TeacherStatusApproved),
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
			Name:                e2eStudentName,
			Currency:            "PHP",
			Contact:             sql.NullString{},
			RatePerClass:        500,
			RateDurationMinutes: int64(constants.DefaultRateDurationMinutes),
			ParentName:          sql.NullString{String: e2eParentName, Valid: true},
			ParentRate:          sql.NullFloat64{},
			ParentCurrency:      sql.NullString{},
			AssignedColor:       constants.DefaultAssignedColor,
			Status:              string(constants.StudentStatusActive),
			InactiveReason:      sql.NullString{},
			DeletedReason:       sql.NullString{},
			IsTestSeed:          1,
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

		siblingID, err := q.InsertStudent(ctx, queries.InsertStudentParams{
			Name:                e2eSiblingStudentName,
			Currency:            "PHP",
			Contact:             sql.NullString{},
			RatePerClass:        500,
			RateDurationMinutes: int64(constants.DefaultRateDurationMinutes),
			ParentName:          sql.NullString{String: e2eParentName, Valid: true},
			ParentRate:          sql.NullFloat64{},
			ParentCurrency:      sql.NullString{},
			AssignedColor:       constants.DefaultAssignedColor,
			Status:              string(constants.StudentStatusActive),
			InactiveReason:      sql.NullString{},
			DeletedReason:       sql.NullString{},
		})
		if err != nil {
			panic(fmt.Sprintf("failed to insert sibling student: %v", err))
		}
		if err := q.InsertStudentRelationship(ctx, queries.InsertStudentRelationshipParams{
			StudentID:        studentID,
			RelatedStudentID: siblingID,
			Relationship:     sql.NullString{String: "siblings", Valid: true},
		}); err != nil {
			panic(fmt.Sprintf("failed to insert student relationship: %v", err))
		}
		if err := q.InsertTeacherStudentM2M(ctx, queries.InsertTeacherStudentM2MParams{
			TeacherID: teacher.ID,
			StudentID: siblingID,
		}); err != nil {
			panic(fmt.Sprintf("failed to link teacher and sibling student: %v", err))
		}

		// PHT calendar date vs SQLite datetime('now') (UTC): "today" at 00:00 PHT can be after UTC now.
		effectiveStart := "2020-01-01 00:00:00"
		if _, err := q.InsertTeacherPlanTransaction(ctx, queries.InsertTeacherPlanTransactionParams{
			TeacherID:          teacher.ID,
			Tier:               string(constants.TeacherPlanTierPro),
			BillingKind:        string(constants.TeacherPlanBillingLifetime),
			EffectiveStart:     effectiveStart,
			EffectiveEnd:       sql.NullString{},
			GrantedByTeacherID: sql.NullInt64{},
			GrantedByName:      "e2e-seed",
			Note:               "E2E Teacher Pro",
		}); err != nil {
			panic(fmt.Sprintf("failed to grant teacher pro: %v", err))
		}

		adminHashed, err := bcrypt.GenerateFromPassword([]byte(e2eAdminPassword), bcrypt.DefaultCost)
		if err != nil {
			panic(fmt.Sprintf("failed to hash admin password: %v", err))
		}
		if err := q.InsertTeacher(ctx, queries.InsertTeacherParams{
			FirstName:           "E2E",
			MiddleName:          "",
			LastName:            "Admin",
			Birthdate:           "1990-01-01",
			Address:             "E2E Address",
			JoiningDate:         "2020-01-01",
			MobileNumber:        "+6391700000002",
			Email:               e2eAdminEmail,
			Certifications:      sql.NullString{},
			AssignedColor:       constants.DefaultTeacherAssignedColor,
			RatePerClass:        500,
			RateDurationMinutes: int64(constants.DefaultRateDurationMinutes),
			Currency:            "PHP",
			DriveUrl:            "",
			Sex:                 sql.NullString{},
			Password:            string(adminHashed),
			Status:              string(constants.TeacherStatusApproved),
		}); err != nil {
			panic(fmt.Sprintf("failed to insert admin teacher: %v", err))
		}
		adminTeacher, err := q.GetTeacherByEmail(ctx, e2eAdminEmail)
		if err != nil {
			panic(fmt.Sprintf("failed to load admin teacher: %v", err))
		}
		if err := q.InsertTeacherRole(ctx, queries.InsertTeacherRoleParams{
			TeacherID: adminTeacher.ID,
			Role:      string(constants.TeacherRoleAdmin),
		}); err != nil {
			panic(fmt.Sprintf("failed to insert admin role: %v", err))
		}
		if err := q.InsertTeacherRole(ctx, queries.InsertTeacherRoleParams{
			TeacherID: adminTeacher.ID,
			Role:      string(constants.TeacherRoleTeacher),
		}); err != nil {
			panic(fmt.Sprintf("failed to insert admin teacher role: %v", err))
		}

		todayPHT, err := time.ParseInLocation(constants.DateLayout, utils.TodayPHT(), constants.LocationPHT)
		if err != nil {
			panic(fmt.Sprintf("failed to parse today PHT: %v", err))
		}
		overdueDate := e2eOverdueScheduledDate(todayPHT)
		conductedDate := utils.TodayPHT()
		_, err = q.InsertClassRecord(ctx, queries.InsertClassRecordParams{
			StudentID:       studentID,
			TeacherID:       teacher.ID,
			Date:            conductedDate,
			StartTime:       sql.NullString{String: "09:00", Valid: true},
			EndTime:         sql.NullString{String: "10:00", Valid: true},
			DurationMinutes: 60,
			Rate:            500,
			Currency:        "PHP",
			IsTrialClass:    0,
			Status:          string(constants.ClassStatusConducted),
			Reason:          sql.NullString{},
			Notes:           sql.NullString{},
			RecordedByRole:  string(auth.RoleTeacher),
		})
		if err != nil {
			panic(fmt.Sprintf("failed to insert conducted class: %v", err))
		}
		_, err = q.InsertScheduledClass(ctx, queries.InsertScheduledClassParams{
			StudentID:       siblingID,
			TeacherID:       teacher.ID,
			ScheduledDate:   overdueDate,
			StartTime:       sql.NullString{String: "06:00", Valid: true},
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
		superEmail := os.Getenv("SUPERUSER_USERNAME")
		superPassword := os.Getenv("SUPERUSER_PASSWORD")
		if superEmail == "" {
			superEmail = cfg.SuperuserUsername
		}
		if superPassword == "" {
			superPassword = cfg.SuperuserPassword
		}
		creds := e2eCredentials{
			Email:                e2eTeacherEmail,
			Password:             e2eTeacherPassword,
			TeacherID:            teacher.ID,
			StudentName:          e2eStudentName,
			StudentID:            studentID,
			StudentRate:          500,
			Currency:             "PHP",
			SuperuserEmail:       superEmail,
			SuperuserPassword:    superPassword,
			OverdueScheduledDate: overdueDate,
			AdminEmail:           e2eAdminEmail,
			AdminPassword:        e2eAdminPassword,
			SiblingStudentName:   e2eSiblingStudentName,
			ParentName:           e2eParentName,
		}
		if err := writeE2ECredentials(creds); err != nil {
			panic(fmt.Sprintf("failed to write credentials: %v", err))
		}
		fmt.Printf("e2e seed ok: teacher %s, student %q (id %d)\n", e2eTeacherEmail, e2eStudentName, studentID)
	},
}

// e2eOverdueScheduledDate picks a calendar date inside the dashboard week range (local TZ)
// so the persistent overdue panel can list the seeded class (Monday: yesterday is often outside the week).
func e2eOverdueScheduledDate(todayPHT time.Time) string {
	weekStart, weekEnd := weekRange()
	todayLocal := time.Now().Format(constants.DateLayout)
	yesterday := todayPHT.AddDate(0, 0, -1).Format(constants.DateLayout)
	for _, candidate := range []string{yesterday, todayLocal, todayPHT.Format(constants.DateLayout), weekStart} {
		if candidate >= weekStart && candidate <= weekEnd {
			return candidate
		}
	}
	return weekStart
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
