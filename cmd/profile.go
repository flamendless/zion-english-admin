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
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"zion-english/frontend"
	"zion-english/internal/auth"
	"zion-english/internal/notifications"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/storage"
	"zion-english/internal/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxAvatarBytes    = 2 << 20
	avatarCacheMaxAge = 3600
)

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
		assignedColor = constants.DefaultTeacherAssignedColor
	}
	displayName := utils.ComposePersonName(row.FirstName, row.MiddleName, row.LastName)
	return frontend.AvatarProps{
		Size:          "xl",
		Initials:      utils.PersonInitials(row.FirstName, row.MiddleName, row.LastName, displayName),
		AssignedColor: assignedColor,
		PictureURL:    teacherPictureURL(row.ID, hasPicture),
		HasPicture:    hasPicture,
		Alt:           displayName + " avatar",
	}
}

func buildTeacherListAvatarProps(teacherID int64, firstName, middleName, lastName, assignedColor string, profilePicture sql.NullString) frontend.AvatarProps {
	hasPicture := profilePicture.Valid && profilePicture.String != ""
	if assignedColor == "" {
		assignedColor = constants.DefaultTeacherAssignedColor
	}
	displayName := utils.ComposePersonName(firstName, middleName, lastName)
	return frontend.AvatarProps{
		Size:          "sm",
		Initials:      utils.PersonInitials(firstName, middleName, lastName, displayName),
		AssignedColor: assignedColor,
		PictureURL:    teacherPictureURL(teacherID, hasPicture),
		HasPicture:    hasPicture,
		Alt:           displayName + " avatar",
	}
}

func buildSuperuserAvatarProps(user auth.User) frontend.AvatarProps {
	return frontend.AvatarProps{
		Size:          "xl",
		Initials:      utils.PersonInitials("", "", "", user.Name),
		AssignedColor: constants.DefaultAssignedColor,
		HasPicture:    false,
		Alt:           user.Name + " avatar",
	}
}

func buildHeaderAvatarProps(ctx context.Context, user auth.User, role auth.Role) (frontend.AvatarProps, error) {
	if role == auth.RoleSuperuser {
		props := frontend.WithSuperuserBadge(buildSuperuserAvatarProps(user))
		props.Size = "nav"
		return props, nil
	}

	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, user.ID)
	if err != nil {
		return frontend.AvatarProps{}, err
	}
	props := buildTeacherAvatarProps(row)
	props.Size = "nav"

	roleStrings, err := loadTeacherRoles(ctx, user.ID)
	if err != nil {
		return frontend.AvatarProps{}, err
	}
	return avatarWithTeacherRoles(props, roleStrings), nil
}

func handleHeaderAvatar(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
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

	writeHTML(w)
	if err := frontend.HeaderAvatar(props).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
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
			Avatar:      frontend.WithSuperuserBadge(buildSuperuserAvatarProps(user)),
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
	blockingDocs, err := dbRO.GetQueries().HasBlockingTeacherDocument(ctx, queries.HasBlockingTeacherDocumentParams{
		TeacherID: user.ID,
		Type:      string(constants.TeacherDocumentTypeDocument),
	})
	if err != nil {
		logs.Log().Error("check blocking teacher document", zap.Error(err))
		blockingDocs = 0
	}
	canUploadResume := true
	resumeUploadDays := 0
	hasResume := false
	resumeFilename := ""
	resumeViewURL := ""
	resumeDoc, resumeErr := dbRO.GetQueries().GetLatestTeacherDocumentByTeacherIDAndType(ctx, queries.GetLatestTeacherDocumentByTeacherIDAndTypeParams{
		TeacherID: user.ID,
		Type:      string(constants.TeacherDocumentTypeResume),
	})
	if resumeErr == nil {
		canUploadResume, resumeUploadDays = utils.ResumeUploadAllowed(resumeDoc.UploadedAt, now)
		hasResume = true
		resumeFilename = resumeDoc.OriginalFilename
		resumeViewURL = utils.URL(fmt.Sprintf("/documents/%d/file", resumeDoc.ID))
	} else if !errors.Is(resumeErr, sql.ErrNoRows) {
		logs.Log().Error("get latest teacher resume", zap.Error(resumeErr))
	}
	blockingIntroVideo, err := dbRO.GetQueries().HasBlockingTeacherIntroVideo(ctx, user.ID)
	if err != nil {
		logs.Log().Error("check blocking teacher intro video", zap.Error(err))
		blockingIntroVideo = 0
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

	roleStrings, err := loadTeacherRoles(ctx, user.ID)
	if err != nil {
		logs.Log().Error("get teacher roles", zap.Error(err))
		HttpError(w, "Failed to load profile", http.StatusInternalServerError)
		return
	}

	data := frontend.ProfileData{
		IsSuperuser:           false,
		Name:                  utils.ComposePersonName(row.FirstName, row.MiddleName, row.LastName),
		Email:                 row.Email,
		Roles:                 roleStrings,
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
		Status:                constants.TeacherStatus(row.Status),
		HasProfilePicture:     row.ProfilePicture.Valid && row.ProfilePicture.String != "",
		Avatar:                avatarWithTeacherRoles(buildTeacherAvatarProps(row), roleStrings),
		CanChangeMobile:       canChangeMobile,
		MobileDaysRemaining:   mobileDays,
		CanChangePassword:     canChangePassword,
		PasswordDaysRemaining: passwordDays,
		CanEditFirstName:      utils.ProfileNameEditable(row.FirstName),
		CanEditMiddleName:     utils.ProfileNameEditable(row.MiddleName),
		CanEditLastName:       utils.ProfileNameEditable(row.LastName),
		CanUploadDocument:     blockingDocs == 0,
		CanUploadResume:           canUploadResume,
		ResumeUploadDaysRemaining: resumeUploadDays,
		HasResume:                 hasResume,
		ResumeFilename:            resumeFilename,
		ResumeViewURL:             resumeViewURL,
	}
	if introVideo, err := dbRO.GetQueries().GetLatestTeacherIntroVideoByTeacherID(ctx, user.ID); err == nil {
		data.HasIntroVideo = true
		data.IntroVideoStatus = constants.TeacherIntroVideoStatus(introVideo.Status)
		data.IntroVideoViewURL = introVideoViewURL(introVideo.ID, introVideo.Url)
		data.IntroVideoSourceType = introVideoSourceType(introVideo.SourceType)
		if introVideo.RejectReason.Valid {
			data.IntroVideoRejectReason = introVideo.RejectReason.String
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		logs.Log().Error("get latest teacher intro video", zap.Error(err))
	}
	introVisible, introUploadAllowed := introVideoUploadAccessForViewer(ctx)
	data.IntroVideoUploadVisible = data.HasIntroVideo || introVisible
	data.IntroVideoUploadsAllowed = introUploadAllowed
	data.CanUploadIntroVideo = introUploadAllowed && blockingIntroVideo == 0
	zoomConnected, zoomConfigured, zoomVisible, zoomConnectionsAllowed := profileZoomStatus(ctx, user.ID)
	data.ZoomConfigured = zoomConfigured
	data.ZoomConnected = zoomConnected
	data.ZoomIntegrationVisible = zoomVisible
	data.ZoomConnectionsAllowed = zoomConnectionsAllowed
	data.ZoomConnectURL = utils.URL("/profile/zoom/connect")
	data.ZoomDisconnectURL = utils.URL("/profile/zoom/disconnect")
	data.ZoomGuideURL = utils.URL("/guides/" + string(constants.GuideSlugConnectZoom))
	data.ZoomStatusMessage = profileZoomFlashMessage(r.URL.Query())

	googleCalendarConnected, googleCalendarConfigured, googleCalendarVisible, googleCalendarConnectionsAllowed := profileGoogleCalendarStatus(ctx, user.ID)
	data.GoogleCalendarConfigured = googleCalendarConfigured
	data.GoogleCalendarConnected = googleCalendarConnected
	data.GoogleCalendarIntegrationVisible = googleCalendarVisible
	data.GoogleCalendarConnectionsAllowed = googleCalendarConnectionsAllowed
	data.GoogleCalendarConnectURL = utils.URL("/profile/google-calendar/connect")
	data.GoogleCalendarDisconnectURL = utils.URL("/profile/google-calendar/disconnect")
	data.GoogleCalendarGuideURL = utils.URL("/guides/" + string(constants.GuideSlugConnectGoogleCalendar))
	data.GoogleCalendarStatusMessage = profileGoogleCalendarFlashMessage(r.URL.Query())

	if err := frontend.Profile(data).Render(ctx, w); err != nil {
		HttpError(w, err.Error(), http.StatusInternalServerError)
	}
}

func profileZoomFlashMessage(query url.Values) string {
	if query.Get("zoom_connected") == "1" {
		return "Zoom account connected successfully."
	}
	if query.Get("zoom_disconnected") == "1" {
		return "Zoom account disconnected."
	}
	switch constants.IntegrationOAuthError(query.Get("zoom_error")) {
	case constants.IntegrationOAuthErrorInvalidState:
		return "Zoom connection failed: invalid session. Please try again."
	case constants.IntegrationOAuthErrorMissingCode:
		return "Zoom connection failed: authorization was not completed."
	case constants.IntegrationOAuthErrorExchangeFailed:
		return "Zoom connection failed: could not exchange authorization code."
	case constants.IntegrationOAuthErrorSaveFailed:
		return "Zoom connection failed: could not save account."
	case "":
		return ""
	default:
		return "Zoom connection failed: " + query.Get("zoom_error")
	}
}

func profileGoogleCalendarFlashMessage(query url.Values) string {
	if query.Get("google_calendar_connected") == "1" {
		return "Google Calendar connected successfully."
	}
	if query.Get("google_calendar_disconnected") == "1" {
		return "Google Calendar disconnected."
	}
	switch constants.IntegrationOAuthError(query.Get("google_calendar_error")) {
	case constants.IntegrationOAuthErrorInvalidState:
		return "Google Calendar connection failed: invalid session. Please try again."
	case constants.IntegrationOAuthErrorMissingCode:
		return "Google Calendar connection failed: authorization was not completed."
	case constants.IntegrationOAuthErrorExchangeFailed:
		return "Google Calendar connection failed: could not exchange authorization code."
	case constants.IntegrationOAuthErrorSaveFailed:
		return "Google Calendar connection failed: could not save account."
	case "":
		return ""
	default:
		return "Google Calendar connection failed: " + query.Get("google_calendar_error")
	}
}

func handleProfileMobile(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.IsTeacherScoped(auth.GetRole(ctx)) {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
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
	if !utils.ValidMobileNumber(mobileNumber) {
		setErrorFlash(w, ErrInvalidMobileNumber.Error())
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
	notifySuperuser(ctx, user, notifications.KindProfileUpdated, fmt.Sprintf("Teacher '%s' updated their mobile number", user.Name), "")
	setSuccessFlash(w, "Mobile number updated successfully.")
	HttpRedirect(w, r, "/profile")
}

func handleProfileNames(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.IsTeacherScoped(auth.GetRole(ctx)) {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
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
	notifySuperuser(ctx, user, notifications.KindProfileUpdated, fmt.Sprintf("Teacher '%s' updated their name", name), "")
	setSuccessFlash(w, "Name updated successfully.")
	HttpRedirect(w, r, "/profile")
}

func handleProfilePassword(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.IsTeacherScoped(auth.GetRole(ctx)) {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
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
	notifySuperuser(ctx, user, notifications.KindProfileUpdated, fmt.Sprintf("Teacher '%s' updated their password", user.Name), "")
	setSuccessFlash(w, "Password updated successfully.")
	HttpRedirect(w, r, "/profile")
}

func handleProfileAvatar(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.IsTeacherScoped(auth.GetRole(ctx)) {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
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
	store := storage.Default()

	if row.ProfilePicture.Valid && row.ProfilePicture.String != "" && row.ProfilePicture.String != filename {
		_ = store.Delete(ctx, storage.CategoryAvatars, row.ProfilePicture.String)
	}

	if err := store.Put(ctx, storage.CategoryAvatars, filename, file, avatarContentType(filename)); err != nil {
		logs.Log().Error("write avatar file", zap.Error(err))
		insertUploadLog(ctx, user, "profile", constants.SystemLogUploadOutcomeFailed, fmt.Sprintf("profile picture storage failed for teacher '%s' (id %d), file '%s': %v", user.Name, user.ID, filepath.Base(header.Filename), err))
		setErrorFlash(w, "Failed to save profile picture")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := dbRW.GetQueries().UpdateTeacherProfilePicture(ctx, queries.UpdateTeacherProfilePictureParams{
		ProfilePicture: sql.NullString{String: filename, Valid: true},
		ID:             user.ID,
	}); err != nil {
		logs.Log().Error("update teacher profile picture", zap.Error(err))
		_ = store.Delete(ctx, storage.CategoryAvatars, filename)
		insertUploadLog(ctx, user, "profile", constants.SystemLogUploadOutcomeFailed, fmt.Sprintf("profile picture database update failed for teacher '%s' (id %d), file '%s': %v", user.Name, user.ID, filepath.Base(header.Filename), err))
		setErrorFlash(w, "Failed to update profile picture")
		HttpRedirect(w, r, "/profile")
		return
	}

	if err := logAvatarDocument(ctx, user.ID, header.Filename, filename, ext, header.Size); err != nil {
		logs.Log().Error("log avatar document", zap.Error(err))
	}

	insertUploadLog(ctx, user, "profile", constants.SystemLogUploadOutcomeSucceeded, fmt.Sprintf("profile picture for teacher '%s' (id %d), file '%s'", user.Name, user.ID, filepath.Base(header.Filename)))
	insertAuditLogAs(ctx, user, "profile", fmt.Sprintf("updated profile picture for teacher '%s'", user.Name))
	notifySuperuser(ctx, user, notifications.KindProfileUpdated, fmt.Sprintf("Teacher '%s' updated their profile picture", user.Name), "")
	setSuccessFlash(w, "Profile picture updated successfully.")
	HttpRedirect(w, r, "/profile")
}

func handleProfilePicture(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	ctx := r.Context()
	user := auth.GetUser(ctx)
	if !auth.IsTeacherScoped(auth.GetRole(ctx)) {
		HttpError(w, MsgAccessDenied, http.StatusForbidden)
		return
	}

	row, err := dbRO.GetQueries().GetTeacherProfileByID(ctx, user.ID)
	if err != nil || !row.ProfilePicture.Valid || row.ProfilePicture.String == "" {
		HttpError(w, "Profile picture not found", http.StatusNotFound)
		return
	}

	obj, err := storage.Default().Get(ctx, storage.CategoryAvatars, row.ProfilePicture.String)
	if err != nil {
		HttpError(w, "Profile picture not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", avatarCacheMaxAge))
	serveStorageObject(w, obj, avatarContentType(row.ProfilePicture.String), nil)
}

func handleTeacherPicture(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
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

	obj, err := storage.Default().Get(ctx, storage.CategoryAvatars, row.ProfilePicture.String)
	if err != nil {
		HttpError(w, "Profile picture not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", avatarCacheMaxAge))
	serveStorageObject(w, obj, avatarContentType(row.ProfilePicture.String), nil)
}

func validateAvatarUpload(file io.ReadSeeker, size int64) (string, error) {
	if size <= 0 {
		return "", ErrAvatarUploadedFileEmpty
	}
	if size > maxAvatarBytes {
		return "", ErrAvatarFileTooLarge
	}

	cfg, format, err := image.DecodeConfig(file)
	if err != nil {
		return "", ErrInvalidAvatarImage
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", ErrFailedToReadUploadedImage
	}
	_ = cfg

	switch format {
	case "jpeg":
		return ".jpg", nil
	case "png":
		return ".png", nil
	default:
		return "", ErrUnsupportedAvatarImageFormat
	}
}

