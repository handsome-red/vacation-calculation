package middleware

import (
	"testing"
	"time"
)

func TestLoginLimiterBlocksEmailAfterFiveFailures(t *testing.T) {
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	l := NewLoginLimiter()
	l.now = func() time.Time { return now }

	for i := 0; i < loginEmailLimit; i++ {
		if _, blocked := l.Blocked("1.1.1.1", "a@b.ru"); blocked {
			t.Fatalf("blocked after %d failures", i)
		}
		l.Fail("1.1.1.1", "a@b.ru")
		now = now.Add(time.Minute)
	}
	wait, blocked := l.Blocked("2.2.2.2", "A@B.ru ")
	if !blocked {
		t.Fatal("e-mail must be blocked from any address")
	}
	// Первая неудача была в 10:00, сейчас 10:05 — ждать до 10:10.
	if wait != 5*time.Minute {
		t.Errorf("wait = %v, want 5m", wait)
	}
	if _, blocked := l.Blocked("1.1.1.1", "other@b.ru"); blocked {
		t.Error("other e-mail from the same address must not be blocked yet")
	}

	now = now.Add(5 * time.Minute)
	if _, blocked := l.Blocked("1.1.1.1", "a@b.ru"); blocked {
		t.Error("must unblock when the oldest failure leaves the window")
	}
}

func TestLoginLimiterBlocksAddressAndSuccessKeepsIt(t *testing.T) {
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	l := NewLoginLimiter()
	l.now = func() time.Time { return now }

	for i := 0; i < loginIPLimit; i++ {
		l.Fail("1.1.1.1", "user"+string(rune('a'+i))+"@b.ru")
	}
	l.Success("usera@b.ru")
	if _, blocked := l.Blocked("1.1.1.1", "new@b.ru"); !blocked {
		t.Error("address must stay blocked after a success")
	}
	if _, blocked := l.Blocked("3.3.3.3", "new@b.ru"); blocked {
		t.Error("other address must not be blocked")
	}
}
