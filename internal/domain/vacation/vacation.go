package vacation

import (
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

type Vacation struct {
	id        VacationID
	status    VacationStatus
	userID    user.UserID
	startDate Date
	endDate   Date
	createdAt time.Time
	updatedAt time.Time
	// TODO
}

type VacationParams struct {
	ID             VacationID
	VacationStatus VacationStatus
	UserID         user.UserID
	StartDate      Date
	EndDate        Date
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Конструктор - создает валидный отпуск
func NewVacation(
	id VacationID,
	userID user.UserID,
	startDate Date,
	endDate Date,
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

	return &Vacation{
		id:        id,
		status:    StatusDraft,
		userID:    userID,
		startDate: startDate,
		endDate:   endDate,
		createdAt: now.UTC(),
		updatedAt: now.UTC(),
	}, nil
}

func (v *Vacation) ChangePeriod(start, end Date, now time.Time) error {
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
	v.status = StatusRejected
	v.updatedAt = now.UTC()
	return nil
}

func (v *Vacation) ID() VacationID                 { return v.id }
func (v *Vacation) VacationStatus() VacationStatus { return v.Status() }
func (v *Vacation) UserID() user.UserID            { return v.userID }
func (v *Vacation) StartDate() Date                { return v.startDate }
func (v *Vacation) EndDate() Date                  { return v.endDate }
func (v *Vacation) Status() VacationStatus         { return v.status }
func (v *Vacation) CreatedAt() time.Time           { return v.createdAt }
func (v *Vacation) UpdatedAt() time.Time           { return v.updatedAt }

func ReconstructVacation(p VacationParams) *Vacation {
	return &Vacation{
		id:        p.ID,
		status:    p.VacationStatus,
		userID:    p.UserID,
		startDate: p.StartDate,
		endDate:   p.EndDate,
		createdAt: p.CreatedAt,
		updatedAt: p.UpdatedAt,
	}
}

func calculateDays(startDate, endDate time.Time) int {
	return int(endDate.Sub(startDate).Hours() / 24)
}

// func (v *Vacation) Approve() error {

// }
