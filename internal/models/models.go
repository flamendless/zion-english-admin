package models

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
	Name          string  `json:"name"`
	Currency      string  `json:"currency"`
	Contact       string  `json:"contact"`
	RatePerClass  float64 `json:"ratePerClass"`
	ParentName    string  `json:"parentName"`
	AssignedColor string  `json:"assignedColor"`
	Status        string  `json:"status"`
}

type StudentRegisterResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Logs    []string `json:"logs"`
}
