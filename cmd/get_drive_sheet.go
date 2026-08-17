package cmd

import (
	"zion-english/internal/logs"
	"zion-english/internal/sheet"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type GetDriveSheetFlags struct {
	url    string
	output string
}

var getDriveSheetFlags GetDriveSheetFlags

var cmdGetDriveSheet = &cobra.Command{
	Use:   "get_drive_sheet",
	Short: "Download Google Drive Sheet as CSV",
	Run: func(cmd *cobra.Command, args []string) {
		driveURL := getDriveSheetFlags.url
		outputPath := getDriveSheetFlags.output

		logs.Log().Info(
			"Downloading drive sheet",
			zap.String("url", driveURL),
			zap.String("output", outputPath),
		)

		if err := sheet.DownloadDriveSheet(driveURL, outputPath); err != nil {
			panic(err)
		}
		logs.Log().Info("Successfully downloaded file", zap.String("output", outputPath))
	},
}

func init() {
	f := cmdGetDriveSheet.Flags
	f().StringVarP(&getDriveSheetFlags.url, "url", "u", "", "Google Drive Sheets URL")
	f().StringVarP(&getDriveSheetFlags.output, "output", "o", "", "Output path")
	if err := cmdGetDriveSheet.MarkFlagRequired("url"); err != nil {
		panic(err)
	}
	if err := cmdGetDriveSheet.MarkFlagRequired("output"); err != nil {
		panic(err)
	}

	rootCmd.AddCommand(cmdGetDriveSheet)
}
