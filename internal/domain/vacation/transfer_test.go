package vacation

import (
	"errors"
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
)

func TestTransfer_OldVacationLeavesQuotaAndCannotBeReactivated(t *testing.T) {
	uid := user.GenerateUserID()
	admin := user.GenerateUserID()
	year := yearStat(2026, 30, 0, 0)
	v, _ := NewVacation(NewVacationID(), uid, KindBase, ymd(2026, time.June, 1), ymd(2026, time.June, 14), ymd(2026, time.March, 10), time.Now())
	_ = v.Approve(time.Now())

	if _, err := NewTransfer(v, ymd(2026, time.July, 1), ymd(2026, time.July, 14), "   ", admin, time.Now()); !errors.Is(err, ErrTransferReasonEmpty) {
		t.Fatalf("empty reason: err = %v, want ErrTransferReasonEmpty", err)
	}

	tr, err := NewTransfer(v, ymd(2026, time.July, 1), ymd(2026, time.July, 14), " производственная необходимость ", admin, time.Now())
	if err != nil {
		t.Fatalf("NewTransfer: %v", err)
	}
	if tr.Reason() != "производственная необходимость" || tr.OldStart() != ymd(2026, time.June, 1) || tr.VacationID() != v.ID() {
		t.Errorf("transfer = %+v", tr)
	}

	if err := v.MarkTransferred(time.Now()); err != nil {
		t.Fatalf("MarkTransferred: %v", err)
	}
	if used := UsedDaysInYear([]*Vacation{v}, year, Holidays{}); used.Base != 0 {
		t.Errorf("transferred vacation still uses %d days", used.Base)
	}
	if err := v.MarkTransferred(time.Now()); !errors.Is(err, ErrVacationNotTransferable) {
		t.Errorf("second transfer: err = %v, want ErrVacationNotTransferable", err)
	}
	if err := v.ChangeStatus(StatusApproved, time.Now()); !errors.Is(err, ErrStatusInvalid) {
		t.Errorf("reactivate transferred: err = %v, want ErrStatusInvalid", err)
	}

	other, _ := NewVacation(NewVacationID(), uid, KindBase, ymd(2026, time.August, 1), ymd(2026, time.August, 14), ymd(2026, time.March, 10), time.Now())
	if err := other.ChangeStatus(StatusTransferred, time.Now()); !errors.Is(err, ErrStatusInvalid) {
		t.Errorf("set TRANSFERRED via ChangeStatus: err = %v, want ErrStatusInvalid", err)
	}
}
