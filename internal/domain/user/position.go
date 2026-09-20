package user

type Position int

// TODO Какие еще могут быть статусы?
const (
	PositionDirector  Position = iota
	PositionSlave

	PositionUnknown
)

func NewPosition(v int) (Position, error) {
	p := Position(v)
	if !isValid(p) {
		return PositionUnknown, ErrPositionInvalid
	}
	
	return p, nil
}

func isValid(p Position) bool {
	switch p {
	case PositionDirector, PositionSlave:
		return true
	default:
		return false
	}
} 