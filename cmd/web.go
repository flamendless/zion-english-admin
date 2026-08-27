package cmd

import (
	"context"
	"database/sql"
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
	"zion-english/internal/announcements"
	"zion-english/internal/auth"
	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/models"
	"zion-english/internal/notifications"
	"zion-english/internal/processor"
	"zion-english/internal/sheet"
	"zion-english/internal/utils"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var webFlags WebFlags
var dbRW database.Service
var dbRO database.Service

type WebFlags struct {
	port    string
	baseURL string
	https   bool
	address string
}

type requestLogs struct {
	messages []string
}

var cmdWeb = &cobra.Command{
	Use:   "web",
	Short: "Start web server",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := conf.Conf()
		if err := os.MkdirAll("tmp", 0755); err != nil {
			panic(err)
		}
		if err := os.MkdirAll("data/avatars", 0755); err != nil {
			panic(err)
		}

		if err := database.Init("data/zion.db"); err != nil {
			panic(fmt.Sprintf("Failed to initialize database: %v", err))
		}
		defer database.Close()

		dbRW = database.New(database.DB_MODE_RW)
		dbRO = database.New(database.DB_MODE_RO)
		initNotifyService()
		initMeetingService()

		basePath := "/" + strings.TrimPrefix(webFlags.baseURL, "/")
		cfg.BasePath = basePath

		publicMux := http.NewServeMux()
		publicMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, basePath, http.StatusFound)
		})
		publicMux.HandleFunc(basePath+"/auth/login", handleLogin)
		publicMux.HandleFunc(basePath+"/auth/logout", handleLogoutWithAccess)
		publicMux.HandleFunc(basePath+"/auth/forgot-password", handleForgotPassword)
		publicMux.HandleFunc(basePath+"/auth/forgot-password/reset", handleResetPassword)
		publicMux.HandleFunc(basePath+"/teachers/register", handleTeacherRegister)
		publicMux.HandleFunc(basePath+"/health", handleHealth)
		publicMux.HandleFunc(basePath+"/profile/zoom/callback", handleZoomCallback)

		publicMux.Handle(
			basePath+"/static/",
			http.StripPrefix(basePath+"/static/", http.FileServer(http.Dir("static"))),
		)

		authMux := http.NewServeMux()
		authMux.HandleFunc(basePath+"/dashboard", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleHome))
		authMux.HandleFunc(basePath+"/students", auth.RequireRole(auth.AdminAccessRoles()...)(handleStudents))
		authMux.HandleFunc(basePath+"/students/register", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleStudentRegister))
		authMux.HandleFunc(basePath+"/teachers", auth.RequireRole(auth.AdminAccessRoles()...)(handleTeachers))
		authMux.HandleFunc(basePath+"/teachers/approve", auth.RequireRole(auth.AdminAccessRoles()...)(handleTeacherApprove))
		authMux.HandleFunc(basePath+"/teachers/unapprove", auth.RequireRole(auth.AdminAccessRoles()...)(handleTeacherUnapprove))
		authMux.HandleFunc(basePath+"/teachers/delete", auth.RequireRole(auth.AdminAccessRoles()...)(handleTeacherDelete))
		authMux.HandleFunc(basePath+"/students/", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleStudentsPath))
		authMux.HandleFunc(basePath+"/teachers/", auth.RequireRole(auth.AdminAccessRoles()...)(handleTeachersPath))
		authMux.HandleFunc(basePath+"/classes/partials/rows", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleClassRecordsPartial))
		authMux.HandleFunc(basePath+"/classes/", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleClassesPath))
		authMux.HandleFunc(basePath+"/classes/record", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleClassRecord))
		authMux.HandleFunc(basePath+"/classes", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleClasses))
		authMux.HandleFunc(basePath+"/schedule/partials/list", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleScheduleListPartial))
		authMux.HandleFunc(basePath+"/schedule/", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleSchedulePath))
		authMux.HandleFunc(basePath+"/schedule", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleSchedule))
		authMux.HandleFunc(basePath+"/my-students", auth.RequireRole(auth.RoleTeacher)(handleMyStudents))
		authMux.HandleFunc(basePath+"/logs", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleSystemLogs))
		authMux.HandleFunc(basePath+"/changelogs", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleChangelogs))
		authMux.HandleFunc(basePath+"/guides/", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleGuidesPath))
		authMux.HandleFunc(basePath+"/guides", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleGuides))
		authMux.HandleFunc(basePath+"/process-logs", auth.RequireRole(auth.AdminAccessRoles()...)(handleLogs))
		authMux.HandleFunc(basePath+"/process", auth.RequireRole(auth.AdminAccessRoles()...)(handleProcessPage))
		authMux.HandleFunc(basePath+"/reports/partials/rows", auth.RequireRole(auth.AdminAccessRoles()...)(handleReportsPartial))
		authMux.HandleFunc(basePath+"/reports/", auth.RequireRole(auth.AdminAccessRoles()...)(handleReportsPath))
		authMux.HandleFunc(basePath+"/reports", auth.RequireRole(auth.AdminAccessRoles()...)(handleReports))
		authMux.HandleFunc(basePath+"/analytics", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleAnalytics))
		authMux.HandleFunc(basePath+"/api/analytics", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleGetAnalytics))
		authMux.HandleFunc(basePath+"/download/processed", auth.RequireRole(auth.AdminAccessRoles()...)(handleDownload))
		authMux.HandleFunc(basePath+"/profile", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleProfile))
		authMux.HandleFunc(basePath+"/profile/mobile", auth.RequireRole(auth.RoleTeacher, auth.RoleAdmin)(handleProfileMobile))
		authMux.HandleFunc(basePath+"/profile/names", auth.RequireRole(auth.RoleTeacher, auth.RoleAdmin)(handleProfileNames))
		authMux.HandleFunc(basePath+"/profile/password", auth.RequireRole(auth.RoleTeacher, auth.RoleAdmin)(handleProfilePassword))
		authMux.HandleFunc(basePath+"/profile/avatar", auth.RequireRole(auth.RoleTeacher, auth.RoleAdmin)(handleProfileAvatar))
		authMux.HandleFunc(basePath+"/profile/picture", auth.RequireRole(auth.RoleTeacher, auth.RoleAdmin)(handleProfilePicture))
		authMux.HandleFunc(basePath+"/profile/document", auth.RequireRole(auth.RoleTeacher, auth.RoleAdmin)(handleProfileDocument))
		authMux.HandleFunc(basePath+"/profile/zoom/connect", auth.RequireRole(auth.RoleTeacher, auth.RoleAdmin)(handleZoomConnect))
		authMux.HandleFunc(basePath+"/profile/zoom/disconnect", auth.RequireRole(auth.RoleTeacher, auth.RoleAdmin)(handleZoomDisconnect))
		documentsRole := auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)
		authMux.HandleFunc(basePath+"/documents/partials/rows", documentsRole(handleDocumentsPartial))
		authMux.HandleFunc(basePath+"/documents", documentsRole(handleDocuments))
		authMux.HandleFunc(basePath+"/documents/", documentsRole(handleDocumentsPath))
		authMux.HandleFunc(basePath+"/role", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleGetRole))
		authMux.HandleFunc(basePath+"/header-avatar", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleHeaderAvatar))
		authMux.HandleFunc(basePath+"/refresh", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleRefreshPage))
		authMux.HandleFunc(basePath+"/api/teachers", auth.RequireRole(auth.AdminAccessRoles()...)(handleGetTeachers))
		authMux.HandleFunc(basePath+"/api/teacher-row", auth.RequireRole(auth.AdminAccessRoles()...)(handleGetTeacherRow))
		authMux.HandleFunc(basePath+"/api/students", auth.RequireRole(auth.AdminAccessRoles()...)(handleGetStudents))
		authMux.HandleFunc(basePath+"/api/students/search", auth.RequireRole(auth.AdminAccessRoles()...)(handleSearchStudents))
		authMux.HandleFunc(basePath+"/api/me/students", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleGetMyStudents))
		authMux.HandleFunc(basePath+"/api/scheduled-classes", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleGetScheduledClasses))
		authMux.HandleFunc(basePath+"/api/teacher-picture", auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)(handleTeacherPicture))
		authMux.HandleFunc(basePath+"/announcements", auth.RequireRole(auth.AdminAccessRoles()...)(handleAnnouncements))
		authMux.HandleFunc(basePath+"/announcements/register", auth.RequireRole(auth.AdminAccessRoles()...)(handleAnnouncementRegister))
		authMux.HandleFunc(basePath+"/announcements/", auth.RequireRole(auth.AdminAccessRoles()...)(handleAnnouncementsPath))
		notificationsRole := auth.RequireRole(auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher)
		authMux.HandleFunc(basePath+"/notifications", notificationsRole(handleNotifications))
		authMux.HandleFunc(basePath+"/notifications/panel", notificationsRole(handleNotificationsPanel))
		authMux.HandleFunc(basePath+"/notifications/unread-count", notificationsRole(handleNotificationsUnreadCount))
		authMux.HandleFunc(basePath+"/notifications/read-all", notificationsRole(handleNotificationsReadAll))
		authMux.HandleFunc(basePath+"/notifications/", notificationsRole(handleNotificationsPath))

		authHandler := auth.Middleware(cfg, dbRO.GetQueries(), auth.CSRFMiddleware(cfg, announcements.Middleware(dbRO.GetQueries(), authMux)))

		rootMux := http.NewServeMux()

		// public routes
		rootMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, basePath, http.StatusFound)
		})
		rootMux.HandleFunc(basePath, handleLanding)
		rootMux.HandleFunc(basePath+"/privacy", handlePrivacy)
		rootMux.HandleFunc(basePath+"/terms", handleTerms)
		rootMux.HandleFunc(basePath+"/support", handleSupport)
		rootMux.HandleFunc(basePath+"/docs/connect-zoom", handleDocsConnectZoom)
		rootMux.HandleFunc(basePath+"/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != basePath+"/" {
				http.NotFound(w, r)
				return
			}
			http.Redirect(w, r, basePath, http.StatusMovedPermanently)
		})
		rootMux.Handle(basePath+"/auth/", publicMux)
		rootMux.Handle(basePath+"/static/", publicMux)
		rootMux.Handle(basePath+"/teachers/register", publicMux)
		rootMux.Handle(basePath+"/health", publicMux)
		rootMux.Handle(basePath+"/profile/zoom/callback", publicMux)

		// protected routes
		rootMux.Handle(basePath+"/dashboard", authHandler)
		rootMux.Handle(basePath+"/logs", authHandler)
		rootMux.Handle(basePath+"/changelogs", authHandler)
		rootMux.Handle(basePath+"/guides", authHandler)
		rootMux.Handle(basePath+"/guides/", authHandler)
		rootMux.Handle(basePath+"/process-logs", authHandler)
		rootMux.Handle(basePath+"/process", authHandler)
		rootMux.Handle(basePath+"/reports/partials/", authHandler)
		rootMux.Handle(basePath+"/reports", authHandler)
		rootMux.Handle(basePath+"/reports/", authHandler)
		rootMux.Handle(basePath+"/analytics", authHandler)
		rootMux.Handle(basePath+"/api/analytics", authHandler)
		rootMux.Handle(basePath+"/download/processed", authHandler)
		rootMux.Handle(basePath+"/role", authHandler)
		rootMux.Handle(basePath+"/header-avatar", authHandler)
		rootMux.Handle(basePath+"/refresh", authHandler)
		rootMux.Handle(basePath+"/students", authHandler)
		rootMux.Handle(basePath+"/students/", authHandler)
		rootMux.Handle(basePath+"/teachers", authHandler)
		rootMux.Handle(basePath+"/teachers/", authHandler)
		rootMux.Handle(basePath+"/teachers/approve", authHandler)
		rootMux.Handle(basePath+"/teachers/unapprove", authHandler)
		rootMux.Handle(basePath+"/classes/partials/", authHandler)
		rootMux.Handle(basePath+"/classes", authHandler)
		rootMux.Handle(basePath+"/classes/", authHandler)
		rootMux.Handle(basePath+"/schedule/partials/", authHandler)
		rootMux.Handle(basePath+"/schedule", authHandler)
		rootMux.Handle(basePath+"/schedule/", authHandler)
		rootMux.Handle(basePath+"/my-students", authHandler)
		rootMux.Handle(basePath+"/profile", authHandler)
		rootMux.Handle(basePath+"/profile/", authHandler)
		rootMux.Handle(basePath+"/documents", authHandler)
		rootMux.Handle(basePath+"/documents/", authHandler)
		rootMux.Handle(basePath+"/api/teachers", authHandler)
		rootMux.Handle(basePath+"/api/teacher-row", authHandler)
		rootMux.Handle(basePath+"/api/students", authHandler)
		rootMux.Handle(basePath+"/api/students/", authHandler)
		rootMux.Handle(basePath+"/api/scheduled-classes", authHandler)
		rootMux.Handle(basePath+"/api/teacher-picture", authHandler)
		rootMux.Handle(basePath+"/api/me/students", authHandler)
		rootMux.Handle(basePath+"/announcements", authHandler)
		rootMux.Handle(basePath+"/announcements/", authHandler)
		rootMux.Handle(basePath+"/notifications", authHandler)
		rootMux.Handle(basePath+"/notifications/", authHandler)

		handler := logRequests(securityHeaders(rootMux))

		port := webFlags.port
		if !cmd.Flags().Changed("port") {
			port = strconv.Itoa(cfg.Port)
		}
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

func init() {
	f := cmdWeb.Flags
	f().StringVarP(&webFlags.port, "port", "p", "8080", "Port to run web server on")
	f().StringVarP(&webFlags.baseURL, "url", "b", "zion-english-admin", "Base URL")
	f().BoolVar(&webFlags.https, "https", false, "Enable HTTPS")
	f().StringVar(&webFlags.address, "address", "", "Domain address for Let's Encrypt certificates (e.g., flamendless.xyz)")
	rootCmd.AddCommand(cmdWeb)
}

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

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		fields := []zap.Field{
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", wrapped.status),
			zap.Duration("duration", time.Since(start)),
			zap.String("remote", clientIP(r)),
		}
		if query := r.URL.RawQuery; query != "" {
			fields = append(fields, zap.String("query", query))
		}

		logs.Log().Info("request", fields...)
	})
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

func HttpRedirect(w http.ResponseWriter, r *http.Request, url string) {
	target := utils.URL(url)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", target)
		w.WriteHeader(http.StatusFound)
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func (l *requestLogs) add(msg string) {
	entry := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
	l.messages = append(l.messages, entry)
	logs.Log().Info(msg)
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

	processMsg := fmt.Sprintf(
		"processed sheet for teacher '%s': %d records (%s to %s), %s",
		req.Name, len(records), req.StartDate, req.EndDate, filename,
	)
	if err := dbRW.GetQueries().UpdateTeacherTemplate(ctx, queries.UpdateTeacherTemplateParams{
		ID:       teacherID,
		Template: sql.NullString{String: req.Template, Valid: true},
	}); err != nil {
		logs.Log().Error("update teacher template", zap.Error(err))
	} else {
		processMsg += fmt.Sprintf(", template updated to '%s'", req.Template)
	}
	insertAuditLogAs(ctx, auth.GetUser(ctx), "process", processMsg)

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

	auditReportDownload(r.Context(), cleanPath)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	http.ServeFile(w, r, fullPath)
}

func validateProcessRequest(req *models.ProcessRequest) error {
	if err := utils.ValidateDriveSpreadsheetURL(req.DriveURL); err != nil {
		if errors.Is(err, utils.ErrInvalidDriveSpreadsheetURL) {
			return errors.New("invalid Google Spreadsheet URL")
		}
		return errors.New("google spreadsheet url is required")
	}

	// Validate name
	matched, _ := regexp.MatchString(`^[a-zA-Z.\s-]+$`, req.Name)
	if !matched {
		return errors.New("name must contain only letters, periods, dashes, and spaces")
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

func formatStudentRelationships(rels []queries.GetAllStudentRelationshipsRow) string {
	if len(rels) == 0 {
		return ""
	}
	parts := make([]string, len(rels))
	for i, r := range rels {
		if r.Relationship.Valid && r.Relationship.String != "" {
			parts[i] = fmt.Sprintf("%s → %s", r.Relationship.String, r.RelatedStudentName)
		} else {
			parts[i] = r.RelatedStudentName
		}
	}
	return strings.Join(parts, "; ")
}

func handleStudentRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)
	isSuperuser := auth.HasAdminAccess(role)

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		frontend.RegisterStudent(frontend.RegisterStudentData{
			IsSuperuser: isSuperuser,
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

	ratePerClass, err := requireFloat64(r.FormValue("ratePerClass"))
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	var teacherIDs []int64
	if isSuperuser {
		teacherIDs, err = parseAssignedTeacherIDs(ctx, dbRO.GetQueries(), r.Form["teachers"])
		if err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	} else {
		teacherIDs, err = parseAssignedTeacherIDs(ctx, dbRO.GetQueries(), []string{strconv.FormatInt(user.ID, 10)})
		if err != nil {
			sendErrorLog(w, err.Error())
			return
		}
	}

	relatedStudentID := int64(0)
	if isSuperuser {
		if relatedStudentValue := r.FormValue("relatedStudentId"); relatedStudentValue != "" {
			relatedStudentID, err = strconv.ParseInt(relatedStudentValue, 10, 64)
			if err != nil {
				sendErrorLog(w, "invalid related student")
				return
			}
		}
	}

	req := models.StudentRegisterRequest{
		Name:         r.FormValue("name"),
		Currency:     r.FormValue("currency"),
		RatePerClass: ratePerClass,
		Status:       r.FormValue("status"),
		ParentName:   r.FormValue("parentName"),
	}
	if isSuperuser {
		req.Contact = r.FormValue("contact")
		req.AssignedColor = r.FormValue("assignedColor")
		req.InactiveReason = r.FormValue("inactiveReason")
		req.Relationship = r.FormValue("relationship")
		req.RelatedStudentID = relatedStudentID
	} else {
		req.AssignedColor = "#90C020"
	}

	if err := validateStudentRequest(&req); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	rl.add(fmt.Sprintf("Registering student: %s", req.Name))

	if req.AssignedColor == "" {
		req.AssignedColor = "#B9D283"
	}

	if req.RelatedStudentID > 0 {
		_, err := dbRO.GetQueries().GetStudentByID(r.Context(), req.RelatedStudentID)
		if err != nil {
			sendErrorLog(w, "related student not found")
			return
		}
	}

	tx, err := dbRW.GetDB().BeginTx(r.Context(), nil)
	if err != nil {
		sendErrorLog(w, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	qtx := dbRW.GetQueries().WithTx(tx)

	studentID, err := qtx.InsertStudent(r.Context(), queries.InsertStudentParams{
		Name:           req.Name,
		Currency:       req.Currency,
		Contact:        sql.NullString{String: req.Contact, Valid: req.Contact != ""},
		RatePerClass:   req.RatePerClass,
		ParentName:     sql.NullString{String: req.ParentName, Valid: req.ParentName != ""},
		AssignedColor:  req.AssignedColor,
		Status:         req.Status,
		InactiveReason: sql.NullString{String: req.InactiveReason, Valid: req.InactiveReason != ""},
	})
	if err != nil {
		sendErrorLog(w, "Failed to register student")
		return
	}

	for _, tid := range teacherIDs {
		err = qtx.InsertTeacherStudentM2M(r.Context(), queries.InsertTeacherStudentM2MParams{
			TeacherID: tid,
			StudentID: studentID,
		})
		if err != nil {
			sendErrorLog(w, "Failed to assign teacher")
			return
		}
	}

	if req.RelatedStudentID > 0 {
		if req.RelatedStudentID == studentID {
			sendErrorLog(w, "a student cannot be related to themselves")
			return
		}

		err = qtx.InsertStudentRelationship(r.Context(), queries.InsertStudentRelationshipParams{
			StudentID:        studentID,
			RelatedStudentID: req.RelatedStudentID,
			Relationship:     sql.NullString{String: req.Relationship, Valid: req.Relationship != ""},
		})
		if err != nil {
			sendErrorLog(w, "Failed to save student relationship")
			return
		}
		rl.add(fmt.Sprintf("Linked to related student ID: %d", req.RelatedStudentID))
	}

	if err := tx.Commit(); err != nil {
		sendErrorLog(w, "Failed to register student")
		return
	}
	teacherIDStrs := make([]string, len(teacherIDs))
	for i, tid := range teacherIDs {
		teacherIDStrs[i] = strconv.FormatInt(tid, 10)
		rl.add(fmt.Sprintf("Assigned student to teacher ID: %d", tid))
	}

	rl.add(fmt.Sprintf("Successfully registered student: %s", req.Name))

	insertAuditLogAs(r.Context(), auth.GetUser(r.Context()), "students", fmt.Sprintf("registered student '%s' (teacher ids %s)", req.Name, strings.Join(teacherIDStrs, ",")))
	notifyTeachers(r.Context(), teacherIDs, teacherNamesMap(r.Context(), teacherIDs), auth.GetUser(r.Context()), notifications.KindStudentRegistered,
		fmt.Sprintf("New student '%s' was assigned to you", req.Name))

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

	if !constants.ValidCurrency(req.Currency) {
		return errors.New("invalid currency. Must be KRW, CAD, YEN, or PHP")
	}

	if req.RatePerClass < 0 {
		return errors.New("rate per class cannot be negative")
	}

	if req.AssignedColor == "" {
		return errors.New("assigned color is required")
	}

	if !constants.ValidStudentStatus(req.Status) {
		return errors.New("invalid status. Must be active or inactive")
	}

	req.InactiveReason = strings.TrimSpace(req.InactiveReason)
	if req.Status == "inactive" && req.InactiveReason == "" {
		return errors.New("inactive reason is required when status is inactive")
	}
	if req.Status == "active" {
		req.InactiveReason = ""
	}

	return nil
}

func handleGetTeacherRow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := frontend.TeacherAssignRow("", true).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
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

	isSuperuser := auth.HasAdminAccess(role)

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

	if !isSuperuser {
		ip := clientIP(r)
		if !auth.RegistrationAllowed(ip) {
			sendErrorLog(w, "Registration limit reached for this network. Please try again later or contact an administrator.")
			return
		}
	}

	rl := &requestLogs{}

	if err := r.ParseForm(); err != nil {
		sendErrorLog(w, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	req, err := parseTeacherForm(r, true)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if req.JoiningDate == "" {
		req.JoiningDate = time.Now().Format("2006-01-02")
	}

	if err := validateTeacherRegisterRequest(&req); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	rl.add(fmt.Sprintf("Registering teacher: %s", req.Name))

	mobileCount, err := dbRW.GetQueries().GetTeacherCountByMobile(r.Context(), req.MobileNumber)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	if mobileCount > 0 {
		existing, lookupErr := dbRW.GetQueries().GetTeacherByMobile(r.Context(), req.MobileNumber)
		if lookupErr == nil && existing.Status == string(constants.TeacherStatusPending) {
			sendErrorLog(w, "An account with this mobile number is awaiting approval")
			return
		}
		sendErrorLog(w, "A teacher with this mobile number already exists")
		return
	}

	emailCount, err := dbRW.GetQueries().GetTeacherCountByEmail(r.Context(), req.Email)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	if emailCount > 0 {
		existing, lookupErr := dbRW.GetQueries().GetTeacherByEmail(r.Context(), req.Email)
		if lookupErr == nil && existing.Status == string(constants.TeacherStatusPending) {
			sendErrorLog(w, "An account with this email is awaiting approval")
			return
		}
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
		FirstName:      req.FirstName,
		MiddleName:     req.MiddleName,
		LastName:       req.LastName,
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

	insertedTeacher, err := dbRW.GetQueries().GetTeacherByEmail(r.Context(), req.Email)
	if err != nil {
		sendErrorLog(w, err.Error())
		return
	}
	if err := dbRW.GetQueries().InsertTeacherRole(r.Context(), queries.InsertTeacherRoleParams{
		TeacherID: insertedTeacher.ID,
		Role:      string(constants.TeacherRoleTeacher),
	}); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	if !isSuperuser {
		auth.RecordRegistration(clientIP(r))
	}

	rl.add(fmt.Sprintf("Successfully registered teacher: %s", req.Name))

	auditActor := auth.GetUser(r.Context())
	if auditActor.Name == "" {
		if loggedIn {
			auditActor = user
		} else {
			auditActor = auth.User{Name: req.Name}
		}
	}
	insertAuditLogAs(r.Context(), auditActor, "teachers", fmt.Sprintf("registered teacher '%s' (status %s)", req.Name, teacherStatus))
	if !isSuperuser {
		notifySuperuser(r.Context(), auditActor, notifications.KindTeacherRegistered,
			fmt.Sprintf("New teacher '%s' registered (status %s)", req.Name, teacherStatus), "")
	}

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
	HttpRedirect(w, r, "/auth/login")
}

func parseTeacherForm(r *http.Request, includePassword bool) (models.TeacherRegisterRequest, error) {
	ratePerClass, err := requireFloat64(r.FormValue("ratePerClass"))
	if err != nil {
		return models.TeacherRegisterRequest{}, err
	}
	req := models.TeacherRegisterRequest{
		FirstName:      strings.TrimSpace(r.FormValue("firstName")),
		MiddleName:     strings.TrimSpace(r.FormValue("middleName")),
		LastName:       strings.TrimSpace(r.FormValue("lastName")),
		Birthdate:      r.FormValue("birthdate"),
		Address:        strings.TrimSpace(r.FormValue("address")),
		JoiningDate:    r.FormValue("joiningDate"),
		MobileNumber:   strings.TrimSpace(r.FormValue("mobileNumber")),
		Email:          utils.NormalizeEmail(r.FormValue("email")),
		Certifications: r.FormValue("certifications"),
		AssignedColor:  r.FormValue("assignedColor"),
		RatePerClass:   ratePerClass,
		Currency:       r.FormValue("currency"),
		DriveUrl:       r.FormValue("driveUrl"),
		Sex:            r.FormValue("sex"),
	}
	if includePassword {
		req.Password = r.FormValue("password")
		req.RetypePassword = r.FormValue("retypePassword")
	}
	req.Name = utils.ComposePersonName(req.FirstName, req.MiddleName, req.LastName)
	return req, nil
}

func validateTeacherFields(req *models.TeacherRegisterRequest) error {
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.MiddleName = strings.TrimSpace(req.MiddleName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Name = utils.ComposePersonName(req.FirstName, req.MiddleName, req.LastName)

	if utils.IsBlank(req.FirstName) {
		return errors.New("first name is required")
	}
	if utils.IsBlank(req.LastName) {
		return errors.New("last name is required")
	}
	if req.Name == "" {
		return errors.New("name is required")
	}

	if req.Birthdate == "" {
		return errors.New("birthdate is required")
	}

	if utils.IsBlank(req.Address) {
		return errors.New("address is required")
	}

	if utils.IsBlank(req.MobileNumber) {
		return errors.New("mobile number is required")
	}

	if req.Email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errors.New("invalid email address")
	}

	if !constants.ValidCurrency(req.Currency) {
		return errors.New("invalid currency. Must be KRW, CAD, YEN, or PHP")
	}

	if req.RatePerClass < 0 {
		return errors.New("rate per class cannot be negative")
	}

	if err := utils.ValidateDriveSpreadsheetURL(req.DriveUrl); err != nil {
		return err
	}

	if !constants.ValidSex(req.Sex) {
		return errors.New("invalid sex. Must be M or F")
	}

	return nil
}

func validateTeacherRegisterRequest(req *models.TeacherRegisterRequest) error {
	if err := validateTeacherFields(req); err != nil {
		return err
	}

	if req.Password == "" {
		return errors.New("password is required")
	}

	if !constants.ValidPassword(req.Password) {
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

	ctx := r.Context()
	existing, err := dbRO.GetQueries().GetTeacherFullByID(ctx, teacherID)
	if err != nil {
		sendErrorLog(w, "Teacher not found")
		return
	}
	if existing.Deleted != 0 {
		sendErrorLog(w, "Teacher is already deleted")
		return
	}
	if existing.Status != string(constants.TeacherStatusPending) {
		sendErrorLog(w, "Only pending teachers can be approved")
		return
	}

	if err := dbRW.GetQueries().ApproveTeacher(ctx, teacherID); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "teachers", fmt.Sprintf("approved teacher '%s' (id %d)", utils.ComposePersonName(existing.FirstName, existing.MiddleName, existing.LastName), teacherID))
	notifyTeacher(ctx, teacherID, utils.ComposePersonName(existing.FirstName, existing.MiddleName, existing.LastName), auth.GetUser(ctx), notifications.KindTeacherApproved,
		"Your teacher account has been approved", "")
	setSuccessFlash(w, "Teacher approved successfully.")
	HttpRedirect(w, r, "/teachers")
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

	ctx := r.Context()
	existing, err := dbRO.GetQueries().GetTeacherFullByID(ctx, teacherID)
	if err != nil {
		sendErrorLog(w, "Teacher not found")
		return
	}
	if existing.Deleted != 0 {
		sendErrorLog(w, "Teacher is already deleted")
		return
	}
	if existing.Status != string(constants.TeacherStatusApproved) {
		sendErrorLog(w, "Only approved teachers can be unapproved")
		return
	}

	if err := dbRW.GetQueries().UnapproveTeacher(ctx, teacherID); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "teachers", fmt.Sprintf("unapproved teacher '%s' (id %d)", utils.ComposePersonName(existing.FirstName, existing.MiddleName, existing.LastName), teacherID))
	setSuccessFlash(w, "Teacher set to pending.")
	HttpRedirect(w, r, "/teachers")
}

func handleTeacherDelete(w http.ResponseWriter, r *http.Request) {
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

	ctx := r.Context()
	existing, err := dbRO.GetQueries().GetTeacherFullByID(ctx, teacherID)
	if err != nil {
		sendErrorLog(w, "Teacher not found")
		return
	}
	if existing.Deleted != 0 {
		sendErrorLog(w, "Teacher is already deleted")
		return
	}
	if existing.Status != string(constants.TeacherStatusPending) {
		sendErrorLog(w, "Only pending teachers can be deleted")
		return
	}

	if err := dbRW.GetQueries().SoftDeleteTeacher(ctx, teacherID); err != nil {
		sendErrorLog(w, err.Error())
		return
	}

	insertAuditLogAs(ctx, auth.GetUser(ctx), "teachers", fmt.Sprintf("deleted teacher '%s' (id %d)", utils.ComposePersonName(existing.FirstName, existing.MiddleName, existing.LastName), teacherID))
	setSuccessFlash(w, "Teacher deleted successfully.")
	HttpRedirect(w, r, "/teachers")
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
			Name:         utils.ComposePersonName(t.FirstName, t.MiddleName, t.LastName),
			DriveUrl:     t.DriveUrl,
			RatePerClass: t.RatePerClass,
			Template:     template,
		})
	}
	if err := frontend.TeacherOptions(teacherResponses, r.URL.Query().Get("selected")).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleLanding(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := frontend.Landing().Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Handle Login]"
	if user, loggedIn := auth.UserFromRequest(r, conf.Conf()); loggedIn {
		if auth.SessionUserValid(r.Context(), dbRO.GetQueries(), user) {
			HttpRedirect(w, r, "/dashboard")
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

	email := utils.NormalizeEmail(r.FormValue("email"))
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

	switch user.Role {
	case auth.RoleSuperuser, auth.RoleAdmin, auth.RoleTeacher:
		insertAuditLogAs(r.Context(), user, "auth", fmt.Sprintf("logged in (%s)", user.Email))
	}

	if ua := r.UserAgent(); ua != "" && (user.Role == auth.RoleTeacher || user.Role == auth.RoleAdmin) && user.ID != 0 {
		useragentID := getOrCreateUserAgentID(r.Context(), dbRW, ua)
		if _, err := dbRW.GetQueries().CreateAccess(r.Context(), queries.CreateAccessParams{
			TeacherID:   user.ID,
			UseragentID: useragentID,
		}); err != nil {
			logs.Log().Warn(logtag, zap.Error(err))
		}
	}

	HttpRedirect(w, r, "/dashboard")
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
			HttpRedirect(w, r, "/dashboard")
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

	email := utils.NormalizeEmail(r.FormValue("email"))
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

	insertAuditLogAs(ctx, auth.User{ID: teacher.ID, Name: utils.ComposePersonName(teacher.FirstName, teacher.MiddleName, teacher.LastName)}, "auth", fmt.Sprintf("requested password reset for '%s'", email))
	HttpRedirect(w, r, "/auth/forgot-password/reset?token="+url.QueryEscape(resetToken))
}

func handleResetPassword(w http.ResponseWriter, r *http.Request) {
	const logtag = "[Handle Reset Password]"
	ctx := r.Context()

	if user, loggedIn := auth.UserFromRequest(r, conf.Conf()); loggedIn {
		if auth.SessionUserValid(ctx, dbRO.GetQueries(), user) {
			HttpRedirect(w, r, "/dashboard")
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
	email := utils.NormalizeEmail(r.FormValue("email"))
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
	if !constants.ValidPassword(password) {
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

	resetActor := auth.User{ID: row.TeacherID.Int64}
	if teacher, err := dbRO.GetQueries().GetTeacherFullByID(ctx, row.TeacherID.Int64); err == nil {
		resetActor.Name = utils.ComposePersonName(teacher.FirstName, teacher.MiddleName, teacher.LastName)
	}
	insertAuditLogAs(ctx, resetActor, "auth", fmt.Sprintf("completed password reset for '%s'", email))

	setSuccessFlash(w, "Password reset successfully. Please log in.")
	HttpRedirect(w, r, "/auth/login")
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
	case auth.RoleAdmin:
		user := auth.GetUser(ctx)
		greetings = fmt.Sprintf("Welcome, admin %s!", user.Name)
	case auth.RoleTeacher:
		user := auth.GetUser(ctx)
		greetings = fmt.Sprintf("Welcome, teacher %s!", user.Name)
	}

	if _, err := fmt.Fprint(w, greetings); err != nil {
		logs.Log().Error("handle get role", zap.Error(err))
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleSearchStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	w.Header().Set("Content-Type", "text/html")
	if q == "" {
		frontend.StudentSearchResults(nil).Render(r.Context(), w)
		return
	}

	students, err := dbRO.GetQueries().SearchStudentsByName(r.Context(), sql.NullString{String: q, Valid: true})
	if err != nil {
		HttpError(w, "Failed to search students", http.StatusInternalServerError)
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

	if err := frontend.StudentSearchResults(studentResponses).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleGetStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	teacherIDStr := r.URL.Query().Get("teacher")
	if teacherIDStr == "" {
		teacherIDStr = r.URL.Query().Get("teacherId")
	}

	var studentResponses []models.StudentAPIResponse
	if teacherIDStr != "" {
		teacherID, parseErr := strconv.ParseInt(teacherIDStr, 10, 64)
		if parseErr != nil || teacherID <= 0 {
			HttpError(w, "Invalid teacher ID", http.StatusBadRequest)
			return
		}
		role := auth.GetRole(ctx)
		if role == auth.RoleTeacher {
			user := auth.GetUser(ctx)
			if teacherID != user.ID {
				HttpError(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
		students, err := dbRO.GetQueries().GetStudentsByTeacherID(ctx, teacherID)
		if err != nil {
			HttpError(w, "Failed to fetch students", http.StatusInternalServerError)
			return
		}
		for _, s := range students {
			studentResponses = append(studentResponses, models.StudentAPIResponse{
				ID:           s.ID,
				Name:         s.Name,
				Currency:     s.Currency,
				RatePerClass: s.RatePerClass,
			})
		}
	} else {
		students, err := dbRO.GetQueries().GetActiveStudents(ctx)
		if err != nil {
			HttpError(w, "Failed to fetch students", http.StatusInternalServerError)
			return
		}
		for _, s := range students {
			studentResponses = append(studentResponses, models.StudentAPIResponse{
				ID:           s.ID,
				Name:         s.Name,
				Currency:     s.Currency,
				RatePerClass: s.RatePerClass,
			})
		}
	}

	w.Header().Set("Content-Type", "text/html")

	selectedID := r.URL.Query().Get("selected")

	if err := frontend.StudentOptions(studentResponses, selectedID).Render(ctx, w); err != nil {
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

	if err := frontend.StudentOptions(studentResponses, r.URL.Query().Get("selected")).Render(r.Context(), w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
