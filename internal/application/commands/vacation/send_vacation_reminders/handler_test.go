package send_vacation_reminders

import (
	"strings"
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

func TestMessage(t *testing.T) {
	today := date.New(2026, time.October, 29)
	subject, body := Message("Пётр", "Александрович", date.New(2026, time.November, 26), date.New(2026, time.December, 9), today)

	if subject != "Напоминание о начале отпуска с 26.11.2026" {
		t.Errorf("subject = %q", subject)
	}
	for _, want := range []string{
		"Здравствуйте, Пётр Александрович!",
		"с 26.11.2026 по 09.12.2026",
		"осталось 28 календарных дней",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q:\n%s", want, body)
		}
	}
}

func TestHRMessage(t *testing.T) {
	today := date.New(2026, time.October, 29)
	subject, body := HRMessage(Employee{
		FullName: "Иванов Пётр Александрович", Initials: "Иванов П.А.", Position: "Инженер-программист",
		Department: "Отдел эпидемиологического надзора", District: "г. Казань", Email: "petr@example.com",
	}, date.New(2026, time.November, 26), date.New(2026, time.December, 9), today)

	if subject != "Отпуск сотрудника Иванов П.А. с 26.11.2026" {
		t.Errorf("subject = %q", subject)
	}
	for _, want := range []string{
		"Сотрудник: Иванов Пётр Александрович",
		"Должность: Инженер-программист",
		"Период отпуска: с 26.11.2026 по 09.12.2026",
		"осталось 28 календарных дней",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q:\n%s", want, body)
		}
	}
}

func TestPlural(t *testing.T) {
	for n, want := range map[int]string{1: "день", 2: "дня", 5: "дней", 11: "дней", 21: "день", 22: "дня", 28: "дней"} {
		if got := plural(n, "день", "дня", "дней"); got != want {
			t.Errorf("plural(%d) = %q, want %q", n, got, want)
		}
	}
}
