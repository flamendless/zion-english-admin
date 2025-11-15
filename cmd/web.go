package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"zion-english/internal/logs"
	"zion-english/internal/processor"
	"zion-english/internal/sheet"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type WebFlags struct {
	port    string
	baseURL string
	https   bool
	address string
}

var webFlags WebFlags

func init() {
	f := cmdWeb.Flags
	f().StringVarP(&webFlags.port, "port", "p", "8080", "Port to run web server on")
	f().StringVarP(&webFlags.baseURL, "url", "b", "zion-english-admin", "Base URL")
	f().BoolVar(&webFlags.https, "https", false, "Enable HTTPS")
	f().StringVar(&webFlags.address, "address", "", "Domain address for Let's Encrypt certificates (e.g., flamendless.xyz)")
	rootCmd.AddCommand(cmdWeb)
}

var cmdWeb = &cobra.Command{
	Use:   "web",
	Short: "Start web server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := os.MkdirAll("tmp", 0755); err != nil {
			panic(err)
		}

		basePath := "/" + strings.TrimPrefix(webFlags.baseURL, "/")
		http.HandleFunc(basePath, handleIndex)
		http.HandleFunc(basePath+"/process", handleProcess)
		http.HandleFunc(basePath+"/download/", handleDownload)

		port := webFlags.port
		if !strings.HasPrefix(port, ":") {
			port = ":" + port
		}

		logs.Log().Info(
			"Starting web server",
			zap.String("port", port),
			zap.String("base URL", webFlags.baseURL),
			zap.Bool("https", webFlags.https),
		)

		var err error
		if webFlags.https {
			if webFlags.address == "" {
				panic("--address flag is required when --https is enabled")
			}
			certFile := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", webFlags.address)
			keyFile := fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", webFlags.address)
			logs.Log().Info(
				"Starting HTTPS server",
				zap.String("address", webFlags.address),
				zap.String("cert", certFile),
				zap.String("key", keyFile),
			)
			err = http.ListenAndServeTLS(port, certFile, keyFile, nil)
		} else {
			err = http.ListenAndServe(port, nil)
		}

		if err != nil {
			panic(err)
		}
	},
}

type ProcessRequest struct {
	DriveURL    string `json:"driveUrl"`
	Name        string `json:"name"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	NameCol     string `json:"nameCol"`
	DurationCol string `json:"durationCol"`
	RateCol     string `json:"rateCol"`
	StatusCol   string `json:"statusCol"`
}

type ProcessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Logs    []string    `json:"logs"`
	Records []RecordRow `json:"records"`
	Total   float64     `json:"total"`
}

type RecordRow struct {
	Name      string  `json:"name"`
	Date      string  `json:"date"`
	Duration  int     `json:"duration"`
	Rate      float64 `json:"rate"`
	StartTime string  `json:"startTime,omitempty"`
	EndTime   string  `json:"endTime,omitempty"`
	Link      string  `json:"link,omitempty"`
	Status    string  `json:"status"`
}

var logMessages []string

func addLog(msg string) {
	logMessages = append(logMessages, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
	logs.Log().Info(msg)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading template: %v", err), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
		return
	}
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logMessages = []string{} // Reset logs for this request

	var req ProcessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	// Validation
	if err := validateRequest(&req); err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	addLog(fmt.Sprintf("Processing request for: %s", req.Name))

	// Download file
	inputPath := filepath.Join("tmp", fmt.Sprintf("%s_input.csv", req.Name))
	if err := sheet.DownloadDriveSheet(req.DriveURL, inputPath); err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to download file: %v", err))
		return
	}

	addLog(fmt.Sprintf("Downloaded file to: %s", inputPath))

	// Parse dates
	parsedStartDate, err := processor.ParseDateString(req.StartDate)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Invalid start date: %v", err))
		return
	}
	targetStartDate := *parsedStartDate

	parsedEndDate, err := processor.ParseDateString(req.EndDate)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Invalid end date: %v", err))
		return
	}
	targetEndDate := *parsedEndDate

	// Process CSV
	colIndices := processor.ColumnIndices{
		Name:      processor.ColumnLetterToIndex(req.NameCol),
		Duration:  processor.ColumnLetterToIndex(req.DurationCol),
		Rate:      processor.ColumnLetterToIndex(req.RateCol),
		Status:    processor.ColumnLetterToIndex(req.StatusCol),
		StartTime: -1,
		EndTime:   -1,
		Link:      -1,
	}

	records, err := processor.ProcessCSVFile(inputPath, targetStartDate, targetEndDate, colIndices)
	if err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to process CSV: %v", err))
		return
	}

	addLog(fmt.Sprintf("Processed %d records", len(records)))

	// Save output
	outputPath := filepath.Join("tmp", fmt.Sprintf("%s_output.csv", req.Name))
	if err := processor.SaveRecordsToCSV(records, outputPath, colIndices); err != nil {
		sendErrorResponse(w, fmt.Sprintf("Failed to save output: %v", err))
		return
	}

	addLog(fmt.Sprintf("Saved output to: %s", outputPath))

	// Calculate total
	var total float64
	for _, rec := range records {
		total += rec.Rate
	}

	// Convert to response format
	responseRecords := make([]RecordRow, len(records))
	for i, rec := range records {
		responseRecords[i] = RecordRow{
			Name:     rec.Name,
			Date:     rec.Date,
			Duration: rec.DurationInMinutes,
			Rate:     rec.Rate,
			Status:   rec.Status,
		}
		if rec.StartTime != nil {
			responseRecords[i].StartTime = rec.StartTime.Format("15:04")
		}
		if rec.EndTime != nil {
			responseRecords[i].EndTime = rec.EndTime.Format("15:04")
		}
		responseRecords[i].Link = rec.GoogleLink
	}

	response := ProcessResponse{
		Success: true,
		Message: "Processing completed successfully",
		Logs:    logMessages,
		Records: responseRecords,
		Total:   total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	var name string
	for i, part := range parts {
		if part == "download" && i+1 < len(parts) {
			name = parts[i+1]
			break
		}
	}

	if name == "" {
		http.Error(w, "Missing filename", http.StatusBadRequest)
		return
	}

	// Validate name to prevent directory traversal
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	if !matched {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	outputPath := filepath.Join("tmp", fmt.Sprintf("%s_output.csv", name))

	// Check if file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", name))
	http.ServeFile(w, r, outputPath)
}

func validateRequest(req *ProcessRequest) error {
	if req.DriveURL == "" {
		return errors.New("google drive uRL is required")
	}
	if !strings.Contains(req.DriveURL, "docs.google.com") {
		return errors.New("invalid Google Drive URL")
	}

	// Validate name
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, req.Name)
	if !matched {
		return errors.New("name must contain only letters, numbers, dashes, and underscores")
	}

	// Validate dates
	if req.StartDate == "" {
		return errors.New("start date is required")
	}
	if req.EndDate == "" {
		return errors.New("end date is required")
	}

	parsedStartDate, err := processor.ParseDateString(req.StartDate)
	if err != nil {
		return errors.New("invalid start date format")
	}

	parsedEndDate, err := processor.ParseDateString(req.EndDate)
	if err != nil {
		return errors.New("invalid end date format")
	}

	if !parsedEndDate.After(*parsedStartDate) {
		return errors.New("end date must be after start date")
	}

	return nil
}

func sendErrorResponse(w http.ResponseWriter, message string) {
	response := ProcessResponse{
		Success: false,
		Message: message,
		Logs:    logMessages,
		Records: []RecordRow{},
		Total:   0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}
