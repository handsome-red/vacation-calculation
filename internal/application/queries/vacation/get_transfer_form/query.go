package get_transfer_form

type Query struct {
	UserID     string
	VacationID string
}

type Result struct {
	UserID     string
	VacationID string
	StartDate  string // ISO — значение по умолчанию для полей формы
	EndDate    string
	Period     string // ДД.ММ.ГГГГ — ДД.ММ.ГГГГ
	Kind       string
	Status     string
	// Transferable — отпуск можно перенести (черновик или утверждённый).
	Transferable bool
}
