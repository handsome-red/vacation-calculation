package auth


type Role struct {
	value string
}

func NewRole(s string) (Role, error) {
	return Role{
		value: s,
	}, nil
}

// TODO
func(r Role) IsValid() bool {
	return true
}