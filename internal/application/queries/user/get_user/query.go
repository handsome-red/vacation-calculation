package get_user

type Query struct {
	UserID string
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
}
