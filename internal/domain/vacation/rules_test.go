package vacation

import (
	"errors"
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

func TestWorkedMonths(t *testing.T) {
	from := time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC)
	day := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }
	for _, tt := range []struct {
		name string
		now  time.Time
		excl int
		want int
	}{
		{"первый день", day(time.March, 10), 0, 0},
		{"14 дней — вниз", day(time.March, 23), 0, 0},
		{"15 дней — вверх", day(time.March, 24), 0, 1},
		{"ровно месяц", day(time.April, 9), 0, 1},
		{"7 мес. 14 дн. — 7", day(time.October, 23), 0, 7},
		{"7 мес. 15 дн. — 8", day(time.October, 24), 0, 8},
		{"исключаемые дни уменьшают стаж", day(time.October, 24), 1, 7},
		{"не больше 12", day(time.December, 31).AddDate(1, 0, 0), 0, 12},
	} {
		if got := WorkedMonths(from, tt.now, tt.excl); got != tt.want {
			t.Errorf("%s: WorkedMonths = %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestSeniorityOnlyForCivilServants(t *testing.T) {
	hired := time.Date(2020, time.March, 10, 0, 0, 0, 0, time.UTC)
	calc := NewWorkYearCalculator()
	at := func(y int) time.Time { return hired.AddDate(y, 0, 0) }

	civil := calc.SeniorityAtFromHired(hired, 0, true)
	for years, want := range map[int]int{0: 1, 1: 1, 4: 1, 5: 5, 9: 5, 10: 7, 14: 7, 15: 10, 30: 10} {
		if got := civil(at(years)); got != want {
			t.Errorf("госслужащий, стаж %d лет: %d дн., want %d", years, got, want)
		}
	}
	// стаж до приёма учитывается: 3 года работы + 2 года (730 дн.) до приёма = 5 лет
	if got := calc.SeniorityAtFromHired(hired, 730, true)(at(3)); got != 5 {
		t.Errorf("со стажем до приёма: %d, want 5", got)
	}
	if got := calc.SeniorityAtFromHired(hired, 0, false)(at(20)); got != 0 {
		t.Errorf("прочий персонал: %d, want 0", got)
	}

	if !user.PositionSpecialistExpert.IsCivilServant() || user.PositionDriver.IsCivilServant() || user.PositionSoftwareEngineer.IsCivilServant() {
		t.Error("IsCivilServant: специалист-эксперт — госслужащий, водитель и инженер-программист — нет")
	}
}

func TestUnpaidLeaveShiftsWorkYearAndWarns(t *testing.T) {
	uid := user.GenerateUserID()
	leave, _ := NewVacation(NewVacationID(), uid, KindUnpaid, ymd(2026, time.June, 1), ymd(2026, time.June, 20), ymd(2026, time.March, 10), time.Now())
	draft, _ := NewVacation(NewVacationID(), uid, KindUnpaid, ymd(2026, time.July, 1), ymd(2026, time.July, 30), ymd(2026, time.March, 10), time.Now())
	_ = leave.Approve(time.Now())

	shifts := WithUnpaidLeaves(nil, []*Vacation{leave, draft})
	if len(shifts) != 1 {
		t.Fatalf("only approved unpaid leave shifts the year, got %d shifts", len(shifts))
	}

	stats, err := NewWorkYearCalculator().FindYearStat(
		time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC), time.Date(2026, time.October, 30, 0, 0, 0, 0, time.UTC),
		false, shifts, func(time.Time) int { return 0 })
	if err != nil {
		t.Fatal(err)
	}
	// 20 дней без сохранения: 14 не сдвигают, 6 — сдвигают рабочий год
	if want := time.Date(2027, time.March, 16, 0, 0, 0, 0, time.UTC); !stats[0].To.Equal(want) {
		t.Errorf("year end = %s, want %s", stats[0].To.Format("02.01.2006"), want.Format("02.01.2006"))
	}
	w := UnpaidWarnings(stats, shifts)
	if len(w) != 1 || w[0].Days != 20 || w[0].Shifted != 6 {
		t.Fatalf("warnings = %+v, want 20 days, shifted 6", w)
	}

	if used := UsedDaysInYear([]*Vacation{leave}, stats[0], Holidays{}); used.Base+used.Irregular+used.Seniority != 0 {
		t.Error("unpaid leave must not use annual quota")
	}
}

func TestValidateCreationParts(t *testing.T) {
	year := yearStat(2026, 30, 0, 0)
	today := ymd(2026, time.April, 1)
	may := Holidays{ymd(2026, time.May, 1): {}, ymd(2026, time.May, 9): {}}

	// 01.05–07.05: 7 календарных дней, но 1 мая — праздник → 6 дней отпуска
	v, err := ValidateCreation([]CreationPeriod{{Start: ymd(2026, time.May, 1), End: ymd(2026, time.May, 7)}}, year, nil, may, today, false)
	if err != nil || len(v) == 0 || !errors.Is(v[0], ErrTooShort) {
		t.Errorf("праздник не входит в длительность части: violations=%v err=%v", v, err)
	}

	// одна часть 14 дней — допустимо
	v, _ = ValidateCreation([]CreationPeriod{{Start: ymd(2026, time.June, 1), End: ymd(2026, time.June, 14)}}, year, nil, Holidays{}, today, false)
	if len(v) != 0 {
		t.Errorf("14 дней одним куском: %v", v)
	}

	// прошедшие даты
	v, _ = ValidateCreation([]CreationPeriod{{Start: ymd(2026, time.March, 20), End: ymd(2026, time.April, 5)}}, year, nil, Holidays{}, today, false)
	if len(v) == 0 || !errors.Is(v[0], ErrInPast) {
		t.Errorf("прошедшие даты: %v", v)
	}

	// перерыв 13 дней между частями — нарушение (нужно 14)
	uid := user.GenerateUserID()
	existing, _ := NewVacation(NewVacationID(), uid, KindBase, ymd(2026, time.June, 1), ymd(2026, time.June, 14), ymd(2026, time.March, 10), time.Now())
	v, _ = ValidateCreation([]CreationPeriod{{Start: ymd(2026, time.June, 28), End: ymd(2026, time.July, 4)}}, year, []*Vacation{existing}, Holidays{}, today, false)
	if len(v) == 0 || !errors.Is(v[0], ErrGapTooSmall) {
		t.Errorf("перерыв 13 дней: %v", v)
	}
	v, _ = ValidateCreation([]CreationPeriod{{Start: ymd(2026, time.June, 29), End: ymd(2026, time.July, 5)}}, year, []*Vacation{existing}, Holidays{}, today, false)
	if len(v) != 0 {
		t.Errorf("перерыв 14 дней допустим: %v", v)
	}

	// административный отпуск: короткий допустим, пересечение — нет
	v, err = ValidateCreation([]CreationPeriod{{Start: ymd(2026, time.June, 20), End: ymd(2026, time.June, 21)}}, year, []*Vacation{existing}, Holidays{}, today, true)
	if err != nil || len(v) != 0 {
		t.Errorf("административный 2 дня: violations=%v err=%v", v, err)
	}
	if _, err = ValidateCreation([]CreationPeriod{{Start: ymd(2026, time.June, 10), End: ymd(2026, time.June, 21)}}, year, []*Vacation{existing}, Holidays{}, today, true); !errors.Is(err, ErrOverlapsExisting) {
		t.Errorf("административный поверх отпуска: err=%v", err)
	}
}
