package vacation

import "github.com/google/uuid"

type VacationID uuid.UUID

func GenerateVacationID() VacationID {
	return VacationID(uuid.New())
}

func NewVacationID(id uuid.UUID) (VacationID, error) {
	if id == uuid.Nil {
		return VacationID(uuid.Nil), ErrIDEmpty
	}

	return VacationID(id), nil
}
