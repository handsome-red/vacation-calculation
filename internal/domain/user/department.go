package user

type Department struct {
	department string
}

func NewDepartment(department string) (Department, error) {
	return Department{department: department}, nil
}

func (d Department) String() string {
	return d.department
}
