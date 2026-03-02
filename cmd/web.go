package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/conf"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/models"
	"zion-english/internal/processor"
	"zion-english/internal/sheet"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type WebFlags struct {
	port    string
	baseURL string
	https   bool
	address string
}

var webFlags WebFlags

var dbRW database.Service
var dbRO database.Service

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
		cfg := conf.Conf()
		if err := os.MkdirAll("tmp", 0755); err != nil {
			panic(err)
		}

		if err := database.Init("data/zion.db"); err != nil {
			panic(fmt.Sprintf("Failed to initialize database: %v", err))
		}
		defer database.Close()

		dbRW = database.New(database.DB_MODE_RW)
		dbRO = database.New(database.DB_MODE_RO)

		basePath := "/" + strings.TrimPrefix(webFlags.baseURL, "/")

		mux := http.NewServeMux()
		mux.HandleFunc(basePath, handleHome)
		mux.HandleFunc("/", handleHome)
		mux.HandleFunc(basePath+"/process", handleProcessPage)
		mux.HandleFunc(basePath+"/finalize", handleFinalize)
		mux.HandleFunc(basePath+"/download/", handleDownload)
		mux.HandleFunc(basePath+"/api/teachers", handleGetTeachers)
		mux.Handle(basePath+"/static/", http.StripPrefix(basePath+"/static/", http.FileServer(http.Dir("static"))))

		authMux := http.NewServeMux()
		authMux.HandleFunc(basePath+"/logs", handleLogs)
		authMux.HandleFunc(basePath+"/students", handleStudents)
		authMux.HandleFunc(basePath+"/students/register", handleStudentRegister)
		authMux.HandleFunc(basePath+"/teachers", handleTeachers)
		authMux.HandleFunc(basePath+"/teachers/register", handleTeacherRegister)
		authHandler := auth.Middleware(cfg.AdminUsername, cfg.AdminPassword, authMux)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, basePath+"/logs") || strings.HasPrefix(r.URL.Path, basePath+"/students") || strings.HasPrefix(r.URL.Path, basePath+"/teachers") {
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

var logMessages []string

func addLog(msg string) {
	logMessages = append(logMessages, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
	logs.Log().Info(msg)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	frontend.Home().Render(r.Context(), w)
}

func handleProcessPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		frontend.Index().Render(r.Context(), w)
		return
	}
	if r.Method == http.MethodPost {
		handleProcess(w, r)
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logMessages = []string{}
	var req models.ProcessRequest
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

	responseRecords := make([]models.RecordRow, len(records))
	for i, rec := range records {
		responseRecords[i] = models.RecordRow{
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

	logToDB(dbRW, &req, r.UserAgent(), outputPath, "")

	response := models.ProcessResponse{
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

type FinalizeRequest struct {
	Name     string `json:"name"`
	DriveURL string `json:"driveUrl"`
}

func handleFinalize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FinalizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.DriveURL == "" {
		http.Error(w, "Missing name or drive URL", http.StatusBadRequest)
		return
	}

	outputPath := filepath.Join("tmp", fmt.Sprintf("%s_output.xlsx", req.Name))
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		http.Error(w, "Output file not found. Please process CSV first.", http.StatusNotFound)
		return
	}

	f, err := excelize.OpenFile(outputPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to open output file: %v", err), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	sheetName := "Sheet1"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read sheet: %v", err), http.StatusInternalServerError)
		return
	}

	rowNum := 0
	for _, row := range rows {
		rowNum++
		if rowNum < 4 {
			continue
		}
		if len(row) < 4 {
			continue
		}

		studentName := row[0]
		date := row[1]
		if studentName == "" || date == "" {
			continue
		}

		durationMinutes := row[2]
		rate := row[3]
		status := row[len(row)-1]

		durationFloat := 0.0
		if durationMinutes != "" {
			durationFloat, _ = strconv.ParseFloat(durationMinutes, 64)
		}

		rateFloat := 0.0
		if rate != "" {
			rateFloat, _ = strconv.ParseFloat(rate, 64)
		}

		err = dbRW.GetQueries().InsertRecord(r.Context(), queries.InsertRecordParams{
			GoogleDriveUrl:  req.DriveURL,
			StudentName:     studentName,
			Date:            date,
			DurationMinutes: durationFloat,
			Rate:            rateFloat,
			Status:          status,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert record: %v", err), http.StatusInternalServerError)
			return
		}
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully finalized %d records", rowNum-3),
		"count":   rowNum - 3,
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

func validateRequest(req *models.ProcessRequest) error {
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

func sendErrorResponse(w http.ResponseWriter, message string, req *models.ProcessRequest, userAgent, outputPath, errMsg string) {
	if req != nil {
		logToDB(dbRW, req, userAgent, outputPath, errMsg)
	}

	response := models.ProcessResponse{
		Success:  false,
		Message:  message,
		Logs:     logMessages,
		Records:  []models.RecordRow{},
		Total:    0,
		RowRange: "",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func logToDB(dbService database.Service, req *models.ProcessRequest, userAgent, outputPath, errMsg string) {
	if req == nil {
		return
	}

	procLog := &database.ProcessingLog{
		GoogleDriveUrl: req.DriveURL,
		Name:           req.Name,
		Template:       sql.NullString{String: req.Template, Valid: req.Template != ""},
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		ExcludedRows:   sql.NullString{String: req.ExcludedRows, Valid: req.ExcludedRows != ""},
		Useragent:      sql.NullString{String: userAgent, Valid: userAgent != ""},
		OutputPath:     sql.NullString{String: outputPath, Valid: outputPath != ""},
		Errors:         sql.NullString{String: errMsg, Valid: errMsg != ""},
	}

	if err := database.InsertProcessingLog(dbService, procLog); err != nil {
		logs.Log().Error("Failed to insert processing log", zap.Error(err))
	}
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logs, err := database.GetAllProcessingLogs(dbRO)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch logs: %v", err), http.StatusInternalServerError)
		return
	}

	viewLogs := make([]frontend.LogItem, len(logs))
	for i, l := range logs {
		viewLogs[i] = frontend.LogItem{
			ID:             strconv.FormatInt(l.ID, 10),
			GoogleDriveURL: l.GoogleDriveUrl,
			Name:           l.Name,
			Template:       l.Template.String,
			StartDate:      l.StartDate,
			EndDate:        l.EndDate,
			ExcludedRows:   l.ExcludedRows.String,
			UserAgent:      l.Useragent.String,
			OutputPath:     l.OutputPath.String,
			Errors:         l.Errors.String,
			CreatedAt:      l.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		}
	}

	w.Header().Set("Content-Type", "text/html")
	frontend.Logs(frontend.LogData{Logs: viewLogs}).Render(r.Context(), w)
}

func handleStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	students, err := database.GetAllStudents(dbRO)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch students: %v", err), http.StatusInternalServerError)
		return
	}

	viewStudents := make([]frontend.StudentItem, len(students))
	for i, s := range students {
		viewStudents[i] = frontend.StudentItem{
			ID:            strconv.FormatInt(s.ID, 10),
			Name:          s.Name,
			Currency:      s.Currency,
			Contact:       s.Contact.String,
			RatePerClass:  s.RatePerClass,
			ParentName:    s.ParentName.String,
			AssignedColor: s.AssignedColor,
			Status:        s.Status,
			CreatedAt:     s.CreatedAt.Time.Format("2006-01-02 15:04:05"),
			UpdatedAt:     s.UpdatedAt.Time.Format("2006-01-02 15:04:05"),
		}
	}

	w.Header().Set("Content-Type", "text/html")
	frontend.Students(frontend.StudentData{Students: viewStudents}).Render(r.Context(), w)
}

func handleStudentRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		frontend.RegisterStudent().Render(r.Context(), w)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logMessages = []string{}
	var req models.StudentRegisterRequest
	var errMsg string

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errMsg = fmt.Sprintf("Invalid request: %v", err)
		sendStudentErrorResponse(w, errMsg)
		return
	}

	if err := validateStudentRequest(&req); err != nil {
		errMsg = err.Error()
		sendStudentErrorResponse(w, errMsg)
		return
	}

	addLog(fmt.Sprintf("Registering student: %s", req.Name))

	existingCount, err := dbRW.GetQueries().GetStudentByName(r.Context(), req.Name)
	if err != nil {
		errMsg = fmt.Sprintf("Database error: %v", err)
		sendStudentErrorResponse(w, errMsg)
		return
	}

	if existingCount > 0 {
		errMsg = "A student with this name already exists"
		sendStudentErrorResponse(w, errMsg)
		return
	}

	err = dbRW.GetQueries().InsertStudent(r.Context(), queries.InsertStudentParams{
		Name:          req.Name,
		Currency:      req.Currency,
		Contact:       sql.NullString{String: req.Contact, Valid: req.Contact != ""},
		RatePerClass:  req.RatePerClass,
		ParentName:    sql.NullString{String: req.ParentName, Valid: req.ParentName != ""},
		AssignedColor: req.AssignedColor,
		Status:        req.Status,
	})
	if err != nil {
		errMsg = fmt.Sprintf("Failed to register student: %v", err)
		sendStudentErrorResponse(w, errMsg)
		return
	}

	addLog(fmt.Sprintf("Successfully registered student: %s", req.Name))

	response := models.StudentRegisterResponse{
		Success: true,
		Message: fmt.Sprintf("Student '%s' registered successfully", req.Name),
		Logs:    logMessages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func validateStudentRequest(req *models.StudentRegisterRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}

	validCurrencies := map[string]bool{"KRW": true, "CAD": true, "YEN": true, "PHP": true}
	if !validCurrencies[req.Currency] {
		return errors.New("invalid currency. Must be KRW, CAD, YEN, or PHP")
	}

	if req.RatePerClass < 0 {
		return errors.New("rate per class cannot be negative")
	}

	if req.AssignedColor == "" {
		return errors.New("assigned color is required")
	}

	validStatuses := map[string]bool{"active": true, "inactive": true}
	if !validStatuses[req.Status] {
		return errors.New("invalid status. Must be active or inactive")
	}

	return nil
}

func sendStudentErrorResponse(w http.ResponseWriter, message string) {
	response := models.StudentRegisterResponse{
		Success: false,
		Message: message,
		Logs:    logMessages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func handleTeachers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teachers, err := database.GetAllTeachers(dbRO)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch teachers: %v", err), http.StatusInternalServerError)
		return
	}

	viewTeachers := make([]frontend.TeacherItem, len(teachers))
	for i, t := range teachers {
		viewTeachers[i] = frontend.TeacherItem{
			ID:             strconv.FormatInt(t.ID, 10),
			Name:           t.Name,
			Birthdate:      t.Birthdate.String,
			Address:        t.Address.String,
			JoiningDate:    t.JoiningDate,
			MobileNumber:   t.MobileNumber.String,
			Email:          t.Email.String,
			Certifications: t.Certifications.String,
			AssignedColor:  t.AssignedColor,
			RatePerClass:   t.RatePerClass,
			Currency:       t.Currency,
			DriveUrl:       t.DriveUrl,
			Sex:            t.Sex.String,
			CreatedAt:      t.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		}
	}

	w.Header().Set("Content-Type", "text/html")
	frontend.Teachers(frontend.TeacherData{Teachers: viewTeachers}).Render(r.Context(), w)
}

func handleTeacherRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		frontend.RegisterTeacher().Render(r.Context(), w)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logMessages = []string{}
	var req models.TeacherRegisterRequest
	var errMsg string

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errMsg = fmt.Sprintf("Invalid request: %v", err)
		sendTeacherErrorResponse(w, errMsg)
		return
	}

	if err := validateTeacherRequest(&req); err != nil {
		errMsg = err.Error()
		sendTeacherErrorResponse(w, errMsg)
		return
	}

	addLog(fmt.Sprintf("Registering teacher: %s", req.Name))

	existingCount, err := dbRW.GetQueries().GetTeacherByName(r.Context(), req.Name)
	if err != nil {
		errMsg = fmt.Sprintf("Database error: %v", err)
		sendTeacherErrorResponse(w, errMsg)
		return
	}

	if existingCount > 0 {
		errMsg = "A teacher with this name already exists"
		sendTeacherErrorResponse(w, errMsg)
		return
	}

	err = dbRW.GetQueries().InsertTeacher(r.Context(), queries.InsertTeacherParams{
		Name:           req.Name,
		Birthdate:      sql.NullString{String: req.Birthdate, Valid: req.Birthdate != ""},
		Address:        sql.NullString{String: req.Address, Valid: req.Address != ""},
		JoiningDate:    req.JoiningDate,
		MobileNumber:   sql.NullString{String: req.MobileNumber, Valid: req.MobileNumber != ""},
		Email:          sql.NullString{String: req.Email, Valid: req.Email != ""},
		Certifications: sql.NullString{String: req.Certifications, Valid: req.Certifications != ""},
		AssignedColor:  req.AssignedColor,
		RatePerClass:   req.RatePerClass,
		Currency:       req.Currency,
		DriveUrl:       req.DriveUrl,
		Sex:            sql.NullString{String: req.Sex, Valid: req.Sex != ""},
	})
	if err != nil {
		errMsg = fmt.Sprintf("Failed to register teacher: %v", err)
		sendTeacherErrorResponse(w, errMsg)
		return
	}

	addLog(fmt.Sprintf("Successfully registered teacher: %s", req.Name))

	response := models.TeacherRegisterResponse{
		Success: true,
		Message: fmt.Sprintf("Teacher '%s' registered successfully", req.Name),
		Logs:    logMessages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func validateTeacherRequest(req *models.TeacherRegisterRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}

	if req.JoiningDate == "" {
		return errors.New("joining date is required")
	}

	validCurrencies := map[string]bool{"KRW": true, "CAD": true, "YEN": true, "PHP": true}
	if !validCurrencies[req.Currency] {
		return errors.New("invalid currency. Must be KRW, CAD, YEN, or PHP")
	}

	if req.RatePerClass < 0 {
		return errors.New("rate per class cannot be negative")
	}

	if req.AssignedColor == "" {
		return errors.New("assigned color is required")
	}

	if req.DriveUrl == "" {
		return errors.New("drive URL is required")
	}

	validSex := map[string]bool{"M": true, "F": true}
	if req.Sex != "" && !validSex[req.Sex] {
		return errors.New("invalid sex. Must be M or F")
	}

	return nil
}

func sendTeacherErrorResponse(w http.ResponseWriter, message string) {
	response := models.TeacherRegisterResponse{
		Success: false,
		Message: message,
		Logs:    logMessages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func handleGetTeachers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teachers, err := dbRO.GetQueries().GetAllTeachers(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch teachers", http.StatusInternalServerError)
		return
	}

	var response []models.TeacherAPIResponse
	for _, t := range teachers {
		response = append(response, models.TeacherAPIResponse{
			ID:       t.ID,
			Name:     t.Name,
			DriveUrl: t.DriveUrl,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
