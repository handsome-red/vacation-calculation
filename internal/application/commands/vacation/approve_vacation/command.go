package approve_vacation

type Command struct {
	UserID     string
	VacationID string
	IsAdmin    bool
}

type Result struct {
	UserID     string
	VacationID string
}
