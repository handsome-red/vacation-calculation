package user

import (
	"slices"
	"strings"
)

type Position string

// TODO: Переписать под map
const (
	PositionUnknown  Position = ""
	PositionDirector Position = "director"
	PositionEmployee Position = "employee"
)

var allPositions = []Position{
	PositionDirector,
	PositionEmployee,
}

func NewPosition(s string) (Position, error) {
	p := Position(strings.ToLower(strings.TrimSpace(s)))
	if !p.isValid() {
		return PositionUnknown, ErrPositionInvalid
	}
	return p, nil
}

func (p Position) isValid() bool {
	for _, valid := range allPositions {
		if p == valid {
			return true
		}
	}

	return false
}

func (p Position) String() string { return string(p) }

// Возвращает название должности на русском
func (p Position) Title() string {
	switch p {
	case PositionDirector:
		return "Директор"
	case PositionEmployee:
		return "Сотрудник"
	default:
		return ""
	}
}

func AllPositions() []Position {
	return slices.Clone(allPositions)
}
