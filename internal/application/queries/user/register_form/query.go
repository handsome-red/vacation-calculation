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

type Result struct {
	Positions   []PositionOption
	Districts   []DistrictOption
	Departments []DepartmentOption
}
