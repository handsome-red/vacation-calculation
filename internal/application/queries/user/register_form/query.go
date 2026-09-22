package register_form

type Query struct{}

type PositionOption struct {
	Value string
	Title string
}

type Result struct{ Positions []PositionOption }
