package notifications

import "testing"

func TestCategoryForKindPaymentKinds(t *testing.T) {
	for _, kind := range []string{KindPaymentSent, KindPaymentReceived, KindPaymentConfirmed} {
		cat, ok := CategoryForKind(kind)
		if !ok || cat != CategoryPayments {
			t.Fatalf("kind %q: got category %q ok=%v", kind, cat, ok)
		}
	}
}
