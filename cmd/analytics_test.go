package cmd

import "testing"

func TestCancellationRate(t *testing.T) {
	if got := cancellationRate(2, 8, 0); got != 20 {
		t.Fatalf("expected 20, got %v", got)
	}
	if got := cancellationRate(0, 0, 0); got != 0 {
		t.Fatalf("expected 0 for empty total, got %v", got)
	}
}

func TestUtilizationPct(t *testing.T) {
	if got := utilizationPct(45, 60); got != 75 {
		t.Fatalf("expected 75, got %v", got)
	}
	if got := utilizationPct(10, 0); got != 0 {
		t.Fatalf("expected 0 when scheduled is zero, got %v", got)
	}
}

func TestMedianInt64(t *testing.T) {
	if got := medianInt64([]int64{10, 30, 20}); got != 20 {
		t.Fatalf("expected 20, got %d", got)
	}
	if got := medianInt64([]int64{10, 20}); got != 15 {
		t.Fatalf("expected 15, got %d", got)
	}
	if got := medianInt64(nil); got != 0 {
		t.Fatalf("expected 0 for empty slice, got %d", got)
	}
}
