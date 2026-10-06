package shift_form

type Query struct {
}

type Result struct {
	Kinds []ShiftKindOption
}

type ShiftKindOption struct {
	Code  string
	Title string
}
