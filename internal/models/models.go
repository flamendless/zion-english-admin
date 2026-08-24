package models

type ProcessRequest struct {
	TeacherID    string `json:"teacherID"`
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

type LogView struct {
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

type StudentRegisterRequest struct {
	Name             string  `json:"name"`
	Currency         string  `json:"currency"`
	Contact          string  `json:"contact"`
	RatePerClass     float64 `json:"ratePerClass"`
	ParentName       string  `json:"parentName"`
	AssignedColor    string  `json:"assignedColor"`
	Status           string  `json:"status"`
	InactiveReason   string  `json:"inactiveReason"`
	Relationship     string  `json:"relationship"`
	RelatedStudentID int64   `json:"relatedStudentId"`
}

type StudentRegisterResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Logs    []string `json:"logs"`
}

type TeacherRegisterRequest struct {
	FirstName      string  `json:"firstName"`
	MiddleName     string  `json:"middleName"`
	LastName       string  `json:"lastName"`
	Name           string  `json:"name"`
	Birthdate      string  `json:"birthdate"`
	Address        string  `json:"address"`
	JoiningDate    string  `json:"joiningDate"`
	MobileNumber   string  `json:"mobileNumber"`
	Email          string  `json:"email"`
	Certifications string  `json:"certifications"`
	AssignedColor  string  `json:"assignedColor"`
	RatePerClass   float64 `json:"ratePerClass"`
	Currency       string  `json:"currency"`
	DriveUrl       string  `json:"driveUrl"`
	Sex            string  `json:"sex"`
	Password       string  `json:"password"`
	RetypePassword string  `json:"retypePassword"`
}

type TeacherRegisterResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Logs    []string `json:"logs"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

type TeacherAPIResponse struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	DriveUrl     string  `json:"driveUrl"`
	RatePerClass float64 `json:"ratePerClass"`
	Template     string  `json:"template"`
}

type StudentAPIResponse struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Currency     string  `json:"currency"`
	RatePerClass float64 `json:"ratePerClass"`
}

type ClassRecordRequest struct {
	StudentID       int64   `json:"studentId"`
	TeacherID       int64   `json:"teacherId"`
	Date            string  `json:"date"`
	StartTime       string  `json:"startTime"`
	EndTime         string  `json:"endTime"`
	DurationMinutes int64   `json:"durationMinutes"`
	Rate            float64 `json:"rate"`
	Currency        string  `json:"currency"`
	Status          string  `json:"status"`
	Reason          string  `json:"reason"`
	Notes           string  `json:"notes"`
}

type ClassRecordResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Logs    []string `json:"logs"`
}

type ClassRecordView struct {
	ID              int64   `json:"id"`
	StudentID       int64   `json:"studentId"`
	TeacherID       int64   `json:"teacherId"`
	StudentName     string  `json:"studentName"`
	TeacherName     string  `json:"teacherName"`
	Date            string  `json:"date"`
	StartTime       string  `json:"startTime"`
	EndTime         string  `json:"endTime"`
	DurationMinutes int64   `json:"durationMinutes"`
	Rate            float64 `json:"rate"`
	Currency        string  `json:"currency"`
	Status          string  `json:"status"`
	Reason          string  `json:"reason"`
	Notes           string  `json:"notes"`
	CreatedAt       string  `json:"createdAt"`
}

type ScheduledClassRequest struct {
	StudentID       int64   `json:"studentId"`
	TeacherID       int64   `json:"teacherId"`
	ScheduledDate   string  `json:"scheduledDate"`
	StartTime       string  `json:"startTime"`
	DurationMinutes int64   `json:"durationMinutes"`
	Rate            float64 `json:"rate"`
	Currency        string  `json:"currency"`
}

type AvatarView struct {
	Initials      string `json:"initials"`
	AssignedColor string `json:"assignedColor"`
	HasPicture    bool   `json:"hasPicture"`
	PictureURL    string `json:"pictureURL"`
	Alt           string `json:"alt"`
}

type ScheduledClassView struct {
	ID              int64   `json:"id"`
	StudentID       int64   `json:"studentId"`
	TeacherID       int64   `json:"teacherId"`
	StudentName     string  `json:"studentName"`
	TeacherName     string  `json:"teacherName"`
	TeacherAvatar   AvatarView `json:"teacherAvatar"`
	ScheduledDate   string  `json:"scheduledDate"`
	StartTime       string  `json:"startTime"`
	EndTime         string  `json:"endTime"`
	DurationMinutes int64   `json:"durationMinutes"`
	Rate            float64 `json:"rate"`
	Currency        string  `json:"currency"`
	Status          string  `json:"status"`
	Reason          string  `json:"reason"`
	CreatedAt       string  `json:"createdAt"`
	RoomURL         string  `json:"roomUrl"`
	RoomPasscode    string  `json:"roomPasscode"`
	MeetingService  string  `json:"meetingService"`
}

type RecordClassPrefill struct {
	FromSchedule    string
	StudentID       string
	TeacherID       string
	StudentName     string
	TeacherName     string
	Date            string
	StartTime       string
	EndTime         string
	DurationMinutes string
	Rate            string
	Currency        string
	Status          string
	HasPrefill      bool
}
