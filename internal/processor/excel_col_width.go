package processor

import (
	"fmt"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
)

const colWidthPadding = 1.15

type colWidthTracker struct {
	maxLen []float64
	minW   float64
	maxW   []float64
}

func newColWidthTracker(cols int, minW float64, maxW []float64) *colWidthTracker {
	if len(maxW) != cols {
		panic("colWidthTracker: maxW length must match cols")
	}
	return &colWidthTracker{
		maxLen: make([]float64, cols),
		minW:   minW,
		maxW:   maxW,
	}
}

func (t *colWidthTracker) Note(col int, text string) {
	if t == nil || col < 1 || col > len(t.maxLen) {
		return
	}
	n := float64(utf8.RuneCountInString(text))
	if n > t.maxLen[col-1] {
		t.maxLen[col-1] = n
	}
}

func (t *colWidthTracker) NoteFloat(col int, v float64) {
	t.Note(col, fmt.Sprintf("%v", v))
}

func newNameColumnWidthTracker(cols int) *colWidthTracker {
	maxW := make([]float64, cols)
	for i := range maxW {
		maxW[i] = 28
	}
	if cols > 0 {
		maxW[0] = 48
	}
	return newColWidthTracker(cols, 10, maxW)
}

func (t *colWidthTracker) Apply(f *excelize.File, sheet string) error {
	for i, maxLen := range t.maxLen {
		col := i + 1
		width := maxLen*colWidthPadding + 1
		if width < t.minW {
			width = t.minW
		}
		if cap := t.maxW[i]; width > cap {
			width = cap
		}
		colName, err := excelize.ColumnNumberToName(col)
		if err != nil {
			return err
		}
		if err := f.SetColWidth(sheet, colName, colName, width); err != nil {
			return err
		}
	}
	return nil
}
