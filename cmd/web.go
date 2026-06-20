package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/mail"
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

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func requireFloat64(n string) (float64, error) {
	if n == "" {
		return 0, errors.New("missing numeric value")
	}
	v, err := strconv.ParseFloat(n, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", n)
	}
	return v, nil
}

func requireInt64(n string) (int64, error) {
	if n == "" {
		return 0, errors.New("missing integer value")
	}
	v, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %s", n)
	}
	return v, nil
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if ip, _, ok := strings.Cut(forwarded, ","); ok {
			return strings.TrimSpace(ip)
		}
		return strings.TrimSpace(forwarded)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
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

func setSuccessFlash(w http.ResponseWriter, msg string) {
	cfg := conf.Conf()
	cookie := &http.Cookie{
		Name:     "success_flash",
		Value:    url.QueryEscape(msg),
		Path:     cfg.BasePath,
		SameSite: http.SameSiteStrictMode,
	}
	if cfg.IsProd() {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
}

func readFlashCookie(w http.ResponseWriter, r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil || cookie.Value == "" {
		return ""
	}

	msg, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		msg = cookie.Value
	}

	cfg := conf.Conf()
	clearCookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     cfg.BasePath,
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
	}
	if cfg.IsProd() {
		clearCookie.Secure = true
	}
	http.SetCookie(w, clearCookie)

	return msg
}

func HttpError(w http.ResponseWriter, msg string, code int) {
	cfg := conf.Conf()
	cookie := &http.Cookie{
		Name:     "error_flash",
		Value:    url.QueryEscape(msg),
		Path:     cfg.BasePath,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	if cfg.IsProd() {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
	http.Error(w, msg, code)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")
		next.ServeHTTP(w, r)
	})
}

func HttpRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Redirect", utils.URL(url))
	w.WriteHeader(http.StatusFound)
}

func redirectToPortal(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") == "true" {
		HttpRedirect(w, "/")
		return
	}
	http.Redirect(w, r, utils.URL("/"), http.StatusFound)
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
		publicMux.HandleFunc(basePath+"/auth/login", handleLogin)
		publicMux.HandleFunc(basePath+"/auth/logout", handleLogout)
		publicMux.HandleFunc(basePath+"/auth/forgot-password", handleForgotPassword)
		publicMux.HandleFunc(basePath+"/auth/forgot-password/reset", handleResetPassword)
		publicMux.HandleFunc(basePath+"/teachers/register", handleTeacherRegister)

		publicMux.Handle(
			basePath+"/static/",
			http.StripPrefix(basePath+"/static/", http.FileServer(http.Dir("static"))),
		)

		authMux := http.NewServeMux()
		authMux.HandleFunc(basePath, auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleHome))
		authMux.HandleFunc(basePath+"/", auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleHome))
		authMux.HandleFunc(basePath+"/students", auth.RequireRole(auth.RoleSuperuser)(handleStudents))
		authMux.HandleFunc(basePath+"/students/register", auth.RequireRole(auth.RoleSuperuser)(handleStudentRegister))
		authMux.HandleFunc(basePath+"/teachers", auth.RequireRole(auth.RoleSuperuser)(handleTeachers))
		authMux.HandleFunc(basePath+"/teachers/approve", auth.RequireRole(auth.RoleSuperuser)(handleTeacherApprove))
		authMux.HandleFunc(basePath+"/teachers/unapprove", auth.RequireRole(auth.RoleSuperuser)(handleTeacherUnapprove))
		authMux.HandleFunc(basePath+"/classes/record", auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleClassRecord))
		authMux.HandleFunc(basePath+"/classes", auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleClasses))
		authMux.HandleFunc(basePath+"/logs", auth.RequireRole(auth.RoleSuperuser)(handleSystemLogs))
		authMux.HandleFunc(basePath+"/process-logs", auth.RequireRole(auth.RoleSuperuser)(handleLogs))
		authMux.HandleFunc(basePath+"/process", auth.RequireRole(auth.RoleSuperuser)(handleProcessPage))
		authMux.HandleFunc(basePath+"/download/processed", auth.RequireRole(auth.RoleSuperuser)(handleDownload))
		authMux.HandleFunc(basePath+"/role", auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleGetRole))
		authMux.HandleFunc(basePath+"/refresh", auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleRefreshPage))
		authMux.HandleFunc(basePath+"/api/teachers", auth.RequireRole(auth.RoleSuperuser)(handleGetTeachers))
		authMux.HandleFunc(basePath+"/api/students", auth.RequireRole(auth.RoleSuperuser)(handleGetStudents))
		authMux.HandleFunc(basePath+"/api/me/students", auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleGetMyStudents))
		authMux.HandleFunc(basePath+"/api/class-records", auth.RequireRole(auth.RoleSuperuser, auth.RoleTeacher)(handleGetClassRecords))

		authHandler := auth.Middleware(cfg, dbRO.GetQueries(), authMux)

		rootMux := http.NewServeMux()

		// public routes
		rootMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, basePath, http.StatusFound)
		})
		rootMux.Handle(basePath+"/auth/", publicMux)
		rootMux.Handle(basePath+"/static/", publicMux)
		rootMux.Handle(basePath+"/teachers/register", publicMux)

		// protected routes
		rootMux.Handle(basePath, authHandler)
		rootMux.Handle(basePath+"/logs", authHandler)
		rootMux.Handle(basePath+"/process-logs", authHandler)
		rootMux.Handle(basePath+"/process", authHandler)
		rootMux.Handle(basePath+"/download/processed", authHandler)
		rootMux.Handle(basePath+"/role", authHandler)
		rootMux.Handle(basePath+"/refresh", authHandler)
		rootMux.Handle(basePath+"/students", authHandler)
		rootMux.Handle(basePath+"/students/", authHandler)
		rootMux.Handle(basePath+"/teachers", authHandler)
		rootMux.Handle(basePath+"/teachers/approve", authHandler)
		rootMux.Handle(basePath+"/teachers/unapprove", authHandler)
		rootMux.Handle(basePath+"/classes", authHandler)
		rootMux.Handle(basePath+"/classes/", authHandler)
		rootMux.Handle(basePath+"/api/teachers", authHandler)
		rootMux.Handle(basePath+"/api/students", authHandler)
		rootMux.Handle(basePath+"/api/class-records", authHandler)
		rootMux.Handle(basePath+"/api/me/students", authHandler)

		handler := securityHeaders(rootMux)

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

type requestLogs struct {
	messages []string
}

func (l *requestLogs) add(msg string) {
	entry := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
	l.messages = append(l.messages, entry)
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
	ctx := r.Context()
	rl := &requestLogs{}
	var req models.ProcessRequest
	var outputPath string

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	req.TeacherID = r.FormValue("teacherID")
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

	rl.add("Processing request for: " + req.Name)

	inputPath := filepath.Join("tmp", fmt.Sprintf("%s_input.csv", req.Name))
	if err := sheet.DownloadDriveSheet(req.DriveURL, inputPath); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	rl.add("Downloaded file to: " + inputPath)

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

	rl.add(fmt.Sprintf("Processed %d records", len(records)))

	if len(records) == 0 {
		rl.add("No record found...")
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

	rl.add("Saved output to: " + outputPath)

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

	teacherID, err := requireInt64(req.TeacherID)
	if err != nil {
		sendErrorLog(w, "invalid teacher ID")
		return
	}

	updateTeacherErr := dbRW.GetQueries().UpdateTeacherTemplate(ctx, queries.UpdateTeacherTemplateParams{
		ID:       teacherID,
		Template: sql.NullString{String: req.Template, Valid: true},
	})
	if updateTeacherErr != nil {
		logs.Log().Error("update teacher template", zap.Error(updateTeacherErr))
	} else {
		user := auth.GetUser(ctx)
		var createdBy sql.NullInt64
		if user.ID > 0 {
			createdBy = sql.NullInt64{Int64: user.ID, Valid: true}
		}
		if err := dbRW.GetQueries().InsertLog(ctx, queries.InsertLogParams{
			Module:        "process",
			Message:       fmt.Sprintf("update teacher '%s' template with '%s'", req.Name, req.Template),
			CreatedBy:     createdBy,
			CreatedByName: sql.NullString{String: user.Name, Valid: user.Name != ""},
		}); err != nil {
			logs.Log().Info("system logs", zap.Error(err))
		}
	}

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
		rl.messages,
	).Render(ctx, w); err != nil {
		logs.Log().Error("failed to render process response", zap.Error(err))
		sendErrorLog(w, err.Error())
	}
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
	if _, err := utils.DriveURLToExportURL(req.DriveURL, "csv"); err != nil {
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

	viewLogs := make([]frontend.ProcessingLogItem, len(logs))
	for i, l := range logs {
		viewLogs[i] = frontend.ProcessingLogItem{
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
	frontend.ProcessingLogs(frontend.ProcessingLogData{Logs: viewLogs}).Render(r.Context(), w)
}

func handleSystemLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logs, err := database.GetAllSystemLogs(dbRO)
	if err != nil {
		HttpError(w, fmt.Sprintf("Failed to fetch logs: %v", err), http.StatusInternalServerError)
		return
	}

	viewLogs := make([]frontend.SystemLogItem, len(logs))
	for i, l := range logs {
		viewLogs[i] = frontend.SystemLogItem{
			ID:        strconv.FormatInt(l.ID, 10),
			Module:    l.Module,
			Message:   l.Message,
			CreatedBy: l.CreatedByName,
			CreatedAt: l.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "text/html")
	frontend.SystemLogs(frontend.SystemLogData{Logs: viewLogs}).Render(r.Context(), w)
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

	rl := &requestLogs{}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	ratePerClass, err := requireFloat64(r.FormValue("ratePerClass"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	teacherID, err := requireInt64(r.FormValue("teacher"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	req := models.StudentRegisterRequest{
		Name:          r.FormValue("name"),
		Currency:      r.FormValue("currency"),
		Contact:       r.FormValue("contact"),
		RatePerClass:  ratePerClass,
		ParentName:    r.FormValue("parentName"),
		AssignedColor: r.FormValue("assignedColor"),
		Status:        r.FormValue("status"),
		TeacherID:     teacherID,
	}

	if err := validateStudentRequest(&req); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	rl.add(fmt.Sprintf("Registering student: %s", req.Name))

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
	rl.add(fmt.Sprintf("Assigned student to teacher ID: %d", req.TeacherID))

	rl.add(fmt.Sprintf("Successfully registered student: %s", req.Name))

	if _, err := fmt.Fprintf(w, "Student '%s' registered successfully\n", req.Name); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	for _, log := range rl.messages {
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
			Status:         t.Status,
			CreatedAt:      t.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		}
	}

	w.Header().Set("Content-Type", "text/html")
	frontend.Teachers(frontend.TeacherData{Teachers: viewTeachers}).Render(r.Context(), w)
}

func handleTeacherRegister(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := auth.UserFromRequest(r, conf.Conf())
	var role auth.Role
	if loggedIn && auth.SessionUserValid(r.Context(), dbRO.GetQueries(), user) {
		role = user.Role
		if role == auth.RoleTeacher {
			HttpError(w, "Access denied", http.StatusForbidden)
			return
		}
	}

	isSuperuser := role == auth.RoleSuperuser

	if r.Method == http.MethodGet {
		if err := frontend.RegisterTeacher(isSuperuser, role).Render(r.Context(), w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rl := &requestLogs{}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	ratePerClass, err := requireFloat64(r.FormValue("ratePerClass"))
	if err != nil {
		sendErrorLog(w, err.Error())
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
		RatePerClass:   ratePerClass,
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

	rl.add(fmt.Sprintf("Registering teacher: %s", req.Name))

	existingCount, err := dbRW.GetQueries().GetTeacherByName(r.Context(), req.Name)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if existingCount > 0 {
		sendErrorLog(w, "A teacher with this name already exists")
		return
	}

	emailCount, err := dbRW.GetQueries().GetTeacherCountByEmail(r.Context(), req.Email)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	if emailCount > 0 {
		sendErrorLog(w, "A teacher with this email already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		sendErrorLog(w, "Failed to hash password")
		return
	}

	if req.AssignedColor == "" {
		req.AssignedColor = "#B9D283"
	}

	teacherStatus := string(constants.TeacherStatusPending)
	if isSuperuser {
		teacherStatus = string(constants.TeacherStatusApproved)
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
		Status:         teacherStatus,
	})
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	rl.add(fmt.Sprintf("Successfully registered teacher: %s", req.Name))

	if isSuperuser {
		if _, err := fmt.Fprintf(w, "Teacher '%s' registered successfully\n", req.Name); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
		for _, log := range rl.messages {
			if _, err := fmt.Fprint(w, log+"\n"); err != nil {
				sendErrorLog(w, err.Error())
				return
			}
		}
		return
	}

	setSuccessFlash(w, "Registration successful. Please wait for admin approval before logging in.")
	HttpRedirect(w, "/auth/login")
}

func validateTeacherRequest(req *models.TeacherRegisterRequest) error {
	if req.Name == "" {
		return errors.New("name is required")
	}

	if req.Email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errors.New("invalid email address")
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

	if req.DriveUrl == "" {
		return errors.New("spreadsheet URL is required")
	}
	if _, err := utils.DriveURLToExportURL(req.DriveUrl, "csv"); err != nil {
		return errors.New("invalid spreadsheet URL")
	}

	validSex := map[string]bool{"M": true, "F": true}
	if req.Sex != "" && !validSex[req.Sex] {
		return errors.New("invalid sex. Must be M or F")
	}

	if req.Password == "" {
		return errors.New("password is required")
	}

	if !validatePassword(req.Password) {
		return errors.New("password must be 8-32 characters with uppercase, lowercase, number, and symbol (!@#$%^&*?)")
	}

	if req.Password != req.RetypePassword {
		return errors.New("passwords do not match")
	}

	return nil
}

func handleTeacherApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	teacherID, err := requireInt64(r.FormValue("id"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if err := dbRW.GetQueries().ApproveTeacher(r.Context(), teacherID); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	setSuccessFlash(w, "Teacher approved successfully.")
	HttpRedirect(w, "/teachers")
}

func handleTeacherUnapprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	teacherID, err := requireInt64(r.FormValue("id"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if err := dbRW.GetQueries().UnapproveTeacher(r.Context(), teacherID); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	setSuccessFlash(w, "Teacher set to pending.")
	HttpRedirect(w, "/teachers")
}

func sendErrorLog(w http.ResponseWriter, message string) {
	logs.Log().Error(message)
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

	teachers, err := dbRO.GetQueries().GetApprovedTeachers(r.Context())
	if err != nil {
		HttpError(w, "Failed to fetch teachers", http.StatusInternalServerError)
		return
	}

	var teacherResponses []models.TeacherAPIResponse
	for _, t := range teachers {
		template := ""
		if t.Template.Valid {
			template = t.Template.String
		}
		teacherResponses = append(teacherResponses, models.TeacherAPIResponse{
			ID:           t.ID,
			Name:         t.Name,
			DriveUrl:     t.DriveUrl,
			RatePerClass: t.RatePerClass,
			Template:     template,
		})
	}
	if err := frontend.TeacherOptions(teacherResponses).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Handle Login]"
	if user, loggedIn := auth.UserFromRequest(r, conf.Conf()); loggedIn {
		if auth.SessionUserValid(r.Context(), dbRO.GetQueries(), user) {
			redirectToPortal(w, r)
			return
		}
		auth.ClearSession(w)
	}

	if r.Method == http.MethodGet {
		successMsg := readFlashCookie(w, r, "success_flash")
		if err := frontend.Login(successMsg).Render(r.Context(), w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ip := clientIP(r)
	if !auth.LoginAllowed(ip) {
		fmt.Fprint(w, "Too many login attempts. Please try again later.")
		return
	}

	if err := r.ParseForm(); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		auth.RecordLoginFailure(ip)
		fmt.Fprint(w, "Invalid email or password")
		return
	}

	user, err := auth.Login(w, r, conf.Conf(), dbRO.GetQueries(), email, password)
	if err != nil {
		auth.RecordLoginFailure(ip)
		if errors.Is(err, auth.ErrTeacherPendingApproval) {
			fmt.Fprint(w, err.Error())
			return
		}
		fmt.Fprint(w, "Invalid email or password")
		return
	}
	auth.ResetLoginFailures(ip)

	if ua := r.UserAgent(); ua != "" && user.Role == auth.RoleTeacher && user.ID != 0 {
		useragentID := getOrCreateUserAgentID(r.Context(), dbRW, ua)
		if _, err := dbRW.GetQueries().CreateAccess(r.Context(), queries.CreateAccessParams{
			TeacherID:   user.ID,
			UseragentID: useragentID,
		}); err != nil {
			logs.Log().Warn(logtag, zap.Error(err))
		}
	}

	HttpRedirect(w, "/")
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	auth.Logout(w)
	HttpRedirect(w, "/")
}

const passwordResetTokenTTL = 30 * time.Minute

func insertPasswordResetEvent(ctx context.Context, email, ip, status, event string, teacherID sql.NullInt64, token sql.NullString, expires sql.NullString) error {
	return dbRW.GetQueries().InsertPasswordResetEvent(ctx, queries.InsertPasswordResetEventParams{
		Email:      email,
		IpAddress:  ip,
		TeacherID:  teacherID,
		ResetToken: token,
		Status:     status,
		Event:      event,
		ExpiresAt:  expires,
	})
}

func passwordResetTokenValid(ctx context.Context, token string) (queries.TblPasswordResetEvent, error) {
	tokenNull := sql.NullString{String: token, Valid: token != ""}
	if token == "" {
		return queries.TblPasswordResetEvent{}, errors.New("missing token")
	}

	completed, err := dbRO.GetQueries().HasCompletedPasswordResetForToken(ctx, tokenNull)
	if err != nil {
		return queries.TblPasswordResetEvent{}, err
	}
	if completed > 0 {
		return queries.TblPasswordResetEvent{}, errors.New("token already used")
	}

	row, err := dbRO.GetQueries().GetPasswordResetByToken(ctx, tokenNull)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return queries.TblPasswordResetEvent{}, errors.New("invalid token")
		}
		return queries.TblPasswordResetEvent{}, err
	}

	if !row.ExpiresAt.Valid {
		return queries.TblPasswordResetEvent{}, errors.New("token expired")
	}
	expires, err := time.Parse("2006-01-02 15:04:05", row.ExpiresAt.String)
	if err != nil {
		return queries.TblPasswordResetEvent{}, errors.New("token expired")
	}
	if time.Now().After(expires) {
		_ = insertPasswordResetEvent(ctx, row.Email, row.IpAddress, "failed", "token_expired", row.TeacherID, tokenNull, row.ExpiresAt)
		return queries.TblPasswordResetEvent{}, errors.New("token expired")
	}

	return row, nil
}

func handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Handle Forgot Password]"
	ctx := r.Context()

	if user, loggedIn := auth.UserFromRequest(r, conf.Conf()); loggedIn {
		if auth.SessionUserValid(ctx, dbRO.GetQueries(), user) {
			redirectToPortal(w, r)
			return
		}
		auth.ClearSession(w)
	}

	if r.Method == http.MethodGet {
		successMsg := readFlashCookie(w, r, "success_flash")
		if err := frontend.ForgotPassword(successMsg).Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ip := clientIP(r)
	if err := r.ParseForm(); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		if err := frontend.ForgotPasswordError("Email is required").Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if _, err := mail.ParseAddress(email); err != nil {
		if err := frontend.ForgotPasswordError("Invalid email address").Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if !auth.ResetRequestAllowed(ip) {
		if err := insertPasswordResetEvent(ctx, email, ip, "blocked", "ip_rate_limited", sql.NullInt64{}, sql.NullString{}, sql.NullString{}); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
		}
		if err := frontend.ForgotPasswordError("Too many attempts. Please try again in 30 minutes.").Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	auth.RecordResetRequest(ip)

	if err := insertPasswordResetEvent(ctx, email, ip, "requested", "request_submitted", sql.NullInt64{}, sql.NullString{}, sql.NullString{}); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
	}

	teacher, err := dbRO.GetQueries().GetTeacherByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err := insertPasswordResetEvent(ctx, email, ip, "requested", "email_not_found", sql.NullInt64{}, sql.NullString{}, sql.NullString{}); err != nil {
				logs.Log().Error(logtag, zap.Error(err))
			}
			if err := frontend.ForgotPasswordSuccess().Render(ctx, w); err != nil {
				HttpError(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		logs.Log().Error(logtag, zap.Error(err))
		HttpError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	if teacher.Status != "approved" {
		if err := insertPasswordResetEvent(ctx, email, ip, "requested", "teacher_pending", sql.NullInt64{Int64: teacher.ID, Valid: true}, sql.NullString{}, sql.NullString{}); err != nil {
			logs.Log().Error(logtag, zap.Error(err))
		}
		if err := frontend.ForgotPasswordSuccess().Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	resetToken := uuid.New().String()
	expiresAt := time.Now().Add(passwordResetTokenTTL).Format("2006-01-02 15:04:05")
	if err := insertPasswordResetEvent(ctx, email, ip, "token_issued", "token_issued", sql.NullInt64{Int64: teacher.ID, Valid: true}, sql.NullString{String: resetToken, Valid: true}, sql.NullString{String: expiresAt, Valid: true}); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		HttpError(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	HttpRedirect(w, "/auth/forgot-password/reset?token="+url.QueryEscape(resetToken))
}

func handleResetPassword(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Handle Reset Password]"
	ctx := r.Context()

	if user, loggedIn := auth.UserFromRequest(r, conf.Conf()); loggedIn {
		if auth.SessionUserValid(ctx, dbRO.GetQueries(), user) {
			redirectToPortal(w, r)
			return
		}
		auth.ClearSession(w)
	}

	if r.Method == http.MethodGet {
		token := r.URL.Query().Get("token")
		row, err := passwordResetTokenValid(ctx, token)
		if err != nil {
			setSuccessFlash(w, "Password reset sent")
			http.Redirect(w, r, utils.URL("/auth/forgot-password"), http.StatusFound)
			return
		}
		if err := frontend.ResetPassword(row.Email, token).Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	token := r.FormValue("token")
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")

	row, err := passwordResetTokenValid(ctx, token)
	if err != nil {
		fmt.Fprint(w, "Invalid or expired reset link. Please request a new password reset.")
		return
	}
	if row.Email != email {
		fmt.Fprint(w, "Invalid or expired reset link. Please request a new password reset.")
		return
	}
	if !row.TeacherID.Valid {
		fmt.Fprint(w, "Invalid or expired reset link. Please request a new password reset.")
		return
	}

	if password == "" {
		fmt.Fprint(w, "Password is required")
		return
	}
	if !validatePassword(password) {
		fmt.Fprint(w, "Password must be 8-32 characters with uppercase, lowercase, number, and symbol (!@#$%^&*?)")
		return
	}
	if password != confirmPassword {
		fmt.Fprint(w, "Passwords do not match")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		fmt.Fprint(w, "Something went wrong")
		return
	}

	if err := dbRW.GetQueries().UpdateTeacherPassword(ctx, queries.UpdateTeacherPasswordParams{
		Password: string(hashedPassword),
		ID:       row.TeacherID.Int64,
	}); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
		fmt.Fprint(w, "Something went wrong")
		return
	}

	tokenNull := sql.NullString{String: token, Valid: true}
	if err := insertPasswordResetEvent(ctx, email, row.IpAddress, "completed", "password_updated", row.TeacherID, tokenNull, row.ExpiresAt); err != nil {
		logs.Log().Error(logtag, zap.Error(err))
	}

	setSuccessFlash(w, "Password reset successfully. Please log in.")
	HttpRedirect(w, "/auth/login")
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

	if auth.GetRole(r.Context()) == auth.RoleTeacher {
		user := auth.GetUser(r.Context())
		if teacherID != user.ID {
			HttpError(w, "Forbidden", http.StatusForbidden)
			return
		}
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
	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		frontend.RecordClass(frontend.RecordClassData{
			IsSuperuser: role == auth.RoleSuperuser,
		}).Render(ctx, w)
		return
	}

	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if role == auth.RoleTeacher && user.ID == 0 {
		HttpError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rl := &requestLogs{}
	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	studentID, err := requireInt64(r.FormValue("student"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	duration, err := requireInt64(r.FormValue("duration"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	rate, err := requireFloat64(r.FormValue("rate"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	teacherID := user.ID
	if role == auth.RoleSuperuser {
		teacherID, err = requireInt64(r.FormValue("teacher"))
		if err != nil {
			sendErrorLog(w, "teacher is required")
			return
		}
	}

	req := models.ClassRecordRequest{
		StudentID:       studentID,
		TeacherID:       teacherID,
		Date:            r.FormValue("date"),
		DurationMinutes: duration,
		Rate:            rate,
		Currency:        r.FormValue("currency"),
		Status:          r.FormValue("status"),
		Reason:          r.FormValue("reason"),
	}

	if err := validateClassRecordRequest(&req); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if role == auth.RoleTeacher {
		assigned, err := dbRO.GetQueries().IsStudentAssignedToTeacher(ctx, queries.IsStudentAssignedToTeacherParams{
			TeacherID: teacherID,
			StudentID: studentID,
		})
		if err != nil {
			sendErrorLog(w, "Failed to verify student assignment")
			return
		}
		if assigned == 0 {
			sendErrorLog(w, "student is not assigned to this teacher")
			return
		}
	}

	err = dbRW.GetQueries().InsertClassRecord(ctx, queries.InsertClassRecordParams{
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

	rl.add("Class recorded successfully")
	for _, log := range rl.messages {
		if _, err := fmt.Fprint(w, log+"\n"); err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}
}

func handleClasses(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := frontend.ClassesData{}
		if auth.GetRole(r.Context()) == auth.RoleTeacher {
			user := auth.GetUser(r.Context())
			data = frontend.ClassesData{
				LockTeacher: true,
				TeacherID:   strconv.FormatInt(user.ID, 10),
				TeacherName: user.Name,
			}
		}
		w.Header().Set("Content-Type", "text/html")
		frontend.Classes(data).Render(r.Context(), w)
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
