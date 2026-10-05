package list_users

type Result struct {
	Users []UserListItem
	Page  int
	Size  int
	Total int
}

type UserListItem struct {
	ID              string
	Status          string
	FirstName       string
	LastName        string
	MiddleName      string
	BirthDate       string
	Position        string
	HiredAt         string
	DepartmentTitle string
	DistrictTitle   string
	WorkdayDuration int
	Email           string
	IsInvalid       bool
	TotalExperience string
}
