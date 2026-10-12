package cmd

import (
	"context"
	"database/sql"
	"fmt"

	"zion-english/internal/database"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

var cmdE2ERestoreTeacher = &cobra.Command{
	Use:   "e2e-restore-teacher-password",
	Short: "Reset E2E teacher password and password_changed_at (for Playwright teardown)",
	Run: func(cmd *cobra.Command, args []string) {
		dbPath := database.E2EDBPath()
		hashed, err := bcrypt.GenerateFromPassword([]byte(e2eTeacherPassword), bcrypt.DefaultCost)
		if err != nil {
			panic(fmt.Sprintf("failed to hash password: %v", err))
		}

		// Short-lived connection only: do not run goose while the E2E web server holds the DB.
		dsn := dbPath + "?_busy_timeout=30000&_journal_mode=WAL"
		db, err := sql.Open("sqlite3", dsn)
		if err != nil {
			panic(fmt.Sprintf("failed to open e2e database: %v", err))
		}
		defer db.Close()

		ctx := context.Background()
		var teacherID int64
		if err := db.QueryRowContext(ctx, "SELECT id FROM tbl_teachers WHERE email = ?", e2eTeacherEmail).Scan(&teacherID); err != nil {
			panic(fmt.Sprintf("e2e teacher not found: %v", err))
		}
		if _, err := db.ExecContext(
			ctx,
			"UPDATE tbl_teachers SET password = ?, password_changed_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			string(hashed),
			teacherID,
		); err != nil {
			panic(fmt.Sprintf("failed to restore teacher password: %v", err))
		}
		fmt.Printf("e2e teacher password restored for %s\n", e2eTeacherEmail)
	},
}

func init() {
	rootCmd.AddCommand(cmdE2ERestoreTeacher)
}
