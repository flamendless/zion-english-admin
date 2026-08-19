package cmd

import (
	"context"
	"fmt"
	"sort"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/processor"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type AssignStudentColorsFlags struct {
	dryRun       bool
	overwriteAll bool
}

var assignStudentColorsFlags AssignStudentColorsFlags

var cmdAssignStudentColors = &cobra.Command{
	Use:   "assign_student_colors",
	Short: "Backfill assigned_color for students with placeholder colors",
	Run: func(cmd *cobra.Command, args []string) {
		if err := database.Init("data/zion.db"); err != nil {
			panic(fmt.Sprintf("failed to initialize database: %v", err))
		}
		defer database.Close()

		dbRW := database.New(database.DB_MODE_RW)
		ctx := context.Background()

		students, err := dbRW.GetQueries().GetAllStudents(ctx)
		if err != nil {
			panic(err)
		}

		var targets []queries.TblStudent
		if assignStudentColorsFlags.overwriteAll {
			targets = students
		} else {
			for _, s := range students {
				if processor.IsDefaultStudentColor(s.AssignedColor) {
					targets = append(targets, s)
				}
			}
		}

		sort.Slice(targets, func(i, j int) bool {
			return targets[i].Name < targets[j].Name
		})

		updated := 0
		skipped := 0
		for i, s := range targets {
			newColor := processor.StudentColorHex(i)
			if newColor == s.AssignedColor {
				skipped++
				continue
			}

			logs.Log().Info(
				"assigning student color",
				zap.Int64("id", s.ID),
				zap.String("name", s.Name),
				zap.String("old_color", s.AssignedColor),
				zap.String("new_color", newColor),
				zap.Bool("dry_run", assignStudentColorsFlags.dryRun),
			)

			if assignStudentColorsFlags.dryRun {
				updated++
				continue
			}

			if err := dbRW.GetQueries().UpdateStudent(ctx, queries.UpdateStudentParams{
				Name:          s.Name,
				Currency:      s.Currency,
				Contact:       s.Contact,
				RatePerClass:  s.RatePerClass,
				ParentName:    s.ParentName,
				AssignedColor: newColor,
				Status:        s.Status,
				ID:            s.ID,
			}); err != nil {
				panic(err)
			}
			updated++
		}

		logs.Log().Info(
			"assign student colors complete",
			zap.Int("total_students", len(students)),
			zap.Int("target_students", len(targets)),
			zap.Int("updated", updated),
			zap.Int("skipped", skipped),
			zap.Bool("dry_run", assignStudentColorsFlags.dryRun),
			zap.Bool("overwrite_all", assignStudentColorsFlags.overwriteAll),
		)
	},
}

func init() {
	f := cmdAssignStudentColors.Flags
	f().BoolVar(&assignStudentColorsFlags.dryRun, "dry-run", true, "Log planned changes without writing to the database")
	f().BoolVar(&assignStudentColorsFlags.overwriteAll, "overwrite-all", false, "Reassign colors for all students, not just placeholder colors")
	rootCmd.AddCommand(cmdAssignStudentColors)
}
