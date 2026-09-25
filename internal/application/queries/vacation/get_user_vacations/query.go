package get_user_vacations


type Query struct {
	UserID string
}

type Vacations struct {
	startDate string
	endDate string
}

type Result struct {
	Vacations []Vacations
}
