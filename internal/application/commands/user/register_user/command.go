package register_user

type Command struct {
	Email           string
	Password        string
	FirstName       string
	LastName        string
	MiddleName      string
	DepartmentCode  string
	BirthDate       string
	Position        string
	Experience		int
	HiredAt         string
	DistrictCode    string
	WorkdayDuration int
	IsInvalid       bool
	// Districts — в какие теротделы можно добавить сотрудника (права руководителя);
	// nil — в любые (администратор, первичная настройка).
	Districts []string
}

type Result struct {
	UserID     string
	Email      string
	FirstName  string
	LastName   string
	MiddleName string
	CreatedAt  string
}
