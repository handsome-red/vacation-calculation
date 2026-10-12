package user

type District struct {
code  string
title string
}

func NewDistrict(code, title string) District {
return District{
code:  code,
title: title,
}
}

func (d District) Code() string  { return d.code }
func (d District) Title() string { return d.title }

// String returns the district title for display.
func (d District) String() string {
return d.title
}
