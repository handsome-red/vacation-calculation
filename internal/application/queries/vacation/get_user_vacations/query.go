package get_user_vacations

type Query struct {
	UserID string
	// All — показать все отпуска; иначе только закончившиеся за последние 3 года и будущие.
	All bool
}

type Vacations struct {
	ID         string
	Kind       string
	WorkYear   string // начало рабочего года, из квоты которого взяты дни
	StartDate  string
	EndDate    string
	Status     string
	StatusCode string
	// IsActive — отпуск действует (черновик или утверждён): его можно менять и переносить.
	IsActive bool
	// Transfer — история переноса: для перенесённой записи — куда, для новой — откуда.
	Transfer *TransferInfo
}

type TransferInfo struct {
	Summary string // «Перенесён на …» / «Перенесён с …»
	Reason  string
	Date    string // когда перенесли
}

type Result struct {
	Vacations []Vacations
	// Hidden — сколько более старых отпусков скрыто (при All = false).
	Hidden int
}
