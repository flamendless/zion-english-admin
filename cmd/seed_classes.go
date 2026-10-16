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

		monthLabel := fmt.Sprintf("%04d-%02d", year, month)
		logs.Log().Info("seed_classes summary",
			zap.String("teacher_email", teacherEmail),
			zap.Int64("teacher_id", summary.TeacherID),
			zap.Bool("teacher_created", summary.TeacherCreated),
			zap.Bool("teacher_existing", summary.TeacherExisting),
			zap.Int("students", summary.StudentsCreated),
			zap.Int("conducted", summary.Conducted),
			zap.Int("scheduled", summary.Scheduled),
			zap.Int("cancelled", summary.Cancelled),
			zap.String("month", monthLabel),
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
	f().BoolVar(&seedClasses.dryRun, "dry-run", true, "Log planned inserts without writing")
	rootCmd.AddCommand(cmdSeedClasses)
}

type seedClassesSummary struct {
	TeacherID       int64
	TeacherCreated  bool
	TeacherExisting bool
	StudentsCreated int
	Conducted       int
	Scheduled       int
	Cancelled       int
}

type seedStudentRef struct {
	ID   int64
	Name string
}

func runSeedClasses(ctx context.Context, dbRW database.Service, teacherEmail string, n int, year int, month time.Month, dryRun bool) (seedClassesSummary, error) {
	q := dbRW.GetQueries()
	summary := seedClassesSummary{}
	monthLabel := fmt.Sprintf("%04d-%02d", year, month)

	teacherID, created, existing, teacherPlan, err := planSeedTeacher(ctx, q, teacherEmail)
	if err != nil {
		return summary, err
	}
	summary.TeacherID = teacherID
	summary.TeacherCreated = created
	summary.TeacherExisting = existing
	if dryRun {
		logSeedTeacherPlan(teacherPlan)
	} else if created {
		teacherID, err = insertSeedTeacher(ctx, q, teacherPlan)
		if err != nil {
			return summary, err
		}
		summary.TeacherID = teacherID
	}

	students, err := planSeedStudents(n)
	if err != nil {
		return summary, err
	}
	if dryRun {
		for _, s := range students {
			logSeedStudentPlan(summary.TeacherID, teacherEmail, s)
		}
	} else {
		students, err = insertSeedStudents(ctx, q, summary.TeacherID, students)
		if err != nil {
			return summary, err
		}
	}
	summary.StudentsCreated = len(students)

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

	teacherID = summary.TeacherID
	for i := 0; i < n; i++ {
		student := students[i%len(students)]
		date := pickSeedDate(i, pastFirst, pastSecond)
		start := seedClassStartTime(i, 0)
		end := utils.EndTimeFromStartAndDuration(start, seedClassDurationMins)
		if dryRun {
			logSeedClassRecordPlan(monthLabel, teacherID, teacherEmail, student, date, start, end, string(constants.ClassStatusConducted), "")
			summary.Conducted++
			continue
		}
		if _, err := q.InsertClassRecord(ctx, queries.InsertClassRecordParams{
			StudentID:       student.ID,
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
		student := students[i%len(students)]
		date := pickSeedDate(i, futureFirst, futureSecond)
		start := seedClassStartTime(i, 1)
		if dryRun {
			logSeedScheduledClassPlan(monthLabel, teacherID, teacherEmail, student, date, start)
			summary.Scheduled++
			continue
		}
		if _, err := q.InsertScheduledClass(ctx, queries.InsertScheduledClassParams{
			StudentID:       student.ID,
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
		student := students[i%len(students)]
		date := pickSeedDate(i, pastFirst, pastSecond)
		start := seedClassStartTime(i, 2)
		end := utils.EndTimeFromStartAndDuration(start, seedClassDurationMins)
		if dryRun {
			logSeedClassRecordPlan(monthLabel, teacherID, teacherEmail, student, date, start, end, string(constants.ClassStatusCancelled), "seed_classes")
			summary.Cancelled++
			continue
		}
		if _, err := q.InsertClassRecord(ctx, queries.InsertClassRecordParams{
			StudentID:       student.ID,
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

type seedTeacherPlan struct {
	Action       string
	ID           int64
	Email        string
	FirstName    string
	LastName     string
	MobileNumber string
	JoiningDate  string
	Status       string
	Role         string
}

func planSeedTeacher(ctx context.Context, q *queries.Queries, email string) (teacherID int64, created, existing bool, plan seedTeacherPlan, err error) {
	plan.Email = email
	row, lookupErr := q.GetTeacherByEmail(ctx, email)
	if lookupErr == nil {
		plan.Action = "use_existing"
		plan.ID = row.ID
		return row.ID, false, true, plan, nil
	}
	if !errors.Is(lookupErr, sql.ErrNoRows) {
		return 0, false, false, plan, lookupErr
	}

	local := email
	if at := strings.Index(email, "@"); at > 0 {
		local = email[:at]
	}
	plan.Action = "create"
	plan.FirstName = "Seed"
	plan.LastName = local
	if len(plan.LastName) > 40 {
		plan.LastName = plan.LastName[:40]
	}
	plan.JoiningDate = utils.TodayPHT()
	plan.MobileNumber = seedMobileForEmail(email)
	plan.Status = string(constants.TeacherStatusApproved)
	plan.Role = string(constants.TeacherRoleTeacher)
	return 0, true, false, plan, nil
}

func insertSeedTeacher(ctx context.Context, q *queries.Queries, plan seedTeacherPlan) (int64, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(seedTeacherPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	if err := q.InsertTeacher(ctx, queries.InsertTeacherParams{
		FirstName:           plan.FirstName,
		MiddleName:          "",
		LastName:            plan.LastName,
		Birthdate:           "1990-01-01",
		Address:             "Seed address",
		JoiningDate:         plan.JoiningDate,
		MobileNumber:        plan.MobileNumber,
		Email:               plan.Email,
		Certifications:      sql.NullString{},
		AssignedColor:       constants.DefaultTeacherAssignedColor,
		RatePerClass:        seedDefaultRate,
		RateDurationMinutes: int64(constants.DefaultRateDurationMinutes),
		Currency:            "PHP",
		DriveUrl:            seedTeacherDriveURL,
		Sex:                 sql.NullString{},
		Password:            string(hashed),
		Status:              plan.Status,
	}); err != nil {
		return 0, err
	}

	inserted, err := q.GetTeacherByEmail(ctx, plan.Email)
	if err != nil {
		return 0, err
	}
	if err := q.InsertTeacherRole(ctx, queries.InsertTeacherRoleParams{
		TeacherID: inserted.ID,
		Role:      plan.Role,
	}); err != nil {
		return 0, err
	}
	return inserted.ID, nil
}

func planSeedStudents(n int) ([]seedStudentRef, error) {
	if n < 1 {
		return nil, ErrSeedStudentsRequired
	}
	students := make([]seedStudentRef, 0, n)
	for i := 0; i < n; i++ {
		students = append(students, seedStudentRef{
			Name: fmt.Sprintf("Seed Student %d", i+1),
		})
	}
	return students, nil
}

func insertSeedStudents(ctx context.Context, q *queries.Queries, teacherID int64, students []seedStudentRef) ([]seedStudentRef, error) {
	for i := range students {
		color := processor.StudentColorHex(i)
		studentID, err := q.InsertStudent(ctx, queries.InsertStudentParams{
			Name:                students[i].Name,
			Currency:            "PHP",
			Contact:             sql.NullString{},
			RatePerClass:        seedDefaultRate,
			RateDurationMinutes: int64(constants.DefaultRateDurationMinutes),
			ParentName:          sql.NullString{},
			ParentRate:          sql.NullFloat64{},
			ParentCurrency:      sql.NullString{},
			AssignedColor:       color,
			Status:              string(constants.StudentStatusActive),
			InactiveReason:      sql.NullString{},
			DeletedReason:       sql.NullString{},
			IsTestSeed:          1,
		})
		if err != nil {
			return students, err
		}
		students[i].ID = studentID
		if err := q.InsertTeacherStudentM2M(ctx, queries.InsertTeacherStudentM2MParams{
			TeacherID: teacherID,
			StudentID: studentID,
		}); err != nil {
			return students, err
		}
	}
	return students, nil
}

func logSeedTeacherPlan(plan seedTeacherPlan) {
	if plan.Action == "use_existing" {
		logs.Log().Info("seed_classes plan teacher",
			zap.String("action", "use_existing"),
			zap.Int64("teacher_id", plan.ID),
			zap.String("email", plan.Email),
		)
		return
	}
	logs.Log().Info("seed_classes plan teacher",
		zap.String("action", "create"),
		zap.String("email", plan.Email),
		zap.String("first_name", plan.FirstName),
		zap.String("last_name", plan.LastName),
		zap.String("mobile_number", plan.MobileNumber),
		zap.String("joining_date", plan.JoiningDate),
		zap.String("status", plan.Status),
		zap.String("role", plan.Role),
		zap.String("password", seedTeacherPassword),
		zap.Float64("rate_per_class", seedDefaultRate),
		zap.String("currency", "PHP"),
	)
}

func logSeedStudentPlan(teacherID int64, teacherEmail string, student seedStudentRef) {
	logs.Log().Info("seed_classes plan student",
		zap.String("action", "create"),
		zap.String("name", student.Name),
		zap.String("status", string(constants.StudentStatusActive)),
		zap.Float64("rate_per_class", seedDefaultRate),
		zap.String("currency", "PHP"),
		zap.Int64("assign_teacher_id", teacherID),
		zap.String("assign_teacher_email", teacherEmail),
	)
}

func logSeedClassRecordPlan(monthLabel string, teacherID int64, teacherEmail string, student seedStudentRef, date, start, end, status, reason string) {
	fields := []zap.Field{
		zap.String("action", "create"),
		zap.String("kind", "class_record"),
		zap.String("month", monthLabel),
		zap.Int64("teacher_id", teacherID),
		zap.String("teacher_email", teacherEmail),
		zap.String("student_name", student.Name),
		zap.String("date", date),
		zap.String("start_time", start),
		zap.String("end_time", end),
		zap.Int64("duration_minutes", seedClassDurationMins),
		zap.Float64("rate", seedDefaultRate),
		zap.String("currency", "PHP"),
		zap.String("status", status),
	}
	if student.ID > 0 {
		fields = append(fields, zap.Int64("student_id", student.ID))
	}
	if reason != "" {
		fields = append(fields, zap.String("reason", reason))
	}
	logs.Log().Info("seed_classes plan class", fields...)
}

func logSeedScheduledClassPlan(monthLabel string, teacherID int64, teacherEmail string, student seedStudentRef, date, start string) {
	fields := []zap.Field{
		zap.String("action", "create"),
		zap.String("kind", "scheduled_class"),
		zap.String("month", monthLabel),
		zap.Int64("teacher_id", teacherID),
		zap.String("teacher_email", teacherEmail),
		zap.String("student_name", student.Name),
		zap.String("scheduled_date", date),
		zap.String("start_time", start),
		zap.Int64("duration_minutes", seedClassDurationMins),
		zap.Float64("rate", seedDefaultRate),
		zap.String("currency", "PHP"),
		zap.String("status", string(constants.ScheduledClassStatusScheduled)),
	}
	if student.ID > 0 {
		fields = append(fields, zap.Int64("student_id", student.ID))
	}
	logs.Log().Info("seed_classes plan class", fields...)
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
