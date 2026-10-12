package get_vacation_form

type Query struct {
	UserID string
}

type Result struct {
	UserID        string
	Total         int  // Base + Irregular + Seniority — полное ежегодное право
	Used          int  // уже использовано в текущем рабочем году
	Available     int  // Total - Used — сколько можно взять
	HasActiveYear bool // есть ли активный рабочий год
}
