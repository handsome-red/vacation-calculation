package auth

type Status struct {
	value string
}

func NewStatus(s string) (Status, error) {
	return Status{
		value: s,
	}, nil
}

// TODO
func (s Status) IsValid() bool {
	return true
}
