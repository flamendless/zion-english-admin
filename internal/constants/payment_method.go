package constants

type PaymentMethod string

const (
	PaymentMethodGCash PaymentMethod = "gcash"
)

var PaymentMethods = []PaymentMethod{
	PaymentMethodGCash,
}

func (m PaymentMethod) String() string {
	return string(m)
}

func (m PaymentMethod) Label() string {
	switch m {
	case PaymentMethodGCash:
		return "GCash"
	default:
		return string(m)
	}
}

func ValidPaymentMethod(method string) bool {
	for _, m := range PaymentMethods {
		if string(m) == method {
			return true
		}
	}
	return false
}
