package create_vacation

type Command struct {
	UserID    string
	StartDate string
	EndDate   string
}

type Result struct {
	VacationID string
	Error      error
}
