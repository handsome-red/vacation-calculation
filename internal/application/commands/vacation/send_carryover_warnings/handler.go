package send_carryover_warnings

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/leavecalc"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

// Когда предупреждать о сроке использования неиспользованного отпуска.
const (
	// SoonDays — заранее, за сколько дней до истечения срока.
	SoonDays = 30
	// ExpiredDays — сколько дней после истечения ещё отправлять письмо «срок истёк»
	// (если сервер был выключен); более старые сроки не рассылаются.
	ExpiredDays = 14

	recipientSoon    = "carryover-soon"
	recipientExpired = "carryover-expired"
)

type Result struct {
	Sent   int
	Failed int
}

type Handler struct {
	userRepo     ports.UserRepository
	leave        *leavecalc.Service
	reminderRepo ports.ReminderRepository
	mailer       ports.Mailer
	logger       ports.Logger
}

func NewHandler(
	userRepo ports.UserRepository,
	leave *leavecalc.Service,
	reminderRepo ports.ReminderRepository,
	mailer ports.Mailer,
	logger ports.Logger,
) *Handler {
	return &Handler{userRepo: userRepo, leave: leave, reminderRepo: reminderRepo, mailer: mailer, logger: logger}
}

// Handle предупреждает сотрудников о неиспользованных днях закончившихся рабочих лет:
// за SoonDays дней до истечения срока переноса (12 месяцев после окончания года)
// и при его истечении. Каждое письмо — один раз на сотрудника и рабочий год.
func (h *Handler) Handle(ctx context.Context, now time.Time) (*Result, error) {
	users, _, err := h.userRepo.FindUsers(ctx, ports.Query{
		Page: 1, Size: 100000,
		Filters: ports.UserFilter{Status: string(user.StatusActive)},
	})
	if err != nil {
		return nil, fmt.Errorf("users: %w", err)
	}

	today := vacation.DateOf(now)
	res := &Result{}
	for _, u := range users {
		snap, err := h.leave.Load(ctx, u, today, "")
		if err != nil {
			h.logger.Error(ctx, "carryover warning: load", "user_id", u.ID().String(), "error", err)
			res.Failed++
			continue
		}
		balances, err := snap.Balances(today, true)
		if err != nil {
			h.logger.Error(ctx, "carryover warning: balances", "user_id", u.ID().String(), "error", err)
			res.Failed++
			continue
		}

		for _, w := range vacation.CarryoverWarnings(balances, today) {
			recipient, ok := stage(w, today)
			if !ok {
				continue
			}
			yearStart := vacation.DateOf(w.YearFrom)
			sent, err := h.reminderRepo.WasSent(ctx, u.ID(), yearStart, recipient)
			if err != nil {
				return res, err
			}
			if sent {
				continue
			}

			subject, body := Message(u.FirstName(), u.MiddleName(), w)
			email := u.Email().Value()
			if err := h.mailer.Send(ctx, email, subject, body); err != nil {
				h.logger.Error(ctx, "carryover warning: send", "user_id", u.ID().String(), "email", email, "error", err)
				res.Failed++
				continue
			}
			if err := h.reminderRepo.MarkSent(ctx, u.ID(), yearStart, w.Deadline, recipient, email, now); err != nil {
				return res, err
			}
			h.logger.Info(ctx, "carryover warning sent", "user_id", u.ID().String(), "email", email, "deadline", w.Deadline.ISO())
			res.Sent++
		}
	}
	return res, nil
}

// stage — какое письмо положено на дату today: «скоро истекает» или «истёк».
func stage(w vacation.CarryoverWarning, today date.Date) (string, bool) {
	deadline := w.Deadline.Time()
	t := today.Time()
	switch {
	case !w.Expired && !t.Before(deadline.AddDate(0, 0, -SoonDays)):
		return recipientSoon, true
	case w.Expired && !t.After(deadline.AddDate(0, 0, ExpiredDays)):
		return recipientExpired, true
	}
	return "", false
}

// Message — тема и текст предупреждения о сроке использования отпуска.
func Message(firstName, middleName string, w vacation.CarryoverWarning) (subject, body string) {
	period := w.YearFrom.Format("02.01.2006") + "–" + w.YearTo.AddDate(0, 0, -1).Format("02.01.2006")
	deadline := w.Deadline.Time().Format("02.01.2006")

	greeting := "Здравствуйте!"
	if name := strings.TrimSpace(firstName + " " + middleName); name != "" {
		greeting = "Здравствуйте, " + name + "!"
	}

	if w.Expired {
		subject = "Истёк срок использования отпуска за рабочий год " + period
		body = greeting + "\n\n" +
			fmt.Sprintf("За рабочий год %s у вас осталось неиспользованных дней ежегодного отпуска: %d.\n", period, w.Unused) +
			fmt.Sprintf("Перенесённый отпуск нужно было использовать не позднее %s (12 месяцев после окончания рабочего года).\n\n", deadline) +
			"Обратитесь в отдел кадров, чтобы согласовать даты отпуска.\n\n" +
			"Письмо сформировано автоматически, отвечать на него не нужно.\n"
		return subject, body
	}

	subject = "Срок использования отпуска за рабочий год " + period + " истекает " + deadline
	body = greeting + "\n\n" +
		fmt.Sprintf("За рабочий год %s у вас осталось неиспользованных дней ежегодного отпуска: %d.\n", period, w.Unused) +
		fmt.Sprintf("Перенесённый отпуск можно использовать не позднее %s (12 месяцев после окончания рабочего года).\n\n", deadline) +
		"Запланируйте отпуск или обратитесь в отдел кадров.\n\n" +
		"Письмо сформировано автоматически, отвечать на него не нужно.\n"
	return subject, body
}
