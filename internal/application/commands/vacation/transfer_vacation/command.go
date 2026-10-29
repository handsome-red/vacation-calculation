package transfer_vacation

type Command struct {
	UserID     string
	VacationID string
	NewStart   string // YYYY-MM-DD
	NewEnd     string // YYYY-MM-DD
	Reason     string
}

type Result struct {
	TransferID     string
	NewVacationIDs []string
	// Warnings — правила, обойдённые администратором при переносе.
	Warnings []string
}
