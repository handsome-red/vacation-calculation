package new_calendar_day_form

type Query struct {
	Year int
}

// Day — особый день производственного календаря (праздник, перенос, рабочий выходной).
type Day struct {
	Date      string // ISO
	DateLabel string // ДД.ММ.ГГГГ, день недели
	Type      string
	TypeLabel string
	Name      string
}

type Result struct {
	Year int
	Days []Day
}
