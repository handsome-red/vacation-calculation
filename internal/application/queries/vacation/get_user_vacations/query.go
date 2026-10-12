package get_user_vacations

type Query struct {
	UserID string
}

type Vacations struct {
	ID         string
	Color      int
	StartDate  string
	EndDate    string
	Status     string
	StatusCode string
}

type Result struct {
	Vacations []Vacations
}
