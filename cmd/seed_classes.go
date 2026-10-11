package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"time"
	"zion-english/internal/auth"
	"zion-english/internal/constants"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/processor"
	"zion-english/internal/utils"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	seedTeacherPassword   = "SeedPass1!"
	seedTeacherDriveURL   = "https://docs.google.com/spreadsheets/d/1seed00000000000000000000000000000/edit"
	seedClassDurationMins = 60
	seedDefaultRate       = 500.0
)

var (
	ErrSeedTeacherEmailRequired = errors.New("teacher email is required")
	ErrSeedStudentsRequired     = errors.New("students must be at least 1")
	ErrSeedInvalidMonth         = errors.New("month must be YYYY-MM in Philippines time")
)

type seedClassesFlags struct {
	teacherEmail string
	students     int
	month        string
	dryRun       bool
}

var seedClasses seedClassesFlags

var cmdSeedClasses = &cobra.Command{
	Use:   "seed_classes",
	Short: "Seed a teacher, students, and class records for a calendar month",
	Run: func(cmd *cobra.Command, args []string) {
		teacherEmail := strings.TrimSpace(strings.ToLower(seedClasses.teacherEmail))
		if teacherEmail == "" {
			panic(ErrSeedTeacherEmailRequired)
		}
		if seedClasses.students < 1 {
			panic(ErrSeedStudentsRequired)
		}

		year, month, ok := utils.ParseMonthPHT(seedClasses.month)
		if !ok {
			panic(ErrSeedInvalidMonth)
		}

		if err := database.Init("data/zion.db"); err != nil {
			panic(fmt.Sprintf("failed to initialize database: %v", err))
		}
		defer database.Close()

		dbRW := database.New(database.DB_MODE_RW)
		ctx := context.Background()

		summary, err := runSeedClasses(ctx, dbRW, teacherEmail, seedClasses.students, year, month, seedClasses.dryRun)
		if err != nil {
			panic(err)
		}

		logs.Log().Info("seed_classes complete",
			zap.String("teacher_email", teacherEmail),
			zap.Int64("teacher_id", summary.TeacherID),
			zap.Bool("teacher_created", summary.TeacherCreated),
			zap.Int("students", summary.StudentsCreated),
			zap.Int("conducted", summary.Conducted),
			zap.Int("scheduled", summary.Scheduled),
			zap.Int("cancelled", summary.Cancelled),
			zap.String("month", fmt.Sprintf("%04d-%02d", year, month)),
			zap.Bool("dry_run", seedClasses.dryRun),
		)
		if summary.TeacherCreated && !seedClasses.dryRun {
			logs.Log().Info("seed_classes created teacher login",
				zap.String("email", teacherEmail),
				zap.String("password", seedTeacherPassword),
			)
		}
	},
}

func init() {
	f := cmdSeedClasses.Flags
	_ = cmdSeedClasses.MarkFlagRequired("teacher-email")
	_ = cmdSeedClasses.MarkFlagRequired("students")
	f().StringVar(&seedClasses.teacherEmail, "teacher-email", "", "Teacher email (created if missing)")
	f().IntVar(&seedClasses.students, "students", 0, "Number of students and of each class status to create")
	f().StringVar(&seedClasses.month, "month", utils.CurrentMonthPHT(), "Calendar month (YYYY-MM, PHT)")
	f().BoolVar(&seedClasses.dryRun, "dry-run", false, "Log planned inserts without writing")
	rootCmd.AddCommand(cmdSeedClasses)
}

type seedClassesSummary struct {
	TeacherID       int64
	TeacherCreated  bool
	StudentsCreated int
	Conducted       int
	Scheduled       int
	Cancelled       int
}

func runSeedClasses(ctx context.Context, dbRW database.Service, teacherEmail string, n int, year int, month time.Month, dryRun bool) (seedClassesSummary, error) {
	q := dbRW.GetQueries()
	summary := seedClassesSummary{}

	teacherID, created, err := ensureSeedTeacher(ctx, q, teacherEmail, dryRun)
	if err != nil {
		return summary, err
	}
	summary.TeacherID = teacherID
	summary.TeacherCreated = created

	studentIDs, err := ensureSeedStudents(ctx, q, teacherID, n, dryRun)
	if err != nil {
		return summary, err
	}
	summary.StudentsCreated = len(studentIDs)

	firstDates, secondDates := cutoffDatePools(year, month)
	today := utils.TodayPHT()

	pastFirst := filterSeedDates(firstDates, func(d string) bool { return d < today })
	pastSecond := filterSeedDates(secondDates, func(d string) bool { return d < today })
	futureFirst := filterSeedDates(firstDates, func(d string) bool { return d >= today })
	futureSecond := filterSeedDates(secondDates, func(d string) bool { return d >= today })

	if len(pastFirst) == 0 && len(pastSecond) == 0 {
		pastFirst, pastSecond = firstDates, secondDates
	}
	if len(futureFirst) == 0 && len(futureSecond) == 0 {
		futureFirst, futureSecond = firstDates, secondDates
	}

	for i := 0; i < n; i++ {
		studentID := studentIDs[i%len(studentIDs)]
		date := pickSeedDate(i, pastFirst, pastSecond)
		start := seedClassStartTime(i, 0)
		end := utils.EndTimeFromStartAndDuration(start, seedClassDurationMins)
		if dryRun {
			summary.Conducted++
			continue
		}
		if _, err := q.InsertClassRecord(ctx, queries.InsertClassRecordParams{
			StudentID:       studentID,
			TeacherID:       teacherID,
			Date:            date,
			StartTime:       sql.NullString{String: start, Valid: true},
			EndTime:         sql.NullString{String: end, Valid: true},
			DurationMinutes: seedClassDurationMins,
			Rate:            seedDefaultRate,
			Currency:        "PHP",
			IsTrialClass:    0,
			Status:          string(constants.ClassStatusConducted),
			Reason:          sql.NullString{},
			Notes:           sql.NullString{String: "seed_classes", Valid: true},
			RecordedByRole:  string(auth.RoleSuperuser),
		}); err != nil {
			return summary, err
		}
		summary.Conducted++
	}

	for i := 0; i < n; i++ {
		studentID := studentIDs[i%len(studentIDs)]
		date := pickSeedDate(i, futureFirst, futureSecond)
		start := seedClassStartTime(i, 1)
		if dryRun {
			summary.Scheduled++
			continue
		}
		if _, err := q.InsertScheduledClass(ctx, queries.InsertScheduledClassParams{
			StudentID:       studentID,
			TeacherID:       teacherID,
			ScheduledDate:   date,
			StartTime:       sql.NullString{String: start, Valid: true},
			DurationMinutes: seedClassDurationMins,
			Rate:            seedDefaultRate,
			Currency:        "PHP",
			IsTrialClass:    0,
			SeriesID:        sql.NullInt64{},
			Reason:          sql.NullString{},
			CreatedByRole:   string(auth.RoleSuperuser),
		}); err != nil {
			return summary, err
		}
		summary.Scheduled++
	}

	for i := 0; i < n; i++ {
		studentID := studentIDs[i%len(studentIDs)]
		date := pickSeedDate(i, pastFirst, pastSecond)
		start := seedClassStartTime(i, 2)
		end := utils.EndTimeFromStartAndDuration(start, seedClassDurationMins)
		if dryRun {
			summary.Cancelled++
			continue
		}
		if _, err := q.InsertClassRecord(ctx, queries.InsertClassRecordParams{
			StudentID:       studentID,
			TeacherID:       teacherID,
			Date:            date,
			StartTime:       sql.NullString{String: start, Valid: true},
			EndTime:         sql.NullString{String: end, Valid: true},
			DurationMinutes: seedClassDurationMins,
			Rate:            seedDefaultRate,
			Currency:        "PHP",
			IsTrialClass:    0,
			Status:          string(constants.ClassStatusCancelled),
			Reason:          sql.NullString{String: "seed_classes", Valid: true},
			Notes:           sql.NullString{},
			RecordedByRole:  string(auth.RoleSuperuser),
		}); err != nil {
			return summary, err
		}
		summary.Cancelled++
	}

	return summary, nil
}

func ensureSeedTeacher(ctx context.Context, q *queries.Queries, email string, dryRun bool) (int64, bool, error) {
	existing, err := q.GetTeacherByEmail(ctx, email)
	if err == nil {
		return existing.ID, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}

	local := email
	if at := strings.Index(email, "@"); at > 0 {
		local = email[:at]
	}
	firstName := "Seed"
	lastName := local
	if len(lastName) > 40 {
		lastName = lastName[:40]
	}
	joining := utils.TodayPHT()
	mobile := seedMobileForEmail(email)

	if dryRun {
		return 0, true, nil
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(seedTeacherPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, false, err
	}

	if err := q.InsertTeacher(ctx, queries.InsertTeacherParams{
		FirstName:      firstName,
		MiddleName:     "",
		LastName:       lastName,
		Birthdate:      "1990-01-01",
		Address:        "Seed address",
		JoiningDate:    joining,
		MobileNumber:   mobile,
		Email:          email,
		Certifications: sql.NullString{},
		AssignedColor:  constants.DefaultTeacherAssignedColor,
		RatePerClass:   seedDefaultRate,
		Currency:       "PHP",
		DriveUrl:       seedTeacherDriveURL,
		Sex:            sql.NullString{},
		Password:       string(hashed),
		Status:         string(constants.TeacherStatusApproved),
	}); err != nil {
		return 0, false, err
	}

	inserted, err := q.GetTeacherByEmail(ctx, email)
	if err != nil {
		return 0, false, err
	}
	if err := q.InsertTeacherRole(ctx, queries.InsertTeacherRoleParams{
		TeacherID: inserted.ID,
		Role:      string(constants.TeacherRoleTeacher),
	}); err != nil {
		return 0, false, err
	}
	return inserted.ID, true, nil
}

func ensureSeedStudents(ctx context.Context, q *queries.Queries, teacherID int64, n int, dryRun bool) ([]int64, error) {
	ids := make([]int64, 0, n)
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("Seed Student %d", i+1)
		color := processor.StudentColorHex(i)
		if dryRun {
			ids = append(ids, int64(i+1))
			continue
		}
		studentID, err := q.InsertStudent(ctx, queries.InsertStudentParams{
			Name:           name,
			Currency:       "PHP",
			Contact:        sql.NullString{},
			RatePerClass:   seedDefaultRate,
			ParentName:     sql.NullString{},
			ParentRate:     sql.NullFloat64{},
			ParentCurrency: sql.NullString{},
			AssignedColor:  color,
			Status:         string(constants.StudentStatusActive),
			InactiveReason: sql.NullString{},
			DeletedReason:  sql.NullString{},
		})
		if err != nil {
			return ids, err
		}
		if err := q.InsertTeacherStudentM2M(ctx, queries.InsertTeacherStudentM2MParams{
			TeacherID: teacherID,
			StudentID: studentID,
		}); err != nil {
			return ids, err
		}
		ids = append(ids, studentID)
	}
	return ids, nil
}

func cutoffDatePools(year int, month time.Month) (firstHalf, secondHalf []string) {
	firstPreset, secondPreset := utils.CutoffRangeForMonth(year, month)
	firstStart, firstEnd := utils.CutoffDatesFromPreset(firstPreset)
	secondStart, secondEnd := utils.CutoffDatesFromPreset(secondPreset)
	firstHalf = datesBetweenInclusive(firstStart, firstEnd)
	secondHalf = datesBetweenInclusive(secondStart, secondEnd)
	return firstHalf, secondHalf
}

func datesBetweenInclusive(start, end string) []string {
	startT, err1 := utils.ParseDatePHT(start)
	endT, err2 := utils.ParseDatePHT(end)
	if err1 != nil || err2 != nil || startT == nil || endT == nil {
		return nil
	}
	var out []string
	for d := *startT; !d.After(*endT); d = d.AddDate(0, 0, 1) {
		out = append(out, d.In(constants.LocationPHT).Format(constants.DateLayout))
	}
	return out
}

func filterSeedDates(dates []string, keep func(string) bool) []string {
	var out []string
	for _, d := range dates {
		if keep(d) {
			out = append(out, d)
		}
	}
	return out
}

// pickSeedDate alternates first/second payroll cutoffs and cycles within each pool.
func pickSeedDate(index int, firstHalf, secondHalf []string) string {
	if index%2 == 0 && len(firstHalf) > 0 {
		return firstHalf[(index/2)%len(firstHalf)]
	}
	if len(secondHalf) > 0 {
		return secondHalf[(index/2)%len(secondHalf)]
	}
	if len(firstHalf) > 0 {
		return firstHalf[index%len(firstHalf)]
	}
	return utils.TodayPHT()
}

func seedClassStartTime(classIndex int, kindOffset int) string {
	slot := classIndex*3 + kindOffset
	hour := 8 + (slot % 10)
	minute := (slot * 7) % 60
	return fmt.Sprintf("%02d:%02d", hour, minute)
}

func seedMobileForEmail(email string) string {
	h := fnv.New32a()
	h.Write([]byte(email))
	n := h.Sum32() % 10000000
	return fmt.Sprintf("+63917%07d", n)
}
