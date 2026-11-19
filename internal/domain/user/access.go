package user

import "slices"

// Права доступа. Администратор управляет всеми сотрудниками. Начальник
// территориального отдела (руководитель) — теми же возможностями, но только в своём
// территориальном отделе и не в отношении администраторов. Остальные сотрудники
// видят только себя.

// ManagedDistricts — коды территориальных отделов, которыми сотрудник руководит:
// у начальника территориального отдела — его теротдел.
func (u *User) ManagedDistricts() []string {
	if u == nil || u.superadmin || u.position != PositionHeadOfTerritorialDepartment || u.district.Code() == "" {
		return nil
	}
	return []string{u.district.Code()}
}

// IsManager — руководитель территориального отдела (не администратор).
func (u *User) IsManager() bool {
	return len(u.ManagedDistricts()) > 0
}

// IsStaff — администратор или руководитель: управляет другими сотрудниками.
func (u *User) IsStaff() bool {
	return u.isAdmin() || u.IsManager()
}

// ManagesDistrict — может ли сотрудник управлять территориальным отделом code.
func (u *User) ManagesDistrict(code string) bool {
	return u != nil && ManagesDistrict(u.superadmin, u.ManagedDistricts(), code)
}

// Manages — может ли сотрудник действовать как администратор в отношении target.
func (u *User) Manages(target *User) bool {
	return u != nil && Manages(u.superadmin, u.ManagedDistricts(), target)
}

// ManagesDistrict — правило для прав isAdmin и managed (коды теротделов руководителя).
func ManagesDistrict(isAdmin bool, managed []string, code string) bool {
	return isAdmin || slices.Contains(managed, code)
}

// Manages — правило для прав isAdmin и managed: руководитель управляет сотрудниками
// своего территориального отдела, кроме администраторов.
func Manages(isAdmin bool, managed []string, target *User) bool {
	if target == nil {
		return false
	}
	if isAdmin {
		return true
	}
	return !target.isAdmin() && slices.Contains(managed, target.District().Code())
}

// RoleLabel — роль сотрудника для отображения.
func (u *User) RoleLabel() string {
	switch {
	case u.isAdmin():
		return "Администратор"
	case u.IsManager():
		return "Руководитель территориального отдела"
	}
	return "Сотрудник"
}
