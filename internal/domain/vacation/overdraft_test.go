package vacation

import (
	"strings"
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

func TestDiscrepancies(t *testing.T) {
	uid := user.GenerateUserID()
	stats := []YearStat{yearStat(2025, 28, 0, 0), yearStat(2026, 28, 0, 0)}
	add := func(start, end, yearStart time.Time) *Vacation {
		v, err := NewVacation(NewVacationID(), uid, KindBase, DateOf(start), DateOf(end), DateOf(yearStart), time.Now())
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	d := func(y int, m time.Month, day int) time.Time { return time.Date(y, m, day, 0, 0, 0, 0, time.UTC) }

	vacations := []*Vacation{
		// 2025: 20 + 15 = 35 дней при норме 28.
		add(d(2025, time.June, 2), d(2025, time.June, 21), stats[0].From),
		add(d(2025, time.August, 4), d(2025, time.August, 18), stats[0].From),
		// 2026: в пределах нормы.
		add(d(2026, time.June, 1), d(2026, time.June, 14), stats[1].From),
		// Привязан к несуществующему году.
		add(d(2026, time.September, 1), d(2026, time.September, 7), d(2024, time.February, 25)),
		// Привязан к будущему, ещё не посчитанному году — не расхождение.
		add(d(2027, time.July, 1), d(2027, time.July, 7), d(2027, time.March, 10)),
	}

	got := Discrepancies(stats, vacations, nil)
	if len(got) != 2 {
		t.Fatalf("got %d discrepancies, want 2: %q", len(got), got)
	}
	if !strings.Contains(got[0], "10.03.2025–09.03.2026") || !strings.Contains(got[0], "списано 35 дн. при норме 28 дн.") {
		t.Errorf("overdraft message = %q", got[0])
	}
	if !strings.Contains(got[1], "01.09.2026–07.09.2026") || !strings.Contains(got[1], "25.02.2024") {
		t.Errorf("orphan message = %q", got[1])
	}
}
