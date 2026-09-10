package user

import "fmt"

type Name struct {
	firstName  string
	lastName   string
	middleName string
}

func NewName(
	firstName string,
	lastName string,
	middleName string,
) (Name, error) {
	if firstName == "" || lastName == "" || middleName == "" {
		return Name{}, ErrNameRequired
	}

	return Name{
		firstName:  firstName,
		lastName:   lastName,
		middleName: middleName,
	}, nil
}

func (n Name) FirstName() string {
	return n.firstName
}

func (n Name) LastName() string {
	return n.lastName
}

func (n Name) MiddleName() string {
	return n.middleName
}

func (n Name) FullName() string {
	return fmt.Sprintf("%s %s %s", n.lastName, n.firstName, n.middleName)
}
