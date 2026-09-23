package constants

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusReceived PaymentStatus = "received"
)

var PaymentStatuses = []PaymentStatus{
	PaymentStatusPending,
	PaymentStatusReceived,
}

func (s PaymentStatus) String() string {
	return string(s)
}

func (s PaymentStatus) Label() string {
	switch s {
	case PaymentStatusPending:
		return "Pending"
	case PaymentStatusReceived:
		return "Received"
	default:
		return string(s)
	}
}

func ValidPaymentStatus(status string) bool {
	for _, s := range PaymentStatuses {
		if string(s) == status {
			return true
		}
	}
	return false
}
