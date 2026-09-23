package register_user

type Command struct {
	Email           string
	Password        string
	FirstName       string
	LastName        string
	MiddleName      string
	DepartmentCode  string
	DepartmentTitle string
	Status          string
	BirthDate       string
	Position        string
	HiredAt         string
	DistrictCode    string
	DistrictTitle   string
	WorkdayDuration int
	IsInvalid       bool
}

type Result struct {
	UserID     string
	Email      string
	FirstName  string
	LastName   string
	MiddleName string
	CreatedAt  string
}
