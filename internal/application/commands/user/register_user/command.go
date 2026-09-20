package register_user

type Command struct {
	Email      string
	Password   string
	FirstName  string
	LastName   string
	MiddleName string
	Department string
}

type Result struct {
	UserID    string
	Email     string
	FirstName  string
	LastName  string
	MiddleName  string
	CreatedAt string
}
