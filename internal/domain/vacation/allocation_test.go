package vacation

import (
	"errors"
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

func ymd(y int, m time.Month, d int) date.Date { return date.New(y, m, d) }

func yearStat(y int, base, irregular, seniority int) YearStat {
	return YearStat{
		From:      time.Date(y, time.March, 10, 0, 0, 0, 0, time.UTC),
		To:        time.Date(y+1, time.March, 10, 0, 0, 0, 0, time.UTC),
		Base:      base,
		Irregular: irregular,
		Seniority: seniority,
	}
}

func TestAllocate_OldestYearFirstThenKindOrder(t *testing.T) {
	balances := []YearBalance{
		{Year: yearStat(2024, 30, 3, 1), Remaining: map[Kind]int{KindBase: 2, KindIrregular: 1, KindSeniority: 1}},
		{Year: yearStat(2025, 30, 3, 1), Remaining: map[Kind]int{KindBase: 30, KindIrregular: 3, KindSeniority: 1}},
	}
	periods := []CreationPeriod{{Start: ymd(2026, time.November, 9), End: ymd(2026, time.November, 15)}}

	got, _, err := Allocate(periods, balances, Holidays{}, false)
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}

	want := []Allocation{
		{Kind: KindBase, WorkYearStart: ymd(2024, time.March, 10), Start: ymd(2026, time.November, 9), End: ymd(2026, time.November, 10), Days: 2},
		{Kind: KindIrregular, WorkYearStart: ymd(2024, time.March, 10), Start: ymd(2026, time.November, 11), End: ymd(2026, time.November, 11), Days: 1},
		{Kind: KindSeniority, WorkYearStart: ymd(2024, time.March, 10), Start: ymd(2026, time.November, 12), End: ymd(2026, time.November, 12), Days: 1},
		{Kind: KindBase, WorkYearStart: ymd(2025, time.March, 10), Start: ymd(2026, time.November, 13), End: ymd(2026, time.November, 15), Days: 3},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d allocations, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("allocation[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
	// входные балансы не должны меняться
	if balances[0].Remaining[KindBase] != 2 {
		t.Error("Allocate mutated input balances")
	}
}

func TestAllocate_HolidaysAreNotCharged(t *testing.T) {
	balances := []YearBalance{
		{Year: yearStat(2025, 30, 0, 0), Remaining: map[Kind]int{KindBase: 30}},
	}
	holidays := Holidays{ymd(2026, time.November, 4): {}}
	periods := []CreationPeriod{{Start: ymd(2026, time.November, 2), End: ymd(2026, time.November, 8)}}

	got, _, err := Allocate(periods, balances, holidays, false)
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	if len(got) != 1 || got[0].Days != 6 || got[0].End != ymd(2026, time.November, 8) {
		t.Fatalf("got %+v, want one allocation of 6 days ending 08.11", got)
	}
}

func TestAllocate_NotEnoughDays(t *testing.T) {
	balances := []YearBalance{
		{Year: yearStat(2025, 30, 0, 0), Remaining: map[Kind]int{KindBase: 5}},
	}
	periods := []CreationPeriod{{Start: ymd(2026, time.November, 9), End: ymd(2026, time.November, 15)}}

	if _, _, err := Allocate(periods, balances, Holidays{}, false); !errors.Is(err, ErrExceedsAnnual) {
		t.Fatalf("err = %v, want ErrExceedsAnnual", err)
	}
}

func TestAllocate_OverdraftGoesToLastYearBase(t *testing.T) {
	balances := []YearBalance{
		{Year: yearStat(2024, 30, 0, 0), Remaining: map[Kind]int{KindBase: 2}},
		{Year: yearStat(2025, 30, 3, 0), Remaining: map[Kind]int{KindBase: 1, KindIrregular: 1}},
	}
	periods := []CreationPeriod{{Start: ymd(2026, time.January, 12), End: ymd(2026, time.January, 18)}}

	got, overdraft, err := Allocate(periods, balances, Holidays{}, true)
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	if overdraft != 3 {
		t.Errorf("overdraft = %d, want 3", overdraft)
	}
	total := 0
	for _, a := range got {
		total += a.Days
	}
	if total != 7 {
		t.Errorf("allocated %d days, want 7: %+v", total, got)
	}
	last := got[len(got)-1]
	if last.WorkYearStart != ymd(2025, time.March, 10) {
		t.Errorf("last allocation from %v, want work year 10.03.2025", last.WorkYearStart)
	}
}

func TestValidateCreation_CollectsAllViolationsButOverlapIsHard(t *testing.T) {
	uid := user.GenerateUserID()
	year := yearStat(2026, 30, 0, 0)
	existing, _ := NewVacation(NewVacationID(), uid, KindBase, ymd(2026, time.June, 1), ymd(2026, time.June, 7), ymd(2026, time.March, 10), time.Now())

	// 3 дня (короче минимума) и вплотную к существующему отпуску (разрыв 2 дня)
	violations, err := ValidateCreation([]CreationPeriod{{Start: ymd(2026, time.June, 10), End: ymd(2026, time.June, 12)}}, year, []*Vacation{existing}, Holidays{}, ymd(2026, time.January, 1), false)
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}
	if len(violations) < 2 || !errors.Is(violations[0], ErrTooShort) || !errors.Is(violations[1], ErrGapTooSmall) {
		t.Fatalf("violations = %v, want ErrTooShort and ErrGapTooSmall", violations)
	}

	_, err = ValidateCreation([]CreationPeriod{{Start: ymd(2026, time.June, 5), End: ymd(2026, time.June, 20)}}, year, []*Vacation{existing}, Holidays{}, ymd(2026, time.January, 1), false)
	if !errors.Is(err, ErrOverlapsExisting) {
		t.Fatalf("err = %v, want ErrOverlapsExisting", err)
	}
}

func TestUsedDaysInYear_FutureVacationExcludesHolidays(t *testing.T) {
	uid := user.GenerateUserID()
	year := yearStat(2026, 30, 0, 0)
	v, _ := NewVacation(NewVacationID(), uid, KindBase, ymd(2026, time.November, 2), ymd(2026, time.December, 2), ymd(2026, time.March, 10), time.Now())
	today := ymd(2026, time.October, 29)

	// праздники нужны до конца отпуска, а не только до сегодняшнего дня
	if end := HolidaysRangeEnd([]*Vacation{v}, today); end != ymd(2026, time.December, 2) {
		t.Fatalf("HolidaysRangeEnd = %v, want 02.12.2026", end)
	}
	used := UsedDaysInYear([]*Vacation{v}, year, Holidays{ymd(2026, time.November, 4): {}})
	if used.Base != 30 {
		t.Errorf("used base = %d, want 30 (31 calendar days minus 04.11)", used.Base)
	}
}

func TestBuildBalances_CarryoverAndEarnedCap(t *testing.T) {
	uid := user.GenerateUserID()
	past := yearStat(2024, 30, 0, 1)
	current := yearStat(2025, 30, 0, 1)

	used, _ := NewVacation(NewVacationID(), uid, KindBase, ymd(2024, time.June, 1), ymd(2024, time.June, 14), ymd(2024, time.March, 10), time.Now())
	cancelled, _ := NewVacation(NewVacationID(), uid, KindBase, ymd(2024, time.July, 1), ymd(2024, time.July, 14), ymd(2024, time.March, 10), time.Now())
	_ = cancelled.Reject(time.Now())

	// 10.09.2025 — полгода текущего рабочего года: заработано 31/12*6 = 15.5 → 15 дн.
	at := ymd(2025, time.September, 10)
	balances, err := BuildBalances([]YearStat{past, current}, []*Vacation{used, cancelled}, Holidays{}, at, NewWorkYearCalculator(), true)
	if err != nil {
		t.Fatalf("BuildBalances: %v", err)
	}
	if len(balances) != 2 {
		t.Fatalf("got %d balances, want 2", len(balances))
	}
	if got := balances[0].Remaining; got[KindBase] != 16 || got[KindSeniority] != 1 {
		t.Errorf("past year remaining = %v, want base 16, seniority 1", got)
	}
	if got := balances[1].Total(); got != 15 {
		t.Errorf("current year available = %d, want 15 (earned to date)", got)
	}
}
