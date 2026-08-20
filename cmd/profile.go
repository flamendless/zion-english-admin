package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	avatarDir         = "data/avatars"
	maxAvatarBytes    = 2 << 20
	avatarCacheMaxAge = 3600
)

func ensureAvatarDir() error {
	return os.MkdirAll(avatarDir, 0755)
}

func avatarFilePath(filename string) string {
	return filepath.Join(avatarDir, filepath.Base(filename))
}

func teacherAvatarURL(hasPicture bool) string {
	if !hasPicture {
		return ""
	}
	return utils.URL("/profile/picture")
}

func teacherPictureURL(teacherID int64, hasPicture bool) string {
	if !hasPicture {
		return ""
	}
	return utils.URL(fmt.Sprintf("/api/teacher-picture?id=%d", teacherID))
}

func buildTeacherAvatarProps(row queries.GetTeacherProfileByIDRow) frontend.AvatarProps {
	hasPicture := row.ProfilePicture.Valid && row.ProfilePicture.String != ""
	assignedColor := row.AssignedColor
	if assignedColor == "" {
		assignedColor = "#B9D283"
	}
	return frontend.AvatarProps{
		Size:          "xl",
		Initials:      utils.PersonInitials(row.FirstName, row.MiddleName, row.LastName, row.Name),
		AssignedColor: assignedColor,
		PictureURL:    teacherAvatarURL(hasPicture),
		HasPicture:    hasPicture,
		Alt:           row.Name + " avatar",
	}
}

func buildSuperuserAvatarProps(user auth.User) frontend.AvatarProps {
	return frontend.AvatarProps{
		Size:          "xl",
		Initials:      utils.PersonInitials("", "", "", user.Name),
		AssignedColor: "#90C020",
		HasPicture:    false,
		Alt:           user.Name + " avatar",
	}
}

func buildHeaderAvatarProps(ctx context.Context, user auth.User, role auth.Role) (frontend.AvatarProps, error) {
	if role == auth.RoleSuperuser {
		props := buildSuperuserAvatarProps(user)
		props.Size = "nav"
		return props, nil
	}

	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, user.ID)
	if err != nil {
		return frontend.AvatarProps{}, err
	}
	props := buildTeacherAvatarProps(row)
	props.Size = "nav"
	return props, nil
}

func handleHeaderAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)
	props, err := buildHeaderAvatarProps(ctx, user, role)
	if err != nil {
		logs.Log().Error("build header avatar", zap.Error(err))
		HttpError(w, "Failed to load avatar", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := frontend.HeaderAvatar(props).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	role := auth.GetRole(ctx)
	now := time.Now()

	if role == auth.RoleSuperuser {
		data := frontend.ProfileData{
			IsSuperuser: true,
			Name:        user.Name,
			Email:       user.Email,
			Role:        frontend.ProfileRoleLabel(role),
			Avatar:      buildSuperuserAvatarProps(user),
		}
		if err := frontend.Profile(data).Render(ctx, w); err != nil {
			HttpError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			HttpError(w, "Profile not found", http.StatusNotFound)
			return
		}
		logs.Log().Error("get teacher profile", zap.Error(err))
		HttpError(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	canChangeMobile, mobileDays := utils.SensitiveChangeAllowed(row.MobileChangedAt, now)
	canChangePassword, passwordDays := utils.SensitiveChangeAllowed(row.PasswordChangedAt, now)
	blockingDocs, err := dbRO.GetQueries().HasBlockingTeacherDocument(ctx, user.ID)
	if err != nil {
		logs.Log().Error("check blocking teacher document", zap.Error(err))
		blockingDocs = 0
	}

	certifications := ""
	if row.Certifications.Valid {
		certifications = row.Certifications.String
	}
	sex := ""
	if row.Sex.Valid {
		sex = row.Sex.String
	}
	template := ""
	if row.Template.Valid {
		template = row.Template.String
	}
	_ = template

	data := frontend.ProfileData{
		IsSuperuser:           false,
		Name:                  row.Name,
		Email:                 row.Email,
		Role:                  frontend.ProfileRoleLabel(role),
		FirstName:             row.FirstName,
		MiddleName:            row.MiddleName,
		LastName:              row.LastName,
		Birthdate:             row.Birthdate,
		Address:               row.Address,
		JoiningDate:           row.JoiningDate,
		MobileNumber:          row.MobileNumber,
		Certifications:        certifications,
		AssignedColor:         row.AssignedColor,
		RatePerClass:          row.RatePerClass,
		Currency:              row.Currency,
		DriveUrl:              row.DriveUrl,
		Sex:                   sex,
		Status:                row.Status,
		HasProfilePicture:     row.ProfilePicture.Valid && row.ProfilePicture.String != "",
		Avatar:                buildTeacherAvatarProps(row),
		CanChangeMobile:       canChangeMobile,
		MobileDaysRemaining:   mobileDays,
		CanChangePassword:     canChangePassword,
		PasswordDaysRemaining: passwordDays,
		CanEditFirstName:      utils.ProfileNameEditable(row.FirstName),
		CanEditMiddleName:     utils.ProfileNameEditable(row.MiddleName),
		CanEditLastName:       utils.ProfileNameEditable(row.LastName),
		CanUploadDocument:     blockingDocs == 0,
	}

	if err := frontend.Profile(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleProfileMobile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if auth.GetRole(ctx) != auth.RoleTeacher {
		HttpError(w, "Access denied", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form submission")
		HttpRedirect(w, r, "/profile")
		return
	}

	mobileNumber := strings.TrimSpace(r.FormValue("mobileNumber"))
	if mobileNumber == "" {
		setErrorFlash(w, "Mobile number is required")
		HttpRedirect(w, r, "/profile")
		return
	}

	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, user.ID)
	if err != nil {
		logs.Log().Error("get teacher profile for mobile update", zap.Error(err))
		setErrorFlash(w, "Failed to load profile")
		HttpRedirect(w, r, "/profile")
		return
	}

	if allowed, days := utils.SensitiveChangeAllowed(row.MobileChangedAt, time.Now()); !allowed {
		setErrorFlash(w, fmt.Sprintf("You can change your mobile number again in %d day(s).", days))
		HttpRedirect(w, r, "/profile")
		return
	}

	if mobileNumber == row.MobileNumber {
		setErrorFlash(w, "New mobile number must be different from your current number")
		HttpRedirect(w, r, "/profile")
		return
	}

	count, err := dbRW.GetQueries().GetTeacherCountByMobile(ctx, mobileNumber)
	if err != nil {
		logs.Log().Error("check mobile duplicate", zap.Error(err))
		setErrorFlash(w, "Failed to validate mobile number")
		HttpRedirect(w, r, "/profile")
		return
	}
	if count > 0 {
		setErrorFlash(w, "A teacher with this mobile number already exists")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := dbRW.GetQueries().UpdateTeacherMobile(ctx, queries.UpdateTeacherMobileParams{
		MobileNumber: mobileNumber,
		ID:           user.ID,
	}); err != nil {
		logs.Log().Error("update teacher mobile", zap.Error(err))
		setErrorFlash(w, "Failed to update mobile number")
		HttpRedirect(w, r, "/profile")
		return
	}

	insertAuditLogAs(ctx, user, "profile", fmt.Sprintf("updated mobile number for teacher '%s'", user.Name))
	setSuccessFlash(w, "Mobile number updated successfully.")
	HttpRedirect(w, r, "/profile")
}

func handleProfileNames(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if auth.GetRole(ctx) != auth.RoleTeacher {
		HttpError(w, "Access denied", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form submission")
		HttpRedirect(w, r, "/profile")
		return
	}

	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, user.ID)
	if err != nil {
		logs.Log().Error("get teacher profile for name update", zap.Error(err))
		setErrorFlash(w, "Failed to load profile")
		HttpRedirect(w, r, "/profile")
		return
	}

	canEditFirst := utils.ProfileNameEditable(row.FirstName)
	canEditMiddle := utils.ProfileNameEditable(row.MiddleName)
	canEditLast := utils.ProfileNameEditable(row.LastName)
	if !canEditFirst && !canEditMiddle && !canEditLast {
		setErrorFlash(w, "Your name is already complete and cannot be changed here")
		HttpRedirect(w, r, "/profile")
		return
	}

	firstName := strings.TrimSpace(r.FormValue("firstName"))
	middleName := strings.TrimSpace(r.FormValue("middleName"))
	lastName := strings.TrimSpace(r.FormValue("lastName"))

	newFirst := row.FirstName
	newMiddle := row.MiddleName
	newLast := row.LastName

	if canEditFirst {
		if utils.IsBlank(firstName) {
			setErrorFlash(w, "First name is required")
			HttpRedirect(w, r, "/profile")
			return
		}
		newFirst = firstName
	} else if firstName != "" && firstName != strings.TrimSpace(row.FirstName) {
		setErrorFlash(w, "First name cannot be changed")
		HttpRedirect(w, r, "/profile")
		return
	}

	if canEditMiddle {
		newMiddle = middleName
	} else if middleName != "" && middleName != strings.TrimSpace(row.MiddleName) {
		setErrorFlash(w, "Middle name cannot be changed")
		HttpRedirect(w, r, "/profile")
		return
	}

	if canEditLast {
		if utils.IsBlank(lastName) {
			setErrorFlash(w, "Last name is required")
			HttpRedirect(w, r, "/profile")
			return
		}
		newLast = lastName
	} else if lastName != "" && lastName != strings.TrimSpace(row.LastName) {
		setErrorFlash(w, "Last name cannot be changed")
		HttpRedirect(w, r, "/profile")
		return
	}

	name := utils.ComposePersonName(newFirst, newMiddle, newLast)
	if name == "" {
		setErrorFlash(w, "Name is required")
		HttpRedirect(w, r, "/profile")
		return
	}

	if newFirst == row.FirstName && newMiddle == row.MiddleName && newLast == row.LastName {
		setErrorFlash(w, "No name changes to save")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := dbRW.GetQueries().UpdateTeacherNames(ctx, queries.UpdateTeacherNamesParams{
		Name:       name,
		FirstName:  newFirst,
		MiddleName: newMiddle,
		LastName:   newLast,
		ID:         user.ID,
	}); err != nil {
		logs.Log().Error("update teacher names", zap.Error(err))
		setErrorFlash(w, "Failed to update name")
		HttpRedirect(w, r, "/profile")
		return
	}

	insertAuditLogAs(ctx, user, "profile", fmt.Sprintf("updated name for teacher '%s'", name))
	setSuccessFlash(w, "Name updated successfully.")
	HttpRedirect(w, r, "/profile")
}

func handleProfilePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if auth.GetRole(ctx) != auth.RoleTeacher {
		HttpError(w, "Access denied", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		setErrorFlash(w, "Invalid form submission")
		HttpRedirect(w, r, "/profile")
		return
	}

	currentPassword := r.FormValue("currentPassword")
	newPassword := r.FormValue("newPassword")
	confirmPassword := r.FormValue("confirmPassword")

	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, user.ID)
	if err != nil {
		logs.Log().Error("get teacher profile for password update", zap.Error(err))
		setErrorFlash(w, "Failed to load profile")
		HttpRedirect(w, r, "/profile")
		return
	}

	if allowed, days := utils.SensitiveChangeAllowed(row.PasswordChangedAt, time.Now()); !allowed {
		setErrorFlash(w, fmt.Sprintf("You can change your password again in %d day(s).", days))
		HttpRedirect(w, r, "/profile")
		return
	}

	if currentPassword == "" {
		setErrorFlash(w, "Current password is required")
		HttpRedirect(w, r, "/profile")
		return
	}
	if newPassword == "" {
		setErrorFlash(w, "New password is required")
		HttpRedirect(w, r, "/profile")
		return
	}
	if !constants.ValidPassword(newPassword) {
		setErrorFlash(w, "Password must be 8-32 characters with uppercase, lowercase, number, and symbol (!@#$%^&*?)")
		HttpRedirect(w, r, "/profile")
		return
	}
	if newPassword != confirmPassword {
		setErrorFlash(w, "Passwords do not match")
		HttpRedirect(w, r, "/profile")
		return
	}

	storedPassword, err := dbRO.GetQueries().GetTeacherPasswordByID(ctx, user.ID)
	if err != nil {
		logs.Log().Error("get teacher password", zap.Error(err))
		setErrorFlash(w, "Failed to verify current password")
		HttpRedirect(w, r, "/profile")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(currentPassword)); err != nil {
		setErrorFlash(w, "Current password is incorrect")
		HttpRedirect(w, r, "/profile")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logs.Log().Error("hash password", zap.Error(err))
		setErrorFlash(w, "Failed to update password")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := dbRW.GetQueries().UpdateTeacherPassword(ctx, queries.UpdateTeacherPasswordParams{
		Password: string(hashedPassword),
		ID:       user.ID,
	}); err != nil {
		logs.Log().Error("update teacher password", zap.Error(err))
		setErrorFlash(w, "Failed to update password")
		HttpRedirect(w, r, "/profile")
		return
	}

	insertAuditLogAs(ctx, user, "profile", fmt.Sprintf("updated password for teacher '%s'", user.Name))
	setSuccessFlash(w, "Password updated successfully.")
	HttpRedirect(w, r, "/profile")
}

func handleProfileAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if auth.GetRole(ctx) != auth.RoleTeacher {
		HttpError(w, "Access denied", http.StatusForbidden)
		return
	}

	if err := ensureAvatarDir(); err != nil {
		logs.Log().Error("create avatar dir", zap.Error(err))
		setErrorFlash(w, "Failed to prepare upload")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := r.ParseMultipartForm(maxAvatarBytes); err != nil {
		setErrorFlash(w, "File is too large. Maximum size is 2 MB.")
		HttpRedirect(w, r, "/profile")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		setErrorFlash(w, "Please choose an image to upload")
		HttpRedirect(w, r, "/profile")
		return
	}
	defer file.Close()

	ext, err := validateAvatarUpload(file, header.Size)
	if err != nil {
		setErrorFlash(w, err.Error())
		HttpRedirect(w, r, "/profile")
		return
	}

	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, user.ID)
	if err != nil {
		logs.Log().Error("get teacher profile for avatar", zap.Error(err))
		setErrorFlash(w, "Failed to load profile")
		HttpRedirect(w, r, "/profile")
		return
	}

	filename := fmt.Sprintf("%d%s", user.ID, ext)
	destPath := avatarFilePath(filename)

	if row.ProfilePicture.Valid && row.ProfilePicture.String != "" && row.ProfilePicture.String != filename {
		_ = os.Remove(avatarFilePath(row.ProfilePicture.String))
	}

	out, err := os.Create(destPath)
	if err != nil {
		logs.Log().Error("create avatar file", zap.Error(err))
		setErrorFlash(w, "Failed to save profile picture")
		HttpRedirect(w, r, "/profile")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		logs.Log().Error("write avatar file", zap.Error(err))
		_ = os.Remove(destPath)
		setErrorFlash(w, "Failed to save profile picture")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := dbRW.GetQueries().UpdateTeacherProfilePicture(ctx, queries.UpdateTeacherProfilePictureParams{
		ProfilePicture: sql.NullString{String: filename, Valid: true},
		ID:             user.ID,
	}); err != nil {
		logs.Log().Error("update teacher profile picture", zap.Error(err))
		_ = os.Remove(destPath)
		setErrorFlash(w, "Failed to update profile picture")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := logAvatarDocument(ctx, user.ID, header.Filename, filename, ext, header.Size); err != nil {
		logs.Log().Error("log avatar document", zap.Error(err))
	}

	insertAuditLogAs(ctx, user, "profile", fmt.Sprintf("updated profile picture for teacher '%s'", user.Name))
	setSuccessFlash(w, "Profile picture updated successfully.")
	HttpRedirect(w, r, "/profile")
}

func handleProfilePicture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if auth.GetRole(ctx) != auth.RoleTeacher {
		HttpError(w, "Access denied", http.StatusForbidden)
		return
	}

	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, user.ID)
	if err != nil || !row.ProfilePicture.Valid || row.ProfilePicture.String == "" {
		HttpError(w, "Profile picture not found", http.StatusNotFound)
		return
	}

	path := avatarFilePath(row.ProfilePicture.String)
	if _, err := os.Stat(path); err != nil {
		HttpError(w, "Profile picture not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", avatarCacheMaxAge))
	http.ServeFile(w, r, path)
}

func handleTeacherPicture(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		HttpError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teacherID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil || teacherID <= 0 {
		HttpError(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, teacherID)
	if err != nil || !row.ProfilePicture.Valid || row.ProfilePicture.String == "" {
		HttpError(w, "Profile picture not found", http.StatusNotFound)
		return
	}

	path := avatarFilePath(row.ProfilePicture.String)
	if _, err := os.Stat(path); err != nil {
		HttpError(w, "Profile picture not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", avatarCacheMaxAge))
	http.ServeFile(w, r, path)
}

func validateAvatarUpload(file io.ReadSeeker, size int64) (string, error) {
	if size <= 0 {
		return "", errors.New("Uploaded file is empty")
	}
	if size > maxAvatarBytes {
		return "", errors.New("File is too large. Maximum size is 2 MB.")
	}

	cfg, format, err := image.DecodeConfig(file)
	if err != nil {
		return "", errors.New("Invalid image file. Please upload a JPEG or PNG image.")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", errors.New("Failed to read uploaded image")
	}
	_ = cfg

	switch format {
	case "jpeg":
		return ".jpg", nil
	case "png":
		return ".png", nil
	default:
		return "", errors.New("Unsupported image format. Please upload a JPEG or PNG image.")
	}
}

func setErrorFlash(w http.ResponseWriter, msg string) {
	cfg := conf.Conf()
	cookie := &http.Cookie{
		Name:     "error_flash",
		Value:    url.QueryEscape(msg),
		Path:     cfg.BasePath,
		SameSite: http.SameSiteStrictMode,
	}
	if cfg.IsProd() {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
}
