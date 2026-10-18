package calendar

import (
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

func TestBuildYear(t *testing.T) {
	tests := []struct {
		name        string
		year        int
		exceptions  map[date.Date]Exception
		expectedErr bool
	}{
		{
			name:        "2026 без исключений",
			year:        2026,
			exceptions:  nil,
			expectedErr: false,
		},
		{
			name:        "2026 без исключений (пустая карта)",
			year:        2026,
			exceptions:  map[date.Date]Exception{},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildYear(tt.year, tt.exceptions)
			if (err != nil) != tt.expectedErr {
				t.Fatalf("%s: err = %v, expectedErr = %v", tt.name, err, tt.expectedErr)
			}
			if tt.expectedErr {
				return
			}
			if got.Year != tt.year {
				t.Errorf("%s: got year %d, want %d", tt.name, got.Year, tt.year)
			}
			if len(got.Month) != 12 {
				t.Errorf("%s: expected 12 months, got %d", tt.name, len(got.Month))
			}
			for i, m := range got.Month {
				if m.Number != time.Month(i+1) {
					t.Errorf("%s: month[%d] number = %v, want %v", tt.name, i, m.Number, time.Month(i+1))
				}
			}
		})
	}
}
