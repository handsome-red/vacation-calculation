package vacation

type Holiday struct {
	name  string
	date Date
}

func NewHoliday(name string, date Date) (Holiday, error) {
	if date.IsZero() {
		return Holiday{}, ErrHolidayDateEmpty
	}
	
	if name == "" {
		return Holiday{}, ErrHolidayNameEmpty
	}

	return Holiday{
		name: name,
		date: date,
	}, nil
}

func (h Holiday) Date() Date   { return h.date }
func (h Holiday) Name() string { return h.name }

func (h Holiday) String() string {
	return h.date.String() + " " + h.name
}
