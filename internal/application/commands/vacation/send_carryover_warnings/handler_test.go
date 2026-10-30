package send_carryover_warnings

import (
	"strings"
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

func TestStage(t *testing.T) {
	w := vacation.CarryoverWarning{
		YearFrom: time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC),
		YearTo:   time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC),
		Unused:   12,
		Deadline: date.New(2027, time.March, 9),
	}
	at := func(y int, m time.Month, d int) date.Date { return date.New(y, m, d) }

	for _, tt := range []struct {
		today   date.Date
		expired bool
		want    string
	}{
		{at(2027, time.February, 6), false, ""},            // за 31 день — рано
		{at(2027, time.February, 7), false, recipientSoon}, // за 30 дней
		{at(2027, time.March, 9), false, recipientSoon},    // последний день
		{at(2027, time.March, 10), true, recipientExpired}, // истёк
		{at(2027, time.March, 23), true, recipientExpired},
		{at(2027, time.March, 24), true, ""}, // давно истёк — не рассылаем
	} {
		w.Expired = tt.expired
		got, _ := stage(w, tt.today)
		if got != tt.want {
			t.Errorf("%s: stage = %q, want %q", tt.today.ISO(), got, tt.want)
		}
	}

	subject, body := Message("Пётр", "Александрович", w)
	if !strings.Contains(subject, "10.03.2025–09.03.2026") || !strings.Contains(body, "не позднее 09.03.2027") || !strings.Contains(body, ": 12.") {
		t.Errorf("message: %q\n%s", subject, body)
	}
}
