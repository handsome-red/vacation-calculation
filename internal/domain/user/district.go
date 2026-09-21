package user

import "strings"

type District string

const (
	DistrictUnknown District = ""
	DistrictElabuga District = "elabuga"
	DistrictKazan   District = "kazan"
)

func NewDistrict(s string) (District, error) {
	d := District(strings.ToLower(strings.TrimSpace(s)))
	if d == "" {
		return DistrictUnknown, nil // пусто — допустимо
	}
	if !d.isValid() {
		return DistrictUnknown, ErrDistrictInvalid
	}
	return d, nil
}

func (d District) isValid() bool {
	switch d {
	case DistrictElabuga, DistrictKazan:
		return true
	default:
		return false
	}
}

func (d District) String() string { return string(d) }
