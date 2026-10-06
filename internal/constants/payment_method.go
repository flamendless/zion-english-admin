package constants

type PaymentMethod string

const (
	PaymentMethodGCash      PaymentMethod = "gcash"
	PaymentMethodBPI        PaymentMethod = "bpi"
	PaymentMethodGoTyme     PaymentMethod = "gotyme"
	PaymentMethodPNB        PaymentMethod = "pnb"
	PaymentMethodUnionBank  PaymentMethod = "unionbank"
	PaymentMethodBDO        PaymentMethod = "bdo"
	PaymentMethodMaya       PaymentMethod = "maya"
)

var PaymentMethods = []PaymentMethod{
	PaymentMethodGCash,
	PaymentMethodBPI,
	PaymentMethodGoTyme,
	PaymentMethodPNB,
	PaymentMethodUnionBank,
	PaymentMethodBDO,
	PaymentMethodMaya,
}

func (m PaymentMethod) String() string {
	return string(m)
}

func (m PaymentMethod) Label() string {
	switch m {
	case PaymentMethodGCash:
		return "GCash"
	case PaymentMethodBPI:
		return "BPI"
	case PaymentMethodGoTyme:
		return "GoTyme"
	case PaymentMethodPNB:
		return "PNB"
	case PaymentMethodUnionBank:
		return "UnionBank"
	case PaymentMethodBDO:
		return "BDO"
	case PaymentMethodMaya:
		return "Maya"
	default:
		return string(m)
	}
}

// LogoFilename is served from /static/.
func (m PaymentMethod) LogoFilename() string {
	switch m {
	case PaymentMethodGCash:
		return "payment-gcash.svg"
	case PaymentMethodBPI:
		return "payment-bpi.svg"
	case PaymentMethodGoTyme:
		return "payment-gotyme.svg"
	case PaymentMethodPNB:
		return "payment-pnb.svg"
	case PaymentMethodUnionBank:
		return "payment-unionbank.svg"
	case PaymentMethodBDO:
		return "payment-bdo.svg"
	case PaymentMethodMaya:
		return "payment-maya.svg"
	default:
		return ""
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
