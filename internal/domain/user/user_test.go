// internal/domain/user/experience_test.go
package user

import (
	"testing"
	"time"
)

func d(y int, m time.Month, day int) time.Time {
	return time.Date(y, m, day, 0, 0, 0, 0, time.UTC)
}

func TestSplitExperience(t *testing.T) {
	tests := []struct {
		name                string
		from, to            time.Time
		wantY, wantM, wantD int
	}{
		{
			name: "same day",
			from: d(2024, 1, 15), to: d(2024, 1, 15),
			wantY: 0, wantM: 0, wantD: 0,
		},
		{
			name: "one day",
			from: d(2024, 1, 15), to: d(2024, 1, 16),
			wantY: 0, wantM: 0, wantD: 1,
		},
		{
			name: "one month",
			from: d(2024, 1, 15), to: d(2024, 2, 15),
			wantY: 0, wantM: 1, wantD: 0,
		},
		{
			name: "one year",
			from: d(2023, 3, 1), to: d(2024, 3, 1),
			wantY: 1, wantM: 0, wantD: 0,
		},
		{
			name: "5 months 7 days",
			from: d(2026, 4, 15), to: d(2026, 9, 22),
			wantY: 0, wantM: 5, wantD: 7,
		},
		{
			name: "end of month borrow",
			from: d(2024, 1, 31), to: d(2024, 3, 1),
			wantY: 0, wantM: 1, wantD: 1,
		},
		{
			name: "leap year anniversary not reached",
			from: d(2024, 2, 29), to: d(2025, 2, 28),
			wantY: 0, wantM: 11, wantD: 30,
		},
		{
			name: "leap year anniversary reached",
			from: d(2024, 2, 29), to: d(2025, 3, 1),
			wantY: 1, wantM: 0, wantD: 0,
		},
		{
			name: "to before from",
			from: d(2025, 1, 1), to: d(2024, 1, 1),
			wantY: 0, wantM: 0, wantD: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y, m, day := splitExperience(tt.from, tt.to)
			if y != tt.wantY || m != tt.wantM || day != tt.wantD {
				t.Errorf("got %d/%d/%d, want %d/%d/%d",
					y, m, day, tt.wantY, tt.wantM, tt.wantD)
			}
		})
	}
}
