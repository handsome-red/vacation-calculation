// Package audit — журнал изменений: кто, когда и что изменил.
package audit

import "time"

// Entry — запись журнала. Имена хранятся на момент действия: запись остаётся
// понятной и после переименования или удаления сотрудника.
type Entry struct {
	At        time.Time
	ActorID   string
	ActorName string
	// TargetID, TargetName — сотрудник, над которым совершено действие; пусто —
	// действие общее (например, праздничные дни).
	TargetID   string
	TargetName string
	// DistrictCode — теротдел, к которому относится действие: по нему руководитель
	// видит записи своего теротдела. Пусто — запись видит только администратор.
	DistrictCode string
	Action       string
	Details      string
}
