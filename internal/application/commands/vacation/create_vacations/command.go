package create_vacations

type Period struct {
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
