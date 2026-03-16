package processor

import (
	"bytes"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testYear = 2026

func TestProcessCSV(t *testing.T) {
	tests := []struct {
		name         string
		csvContent   string
		startDate    time.Time
		endDate      time.Time
		expectedDays []int
		expectError  bool
	}{
		{
			name: "input 1-15, start 1, end 15 - returns all dates",
			csvContent: `1/1/26,,,,,,,
John Doe,26,50,,,,,
1/2/26,,,,,,,
Jane Doe,30,60,,,,,
1/3/26,,,,,,,
Bob Smith,26,50,,,,,
1/4/26,,,,,,,
Alice,45,70,,,,,
1/5/26,,,,,,,
Charlie,30,50,,,,,
1/6/26,,,,,,,
David,60,80,,,,,
1/7/26,,,,,,,
Eve,26,50,,,,,
1/8/26,,,,,,,
Frank,30,60,,,,,
1/9/26,,,,,,,
Grace,45,70,,,,,
1/10/26,,,,,,,
Henry,26,50,,,,,
1/11/26,,,,,,,
Ivy,30,60,,,,,
1/12/26,,,,,,,
Jack,45,70,,,,,
1/13/26,,,,,,,
Kate,26,50,,,,,
1/14/26,,,,,,,
Leo,30,60,,,,,
1/15/26,,,,,,,
Mike,45,70,,,,,`,
			startDate:    time.Date(testYear, time.January, 1, 0, 0, 0, 0, time.UTC),
			endDate:      time.Date(testYear, time.January, 15, 0, 0, 0, 0, time.UTC),
			expectedDays: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			expectError:  false,
		},
		{
			name: "input 2-14, start 1, end 15 - returns only existing dates",
			csvContent: `1/2/26,,,,,,,
John Doe,26,50,,,,,
1/3/26,,,,,,,
Jane Doe,30,60,,,,,
1/4/26,,,,,,,
Bob Smith,26,50,,,,,
1/5/26,,,,,,,
Alice,45,70,,,,,
1/6/26,,,,,,,
Charlie,30,50,,,,,
1/7/26,,,,,,,
David,60,80,,,,,
1/8/26,,,,,,,
Eve,26,50,,,,,
1/9/26,,,,,,,
Frank,30,60,,,,,
1/10/26,,,,,,,
Grace,45,70,,,,,
1/11/26,,,,,,,
Henry,26,50,,,,,
1/12/26,,,,,,,
Ivy,30,60,,,,,
1/13/26,,,,,,,
Jack,45,70,,,,,
1/14/26,,,,,,,
Kate,26,50,,,,,`,
			startDate:    time.Date(testYear, time.January, 1, 0, 0, 0, 0, time.UTC),
			endDate:      time.Date(testYear, time.January, 15, 0, 0, 0, 0, time.UTC),
			expectedDays: []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
			expectError:  false,
		},
		{
			name: "input 1-14, start 2, end 15 - returns 2-14",
			csvContent: `1/1/26,,,,,,,
John Doe,26,50,,,,,
1/2/26,,,,,,,
Jane Doe,30,60,,,,,
1/3/26,,,,,,,
Bob Smith,26,50,,,,,
1/4/26,,,,,,,
Alice,45,70,,,,,
1/5/26,,,,,,,
Charlie,30,50,,,,,
1/6/26,,,,,,,
David,60,80,,,,,
1/7/26,,,,,,,
Eve,26,50,,,,,
1/8/26,,,,,,,
Frank,30,60,,,,,
1/9/26,,,,,,,
Grace,45,70,,,,,
1/10/26,,,,,,,
Henry,26,50,,,,,
1/11/26,,,,,,,
Ivy,30,60,,,,,
1/12/26,,,,,,,
Jack,45,70,,,,,
1/13/26,,,,,,,
Kate,26,50,,,,,
1/14/26,,,,,,,
Leo,30,60,,,,,`,
			startDate:    time.Date(testYear, time.January, 2, 0, 0, 0, 0, time.UTC),
			endDate:      time.Date(testYear, time.January, 15, 0, 0, 0, 0, time.UTC),
			expectedDays: []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
			expectError:  false,
		},
		{
			name: "input 0 to 32 - invalid days are skipped",
			csvContent: `1/0/26,,,,,,,
Invalid Day 0,26,50,,,,,
1/1/26,,,,,,,
Valid,30,60,,,,,
1/32/26,,,,,,,
Invalid Day 32,45,70,,,,,`,
			startDate:    time.Date(testYear, time.January, 1, 0, 0, 0, 0, time.UTC),
			endDate:      time.Date(testYear, time.January, 15, 0, 0, 0, 0, time.UTC),
			expectedDays: []int{1, 1, 1},
			expectError:  false,
		},
		{
			name: "input 3-16, start 1, end 15 - returns 3-15",
			csvContent: `1/3/26,,,,,,,
John Doe,26,50,,,,,
1/4/26,,,,,,,
Jane Doe,30,60,,,,,
1/5/26,,,,,,,
Bob Smith,26,50,,,,,
1/6/26,,,,,,,
Alice,45,70,,,,,
1/7/26,,,,,,,
Charlie,30,50,,,,,
1/8/26,,,,,,,
David,60,80,,,,,
1/9/26,,,,,,,
Eve,26,50,,,,,
1/10/26,,,,,,,
Frank,30,60,,,,,
1/11/26,,,,,,,
Grace,45,70,,,,,
1/12/26,,,,,,,
Henry,26,50,,,,,
1/13/26,,,,,,,
Ivy,30,60,,,,,
1/14/26,,,,,,,
Jack,45,70,,,,,
1/15/26,,,,,,,
Kate,26,50,,,,,
1/16/26,,,,,,,
Leo,30,60,,,,,`,
			startDate:    time.Date(testYear, time.January, 1, 0, 0, 0, 0, time.UTC),
			endDate:      time.Date(testYear, time.January, 15, 0, 0, 0, 0, time.UTC),
			expectedDays: []int{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.csvContent))

			colIndices := ColumnIndices{
				Name:     0,
				Duration: 1,
				Rate:     2,
				Status:   -1,
			}

			records, err := ProcessCSV(reader, tt.startDate, tt.endDate, colIndices, nil)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, records, len(tt.expectedDays))

			actualDays := make([]int, len(records))
			for i, rec := range records {
				parsedDate, err := ParseDateString(rec.Date)
				require.NoError(t, err, "failed to parse date %q", rec.Date)
				actualDays[i] = parsedDate.Day()
			}
			sort.Ints(actualDays)
			sort.Ints(tt.expectedDays)

			assert.Equal(t, tt.expectedDays, actualDays)
		})
	}
}

func TestProcessCSVEdgeCases(t *testing.T) {
	t.Run("empty CSV returns empty records", func(t *testing.T) {
		reader := bytes.NewReader([]byte(""))

		colIndices := ColumnIndices{
			Name:      0,
			Duration:  -1,
			Rate:      -1,
			Status:    -1,
			StartTime: -1,
			EndTime:   -1,
			Link:      -1,
		}

		records, err := ProcessCSV(
			reader,
			time.Date(testYear, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(testYear, time.January, 31, 0, 0, 0, 0, time.UTC),
			colIndices,
			nil,
		)

		require.NoError(t, err)
		assert.Empty(t, records)
	})

	t.Run("start date after all CSV dates", func(t *testing.T) {
		csvContent := `1/1/26,,,,,,,
John,26,50,,,,,
1/2/26,,,,,,,
Jane,30,60,,,,,`
		reader := bytes.NewReader([]byte(csvContent))

		colIndices := ColumnIndices{
			Name:      0,
			Duration:  1,
			Rate:      2,
			Status:    -1,
			StartTime: -1,
			EndTime:   -1,
			Link:      -1,
		}

		records, err := ProcessCSV(
			reader,
			time.Date(testYear, time.January, 10, 0, 0, 0, 0, time.UTC),
			time.Date(testYear, time.January, 20, 0, 0, 0, 0, time.UTC),
			colIndices,
			nil,
		)

		require.NoError(t, err)
		assert.Empty(t, records)
	})

	t.Run("end date before all CSV dates", func(t *testing.T) {
		csvContent := `1/10/26,,,,,,,
John,26,50,,,,,
1/11/26,,,,,,,
Jane,30,60,,,,,`
		reader := bytes.NewReader([]byte(csvContent))

		colIndices := ColumnIndices{
			Name:      0,
			Duration:  1,
			Rate:      2,
			Status:    -1,
			StartTime: -1,
			EndTime:   -1,
			Link:      -1,
		}

		records, err := ProcessCSV(
			reader,
			time.Date(testYear, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(testYear, time.January, 5, 0, 0, 0, 0, time.UTC),
			colIndices,
			nil,
		)

		require.NoError(t, err)
		assert.Empty(t, records)
	})
}

func TestProcessCSVExcludedRows(t *testing.T) {
	csvContent := `1/1/26,,,,,,,
John,26,50,,,,,
1/2/26,,,,,,,
Jane,30,60,,,,,
1/3/26,,,,,,,
Bob,45,70,,,,,
1/4/26,,,,,,,
Alice,26,50,,,,,
1/5/26,,,,,,,
Charlie,30,60,,,,,`
	reader := bytes.NewReader([]byte(csvContent))

	colIndices := ColumnIndices{
		Name:      0,
		Duration:  1,
		Rate:      2,
		Status:    -1,
		StartTime: -1,
		EndTime:   -1,
		Link:      -1,
	}

	excludedRows := map[int]bool{
		2: true,
		6: true,
	}

	records, err := ProcessCSV(
		reader,
		time.Date(testYear, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(testYear, time.January, 5, 0, 0, 0, 0, time.UTC),
		colIndices,
		excludedRows,
	)

	require.NoError(t, err)
	assert.Len(t, records, 3)

	expectedNames := []string{"Alice", "Charlie", "Jane"}
	for i, expectedName := range expectedNames {
		assert.Equal(t, expectedName, records[i].Name)
	}
}
