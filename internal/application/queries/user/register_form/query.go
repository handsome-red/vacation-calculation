package register_form

type Query struct{}

type PositionOption struct {
	Code  string
	Title string
}

type DepartmentOption struct {
	Code         string
	Title        string
	DistrictCode string
}

type DistrictOption struct {
	Code  string
	Title string
}

type Form struct {
	LastName        string
	FirstName       string
	MiddleName      string
	Email           string
	BirthDate       string
	HiredAt         string
	WorkdayDuration int
	IsInvalid       bool
}

type Result struct {
	Positions   []PositionOption
	Districts   []DistrictOption
	Departments []DepartmentOption

	Errors    []string
	CSRFToken string
	Today     string
	Form      Form
}