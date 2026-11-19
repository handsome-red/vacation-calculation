package vacation

import (
	"errors"
	"testing"
	"time"
)

// Сохранённые годы берутся как есть, следующие считаются от конца последнего из них.
func TestFindYearStatUsesStoredYears(t *testing.T) {
	hired := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2022, time.February, 1, 0, 0, 0, 0, time.UTC)
	stored := []WorkYear{{Number: 1, Start: ymd(2020, time.January, 1), End: ymd(2021, time.February, 28)}}
	// Прогул в первом году его не продлевает: год сохранён.
	shifts := []Shift{{From: time.Date(2020, time.March, 1, 0, 0, 0, 0, time.UTC), To: time.Date(2020, time.March, 10, 0, 0, 0, 0, time.UTC), Kind: ShiftKindAbsenteeism}}

	stats, err := NewWorkYearCalculator().FindYearStatWithEnds(hired, now, false, shifts, func(time.Time) int { return 0 }, stored, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(stats) != 2 {
		t.Fatalf("years = %d, want 2", len(stats))
	}
	if !stats[0].Stored || DateOf(stats[0].To) != ymd(2021, time.March, 1) {
		t.Errorf("year 1 = %+v", stats[0])
	}
	if stats[1].Stored || DateOf(stats[1].From) != ymd(2021, time.March, 1) || DateOf(stats[1].To) != ymd(2022, time.March, 1) {
		t.Errorf("year 2 = %v..%v", stats[1].From, stats[1].To)
	}
}

func TestValidateWorkYears(t *testing.T) {
	hired := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	cases := map[string][]WorkYear{
		"не с даты приёма": {{Number: 1, Start: ymd(2020, time.January, 2), End: ymd(2020, time.December, 31)}},
		"разрыв":           {{Number: 1, Start: ymd(2020, time.January, 1), End: ymd(2020, time.December, 30)}, {Number: 2, Start: ymd(2021, time.January, 1), End: ymd(2021, time.December, 31)}},
		"пропуск номера":   {{Number: 2, Start: ymd(2020, time.January, 1), End: ymd(2020, time.December, 31)}},
		"конец до начала":  {{Number: 1, Start: ymd(2020, time.January, 1), End: ymd(2019, time.December, 31)}},
	}
	for name, years := range cases {
		if err := ValidateWorkYears(hired, years); !errors.Is(err, ErrWorkYearsInvalid) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	ok := []WorkYear{{Number: 1, Start: ymd(2020, time.January, 1), End: ymd(2020, time.December, 31)}, {Number: 2, Start: ymd(2021, time.January, 1), End: ymd(2021, time.December, 31)}}
	if err := ValidateWorkYears(hired, ok); err != nil {
		t.Errorf("valid years: %v", err)
	}
}
