package vacation

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

// Transfer — запись истории переноса отпуска: какой отпуск, с каких дат,
// на какие, почему и кто перенёс. Новые даты оформлены отдельными записями
// отпуска со ссылкой TransferredFrom на исходный отпуск.
type Transfer struct {
	id         string
	userID     user.UserID
	vacationID VacationID
	oldStart   date.Date
	oldEnd     date.Date
	newStart   date.Date
	newEnd     date.Date
	reason     string
	createdBy  user.UserID
	createdAt  time.Time
}

type TransferParams struct {
	ID         string
	UserID     user.UserID
	VacationID VacationID
	OldStart   date.Date
	OldEnd     date.Date
	NewStart   date.Date
	NewEnd     date.Date
	Reason     string
	CreatedBy  user.UserID
	CreatedAt  time.Time
}

// NewTransfer фиксирует перенос отпуска v на [newStart, newEnd].
func NewTransfer(v *Vacation, newStart, newEnd date.Date, reason string, by user.UserID, now time.Time) (Transfer, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return Transfer{}, ErrTransferReasonEmpty
	}
	if newEnd.Before(newStart) {
		return Transfer{}, ErrEndDateBeforeStart
	}
	return Transfer{
		id:         uuid.NewString(),
		userID:     v.UserID(),
		vacationID: v.ID(),
		oldStart:   v.StartDate(),
		oldEnd:     v.EndDate(),
		newStart:   newStart,
		newEnd:     newEnd,
		reason:     reason,
		createdBy:  by,
		createdAt:  now.UTC(),
	}, nil
}

func ReconstructTransfer(p TransferParams) Transfer {
	return Transfer{
		id:         p.ID,
		userID:     p.UserID,
		vacationID: p.VacationID,
		oldStart:   p.OldStart,
		oldEnd:     p.OldEnd,
		newStart:   p.NewStart,
		newEnd:     p.NewEnd,
		reason:     p.Reason,
		createdBy:  p.CreatedBy,
		createdAt:  p.CreatedAt,
	}
}

func (t Transfer) ID() string             { return t.id }
func (t Transfer) UserID() user.UserID    { return t.userID }
func (t Transfer) VacationID() VacationID { return t.vacationID }
func (t Transfer) OldStart() date.Date    { return t.oldStart }
func (t Transfer) OldEnd() date.Date      { return t.oldEnd }
func (t Transfer) NewStart() date.Date    { return t.newStart }
func (t Transfer) NewEnd() date.Date      { return t.newEnd }
func (t Transfer) Reason() string         { return t.reason }
func (t Transfer) CreatedBy() user.UserID { return t.createdBy }
func (t Transfer) CreatedAt() time.Time   { return t.createdAt }
