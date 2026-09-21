package get_active_users

type Query struct {
	Page int
	Size int
}

type Result struct {
	Users      []UserListItem
	TotalCount int
	Page       int
	PageSize   int
}

type UserListItem struct {
	ID              string
	FirstName       string
	LastName        string
	MiddleName      string
	BirthDate       string
	Position        string
	HiredAt         string
	Department      string
	District        string
	WorkdayDuration int
	Email           string
	IsInvalid       bool
}
