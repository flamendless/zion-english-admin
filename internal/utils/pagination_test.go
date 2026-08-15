package utils

import (
	"testing"
)

func TestPageOffset(t *testing.T) {
	p := Page{Number: 3, Size: 25}
	if p.Offset() != 50 {
		t.Fatalf("expected offset 50, got %d", p.Offset())
	}
}

func TestPageTotalPages(t *testing.T) {
	tests := []struct {
		total int64
		size  int
		want  int
	}{
		{0, 25, 1},
		{25, 25, 1},
		{26, 25, 2},
		{100, 25, 4},
	}
	for _, tc := range tests {
		p := Page{Size: tc.size, Total: tc.total}
		if got := p.TotalPages(); got != tc.want {
			t.Fatalf("total %d size %d: want %d pages, got %d", tc.total, tc.size, tc.want, got)
		}
	}
}

func TestPageHasPrevNext(t *testing.T) {
	p := Page{Number: 2, Size: 25, Total: 75}
	if !p.HasPrev() {
		t.Fatal("expected HasPrev true")
	}
	if !p.HasNext() {
		t.Fatal("expected HasNext true")
	}
	p = Page{Number: 3, Size: 25, Total: 75}
	if p.HasNext() {
		t.Fatal("expected HasNext false on last page")
	}
}
