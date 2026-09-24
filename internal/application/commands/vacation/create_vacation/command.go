package create_vacation

type Command struct {
	Status    string
	UserID    string
	startDate string
	endDate   string
}

type Result struct {
	VacationID string
	Error      error
}
