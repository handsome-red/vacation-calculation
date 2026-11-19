package user

import "testing"

func TestInitials(t *testing.T) {
	cases := []struct {
		last, first, middle, want string
	}{
		{"Иванов", "Иван", "Иванович", "Иванов И.И."},
		{"Smith", "John", "", "Smith J."},
		{"Петров", "", "", "Петров"},
	}
	for _, c := range cases {
		u := ReconstructUser(UserParams{LastName: c.last, FirstName: c.first, MiddleName: c.middle})
		if got := u.Initials(); got != c.want {
			t.Errorf("Initials(%q, %q, %q) = %q, want %q", c.last, c.first, c.middle, got, c.want)
		}
	}
}
