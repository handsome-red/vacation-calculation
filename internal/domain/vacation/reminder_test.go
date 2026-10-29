package vacation

import (
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

func TestUpcomingForReminder(t *testing.T) {
	today := ymd(2026, time.October, 29)
	alice, bob := user.GenerateUserID(), user.GenerateUserID()
	mk := func(uid user.UserID, from, to time.Time, approve bool) *Vacation {
		v, _ := NewVacation(NewVacationID(), uid, KindBase, DateOf(from), DateOf(to), ymd(2026, time.March, 10), time.Now())
		if approve {
			_ = v.Approve(time.Now())
		}
		return v
	}
	day := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }

	vacations := []*Vacation{
		// alice: два куска встык — одно уведомление на весь отпуск; начало ровно через 28 дней
		mk(alice, day(time.November, 30), day(time.December, 6), true),
		mk(alice, day(time.November, 26), day(time.November, 29), true),
		// bob: начинается через 29 дней — рано
		mk(bob, day(time.November, 27), day(time.December, 10), true),
		// bob: черновик через 10 дней — не уведомляем
		mk(bob, day(time.November, 8), day(time.November, 14), false),
		// bob: уже начался — поздно
		mk(bob, day(time.October, 29), day(time.November, 4), true),
	}

	got := UpcomingForReminder(vacations, today, ReminderDaysBefore)
	if len(got) != 1 {
		t.Fatalf("got %d reminders, want 1: %+v", len(got), got)
	}
	if got[0].UserID != alice || got[0].Start != ymd(2026, time.November, 26) || got[0].End != ymd(2026, time.December, 6) {
		t.Errorf("reminder = %+v, want alice 26.11.2026–06.12.2026", got[0])
	}
}
