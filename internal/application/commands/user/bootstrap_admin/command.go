package bootstrap_admin

// Command — профиль первого администратора из переменных окружения ADMIN_*.
type Command struct {
	Email          string
	Password       string
	LastName       string
	FirstName      string
	MiddleName     string
	BirthDate      string
	Position       string
	HiredAt        string
	DistrictCode   string
	DepartmentCode string
}

type Result struct {
	// Created — создан новый пользователь; Promoted — существующему выданы права.
	// Оба false — администратор уже есть, ничего не изменено.
	Created  bool
	Promoted bool
	UserID   string
}
