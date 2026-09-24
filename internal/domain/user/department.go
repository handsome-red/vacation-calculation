package user

type DepartmentWithDistrict struct {
	Department   Department
	DistrictCode string
}

type Department struct {
	Code  string
	Title string
}

func (d Department) String() string { return d.Code }
