package vacation

import (
	"sort"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

// ReminderDaysBefore — за сколько календарных дней до начала отпуска
// сотруднику приходит уведомление.
const ReminderDaysBefore = 28

// UpcomingVacation — непрерывный отпуск сотрудника. Записи, идущие встык
// (дни из разных рабочих лет или видов), склеены в один период.
type UpcomingVacation struct {
	UserID user.UserID
	Start  date.Date
	End    date.Date
}

// UpcomingForReminder возвращает утверждённые отпуска, которые начинаются
// в ближайшие daysBefore календарных дней (today < start <= today+daysBefore).
// Если уведомление не ушло ровно за daysBefore дней (сервер был выключен, отпуск
// оформлен позже), оно уйдёт при первой проверке до начала отпуска.
func UpcomingForReminder(vacations []*Vacation, today date.Date, daysBefore int) []UpcomingVacation {
	byUser := make(map[user.UserID][]*Vacation)
	for _, v := range vacations {
		if v != nil && v.Status() == StatusApproved {
			byUser[v.UserID()] = append(byUser[v.UserID()], v)
		}
	}

	last := DateOf(today.Time().AddDate(0, 0, daysBefore))
	out := make([]UpcomingVacation, 0)
	for uid, list := range byUser {
		sort.Slice(list, func(i, j int) bool { return list[i].StartDate().Before(list[j].StartDate()) })

		cur := UpcomingVacation{UserID: uid, Start: list[0].StartDate(), End: list[0].EndDate()}
		flush := func() {
			if cur.Start.After(today) && !cur.Start.After(last) {
				out = append(out, cur)
			}
		}
		for _, v := range list[1:] {
			if v.StartDate() == cur.End.AddDay() {
				cur.End = v.EndDate()
				continue
			}
			flush()
			cur = UpcomingVacation{UserID: uid, Start: v.StartDate(), End: v.EndDate()}
		}
		flush()
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })
	return out
}
