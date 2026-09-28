package get_user_vacations


type Query struct {
	UserID string
}

type Vacations struct {
	StartDate string
	EndDate string
}

type Result struct {
	Vacations []Vacations
}
