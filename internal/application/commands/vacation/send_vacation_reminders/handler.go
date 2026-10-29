package send_vacation_reminders

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

// recipientEmployee — ключ получателя «сотрудник» в журнале; отдел кадров — "hr:<адрес>".
const recipientEmployee = "employee"

type Result struct {
	Sent   int
	Failed int
}

type Handler struct {
	vacationRepo ports.VacationRepository
	userRepo     ports.UserRepository
	reminderRepo ports.ReminderRepository
	mailer       ports.Mailer
	hrEmails     []string
	logger       ports.Logger
}

func NewHandler(
	vacationRepo ports.VacationRepository,
	userRepo ports.UserRepository,
	reminderRepo ports.ReminderRepository,
	mailer ports.Mailer,
	hrEmails []string,
	logger ports.Logger,
) *Handler {
	return &Handler{
		vacationRepo: vacationRepo,
		userRepo:     userRepo,
		reminderRepo: reminderRepo,
		mailer:       mailer,
		hrEmails:     hrEmails,
		logger:       logger,
	}
}

type letter struct {
	recipient string // ключ в журнале
	email     string
	subject   string
	body      string
}

// Handle отправляет уведомления об утверждённых отпусках, которые начинаются
// в ближайшие vacation.ReminderDaysBefore календарных дней: сотруднику и отделу кадров.
// Каждому получателю письмо уходит один раз; неудачная отправка повторится при следующем запуске.
func (h *Handler) Handle(ctx context.Context, now time.Time) (*Result, error) {
	today := vacation.DateOf(now)

	vacations, err := h.vacationRepo.FindApprovedEndingFrom(ctx, today)
	if err != nil {
		return nil, fmt.Errorf("list approved vacations: %w", err)
	}

	result := &Result{}
	for _, up := range vacation.UpcomingForReminder(vacations, today, vacation.ReminderDaysBefore) {
		var u *user.User

		for _, l := range h.letters() {
			sent, err := h.reminderRepo.WasSent(ctx, up.UserID, up.Start, l.recipient)
			if err != nil {
				return result, err
			}
			if sent {
				continue
			}

			if u == nil {
				if u, err = h.userRepo.FindByID(ctx, up.UserID); err != nil {
					h.logger.Error(ctx, "vacation reminder: load user", "user_id", up.UserID.String(), "error", err)
					result.Failed++
					break
				}
				if !u.Status().IsActive() {
					break
				}
			}

			l = h.fill(l, u, up, today)
			if err := h.mailer.Send(ctx, l.email, l.subject, l.body); err != nil {
				h.logger.Error(ctx, "vacation reminder: send", "user_id", up.UserID.String(), "email", l.email, "error", err)
				result.Failed++
				continue
			}
			if err := h.reminderRepo.MarkSent(ctx, up.UserID, up.Start, up.End, l.recipient, l.email, now); err != nil {
				return result, err
			}
			h.logger.Info(ctx, "vacation reminder sent", "user_id", up.UserID.String(), "email", l.email, "start", up.Start.ISO())
			result.Sent++
		}
	}

	return result, nil
}

// letters — получатели уведомления об отпуске (адрес сотрудника и текст заполняются в fill).
func (h *Handler) letters() []letter {
	out := make([]letter, 0, 1+len(h.hrEmails))
	out = append(out, letter{recipient: recipientEmployee})
	for _, e := range h.hrEmails {
		out = append(out, letter{recipient: "hr:" + strings.ToLower(e), email: e})
	}
	return out
}

func (h *Handler) fill(l letter, u *user.User, up vacation.UpcomingVacation, today date.Date) letter {
	if l.recipient == recipientEmployee {
		l.email = u.Email().Value()
		l.subject, l.body = Message(u.FirstName(), u.MiddleName(), up.Start, up.End, today)
		return l
	}
	position := ""
	if p := u.Position(); p != "" {
		position = p.Title()
	}
	l.subject, l.body = HRMessage(Employee{
		FullName:   strings.Join(strings.Fields(u.LastName()+" "+u.FirstName()+" "+u.MiddleName()), " "),
		Initials:   u.Initials(),
		Position:   position,
		Department: u.Department().String(),
		District:   u.District().Title(),
		Email:      u.Email().Value(),
	}, up.Start, up.End, today)
	return l
}

// Message — тема и текст уведомления сотруднику о предстоящем отпуске.
func Message(firstName, middleName string, start, end, today date.Date) (subject, body string) {
	from, to, left := start.Time().Format("02.01.2006"), end.Time().Format("02.01.2006"), daysLeft(start, today)

	name := strings.TrimSpace(firstName + " " + middleName)
	greeting := "Здравствуйте!"
	if name != "" {
		greeting = "Здравствуйте, " + name + "!"
	}

	subject = "Напоминание о начале отпуска с " + from
	body = greeting + "\n\n" +
		fmt.Sprintf("Напоминаем, что с %s по %s у вас запланирован ежегодный оплачиваемый отпуск.\n", from, to) +
		fmt.Sprintf("До начала отпуска %s.\n\n", leftText(left)) +
		"Если даты отпуска нужно изменить, обратитесь в отдел кадров.\n\n" +
		"Письмо сформировано автоматически, отвечать на него не нужно.\n"
	return subject, body
}

// Employee — данные сотрудника для письма в отдел кадров.
type Employee struct {
	FullName   string
	Initials   string
	Position   string
	Department string
	District   string
	Email      string
}

// HRMessage — тема и текст уведомления отделу кадров о предстоящем отпуске сотрудника.
func HRMessage(e Employee, start, end, today date.Date) (subject, body string) {
	from, to, left := start.Time().Format("02.01.2006"), end.Time().Format("02.01.2006"), daysLeft(start, today)

	subject = "Отпуск сотрудника " + e.Initials + " с " + from
	var b strings.Builder
	b.WriteString("Уведомление о предстоящем ежегодном оплачиваемом отпуске сотрудника.\n\n")
	b.WriteString("Сотрудник: " + e.FullName + "\n")
	if e.Position != "" {
		b.WriteString("Должность: " + e.Position + "\n")
	}
	if e.District != "" {
		b.WriteString("Территориальный отдел: " + e.District + "\n")
	}
	if e.Department != "" {
		b.WriteString("Район/отдел: " + e.Department + "\n")
	}
	b.WriteString("E-mail: " + e.Email + "\n")
	b.WriteString(fmt.Sprintf("Период отпуска: с %s по %s\n", from, to))
	b.WriteString(fmt.Sprintf("До начала отпуска %s.\n\n", leftText(left)))
	b.WriteString("Сотруднику направлено такое же уведомление.\n")
	b.WriteString("Письмо сформировано автоматически, отвечать на него не нужно.\n")
	return subject, b.String()
}

func daysLeft(start, today date.Date) int {
	return int(start.Time().Sub(today.Time()).Hours() / 24)
}

func leftText(n int) string {
	return fmt.Sprintf("%s %d %s", plural(n, "остался", "осталось", "осталось"), n,
		plural(n, "календарный день", "календарных дня", "календарных дней"))
}

// plural выбирает форму слова для числа n: 1 день, 2 дня, 5 дней.
func plural(n int, one, few, many string) string {
	n %= 100
	if n >= 11 && n <= 14 {
		return many
	}
	switch n % 10 {
	case 1:
		return one
	case 2, 3, 4:
		return few
	default:
		return many
	}
}
