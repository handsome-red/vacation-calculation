// domain/vacation/vacation_id.go
package vacation

import (
	"fmt"

	"github.com/google/uuid"
)

type VacationID struct {
	value uuid.UUID
}

func NewVacationID() VacationID {
	return VacationID{value: uuid.New()}
}

func VacationIDFromUUID(id uuid.UUID) (VacationID, error) {
	if id == uuid.Nil {
		return VacationID{}, ErrVacationIDEmpty
	}
	return VacationID{value: id}, nil
}

func ParseVacationID(s string) (VacationID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return VacationID{}, fmt.Errorf("parse vacation id: %w", err)
	}
	return VacationIDFromUUID(id)
}

func GenerateVacationID() VacationID {
	return VacationID{value: uuid.New()}
}

func (v VacationID) IsZero() bool    { return v.value == uuid.Nil }
func (v VacationID) String() string  { return v.value.String() }
func (v VacationID) UUID() uuid.UUID { return v.value }
