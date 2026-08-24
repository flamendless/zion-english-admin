package meetings

const (
	ServiceZoom = "zoom"

	MaxAutoZoomDurationMinutes int64 = 40
)

const ManualZoomWarning = "Classes longer than 40 minutes cannot receive an automatic Zoom meeting room on a Basic account. Please create your Zoom meeting manually and share the link with your student."

func SupportsAutoRoom(durationMinutes int64) bool {
	return durationMinutes > 0 && durationMinutes <= MaxAutoZoomDurationMinutes
}
