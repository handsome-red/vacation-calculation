package user

type Department struct {
	code  string
	title string
}

func NewDepartment(code, title string) Department {
	return Department{
		code:  code,
		title: title,
	}
}

// Возвращает название района/отдела на русском
func (d Department) Title() string { return d.title }

func (d Department) Code() string { return d.code }
