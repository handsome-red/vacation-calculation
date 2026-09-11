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
	ID         string
	Email      string
	FirstName  string
	LastName   string
	MiddleName string
	Department string
}
