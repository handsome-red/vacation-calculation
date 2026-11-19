package ports

import (
	"context"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/audit"
	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/calendar"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	// "github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type UserFilter struct {
	Status          string
	FirstName       string
	LastName        string
	MiddleName      string
	BirthDate       string
	Position        string
	HiredAt         string
	Email           string
	DistrictTitle   string
	DepartmentTitle string
	TotalExperience string

	DistrictCode   string
	// DistrictCodes — только эти территориальные отделы (права руководителя);
	// nil — без ограничения, пустой — ни одного.
	DistrictCodes  []string
	// NoAdmins — без администраторов: руководитель ими не управляет.
	NoAdmins       bool
	DepartmentCode string
	// Positions — любая из должностей (например, все должности категории).
	Positions []string
}

type Query struct {
	Page    int
	Size    int
	Filters UserFilter
}

type UserCommandRepository interface {
	Save(ctx context.Context, user *user.User) error
	Delete(ctx context.Context, id user.UserID) error
}

type UserQueryRepository interface {
	FindByID(ctx context.Context, id user.UserID) (*user.User, error)
	FindByEmail(ctx context.Context, email user.Email) (*user.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsAdmin(ctx context.Context) (bool, error)
	FindUsers(ctx context.Context, q Query) ([]*user.User, int, error)
}

type UserRepository interface {
	UserCommandRepository
	UserQueryRepository
}

type DistrictRepository interface {
	List(ctx context.Context) ([]user.District, error)
	FindByCode(ctx context.Context, code string) (user.District, error)
}

type DepartmentRepository interface {
	List(ctx context.Context) ([]user.Department, error)
	ListWithDistrict(ctx context.Context) ([]user.DepartmentWithDistrict, error)
}

type ShiftRepository interface {
	ListByUser(ctx context.Context, userID user.UserID) ([]vacation.Shift, error)
	ListByUserInRange(ctx context.Context, userID user.UserID, from, to time.Time) ([]vacation.Shift, error)
	Save(ctx context.Context, userID user.UserID, shift vacation.Shift) error
}

type VacationRepository interface {
	Save(ctx context.Context, vacation *vacation.Vacation) error
	FindByID(ctx context.Context, userID user.UserID, id vacation.VacationID) (*vacation.Vacation, error)
	FindByUserID(ctx context.Context, userID user.UserID) ([]*vacation.Vacation, error)
	// FindApprovedEndingFrom — утверждённые отпуска всех сотрудников, не закончившиеся к from.
	FindApprovedEndingFrom(ctx context.Context, from date.Date) ([]*vacation.Vacation, error)
	FindByUserIDInRange(
		ctx context.Context,
		userID user.UserID,
		from, to date.Date,
	) ([]*vacation.Vacation, error)
	// Delete удаляет запись отпуска вместе с историей её переносов. Записи, созданные
	// переносом этого отпуска, остаются и перестают на него ссылаться.
	Delete(ctx context.Context, userID user.UserID, id vacation.VacationID) error
}

// ReminderRepository — журнал отправленных уведомлений о начале отпуска,
// по каждому получателю (сотрудник, отдел кадров).
type ReminderRepository interface {
	WasSent(ctx context.Context, userID user.UserID, start date.Date, recipient string) (bool, error)
	MarkSent(ctx context.Context, userID user.UserID, start, end date.Date, recipient, email string, at time.Time) error
}

type TransferRepository interface {
	Save(ctx context.Context, t vacation.Transfer) error
	ListByUser(ctx context.Context, userID user.UserID) ([]vacation.Transfer, error)
}

// OpeningUsageRepository — дни, использованные до начала учёта в системе.
type OpeningUsageRepository interface {
	ListByUser(ctx context.Context, userID user.UserID) ([]vacation.OpeningUsage, error)
	Replace(ctx context.Context, userID user.UserID, usages []vacation.OpeningUsage) error
}

// WorkYearRepository — сохранённые рабочие годы: завершившиеся и с границами,
// заданными администратором. Годы идут подряд с первого.
type WorkYearRepository interface {
	ListByUser(ctx context.Context, userID user.UserID) ([]vacation.WorkYear, error)
	// Add сохраняет следующие по номеру годы; уже сохранённые номера не меняются.
	Add(ctx context.Context, userID user.UserID, years []vacation.WorkYear) error
	// Apply одной транзакцией заменяет сохранённые годы и переносит отпуска и
	// использованные дни на новые начала рабочих лет.
	Apply(ctx context.Context, userID user.UserID, years []vacation.WorkYear, vacations []vacation.VacationWorkYear, opening []vacation.OpeningUsage) error
	// DeleteByUser удаляет сохранённые годы: после изменения даты приёма они считаются заново.
	DeleteByUser(ctx context.Context, userID user.UserID) error
}

// SessionRepository — сессии входа. Ключ — хеш токена из куки, не сам токен.
type SessionRepository interface {
	Create(ctx context.Context, tokenHash string, userID user.UserID, expires time.Time) error
	// Find возвращает владельца действующей на момент now сессии; found = false, если
	// сессии нет или она истекла.
	Find(ctx context.Context, tokenHash string, now time.Time) (userID user.UserID, found bool, err error)
	Delete(ctx context.Context, tokenHash string) error
	DeleteExpired(ctx context.Context, now time.Time) error
}

// AuditFilter — какие записи журнала выбрать.
type AuditFilter struct {
	// TargetID — только действия над этим сотрудником; пусто — все.
	TargetID string
	// DistrictCodes — только записи этих теротделов; nil — без ограничения, пустой
	// срез — ничего.
	DistrictCodes []string
	Limit         int
}

type AuditRepository interface {
	Record(ctx context.Context, e audit.Entry) error
	// List возвращает записи от новых к старым.
	List(ctx context.Context, f AuditFilter) ([]audit.Entry, error)
}

type CalendarDayRepository interface {
	Save(ctx context.Context, d calendar.Exception) error
	// ListInRange возвращает все исключения календаря за период [from, to] включительно
	ListInRange(ctx context.Context, from, to date.Date) ([]calendar.Exception, error)
	// ListByYear возвращает все исключения календаря за календарный год
	ListByYear(ctx context.Context, year int) ([]calendar.Exception, error)
	Delete(ctx context.Context, d date.Date) error
}

type AuthRepository interface {
	FindByEmail(ctx context.Context, email auth.Email) (*auth.AuthUser, error)
	Save(ctx context.Context, u *auth.AuthUser) error
}
