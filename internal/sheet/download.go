package sheet

import (
	"os"
	"os/exec"
	"path/filepath"
	"zion-english/internal/utils"
)

func DownloadDriveSheet(driveURL, outputPath string) error {
	exportURL, err := utils.DriveURLToExportURL(driveURL, "csv")
	if err != nil {
		return err
	}

	return downloadFileWithCurl(exportURL, outputPath)
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
