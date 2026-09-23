package register_form

type Query struct{}

type PositionOption struct {
	Value string
	Title string
}

type DepartmentOption struct {
	Value string
	Title string
}

type DistrictsOption struct {
	Value string
	Title string
}

type Result struct {
	Positions   []PositionOption
	Districts   []DistrictsOption
	Departments []DepartmentOption
}
