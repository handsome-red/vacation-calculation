package get_user

type Query struct {
	UserID string
}

type VacationStat struct {
	Year      int
	Base      int
	Irregular int
	Seniority int
}

type VacationBlock struct {
	Stats []VacationStat
}

type Result struct {
	ID              string
	Status          string
	IsActive        bool
	LastName        string
	FirstName       string
	MiddleName      string
	BirthDate       string
	Position        string
	HiredAt         string
	Department      string
	District        string
	WorkdayDuration int
	Email           string
	IsInvalid       bool
	Experience      string
	CreatedAt       string
	UpdatedAt       string

	Initials string
	Today    string
	WorkYear string
	Supposed string
	Earned   float32

	Vacations []VacationStat
}
