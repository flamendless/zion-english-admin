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
	"strconv"
	"strings"
	"time"
	"zion-english/internal/auth"
	"zion-english/internal/database"
	"zion-english/internal/logs"
	"zion-english/internal/processor"
	"zion-english/internal/sheet"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type WebFlags struct {
	port    string
	baseURL string
	https   bool
	address string
}

type AppConfig struct {
	AdminUsername string `env:"ADMIN_USERNAME" required:"true"`
	AdminPassword string `env:"ADMIN_PASSWORD" required:"true"`
}

var appConfig AppConfig

var webFlags WebFlags

func init() {
	f := cmdWeb.Flags
	f().StringVarP(&webFlags.port, "port", "p", "8080", "Port to run web server on")
	f().StringVarP(&webFlags.baseURL, "url", "b", "zion-english-admin", "Base URL")
	f().BoolVar(&webFlags.https, "https", false, "Enable HTTPS")
	f().StringVar(&webFlags.address, "address", "", "Domain address for Let's Encrypt certificates (e.g., flamendless.xyz)")
	rootCmd.AddCommand(cmdWeb)

	if err := cleanenv.ReadEnv(&appConfig); err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}
}

var cmdWeb = &cobra.Command{
	Use:   "web",
	Short: "Start web server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := os.MkdirAll("tmp", 0755); err != nil {
			panic(err)
		}

		if err := database.Init("data/processing_logs.db"); err != nil {
			panic(fmt.Sprintf("Failed to initialize database: %v", err))
		}
		defer database.Close()

		basePath := "/" + strings.TrimPrefix(webFlags.baseURL, "/")

		mux := http.NewServeMux()
		mux.HandleFunc(basePath, handleIndex)
		mux.HandleFunc(basePath+"/process", handleProcess)
		mux.HandleFunc(basePath+"/download/", handleDownload)
		mux.Handle(basePath+"/static/", http.StripPrefix(basePath+"/static/", http.FileServer(http.Dir("static"))))

		authCfg := &auth.Config{
			Username: appConfig.AdminUsername,
			Password: appConfig.AdminPassword,
		}
		authMux := http.NewServeMux()
		authMux.HandleFunc(basePath+"/logs", handleLogs)
		authHandler := auth.Middleware(authCfg, authMux)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, basePath+"/logs") {
				authHandler.ServeHTTP(w, r)
			} else {
				mux.ServeHTTP(w, r)
			}
		})

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
			err = http.ListenAndServeTLS(port, certFile, keyFile, handler)
		} else {
			err = http.ListenAndServe(port, handler)
		}

		if err != nil {
			panic(err)
		}
	},
}

type ProcessRequest struct {
	DriveURL     string `json:"driveUrl"`
	Name         string `json:"name"`
	Template     string `json:"template"`
	StartDate    string `json:"startDate"`
	EndDate      string `json:"endDate"`
	NameCol      string `json:"nameCol"`
	DurationCol  string `json:"durationCol"`
	RateCol      string `json:"rateCol"`
	StatusCol    string `json:"statusCol"`
	ExcludedRows string `json:"excludedRows"`
}

type ProcessResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Logs     []string    `json:"logs"`
	Records  []RecordRow `json:"records"`
	Total    float64     `json:"total"`
	RowRange string      `json:"rowRange,omitempty"`
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

	logMessages = []string{}
	var req ProcessRequest
	var outputPath string
	var errMsg string

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errMsg = fmt.Sprintf("Invalid request: %v", err)
		sendErrorResponse(w, errMsg, &req, r.UserAgent(), "", errMsg)
		return
	}

	if err := validateRequest(&req); err != nil {
		errMsg = err.Error()
		sendErrorResponse(w, errMsg, &req, r.UserAgent(), "", errMsg)
		return
	}

	addLog(fmt.Sprintf("Processing request for: %s", req.Name))

	inputPath := filepath.Join("tmp", fmt.Sprintf("%s_input.csv", req.Name))
	if err := sheet.DownloadDriveSheet(req.DriveURL, inputPath); err != nil {
		errMsg = fmt.Sprintf("Failed to download file: %v", err)
		sendErrorResponse(w, errMsg, &req, r.UserAgent(), "", errMsg)
		return
	}

	addLog(fmt.Sprintf("Downloaded file to: %s", inputPath))

	parsedStartDate, err := processor.ParseDateString(req.StartDate)
	if err != nil {
		errMsg = fmt.Sprintf("Invalid start date: %v", err)
		sendErrorResponse(w, errMsg, &req, r.UserAgent(), "", errMsg)
		return
	}
	targetStartDate := *parsedStartDate

	parsedEndDate, err := processor.ParseDateString(req.EndDate)
	if err != nil {
		errMsg = fmt.Sprintf("Invalid end date: %v", err)
		sendErrorResponse(w, errMsg, &req, r.UserAgent(), "", errMsg)
		return
	}
	targetEndDate := *parsedEndDate

	colIndices := processor.ColumnIndices{
		Name:      processor.ColumnLetterToIndex(req.NameCol),
		Duration:  processor.ColumnLetterToIndex(req.DurationCol),
		Rate:      processor.ColumnLetterToIndex(req.RateCol),
		Status:    processor.ColumnLetterToIndex(req.StatusCol),
		StartTime: -1,
		EndTime:   -1,
		Link:      -1,
	}

	excludedRows := make(map[int]bool)
	if req.ExcludedRows != "" {
		parts := strings.Split(req.ExcludedRows, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				if rowNum, err := strconv.Atoi(part); err == nil && rowNum > 0 {
					excludedRows[rowNum] = true
				}
			}
		}
	}

	records, err := processor.ProcessCSVFile(inputPath, targetStartDate, targetEndDate, colIndices, excludedRows)
	if err != nil {
		errMsg = fmt.Sprintf("Failed to process CSV: %v", err)
		sendErrorResponse(w, errMsg, &req, r.UserAgent(), "", errMsg)
		return
	}

	addLog(fmt.Sprintf("Processed %d records", len(records)))

	outputPath = filepath.Join("tmp", fmt.Sprintf("%s_output.xlsx", req.Name))
	if err := processor.SaveRecords(records, outputPath, colIndices, req.Name); err != nil {
		errMsg = fmt.Sprintf("Failed to save output: %v", err)
		sendErrorResponse(w, errMsg, &req, r.UserAgent(), "", errMsg)
		return
	}

	addLog(fmt.Sprintf("Saved output to: %s", outputPath))

	var total float64
	var minRow, maxRow int
	if len(records) > 0 {
		minRow = records[0].OriginalRowIndex
		maxRow = records[0].OriginalRowIndex
	}
	for _, rec := range records {
		total += rec.Rate
		if rec.OriginalRowIndex < minRow {
			minRow = rec.OriginalRowIndex
		}
		if rec.OriginalRowIndex > maxRow {
			maxRow = rec.OriginalRowIndex
		}
	}

	var rowRange string
	if len(records) > 0 {
		if minRow == maxRow {
			rowRange = fmt.Sprintf("Row: %d", minRow)
		} else {
			rowRange = fmt.Sprintf("Rows: %d-%d", minRow, maxRow)
		}
	}

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

	logToDB(&req, r.UserAgent(), outputPath, "")

	response := ProcessResponse{
		Success:  true,
		Message:  "Processing completed successfully",
		Logs:     logMessages,
		Records:  responseRecords,
		Total:    total,
		RowRange: rowRange,
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

	outputPath := filepath.Join("tmp", fmt.Sprintf("%s_output.xlsx", name))

	// Check if file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.xlsx\"", name))
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

func sendErrorResponse(w http.ResponseWriter, message string, req *ProcessRequest, userAgent, outputPath, errMsg string) {
	if req != nil {
		logToDB(req, userAgent, outputPath, errMsg)
	}

	response := ProcessResponse{
		Success:  false,
		Message:  message,
		Logs:     logMessages,
		Records:  []RecordRow{},
		Total:    0,
		RowRange: "",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func logToDB(req *ProcessRequest, userAgent, outputPath, errMsg string) {
	if req == nil {
		return
	}

	procLog := &database.ProcessingLog{
		GoogleDriveURL: req.DriveURL,
		Name:           req.Name,
		Template:       req.Template,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		ExcludedRows:   req.ExcludedRows,
		UserAgent:      userAgent,
		OutputPath:     outputPath,
		Errors:         errMsg,
	}

	if _, err := database.InsertProcessingLog(procLog); err != nil {
		logs.Log().Error("Failed to insert processing log", zap.Error(err))
	}
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logs, err := database.GetAllProcessingLogs()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch logs: %v", err), http.StatusInternalServerError)
		return
	}

	type logView struct {
		ID             int64  `json:"id"`
		GoogleDriveURL string `json:"google_drive_url"`
		Name           string `json:"name"`
		Template       string `json:"template"`
		StartDate      string `json:"start_date"`
		EndDate        string `json:"end_date"`
		ExcludedRows   string `json:"excluded_rows"`
		UserAgent      string `json:"useragent"`
		OutputPath     string `json:"output_path"`
		Errors         string `json:"errors"`
		CreatedAt      string `json:"created_at"`
	}

	viewLogs := make([]logView, len(logs))
	for i, l := range logs {
		viewLogs[i] = logView{
			ID:             l.ID,
			GoogleDriveURL: l.GoogleDriveURL,
			Name:           l.Name,
			Template:       l.Template,
			StartDate:      l.StartDate,
			EndDate:        l.EndDate,
			ExcludedRows:   l.ExcludedRows,
			UserAgent:      l.UserAgent,
			OutputPath:     l.OutputPath,
			Errors:         l.Errors,
			CreatedAt:      l.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	tmpl, err := template.ParseFiles("templates/logs.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading template: %v", err), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, map[string]interface{}{"Logs": viewLogs}); err != nil {
		http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
		return
	}
}
