package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"zion-english/internal/logs"
	"zion-english/internal/utils"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type GetDriveSheetFlags struct {
	url    string
	output string
}

var getDriveSheetFlags GetDriveSheetFlags

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

		exportURL, err := utils.DriveURLToExportURL(driveURL, "csv")
		if err != nil {
			logs.Log().Error("Failed to convert drive URL to export URL", zap.Error(err))
			return
		}

		logs.Log().Info("Export URL", zap.String("url", exportURL))

		if err := downloadFileWithCurl(exportURL, outputPath); err != nil {
			logs.Log().Error("Failed to download file", zap.Error(err))
			return
		}

		logs.Log().Info("Successfully downloaded file", zap.String("output", outputPath))
	},
}

func downloadFileWithCurl(url, outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}
	}

	cmd := exec.Command("curl", "-L", "-o", outputPath, url)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
