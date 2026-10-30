package list_users

type Query struct {
	Page    int
	Size    int
	Filters Filters
}

// Filters — фильтры списка сотрудников для кадровика.
type Filters struct {
	LastName       string
	FirstName      string
	MiddleName     string
	Email          string
	Status         string
	Category       string // user.CategoryCivilServant | user.CategoryOtherStaff
	Position       string
	DistrictCode   string
	DepartmentCode string
}

type Result struct {
	Users []UserListItem
	Page  int
	Size  int
	Total int
	Pages int

	// Справочники для фильтров.
	Districts   []Option
	Departments []DepartmentOption
	Positions   []Option
	Categories  []Option
}

type Option struct {
	Code  string
	Title string
}

type DepartmentOption struct {
	Code         string
	Title        string
	DistrictCode string
}

type UserListItem struct {
	ID              string
	Status          string
	FirstName       string
	LastName        string
	MiddleName      string
	BirthDate       string
	Position        string
	Category        string
	HiredAt         string
	DepartmentTitle string
	DistrictTitle   string
	WorkdayDuration int
	Email           string
	IsInvalid       bool
	TotalExperience string
}
