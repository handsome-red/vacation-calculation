package get_calendar

type Query struct {
	Year int	
}

type holidayItem struct {
	Date string `json:"date"`
	Name string `json:"name"`
}

type Result struct {
	Year int
	HolidaysJSON string
}