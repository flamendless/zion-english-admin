package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"zion-english/internal/constants"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const testSeedBulkDeleteReason = "test/seed bulk cleanup"

type deleteTestSeedStudentsFlags struct {
	dryRun bool
	hard   bool
}

var deleteTestSeedStudents deleteTestSeedStudentsFlags

var cmdDeleteTestSeedStudents = &cobra.Command{
	Use:   "delete_test_seed_students",
	Short: "Soft-delete or hard-delete all students marked test/seed (is_test_seed = 1)",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := database.DBPathFromEnv()
		if err := database.Init(dbPath); err != nil {
			panic(fmt.Sprintf("failed to initialize database: %v", err))
		}
		defer database.Close()

		dbRW := database.New(database.DB_MODE_RW)
		ctx := context.Background()
		q := dbRW.GetQueries()

		students, err := q.GetTestSeedStudents(ctx)
		if err != nil {
			panic(err)
		}

		fmt.Printf("delete_test_seed_students: db=%s found=%d dry_run=%v hard=%v\n", dbPath, len(students), deleteTestSeedStudents.dryRun, deleteTestSeedStudents.hard)
		for _, s := range students {
			fmt.Printf("  id=%d name=%q status=%s\n", s.ID, s.Name, s.Status)
		}

		if deleteTestSeedStudents.dryRun {
			fmt.Println("dry-run: no changes written")
			return
		}

		var deleted, skipped int
		if deleteTestSeedStudents.hard {
			deleted, err = hardDeleteTestSeedStudents(ctx, dbRW)
			if err != nil {
				panic(err)
			}
		} else {
			deleted, skipped, err = softDeleteTestSeedStudents(ctx, q, students)
			if err != nil {
				panic(err)
			}
		}

		fmt.Printf("done: deleted=%d skipped=%d\n", deleted, skipped)
		logs.Log().Info(
			"delete_test_seed_students complete",
			zap.String("db_path", dbPath),
			zap.Int("found", len(students)),
			zap.Int("deleted", deleted),
			zap.Int("skipped", skipped),
			zap.Bool("hard", deleteTestSeedStudents.hard),
			zap.Bool("dry_run", false),
		)
	},
}

func softDeleteTestSeedStudents(ctx context.Context, q *queries.Queries, students []queries.GetTestSeedStudentsRow) (deleted, skipped int, err error) {
	for _, s := range students {
		if s.Status == string(constants.StudentStatusDeleted) {
			skipped++
			continue
		}
		if err := q.SoftDeleteStudent(ctx, queries.SoftDeleteStudentParams{
			DeletedReason: sql.NullString{String: testSeedBulkDeleteReason, Valid: true},
			ID:            s.ID,
		}); err != nil {
			return deleted, skipped, err
		}
		deleted++
	}
	return deleted, skipped, nil
}

func hardDeleteTestSeedStudents(ctx context.Context, dbRW database.Service) (int, error) {
	students, err := dbRW.GetQueries().GetTestSeedStudents(ctx)
	if err != nil {
		return 0, err
	}
	n := len(students)
	if n == 0 {
		return 0, nil
	}

	tx, err := dbRW.GetDB().BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	qtx := dbRW.GetQueries().WithTx(tx)

	run := func(fn func(context.Context) error) error {
		if err := fn(ctx); err != nil {
			_ = tx.Rollback()
			return err
		}
		return nil
	}

	steps := []func(context.Context) error{
		qtx.HardDeleteTestSeedClassCalendarEvents,
		qtx.HardDeleteTestSeedClassMeetingRooms,
		qtx.HardDeleteTestSeedScheduledClassLearningMaterials,
		qtx.HardDeleteTestSeedScheduledClasses,
		qtx.HardDeleteTestSeedClassRecordLearningMaterials,
		qtx.HardDeleteTestSeedClassRecords,
		qtx.HardDeleteTestSeedStudentRelationships,
		qtx.HardDeleteTestSeedTeacherStudentLinks,
		qtx.HardDeleteTestSeedStudents,
	}
	for _, step := range steps {
		if err := run(step); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return n, nil
}

func init() {
	f := cmdDeleteTestSeedStudents.Flags
	f().BoolVar(&deleteTestSeedStudents.dryRun, "dry-run", true, "List test/seed students without writing")
	f().BoolVar(&deleteTestSeedStudents.hard, "hard", false, "Hard-delete rows and related class data (default: soft delete)")
	rootCmd.AddCommand(cmdDeleteTestSeedStudents)
}
