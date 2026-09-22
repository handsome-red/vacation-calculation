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
