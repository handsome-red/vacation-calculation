package vacation

import (
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Vacation struct {
	id        VacationID
	status    VacationStatus
	kind      Kind
	userID    user.UserID
	startDate date.Date
	endDate   date.Date
	// Начало рабочего года, из квоты которого списаны дни отпуска
	workYearStart date.Date
	// Отпуск, перенесённый на даты этой записи (пусто, если запись не результат переноса)
	transferredFrom VacationID
	createdAt       time.Time
	updatedAt       time.Time
}

type VacationParams struct {
	ID              VacationID
	VacationStatus  VacationStatus
	Kind            Kind
	UserID          user.UserID
	StartDate       date.Date
	EndDate         date.Date
	WorkYearStart   date.Date
	TransferredFrom VacationID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Конструктор - создает валидный отпуск
func NewVacation(
	id VacationID,
	userID user.UserID,
	kind Kind,
	startDate date.Date,
	endDate date.Date,
	workYearStart date.Date,
	now time.Time,
) (*Vacation, error) {
	if id.IsZero() {
		return nil, ErrVacationIDEmpty
	}
	if userID.IsZero() {
		return nil, ErrUserIDEmpty
	}
	if startDate.IsZero() {
		return nil, ErrStartDateEmpty
	}
	if endDate.IsZero() {
		return nil, ErrEndDateEmpty
	}
	if endDate.Before(startDate) {
		return nil, ErrEndDateBeforeStart
	}
	if workYearStart.IsZero() {
		return nil, ErrWorkYearStartEmpty
	}

	return &Vacation{
		id:            id,
		status:        StatusDraft,
		userID:        userID,
		startDate:     startDate,
		endDate:       endDate,
		kind:          kind,
		workYearStart: workYearStart,
		createdAt:     now.UTC(),
		updatedAt:     now.UTC(),
	}, nil
}

func (v *Vacation) ChangePeriod(start, end date.Date, now time.Time) error {
	if v.Status() != StatusDraft {
		return ErrVacationNotEditable
	}
	if end.Before(start) {
		return ErrEndDateBeforeStart
	}
	v.startDate = start
	v.endDate = end
	v.updatedAt = now.UTC()
	return nil
}

func (v *Vacation) Approve(now time.Time) error {
	if v.Status() != StatusDraft {
		return ErrVacationNotApprovable
	}
	v.status = StatusApproved
	v.updatedAt = now.UTC()
	return nil
}

func (v *Vacation) Reject(now time.Time) error {
	if v.Status() != StatusDraft {
		return ErrVacationNotRejectable
	}
	v.status = StatusCancelled
	v.updatedAt = now.UTC()
	return nil
}

// MarkTransferred помечает отпуск перенесённым: запись остаётся в истории,
// но больше не занимает даты и дни квоты.
func (v *Vacation) MarkTransferred(now time.Time) error {
	if !v.status.IsActive() {
		return ErrVacationNotTransferable
	}
	v.status = StatusTransferred
	v.updatedAt = now.UTC()
	return nil
}

// LinkTransferredFrom связывает запись с отпуском, который на неё перенесли.
func (v *Vacation) LinkTransferredFrom(id VacationID) {
	v.transferredFrom = id
}

func (v *Vacation) ChangeStatus(status VacationStatus, now time.Time) error {
	validatedStatus, err := NewVacationStatus(status.ISO())
	if err != nil {
		return err
	}
	// Перенос меняет статус только через MarkTransferred, и перенесённый отпуск
	// не возвращается в работу — иначе разойдётся история переносов.
	if v.status == validatedStatus || validatedStatus == StatusTransferred || v.status == StatusTransferred {
		return ErrStatusInvalid
	}
	v.status = validatedStatus
	v.updatedAt = now.UTC()
	return nil
}

func (v *Vacation) ID() VacationID                 { return v.id }
func (v *Vacation) VacationStatus() VacationStatus { return v.Status() }
func (v *Vacation) UserID() user.UserID            { return v.userID }
func (v *Vacation) StartDate() date.Date           { return v.startDate }
func (v *Vacation) EndDate() date.Date             { return v.endDate }
func (v *Vacation) Status() VacationStatus         { return v.status }
func (v *Vacation) CreatedAt() time.Time           { return v.createdAt }
func (v *Vacation) UpdatedAt() time.Time           { return v.updatedAt }

func (v *Vacation) Kind() Kind { return v.kind }

func (v *Vacation) WorkYearStart() date.Date { return v.workYearStart }

func (v *Vacation) TransferredFrom() VacationID { return v.transferredFrom }

func ReconstructVacation(p VacationParams) *Vacation {
	return &Vacation{
		id:              p.ID,
		status:          p.VacationStatus,
		kind:            p.Kind,
		userID:          p.UserID,
		startDate:       p.StartDate,
		endDate:         p.EndDate,
		workYearStart:   p.WorkYearStart,
		transferredFrom: p.TransferredFrom,
		createdAt:       p.CreatedAt,
		updatedAt:       p.UpdatedAt,
	}
}
