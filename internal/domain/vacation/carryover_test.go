package vacation

import (
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

func TestCarryoverDeadline(t *testing.T) {
	// рабочий год 10.03.2020–09.03.2021 → использовать не позднее 09.03.2022
	if got := CarryoverDeadline(yearStat(2020, 30, 0, 0)); got != ymd(2022, time.March, 9) {
		t.Errorf("deadline = %v, want 09.03.2022", got)
	}
}

func TestCarryoverWarnings(t *testing.T) {
	today := ymd(2026, time.October, 29)
	balances := []YearBalance{
		{Year: yearStat(2024, 30, 0, 0), Remaining: map[Kind]int{KindBase: 5}},  // срок 09.03.2026 — истёк
		{Year: yearStat(2025, 30, 0, 0), Remaining: map[Kind]int{KindBase: 21}}, // закончился 09.03.2026, срок 09.03.2027
		{Year: yearStat(2023, 30, 0, 0), Remaining: map[Kind]int{KindBase: 0}},  // всё использовано
		{Year: yearStat(2026, 30, 0, 0), Remaining: map[Kind]int{KindBase: 12}}, // текущий год — не предупреждаем
	}

	got := CarryoverWarnings(balances, today)
	if len(got) != 2 {
		t.Fatalf("got %d warnings, want 2: %+v", len(got), got)
	}
	if !got[0].Expired || got[0].Unused != 5 {
		t.Errorf("2024: %+v, want expired with 5 days", got[0])
	}
	if got[1].Expired || got[1].Deadline != ymd(2027, time.March, 9) {
		t.Errorf("2025: %+v, want active until 09.03.2027", got[1])
	}
	if want := "За рабочий год 10.03.2025–09.03.2026 не использовано 21 день. Их нужно использовать не позднее 09.03.2027."; got[1].Message() != want {
		t.Errorf("message = %q, want %q", got[1].Message(), want)
	}
}

func TestLateUses(t *testing.T) {
	uid := user.GenerateUserID()
	stats := []YearStat{yearStat(2024, 30, 0, 0)} // срок 09.03.2026
	inTime, _ := NewVacation(NewVacationID(), uid, KindBase, ymd(2026, time.February, 1), ymd(2026, time.February, 14), ymd(2024, time.March, 10), time.Now())
	late, _ := NewVacation(NewVacationID(), uid, KindBase, ymd(2026, time.March, 1), ymd(2026, time.March, 14), ymd(2024, time.March, 10), time.Now())

	got := LateUses([]*Vacation{inTime, late}, stats)
	if len(got) != 1 || got[0].Start != ymd(2026, time.March, 1) || got[0].Deadline != ymd(2026, time.March, 9) {
		t.Fatalf("got %+v, want one late use from 01.03.2026", got)
	}
}
