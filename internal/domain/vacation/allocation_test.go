package vacation

import (
	"errors"
	"strings"
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

	got, err := Allocate(periods, balances, Holidays{})
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

	got, err := Allocate(periods, balances, holidays)
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

	if _, err := Allocate(periods, balances, Holidays{}); !errors.Is(err, ErrExceedsAnnual) {
		t.Fatalf("err = %v, want ErrExceedsAnnual", err)
	}
}

// Сверх остатков всех лет списать нельзя: в ошибке — сколько доступно.
func TestAllocate_NoOverdraftAcrossYears(t *testing.T) {
	balances := []YearBalance{
		{Year: yearStat(2024, 30, 0, 0), Remaining: map[Kind]int{KindBase: 2}},
		{Year: yearStat(2025, 30, 3, 0), Remaining: map[Kind]int{KindBase: 1, KindIrregular: 1}},
	}
	periods := []CreationPeriod{{Start: ymd(2026, time.January, 12), End: ymd(2026, time.January, 18)}}

	_, err := Allocate(periods, balances, Holidays{})
	if !errors.Is(err, ErrExceedsAnnual) || !strings.Contains(err.Error(), "доступно 4 дн.") {
		t.Fatalf("err = %v, want ErrExceedsAnnual with 4 days available", err)
	}
}

func TestAllocate_ChosenKindUsesOnlyItsQuota(t *testing.T) {
	balances := []YearBalance{
		{Year: yearStat(2024, 30, 3, 1), Remaining: map[Kind]int{KindBase: 10, KindIrregular: 1, KindSeniority: 0}},
		{Year: yearStat(2025, 30, 3, 5), Remaining: map[Kind]int{KindBase: 30, KindIrregular: 3, KindSeniority: 5}},
	}
	periods := []CreationPeriod{
		{Kind: KindSeniority, Start: ymd(2026, time.June, 1), End: ymd(2026, time.June, 5)},
		{Kind: KindIrregular, Start: ymd(2026, time.June, 6), End: ymd(2026, time.June, 7)},
	}

	got, err := Allocate(periods, balances, Holidays{})
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	want := []Allocation{
		{Kind: KindSeniority, WorkYearStart: ymd(2025, time.March, 10), Start: ymd(2026, time.June, 1), End: ymd(2026, time.June, 5), Days: 5},
		{Kind: KindIrregular, WorkYearStart: ymd(2024, time.March, 10), Start: ymd(2026, time.June, 6), End: ymd(2026, time.June, 6), Days: 1},
		{Kind: KindIrregular, WorkYearStart: ymd(2025, time.March, 10), Start: ymd(2026, time.June, 7), End: ymd(2026, time.June, 7), Days: 1},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("allocation[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}

	tooMany := []CreationPeriod{{Kind: KindSeniority, Start: ymd(2026, time.June, 1), End: ymd(2026, time.June, 6)}}
	if _, err := Allocate(tooMany, balances, Holidays{}); !errors.Is(err, ErrExceedsAnnual) {
		t.Fatalf("err = %v, want ErrExceedsAnnual", err)
	}
}

func TestAllocate_ChosenKindIsReservedBeforeAuto(t *testing.T) {
	balances := []YearBalance{
		{Year: yearStat(2025, 30, 3, 0), Remaining: map[Kind]int{KindBase: 2, KindIrregular: 3}},
	}
	periods := []CreationPeriod{
		{Start: ymd(2026, time.June, 1), End: ymd(2026, time.June, 3)},
		{Kind: KindIrregular, Start: ymd(2026, time.June, 4), End: ymd(2026, time.June, 5)},
	}

	got, err := Allocate(periods, balances, Holidays{})
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	if len(got) != 3 || got[0].Kind != KindBase || got[0].Days != 2 || got[1].Kind != KindIrregular || got[1].Days != 1 || got[2].Start != ymd(2026, time.June, 4) {
		t.Fatalf("got %+v", got)
	}
}

// Выбранный вид не занимает дни другого вида, даже если их много.
func TestAllocate_ChosenKindDoesNotBorrowOtherKinds(t *testing.T) {
	balances := []YearBalance{
		{Year: yearStat(2025, 30, 3, 1), Remaining: map[Kind]int{KindBase: 30, KindIrregular: 3, KindSeniority: 1}},
	}
	periods := []CreationPeriod{{Kind: KindSeniority, Start: ymd(2026, time.June, 1), End: ymd(2026, time.June, 3)}}

	_, err := Allocate(periods, balances, Holidays{})
	if !errors.Is(err, ErrExceedsAnnual) || !strings.Contains(err.Error(), "доступно 1 дн.") {
		t.Fatalf("err = %v, want ErrExceedsAnnual with 1 day available", err)
	}
}

// В текущем году заработанные дни можно потратить на любой вид, а не только на основной.
func TestAllocate_EarnedDaysCanBeSpentOnChosenKind(t *testing.T) {
	current := yearStat(2025, 30, 3, 1)
	at := ymd(2025, time.September, 10)
	balances, err := BuildBalances([]YearStat{current}, nil, Holidays{}, at, NewWorkYearCalculator(), true)
	if err != nil {
		t.Fatalf("BuildBalances: %v", err)
	}
	if balances[0].Remaining[KindSeniority] != 0 {
		t.Fatalf("Remaining = %v, want earned days in base first", balances[0].Remaining)
	}

	periods := []CreationPeriod{{Kind: KindSeniority, Start: at, End: at}}
	got, err := Allocate(periods, balances, Holidays{})
	if err != nil {
		t.Fatalf("Allocate: %v", err)
	}
	if len(got) != 1 || got[0].Kind != KindSeniority {
		t.Fatalf("got %+v", got)
	}

	// Сверх заработанного (17 дн.) нельзя даже при большой квоте вида.
	base := []CreationPeriod{{Kind: KindBase, Start: at, End: ymd(2025, time.September, 27)}}
	if _, err := Allocate(base, balances, Holidays{}); !errors.Is(err, ErrExceedsAnnual) {
		t.Fatalf("err = %v, want ErrExceedsAnnual", err)
	}
}

func TestValidateCreation_AdjacentKindsCountAsOnePart(t *testing.T) {
	year := yearStat(2026, 30, 3, 5)
	today := ymd(2026, time.January, 1)

	joined := []CreationPeriod{
		{Kind: KindIrregular, Start: ymd(2026, time.June, 1), End: ymd(2026, time.June, 3)},
		{Kind: KindBase, Start: ymd(2026, time.June, 4), End: ymd(2026, time.June, 17)},
	}
	v, err := ValidateCreation(joined, year, nil, Holidays{}, today)
	if err != nil || len(v) != 0 {
		t.Fatalf("violations = %v, err = %v; want none", v, err)
	}

	alone := []CreationPeriod{{Kind: KindIrregular, Start: ymd(2026, time.June, 1), End: ymd(2026, time.June, 3)}}
	v, err = ValidateCreation(alone, year, nil, Holidays{}, today)
	if err != nil || len(v) == 0 || !errors.Is(v[0], ErrTooShort) {
		t.Fatalf("violations = %v, err = %v; want ErrTooShort", v, err)
	}
}

func TestValidateCreation_UnpaidPeriodSkipsAnnualRulesInMixedRequest(t *testing.T) {
	year := yearStat(2026, 30, 3, 5)
	periods := []CreationPeriod{
		{Kind: KindBase, Start: ymd(2026, time.June, 1), End: ymd(2026, time.June, 14)},
		{Kind: KindUnpaid, Start: ymd(2026, time.June, 20), End: ymd(2026, time.June, 21)},
	}
	v, err := ValidateCreation(periods, year, nil, Holidays{}, ymd(2026, time.January, 1))
	if err != nil || len(v) != 0 {
		t.Fatalf("violations = %v, err = %v; want none", v, err)
	}
}

func TestBuildBalances_OpeningUsageReducesRemaining(t *testing.T) {
	year := yearStat(2024, 30, 3, 1)
	stats := []YearStat{year}
	ApplyOpeningUsage(stats, []OpeningUsage{
		{WorkYearStart: ymd(2024, time.March, 10), Days: OpeningDays{Base: 20, Irregular: 3}},
		{WorkYearStart: ymd(2020, time.March, 10), Days: OpeningDays{Base: 30}},
	})
	balances, err := BuildBalances(stats, nil, Holidays{}, ymd(2026, time.January, 1), NewWorkYearCalculator(), true)
	if err != nil {
		t.Fatalf("BuildBalances: %v", err)
	}
	got := balances[0].Remaining
	if got[KindBase] != 10 || got[KindIrregular] != 0 || got[KindSeniority] != 1 {
		t.Errorf("remaining = %v, want base 10, irregular 0, seniority 1", got)
	}
}

func TestValidateCreation_CollectsAllViolationsButOverlapIsHard(t *testing.T) {
	uid := user.GenerateUserID()
	year := yearStat(2026, 30, 0, 0)
	existing, _ := NewVacation(NewVacationID(), uid, KindBase, ymd(2026, time.June, 1), ymd(2026, time.June, 7), ymd(2026, time.March, 10), time.Now())

	// 3 дня (короче минимума) и вплотную к существующему отпуску (разрыв 2 дня)
	violations, err := ValidateCreation([]CreationPeriod{{Start: ymd(2026, time.June, 10), End: ymd(2026, time.June, 12)}}, year, []*Vacation{existing}, Holidays{}, ymd(2026, time.January, 1))
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}
	if len(violations) < 2 || !errors.Is(violations[0], ErrTooShort) || !errors.Is(violations[1], ErrGapTooSmall) {
		t.Fatalf("violations = %v, want ErrTooShort and ErrGapTooSmall", violations)
	}

	_, err = ValidateCreation([]CreationPeriod{{Start: ymd(2026, time.June, 5), End: ymd(2026, time.June, 20)}}, year, []*Vacation{existing}, Holidays{}, ymd(2026, time.January, 1))
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
