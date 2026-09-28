package vacation

import 	"context"


type HolidayRepository interface  {
	Save(ctx context.Context, holiday Holiday) error
	ListByRange(ctx context.Context, from, to Date) ([]Holiday, error)
	ListByYear(ctx context.Context, year int) ([]Holiday, error)
}