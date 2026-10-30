package vacation

import (
	"strings"
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

func TestScheduleParts(t *testing.T) {
	uid := user.GenerateUserID()
	mk := func(from, to time.Time, kind Kind, approve bool) *Vacation {
		v, _ := NewVacation(NewVacationID(), uid, kind, DateOf(from), DateOf(to), ymd(2026, time.March, 10), time.Now())
		if approve {
			_ = v.Approve(time.Now())
		}
		return v
	}
	d := func(m time.Month, day int) time.Time { return time.Date(2027, m, day, 0, 0, 0, 0, time.UTC) }

	vacations := []*Vacation{
		mk(d(time.June, 1), d(time.June, 10), KindBase, true),
		mk(d(time.June, 11), d(time.June, 14), KindIrregular, false), // встык — та же часть
		mk(d(time.August, 1), d(time.August, 7), KindBase, true),
		mk(d(time.September, 1), d(time.September, 20), KindUnpaid, true), // административный — не в графике
	}
	parts := ScheduleParts(vacations, 2027, Holidays{ymd(2027, time.June, 12): {}})
	if len(parts) != 2 {
		t.Fatalf("parts = %+v, want 2", parts)
	}
	if parts[0].End != ymd(2027, time.June, 14) || parts[0].Days != 13 || parts[0].Approved {
		t.Errorf("part 1 = %+v, want 01.06–14.06, 13 days (12.06 holiday), not fully approved", parts[0])
	}

	issues := ScheduleIssues(parts)
	joined := strings.Join(issues, "; ")
	if !strings.Contains(joined, "Нет части отпуска не менее 14") {
		t.Errorf("issues = %v, want missing 14-day part (13 days without holiday)", issues)
	}
	if ScheduleIssues(nil)[0] != "Отпуск в графике не запланирован" {
		t.Error("empty schedule must be reported")
	}
	if ApprovalDeadline(2027) != ymd(2026, time.December, 17) {
		t.Error("deadline for 2027 must be 17.12.2026")
	}
}
