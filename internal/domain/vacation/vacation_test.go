package vacation

import (
	"fmt"
	"testing"
	"time"
)

// func fullYears(from, to time.Time) int {

// func TestFullYears(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		from, to time.Time
// 		expected int
// 	}{
// 		{
// 			name:     "base zero year",
// 			from:     time.Date(2026, time.April, 21, 0, 0, 0, 0, time.Local),
// 			to:       time.Date(2026, time.September, 21, 0, 0, 0, 0, time.Local),
// 			expected: 0,
// 		},
// 		{
// 			name:     "base one year",
// 			from:     time.Date(2026, time.April, 21, 0, 0, 0, 0, time.Local),
// 			to:       time.Date(2027, time.September, 21, 0, 0, 0, 0, time.Local),
// 			expected: 1,
// 		},
// 		{
// 			name:     "base",
// 			from:     time.Date(2026, time.April, 21, 0, 0, 0, 0, time.Local),
// 			to:       time.Date(2027, time.September, 21, 0, 0, 0, 0, time.Local),
// 			expected: 1,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := fullYears(tt.from, tt.to); got != tt.expected {
// 				t.Errorf("got %d; want %v", got, tt.expected)
// 			}
// 		})
// 	}
// }

// func TestFindYearStat(t *testing.T) {
// 	tests := []struct {
// 		name  string
// 		start time.Time
// 		end   time.Time
// 		want  []int
// 	}{
// 		{
// 			name:  "zero years when start equals end",
// 			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
// 			end:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
// 			want:  []int{},
// 		},
// 		{
// 			name:  "one year",
// 			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
// 			end:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 			want:  []int{2024},
// 		},
// 		{
// 			name:  "multiple years",
// 			start: time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC),
// 			end:   time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
// 			want:  []int{2022, 2023, 2024},
// 		},
// 		{
// 			name:  "same year different months",
// 			start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
// 			end:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
// 			want:  []int{2024},
// 		},
// 		{
// 			name:  "start after end returns empty",
// 			start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
// 			end:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
// 			want:  []int{},
// 		},
// 		{
// 			name:  "leap year crossing",
// 			start: time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC),
// 			end:   time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
// 			want:  []int{2020, 2021, 2022},
// 		},
// 		{
// 			name:  "ten years",
// 			start: time.Date(2015, 6, 1, 0, 0, 0, 0, time.UTC),
// 			end:   time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
// 			want:  []int{2015, 2016, 2017, 2018, 2019, 2020, 2021, 2022, 2023, 2024},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got := FindYearStat(tt.start, tt.end)

// 			if len(got) != len(tt.want) {
// 				t.Fatalf("FindYearStat() len = %d, want %d (got: %+v)", len(got), len(tt.want), got)
// 			}

// 			for i, stat := range got {
// 				if stat.Year != tt.want[i] {
// 					t.Errorf("FindYearStat()[%d].Year = %d, want %d", i, stat.Year, tt.want[i])
// 				}
// 			}
// 		})
// 	}
// }

// func TestCalculate(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		hiredAt  time.Time
// 		now      time.Time
// 		interval []Interval
// 		want     []Interval
// 	}{
// 		{
// 			name:    "base",
// 			hiredAt: time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC),
// 			now:     time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC),
// 			interval: []Interval{
// 				{from: time.Date(2024, time.February, 15, 0, 0, 0, 0, time.UTC),
// 					to: time.Date(2024, time.March, 15, 0, 0, 0, 0, time.UTC)},
// 			},
// 			want: []Interval{
// 				{from: time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC),
// 					to: time.Date(2025, time.February, 15, 0, 0, 0, 0, time.UTC)},
// 			},
// 		},
// 	}

// 	wc := WorkYearCalculator{}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			results := wc.Calculate(tt.hiredAt, tt.now, tt.interval)
// 			if len(results) != len(tt.want) {
// 				t.Errorf("Calculate len results = %v, len want %v", len(results), len(tt.want))
// 			}

// 			for i := range results {
// 				if !results[i].from.Equal(tt.want[i].from) || !results[i].to.Equal(tt.want[i].to) {
// 					t.Errorf("Calculate()\nfrom = %v, from %v\nto = %v, to %v", results[i].from, tt.want[i].from, results[i].to, tt.want[i].to)
// 				}
// 			}
// 		})
// 	}
// }

func TestShift_Days(t *testing.T) {
	tests := []struct {
		name string
		from time.Time
		to   time.Time
		want int
	}{
		{"один день", d(2024, 1, 1), d(2024, 1, 1), 1},
		{"два дня", d(2024, 1, 1), d(2024, 1, 2), 2},
		{"неделя", d(2024, 1, 1), d(2024, 1, 7), 7},
		{"через месяц", d(2024, 1, 1), d(2024, 1, 31), 31},
		{"весь февраль невисокосного", d(2023, 2, 1), d(2023, 2, 28), 28},
		{"весь февраль високосного", d(2024, 2, 1), d(2024, 2, 29), 29},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Shift{Kind: ShiftKindAbsenteeism, From: tt.from, To: tt.to}
			if got := s.Days(); got != tt.want {
				t.Errorf("Days() = %d, хотим %d", got, tt.want)
			}
		})
	}
}

func TestShiftDaysForYear_NoShifts(t *testing.T) {
	yearStart := d(2024, 3, 1)
	yearEnd := d(2025, 3, 1) // исключительно

	got := shiftDaysForYear(yearStart, yearEnd, nil)
	if got != 0 {
		t.Errorf("ожидали 0, получили %d", got)
	}
}

func TestShiftDaysForYear_AbsenteeismInside(t *testing.T) {
	yearStart := d(2024, 3, 1)
	yearEnd := d(2025, 3, 1)

	shifts := []Shift{
		mustShift(t, ShiftKindAbsenteeism, d(2024, 5, 10), d(2024, 5, 12)), // 3 дня
	}

	got := shiftDaysForYear(yearStart, yearEnd, shifts)
	if got != 3 {
		t.Errorf("ожидали 3, получили %d", got)
	}
}

func TestShiftDaysForYear_ParentalLeave(t *testing.T) {
	yearStart := d(2024, 3, 1)
	yearEnd := d(2025, 3, 1)

	shifts := []Shift{
		mustShift(t, ShiftKindParentalLeave, d(2024, 6, 1), d(2024, 8, 31)), // 92 дня
	}

	got := shiftDaysForYear(yearStart, yearEnd, shifts)
	if got != 92 {
		t.Errorf("ожидали 92, получили %d", got)
	}
}

func TestShiftDaysForYear_UnpaidUpTo14(t *testing.T) {
	yearStart := d(2024, 3, 1)
	yearEnd := d(2025, 3, 1)

	cases := []struct {
		name string
		days int
	}{
		{"1 день", 1},
		{"13 дней", 13},
		{"ровно 14 дней", 14},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			to := d(2024, 4, 1).AddDate(0, 0, tc.days-1)
			shifts := []Shift{
				mustShift(t, ShiftKindUnpaid, d(2024, 4, 1), to),
			}
			got := shiftDaysForYear(yearStart, yearEnd, shifts)
			if got != 0 {
				t.Errorf("при %d днях UNPAID ожидали 0 сдвига, получили %d", tc.days, got)
			}
		})
	}
}

func TestShiftDaysForYear_UnpaidOver14(t *testing.T) {
	yearStart := d(2024, 3, 1)
	yearEnd := d(2025, 3, 1)

	cases := []struct {
		days int
		want int
	}{
		{15, 1},
		{20, 6},
		{30, 16},
		{45, 31},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%d дней → сдвиг %d", tc.days, tc.want), func(t *testing.T) {
			to := d(2024, 4, 1).AddDate(0, 0, tc.days-1)
			shifts := []Shift{
				mustShift(t, ShiftKindUnpaid, d(2024, 4, 1), to),
			}
			got := shiftDaysForYear(yearStart, yearEnd, shifts)
			if got != tc.want {
				t.Errorf("при %d днях UNPAID ожидали сдвиг %d, получили %d",
					tc.days, tc.want, got)
			}
		})
	}
}

func TestShiftDaysForYear_UnpaidMultiple(t *testing.T) {
	yearStart := d(2024, 3, 1)
	yearEnd := d(2025, 3, 1)

	// 10 + 10 = 20 → превышение 6.
	shifts := []Shift{
		mustShift(t, ShiftKindUnpaid, d(2024, 4, 1), d(2024, 4, 10)), // 10
		mustShift(t, ShiftKindUnpaid, d(2024, 6, 1), d(2024, 6, 10)), // 10
	}

	got := shiftDaysForYear(yearStart, yearEnd, shifts)
	if got != 6 {
		t.Errorf("ожидали 6, получили %d", got)
	}
}

func TestShiftDaysForYear_UnpaidPerYear(t *testing.T) {
	// 10 дней в первом году, 10 во втором.
	// Ни один не превышает 14 → сдвига нет ни в одном.
	shifts := []Shift{
		mustShift(t, ShiftKindUnpaid, d(2024, 5, 1), d(2024, 5, 10)), // 10
		mustShift(t, ShiftKindUnpaid, d(2025, 5, 1), d(2025, 5, 10)), // 10
	}

	year1 := shiftDaysForYear(d(2024, 3, 1), d(2025, 3, 1), shifts)
	if year1 != 0 {
		t.Errorf("год 1: ожидали 0, получили %d", year1)
	}

	year2 := shiftDaysForYear(d(2025, 3, 1), d(2026, 3, 1), shifts)
	if year2 != 0 {
		t.Errorf("год 2: ожидали 0, получили %d", year2)
	}
}

func TestShiftDaysForYear_OverlapBoundaries(t *testing.T) {
	yearStart := d(2024, 3, 1)
	yearEnd := d(2025, 3, 1)

	// Уход с 01.01.2024 по 31.12.2025 — покрывает год целиком + вылезает.
	shifts := []Shift{
		mustShift(t, ShiftKindParentalLeave, d(2024, 1, 1), d(2025, 12, 31)),
	}

	got := shiftDaysForYear(yearStart, yearEnd, shifts)
	// Ожидаем ровно длину года: 01.03.2024 – 28.02.2025 включительно = 365 дней.
	if got != 365 {
		t.Errorf("ожидали 365, получили %d", got)
	}
}

func TestShiftDaysForYear_OutsideYear(t *testing.T) {
	yearStart := d(2024, 3, 1)
	yearEnd := d(2025, 3, 1)

	shifts := []Shift{
		mustShift(t, ShiftKindAbsenteeism, d(2023, 1, 1), d(2023, 12, 31)), // до
		mustShift(t, ShiftKindAbsenteeism, d(2025, 6, 1), d(2025, 6, 10)),  // после
	}

	got := shiftDaysForYear(yearStart, yearEnd, shifts)
	if got != 0 {
		t.Errorf("ожидали 0, получили %d", got)
	}
}

func TestShiftDaysForYear_Combined(t *testing.T) {
	yearStart := d(2024, 3, 1)
	yearEnd := d(2025, 3, 1)

	shifts := []Shift{
		mustShift(t, ShiftKindAbsenteeism, d(2024, 4, 1), d(2024, 4, 3)),    // 3
		mustShift(t, ShiftKindParentalLeave, d(2024, 6, 1), d(2024, 8, 31)), // 92
		mustShift(t, ShiftKindUnpaid, d(2024, 10, 1), d(2024, 10, 20)),      // 20 → +6
	}

	got := shiftDaysForYear(yearStart, yearEnd, shifts)
	// 3 + 92 + (20-14) = 101
	if got != 101 {
		t.Errorf("ожидали 101, получили %d", got)
	}
}

func TestFindYearStat_AbsenteeismShifts(t *testing.T) {
	hiredAt := d(2024, 3, 1)
	now := d(2025, 6, 1)

	shifts := []Shift{
		mustShift(t, ShiftKindAbsenteeism, d(2024, 4, 1), d(2024, 4, 10)), // 10
	}

	got := NewWorkYearCalculator().FindYearStat(hiredAt, now, false, shifts,
		func(at time.Time) int { return 0 })

	if len(got) != 2 {
		t.Fatalf("ожидали 2 года, получили %d", len(got))
	}

	// Год 1: 01.03.2024 – 11.03.2025 (сдвиг +10)
	if !got[0].From.Equal(d(2024, 3, 1)) {
		t.Errorf("начало года 1: ожидали 01.03.2024, получили %s", got[0].From)
	}
	if !got[0].To.Equal(d(2025, 3, 11)) {
		t.Errorf("конец года 1: ожидали 11.03.2025, получили %s", got[0].To)
	}

	// Год 2: 11.03.2025 – 11.03.2026 (сдвигов нет)
	if !got[1].From.Equal(d(2025, 3, 11)) {
		t.Errorf("начало года 2: ожидали 11.03.2025, получили %s", got[1].From)
	}
	if !got[1].To.Equal(d(2026, 3, 11)) {
		t.Errorf("конец года 2: ожидали 11.03.2026, получили %s", got[1].To)
	}
}

func d(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// mustShift — конструктор Shift с проверкой инвариантов.
func mustShift(t *testing.T, kind ShiftKind, from, to time.Time) Shift {
	t.Helper()
	s := Shift{Kind: kind, From: from, To: to}
	if s.To.Before(s.From) {
		t.Fatalf("некорректный shift: To (%s) до From (%s)", s.To, s.From)
	}
	return s
}
