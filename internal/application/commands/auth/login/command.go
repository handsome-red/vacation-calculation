package login

type Command struct {
	Password string
	Email    string
}

type Result struct {
	UserID string
	Email  string
}
