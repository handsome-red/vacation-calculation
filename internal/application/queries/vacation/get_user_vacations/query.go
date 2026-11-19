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
	TotalDays  int // календарных дней, включая первый и последний
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

// OpeningRow — дни рабочего года, использованные до начала учёта (без дат).
type OpeningRow struct {
	WorkYear  string
	Base      int
	Irregular int
	Seniority int
}

type Result struct {
	Vacations []Vacations
	Opening   []OpeningRow
	// Hidden — сколько более старых отпусков скрыто (при All = false).
	Hidden int
}
