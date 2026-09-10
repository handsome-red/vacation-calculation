package get_user

type Query struct {
	UserID string
}

type Result struct {
	ID         string
	Email      string
	FirstName  string
	LastName   string
	MiddleName string
	Department string
	IsActive   bool
	CreatedAt  string
	UpdatedAt  string
}
