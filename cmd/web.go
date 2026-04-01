package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/models"
	"zion-english/internal/processor"
	"zion-english/internal/sheet"
	"zion-english/internal/utils"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func parseFloat64(n string) float64 {
	v, err := strconv.ParseFloat(n, 64)
	if err != nil {
		logs.Log().Error("parse float 64", zap.Error(err), zap.String("n", n))
	}
	return v
}

func parseInt64(n string) int64 {
	v, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		logs.Log().Error("parse int 64", zap.Error(err), zap.String("n", n))
	}
	return v
}

func getOrCreateUserAgentID(ctx context.Context, db database.Service, userAgentStr string) int64 {
	if userAgentStr == "" {
		return 0
	}

	uaInfo := utils.ParseUserAgent(userAgentStr)
	if uaInfo.Browser == "" {
		return 0
	}

	id, err := db.GetQueries().UpsertUserAgent(ctx, queries.UpsertUserAgentParams{
		UserAgent:      userAgentStr,
		Browser:        uaInfo.Browser,
		BrowserVersion: uaInfo.BrowserVersion,
		Os:             uaInfo.OS,
		Device:         uaInfo.Device,
	})
	if err != nil {
		return 0
	}
	return id
}

func validatePassword(p string) bool {
	return constants.ReLength.MatchString(p) &&
		constants.ReLower.MatchString(p) &&
		constants.ReUpper.MatchString(p) &&
		constants.ReDigit.MatchString(p) &&
		constants.ReSpecial.MatchString(p)
}

func HttpError(w http.ResponseWriter, msg string, code int) {
	http.SetCookie(w, &http.Cookie{
		Name:  "error_flash",
		Value: url.QueryEscape(msg),
		Path:  "/",
	})
	http.Error(w, msg, code)
}

func HttpRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Redirect", utils.URL(url))
	w.WriteHeader(http.StatusFound)
}

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
		cfg.BasePath = basePath

		publicMux := http.NewServeMux()
		publicMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, basePath, http.StatusFound)
		})
		publicMux.HandleFunc(basePath+"/process", handleProcessPage)
		publicMux.HandleFunc(basePath+"/download/processed", handleDownload)

		publicMux.HandleFunc(basePath+"/role", handleGetRole)
		publicMux.HandleFunc(basePath+"/refresh", handleRefreshPage)
		publicMux.HandleFunc(basePath+"/auth/login", handleLogin)
		publicMux.HandleFunc(basePath+"/auth/logout", handleLogout)
		publicMux.HandleFunc(basePath+"/teachers/register", handleTeacherRegister)

		publicMux.HandleFunc(basePath+"/api/teachers", handleGetTeachers)
		publicMux.HandleFunc(basePath+"/api/students", handleGetStudents)
		publicMux.HandleFunc(basePath+"/api/class-records", handleGetClassRecords)

		publicMux.Handle(
			basePath+"/static/",
			http.StripPrefix(basePath+"/static/", http.FileServer(http.Dir("static"))),
		)

		authMux := http.NewServeMux()
		authMux.HandleFunc(basePath, auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleHome))
		authMux.HandleFunc(basePath+"/", auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleHome))
		authMux.HandleFunc(basePath+"/students", auth.RequireRole(auth.RoleSuperuser)(handleStudents))
		authMux.HandleFunc(basePath+"/students/register", auth.RequireRole(auth.RoleSuperuser)(handleStudentRegister))
		authMux.HandleFunc(basePath+"/teachers", auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleTeachers))
		authMux.HandleFunc(basePath+"/classes/record", auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleClassRecord))
		authMux.HandleFunc(basePath+"/classes", auth.RequireRole(auth.RoleSuperuser)(handleClasses))
		authMux.HandleFunc(basePath+"/api/me/students", auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleGetMyStudents))

		authHandler := auth.Middleware(cfg, dbRO.GetQueries(), authMux)

		rootMux := http.NewServeMux()

		// public routes
		rootMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, basePath, http.StatusFound)
		})
		rootMux.Handle(basePath+"/download/processed", publicMux)
		rootMux.Handle(basePath+"/process", publicMux)
		rootMux.Handle(basePath+"/role", publicMux)
		rootMux.Handle(basePath+"/auth/", publicMux)
		rootMux.Handle(basePath+"/api/", publicMux)
		rootMux.Handle(basePath+"/static/", publicMux)
		rootMux.Handle(basePath+"/refresh", publicMux)
		rootMux.Handle(basePath+"/teachers/register", publicMux)

		// protected routes
		rootMux.Handle(basePath, authHandler)
		rootMux.Handle(basePath+"/students", authHandler)
		rootMux.Handle(basePath+"/students/", authHandler)
		rootMux.Handle(basePath+"/teachers", authHandler)
		rootMux.Handle(basePath+"/teachers/", authHandler)
		rootMux.Handle(basePath+"/classes", authHandler)
		rootMux.Handle(basePath+"/classes/", authHandler)
		rootMux.Handle(basePath+"/api/me/students", authHandler)

		handler := rootMux

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
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := frontend.Home(auth.GetRole(r.Context())).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleRefreshPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func handleProcessPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if err := frontend.Process().Render(r.Context(), w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if r.Method == http.MethodPost {
		handleProcess(w, r)
		return
	}
	HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logMessages = []string{}
	var req models.ProcessRequest
	var outputPath string

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	req.DriveURL = r.FormValue("driveUrl")
	req.Name = r.FormValue("name")
	req.Template = r.FormValue("templateSelect")
	req.StartDate = r.FormValue("startDate")
	req.EndDate = r.FormValue("endDate")
	req.NameCol = r.FormValue("nameCol")
	req.DurationCol = r.FormValue("durationCol")
	req.RateCol = r.FormValue("rateCol")
	req.StatusCol = r.FormValue("statusCol")
	req.ExcludedRows = r.FormValue("excludedRows")

	if err := validateProcessRequest(&req); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	addLog("Processing request for: " + req.Name)

	inputPath := filepath.Join("tmp", fmt.Sprintf("%s_input.csv", req.Name))
	if err := sheet.DownloadDriveSheet(req.DriveURL, inputPath); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	addLog("Downloaded file to: " + inputPath)

	parsedStartDate, err := processor.ParseDateString(req.StartDate)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	targetStartDate := *parsedStartDate

	parsedEndDate, err := processor.ParseDateString(req.EndDate)
	if err != nil {
		sendErrorLog(w, err.Error())
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
		sendErrorLog(w, err.Error())
		return
	}

	addLog(fmt.Sprintf("Processed %d records", len(records)))

	if len(records) == 0 {
		addLog("No record found...")
		sendErrorLog(w, "no record found")
		return
	}

	safename := utils.SanitizeFilename(req.Name)
	filename := fmt.Sprintf("%s_output_%s.xlsx", safename, utils.RandomString(8))
	outputPath = filepath.Join("tmp", filename)
	if err := processor.SaveRecords(records, outputPath, colIndices, req.Name); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	addLog("Saved output to: " + outputPath)

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

	processRecords := make([]frontend.ProcessRecord, len(responseRecords))
	for i, rec := range responseRecords {
		processRecords[i] = frontend.ProcessRecord{
			Name:     rec.Name,
			Date:     rec.Date,
			Duration: rec.Duration,
			Rate:     rec.Rate,
			Status:   rec.Status,
		}
	}

	if err := frontend.ProcessResponse(
		filename,
		req.DriveURL,
		rowRange,
		total,
		processRecords,
		logMessages,
	).Render(r.Context(), w); err != nil {
		logs.Log().Error("failed to render process response", zap.Error(err))
		sendErrorLog(w, err.Error())
	}
}

type FinalizeRequest struct {
	Name     string `form:"name"`
	DriveURL string `form:"driveUrl"`
}

func handleFinalize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendErrorResponseHTML(w, "Invalid request", http.StatusBadRequest)
		return
	}

	req := FinalizeRequest{
		Name:     r.FormValue("name"),
		DriveURL: r.FormValue("driveUrl"),
	}

	if req.Name == "" || req.DriveURL == "" {
		sendErrorResponseHTML(w, "Missing name or spreadsheet URL", http.StatusBadRequest)
		return
	}

	outputPath := filepath.Join("tmp", fmt.Sprintf("%s_output.xlsx", req.Name))
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		sendErrorResponseHTML(w, "Output file not found. Please process CSV first.", http.StatusNotFound)
		return
	}

	f, err := excelize.OpenFile(outputPath)
	if err != nil {
		sendErrorResponseHTML(w, fmt.Sprintf("Failed to open output file: %v", err), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	sheetName := "Sheet1"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		sendErrorResponseHTML(w, fmt.Sprintf("Failed to read sheet: %v", err), http.StatusInternalServerError)
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
			durationFloat = parseFloat64(durationMinutes)
		}

		rateFloat := 0.0
		if rate != "" {
			rateFloat = parseFloat64(rate)
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
			sendErrorResponseHTML(w, fmt.Sprintf("Failed to insert record: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<div class="log-success">✓ Successfully finalized %d records</div>`, rowNum-3)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorLog(w, "Method not allowed")
		return
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		sendErrorLog(w, "missing filename")
		return
	}

	const baseDir = "tmp"
	cleanPath := filepath.Clean(filename)
	fullPath := filepath.Join(baseDir, cleanPath)
	if !strings.HasPrefix(fullPath, filepath.Clean(baseDir)+string(os.PathSeparator)) {
		sendErrorLog(w, "invalid path")
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		sendErrorLog(w, "file not found")
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	http.ServeFile(w, r, fullPath)
}

func validateProcessRequest(req *models.ProcessRequest) error {
	if req.DriveURL == "" {
		return errors.New("google spreadsheet url is required")
	}
	if !strings.Contains(req.DriveURL, "docs.google.com") {
		return errors.New("invalid Google Spreadsheet URL")
	}

	// Validate name
	matched, _ := regexp.MatchString(`^[a-zA-Z -]+$`, req.Name)
	if !matched {
		return errors.New("name must contain only letters, dashes, and space")
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

func sendErrorResponseHTML(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `<div class="log-error">✗ %s</div>`, message)
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
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
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logs, err := database.GetAllProcessingLogs(dbRO)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to fetch logs: %v", err), http.StatusInternalServerError)
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
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	students, err := database.GetAllStudents(dbRO)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to fetch students: %v", err), http.StatusInternalServerError)
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
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logMessages = []string{}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	req := models.StudentRegisterRequest{
		Name:          r.FormValue("name"),
		Currency:      r.FormValue("currency"),
		Contact:       r.FormValue("contact"),
		RatePerClass:  parseFloat64(r.FormValue("ratePerClass")),
		ParentName:    r.FormValue("parentName"),
		AssignedColor: r.FormValue("assignedColor"),
		Status:        r.FormValue("status"),
		TeacherID:     parseInt64(r.FormValue("teacher")),
	}

	if err := validateStudentRequest(&req); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	addLog(fmt.Sprintf("Registering student: %s", req.Name))

	existingCount, err := dbRW.GetQueries().GetStudentByName(r.Context(), req.Name)
	if err != nil {
		sendErrorLog(w, fmt.Sprintf("Database error: %v", err))
		return
	}

	if existingCount > 0 {
		sendErrorLog(w, "A student with this name already exists")
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
		sendErrorLog(w, fmt.Sprintf("Failed to register student: %v", err))
		return
	}

	studentID, err := dbRW.GetQueries().GetStudentIDByName(r.Context(), req.Name)
	if err != nil {
		sendErrorLog(w, fmt.Sprintf("Failed to get student ID: %v", err))
		return
	}

	err = dbRW.GetQueries().InsertTeacherStudentM2M(r.Context(), queries.InsertTeacherStudentM2MParams{
		TeacherID: req.TeacherID,
		StudentID: studentID,
	})
	if err != nil {
		sendErrorLog(w, fmt.Sprintf("Failed to assign teacher: %v", err))
		return
	}
	addLog(fmt.Sprintf("Assigned student to teacher ID: %d", req.TeacherID))

	addLog(fmt.Sprintf("Successfully registered student: %s", req.Name))

	if _, err := fmt.Fprintf(w, fmt.Sprintf("Student '%s' registered successfully\n", req.Name)); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	for _, log := range logMessages {
		if _, err := fmt.Fprint(w, log+"\n"); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}
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

	if req.TeacherID == 0 {
		return errors.New("assigned teacher is required")
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
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teachers, err := database.GetAllTeachers(dbRO)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to fetch teachers: %v", err), http.StatusInternalServerError)
		return
	}

	viewTeachers := make([]frontend.TeacherItem, len(teachers))
	for i, t := range teachers {
		viewTeachers[i] = frontend.TeacherItem{
			ID:             strconv.FormatInt(t.ID, 10),
			Name:           t.Name,
			Birthdate:      t.Birthdate,
			Address:        t.Address,
			JoiningDate:    t.JoiningDate,
			MobileNumber:   t.MobileNumber,
			Email:          t.Email,
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
		if err := frontend.RegisterTeacher().Render(r.Context(), w); err != nil {
			HttpError(w, err.Error(), http.StatusMethodNotAllowed)
		}
		return
	}

	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logMessages = []string{}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	req := models.TeacherRegisterRequest{
		Name:           r.FormValue("name"),
		Birthdate:      r.FormValue("birthdate"),
		Address:        r.FormValue("address"),
		JoiningDate:    r.FormValue("joiningDate"),
		MobileNumber:   r.FormValue("mobileNumber"),
		Email:          r.FormValue("email"),
		Certifications: r.FormValue("certifications"),
		AssignedColor:  r.FormValue("assignedColor"),
		RatePerClass:   parseFloat64(r.FormValue("ratePerClass")),
		Currency:       r.FormValue("currency"),
		DriveUrl:       r.FormValue("driveUrl"),
		Sex:            r.FormValue("sex"),
		Password:       r.FormValue("password"),
		RetypePassword: r.FormValue("retypePassword"),
	}

	if err := validateTeacherRequest(&req); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	addLog(fmt.Sprintf("Registering teacher: %s", req.Name))

	existingCount, err := dbRW.GetQueries().GetTeacherByName(r.Context(), req.Name)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if existingCount > 0 {
		sendErrorLog(w, "A teacher with this name already exists")
		return
	}

	if req.Password != req.RetypePassword {
		sendErrorLog(w, "Passwords must be the same")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		sendErrorLog(w, "Failed to hash password")
		return
	}

	err = dbRW.GetQueries().InsertTeacher(r.Context(), queries.InsertTeacherParams{
		Name:           req.Name,
		Birthdate:      req.Birthdate,
		Address:        req.Address,
		JoiningDate:    req.JoiningDate,
		MobileNumber:   req.MobileNumber,
		Email:          req.Email,
		Certifications: sql.NullString{String: req.Certifications, Valid: req.Certifications != ""},
		AssignedColor:  req.AssignedColor,
		RatePerClass:   req.RatePerClass,
		Currency:       req.Currency,
		DriveUrl:       req.DriveUrl,
		Sex:            sql.NullString{String: req.Sex, Valid: req.Sex != ""},
		Password:       string(hashedPassword),
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	addLog(fmt.Sprintf("Successfully registered teacher: %s", req.Name))

	response := models.TeacherRegisterResponse{
		Success: true,
		Message: fmt.Sprintf("Teacher '%s' registered successfully", req.Name),
		Logs:    logMessages,
	}

	if _, err := fmt.Fprint(w, response.Message+"\n"); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	for _, log := range logMessages {
		if _, err := fmt.Fprint(w, log+"\n"); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}
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
		return errors.New("spreadsheet URL is required")
	}

	validSex := map[string]bool{"M": true, "F": true}
	if req.Sex != "" && !validSex[req.Sex] {
		return errors.New("invalid sex. Must be M or F")
	}

	if req.Password == "" {
		return errors.New("password is required")
	}

	if !validatePassword(req.Password) {
		return errors.New("password must be 8-32 characters with uppercase, lowercase, number, and symbol (!@#$%^&*)")
	}

	if req.Password != req.RetypePassword {
		return errors.New("passwords do not match")
	}

	return nil
}

func sendErrorLog(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusBadRequest)
	if _, err := fmt.Fprint(w, message+"\n"); err != nil {
		logs.Log().Error("error response", zap.String("message", message), zap.Error(err))
		return
	}
}

func handleGetTeachers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teachers, err := dbRO.GetQueries().GetAllTeachers(r.Context())
	if err != nil {
		HttpError(w, "Failed to fetch teachers", http.StatusInternalServerError)
		return
	}

	var teacherResponses []models.TeacherAPIResponse
	for _, t := range teachers {
		teacherResponses = append(teacherResponses, models.TeacherAPIResponse{
			ID:           t.ID,
			Name:         t.Name,
			DriveUrl:     t.DriveUrl,
			RatePerClass: t.RatePerClass,
		})
	}
	if err := frontend.TeacherOptions(teacherResponses).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Handle Login]"
	ctx := r.Context()
	if r.Method == http.MethodGet {
		if err := frontend.Login().Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	if err := r.ParseForm(); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		fmt.Fprint(w, "Invalid email or password")
		return
	}

	if err := auth.Login(w, r, conf.Conf(), dbRO.GetQueries(), email, password); err != nil {
		fmt.Fprint(w, "Invalid email or password")
		return
	}

	if ua := r.UserAgent(); ua != "" {
		useragentID := getOrCreateUserAgentID(context.Background(), dbRW, ua)
		if _, err := dbRW.GetQueries().CreateAccess(context.Background(), queries.CreateAccessParams{
			TeacherID:   auth.GetUser(ctx).ID,
			UseragentID: useragentID,
		}); err != nil {
			logs.Log().Warn(logtag, zap.Error(err))
		}
	}

	HttpRedirect(w, "/")
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Handle Logout]"
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	auth.Logout(w)
	HttpRedirect(w, "/")
}

func handleGetRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	greetings := "Welcome!"
	role := auth.GetRole(ctx)
	switch role {
	case auth.RoleSuperuser:
		greetings = "Welcome, superuser!"
	case auth.RoleTeacher:
		user := auth.GetUser(ctx)
		greetings = fmt.Sprintf("Welcome, teacher %s!", user.Name)
	}

	if _, err := fmt.Fprint(w, greetings); err != nil {
		logs.Log().Error("handle get role", zap.Error(err))
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleGetStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	students, err := dbRO.GetQueries().GetActiveStudents(r.Context())
	if err != nil {
		HttpError(w, "Failed to fetch students", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	var studentResponses []models.StudentAPIResponse
	for _, s := range students {
		studentResponses = append(studentResponses, models.StudentAPIResponse{
			ID:           s.ID,
			Name:         s.Name,
			Currency:     s.Currency,
			RatePerClass: s.RatePerClass,
		})
	}

	if err := frontend.StudentOptions(studentResponses).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleGetMyStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user.ID == 0 {
		HttpError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	students, err := dbRO.GetQueries().GetStudentsByTeacherID(r.Context(), user.ID)
	if err != nil {
		HttpError(w, "Failed to fetch students", http.StatusInternalServerError)
		return
	}

	var studentResponses []models.StudentAPIResponse
	for _, s := range students {
		studentResponses = append(studentResponses, models.StudentAPIResponse{
			ID:           s.ID,
			Name:         s.Name,
			Currency:     s.Currency,
			RatePerClass: s.RatePerClass,
		})
	}
	fmt.Println(1111, students, studentResponses)

	if err := frontend.StudentOptions(studentResponses).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleGetClassRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teacherIDStr := r.URL.Query().Get("teacherId")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	if teacherIDStr == "" || startDate == "" || endDate == "" {
		HttpError(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	teacherID, err := strconv.ParseInt(teacherIDStr, 10, 64)
	if err != nil {
		HttpError(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	records, err := dbRO.GetQueries().GetClassRecordsByTeacherAndDateRange(r.Context(), queries.GetClassRecordsByTeacherAndDateRangeParams{
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
	})
	if err != nil {
		HttpError(w, "Failed to fetch class records", http.StatusInternalServerError)
		return
	}

	totalRate, err := dbRO.GetQueries().GetTotalRateByTeacherAndDateRange(r.Context(), queries.GetTotalRateByTeacherAndDateRangeParams{
		TeacherID: teacherID,
		Date:      startDate,
		Date_2:    endDate,
	})
	if err != nil {
		HttpError(w, "Failed to fetch total rate", http.StatusInternalServerError)
		return
	}

	var response []models.ClassRecordView
	for _, cr := range records {
		response = append(response, models.ClassRecordView{
			ID:              cr.ID,
			StudentID:       cr.StudentID,
			TeacherID:       cr.TeacherID,
			StudentName:     cr.StudentName,
			TeacherName:     cr.TeacherName,
			Date:            cr.Date,
			DurationMinutes: cr.DurationMinutes,
			Rate:            cr.Rate,
			Currency:        cr.Currency,
			Status:          cr.Status,
			Reason:          cr.Reason.String,
			CreatedAt:       cr.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"records":   response,
		"totalRate": totalRate,
	})
}

func handleClassRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		frontend.RecordClass().Render(r.Context(), w)
		return
	}

	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if user.ID == 0 {
		HttpError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logMessages = []string{}
	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	req := models.ClassRecordRequest{
		StudentID:       parseInt64(r.FormValue("student")),
		TeacherID:       user.ID,
		Date:            r.FormValue("date"),
		DurationMinutes: parseInt64(r.FormValue("duration")),
		Rate:            parseFloat64(r.FormValue("rate")),
		Currency:        r.FormValue("currency"),
		Status:          r.FormValue("status"),
		Reason:          r.FormValue("reason"),
	}

	if err := validateClassRecordRequest(&req); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	err := dbRW.GetQueries().InsertClassRecord(r.Context(), queries.InsertClassRecordParams{
		StudentID:       req.StudentID,
		TeacherID:       req.TeacherID,
		Date:            req.Date,
		DurationMinutes: req.DurationMinutes,
		Rate:            req.Rate,
		Currency:        req.Currency,
		Status:          req.Status,
		Reason:          sql.NullString{String: req.Reason, Valid: req.Reason != ""},
		RecordedByRole:  string(user.Role),
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if _, err := fmt.Fprint(w, "Class recorded successfully!\n"); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	addLog("Class recorded successfully")
	for _, log := range logMessages {
		if _, err := fmt.Fprint(w, log+"\n"); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}
}

func handleClasses(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		frontend.Classes().Render(r.Context(), w)
		return
	}

	HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func validateClassRecordRequest(req *models.ClassRecordRequest) error {
	if req.StudentID == 0 {
		return errors.New("student is required")
	}
	if req.TeacherID == 0 {
		return errors.New("teacher is required")
	}
	if req.Date == "" {
		return errors.New("date is required")
	}
	if req.DurationMinutes == 0 {
		return errors.New("duration is required")
	}
	if req.Rate == 0 {
		return errors.New("rate is required")
	}
	if req.Currency == "" {
		return errors.New("currency is required")
	}
	validStatuses := map[string]bool{"conducted": true, "cancelled": true, "rescheduled": true}
	if !validStatuses[req.Status] {
		return errors.New("invalid status")
	}
	return nil
}

func sendLogs(w http.ResponseWriter) {
	for _, log := range logMessages {
		if _, err := fmt.Fprint(w, log+"\n"); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}
}
