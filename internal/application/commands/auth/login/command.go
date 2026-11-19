package login

type Command struct {
	Password string
	Email    string
}

type Result struct {
	UserID string
	Email  string
	// IsStaff — администратор или руководитель: после входа — на главную.
	IsStaff bool
}
