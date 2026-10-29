package vacation

import (
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/date"
)

func TestMergeCreationPeriods(t *testing.T) {
	d := func(m time.Month, day int) date.Date { return date.New(2026, m, day) }

	t.Run("single period", func(t *testing.T) {
		got := MergeCreationPeriods([]CreationPeriod{{Kind: KindBase, Start: d(11, 9), End: d(11, 13)}})
		if len(got) != 1 || got[0].Start != d(11, 9) || got[0].End != d(11, 13) {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("adjacent periods of same kind are merged", func(t *testing.T) {
		got := MergeCreationPeriods([]CreationPeriod{
			{Kind: KindBase, Start: d(11, 16), End: d(11, 20)},
			{Kind: KindBase, Start: d(11, 9), End: d(11, 15)},
		})
		if len(got) != 1 || got[0].Start != d(11, 9) || got[0].End != d(11, 20) {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("gap or different kind keeps periods apart", func(t *testing.T) {
		got := MergeCreationPeriods([]CreationPeriod{
			{Kind: KindBase, Start: d(11, 9), End: d(11, 13)},
			{Kind: KindBase, Start: d(11, 16), End: d(11, 20)},
			{Kind: KindSeniority, Start: d(11, 21), End: d(11, 22)},
		})
		if len(got) != 3 {
			t.Fatalf("got %d periods, want 3: %+v", len(got), got)
		}
	})
}
