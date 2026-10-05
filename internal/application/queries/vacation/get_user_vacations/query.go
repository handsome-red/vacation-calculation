package get_user_vacations

type Query struct {
	UserID string
}

type Vacations struct {
	Color     int
	StartDate string
	EndDate   string
	Status    string
}

type Result struct {
	Vacations []Vacations
}
