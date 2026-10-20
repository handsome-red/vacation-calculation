package create_vacations

type Period struct {
	Kind      string
	StartDate string
	EndDate   string
}

type Command struct {
	UserID  string
	Periods []Period
}

type Result struct {
	VacationID string
	Error      error
}
