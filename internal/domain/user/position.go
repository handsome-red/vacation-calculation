package user

import "strings"

type Position string

const (
	PositionUnknown  Position = ""
	PositionDirector Position = "director"
	PositionEmployee Position = "employee"
)

func NewPosition(s string) (Position, error) {
	p := Position(strings.ToLower(strings.TrimSpace(s)))
	if !p.isValid() {
		return PositionUnknown, ErrPositionInvalid
	}
	return p, nil
}

func (p Position) isValid() bool {
	switch p {
	case PositionDirector, PositionEmployee:
		return true
	default:
		return false
	}
}

func (p Position) String() string { return string(p) }
