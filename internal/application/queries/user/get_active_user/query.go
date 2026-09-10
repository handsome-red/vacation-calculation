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
	FullName   string
	Department string
}
