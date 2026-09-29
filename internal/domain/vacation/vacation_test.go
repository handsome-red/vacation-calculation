package vacation

import (
	"testing"
	"time"
)

// func fullYears(from, to time.Time) int {

func TestFullYears(t *testing.T) {
	tests := []struct {
		name     string
		from, to time.Time
		expected int
	}{
		{
			name:     "base zero year",
			from:     time.Date(2026, time.April, 21, 0, 0, 0, 0, time.Local),
			to:       time.Date(2026, time.September, 21, 0, 0, 0, 0, time.Local),
			expected: 0,
		},
		{
			name:     "base one year",
			from:     time.Date(2026, time.April, 21, 0, 0, 0, 0, time.Local),
			to:       time.Date(2027, time.September, 21, 0, 0, 0, 0, time.Local),
			expected: 1,
		},
		{
			name:     "base",
			from:     time.Date(2026, time.April, 21, 0, 0, 0, 0, time.Local),
			to:       time.Date(2027, time.September, 21, 0, 0, 0, 0, time.Local),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fullYears(tt.from, tt.to); got != tt.expected {
				t.Errorf("got %d; want %v", got, tt.expected)
			}
		})
	}
}

func TestFindYearStat(t *testing.T) {
	tests := []struct {
		name  string
		start time.Time
		end   time.Time
		want  []int
	}{
		{
			name:  "zero years when start equals end",
			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			want:  []int{},
		},
		{
			name:  "one year",
			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			want:  []int{2024},
		},
		{
			name:  "multiple years",
			start: time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			want:  []int{2022, 2023, 2024},
		},
		{
			name:  "same year different months",
			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			want:  []int{2024},
		},
		{
			name:  "start after end returns empty",
			start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			want:  []int{},
		},
		{
			name:  "leap year crossing",
			start: time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
			want:  []int{2020, 2021, 2022},
		},
		{
			name:  "ten years",
			start: time.Date(2015, 6, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			want:  []int{2015, 2016, 2017, 2018, 2019, 2020, 2021, 2022, 2023, 2024},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindYearStat(tt.start, tt.end)

			if len(got) != len(tt.want) {
				t.Fatalf("FindYearStat() len = %d, want %d (got: %+v)", len(got), len(tt.want), got)
			}

			for i, stat := range got {
				if stat.Year != tt.want[i] {
					t.Errorf("FindYearStat()[%d].Year = %d, want %d", i, stat.Year, tt.want[i])
				}
			}
		})
	}
}
