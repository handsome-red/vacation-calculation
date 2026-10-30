package user

import (
	"sort"
	"strings"
)

type Position string

// TODO: Актуализировать список
const (
	PositionUnknown Position = ""

	PositionDirector       Position = "director"
	PositionHeadOfTerritorialDepartment Position = "head_of_territorial_department"
	PositionDeputyDirector Position = "deputy_director"

	PositionAssistant Position = "assistant"
	PositionDeputyHead Position = "deputy_head"
	PositionLeadSpecialist Position = "lead_specialist"
	PositionSeniorSpecialistSecondClass Position = "senior_specialist_second_class"
	PositionSpecialistFirstClass Position = "specialist_first_class"

	// Специалисты.
	PositionChiefSpecialistExpert Position = "chief_specialist_expert"
	PositionSpecialistExpert      Position = "specialist_expert"
	PositionSeniorSpecialistFirst Position = "senior_specialist_first"

	PositionGarageManager    Position = "garage_manager"
	PositionDriver           Position = "driver"
	PositionWatchman         Position = "watchman"
	PositionFacilityManager  Position = "facility_manager"
	PositionWarehouseManager Position = "warehouse_manager"
	PositionDocumentClerk    Position = "document_clerk"
	PositionRecordClerk      Position = "record_clerk"
	PositionHandyman         Position = "handyman"
	PositionSoftwareEngineer Position = "software_engineer"
)

var positionsMap = map[Position]string{
	PositionDirector: "Руководитель",  // +
	PositionHeadOfTerritorialDepartment:              "Начальник территориального отдела", // +
	PositionDeputyDirector:        "Заместитель начальника",            // +
	PositionChiefSpecialistExpert: "Главный специалист-эксперт",        // +
	PositionSpecialistExpert:      "Специалист-эксперт",                // +
	PositionSeniorSpecialistFirst: "Старший специалист 1 разряда",
	PositionAssistant: "Помощник", //+
	PositionDeputyHead: "Заместитель руководителя", //+
	PositionLeadSpecialist: "Ведущий специалист", //+
	PositionSeniorSpecialistSecondClass: "Старший специалист 2 разряда",
	PositionSpecialistFirstClass: "Специалист 1 разряда ",

	PositionGarageManager:    "Начальник гаража",
	PositionDriver:           "Водитель автомобиля",
	PositionWatchman:         "Сторож",
	PositionFacilityManager:  "Заведующий хозяйством",
	PositionWarehouseManager: "Заведующий складом",
	PositionDocumentClerk:    "Документовед",
	PositionRecordClerk:      "Делопроизводитель",
	PositionHandyman:         "Подсобный рабочий",
	PositionSoftwareEngineer: "Инженер-программист",
}

func NewPosition(s string) (Position, error) {
	p := Position(strings.ToLower(strings.TrimSpace(s)))
	if !p.isValid() {
		return PositionUnknown, ErrPositionInvalid
	}
	return p, nil
}

func (p Position) isValid() bool {
	_, ok := positionsMap[p]

	return ok
}

// isIrregular - является ли должность с ненормированным рабочим днем
func (p Position) isIrregular() bool {
	switch p {
	case
		PositionDirector,
		PositionHeadOfTerritorialDepartment,
		PositionDeputyDirector,
		PositionChiefSpecialistExpert,
		PositionSpecialistExpert,
		PositionAssistant,
		PositionDeputyHead,
		PositionLeadSpecialist		:
		return true
	default:
		return false
	}
}

// IsCivilServant — государственный гражданский служащий (79-ФЗ). Остальные
// должности — прочий персонал по ТК РФ: выслуга лет госслужбы им не начисляется.
func (p Position) IsCivilServant() bool {
	switch p {
	case
		PositionGarageManager,
		PositionDriver,
		PositionWatchman,
		PositionFacilityManager,
		PositionWarehouseManager,
		PositionDocumentClerk,
		PositionRecordClerk,
		PositionHandyman,
		PositionSoftwareEngineer,
		PositionUnknown:
		return false
	default:
		return true
	}
}

// Категория сотрудника по должности.
const (
	CategoryCivilServant = "civil"
	CategoryOtherStaff   = "other"
)

func (p Position) Category() string {
	if p.IsCivilServant() {
		return CategoryCivilServant
	}
	return CategoryOtherStaff
}

func CategoryTitle(category string) string {
	switch category {
	case CategoryCivilServant:
		return "Государственные служащие"
	case CategoryOtherStaff:
		return "Прочий персонал"
	}
	return ""
}

// PositionsOfCategory — коды должностей категории (для фильтра списка сотрудников).
func PositionsOfCategory(category string) []string {
	out := make([]string, 0)
	for p := range positionsMap {
		if p.Category() == category {
			out = append(out, string(p))
		}
	}
	sort.Strings(out)
	return out
}

func (p Position) String() string { return string(p) }

// Возвращает название должности на русском
func (p Position) Title() string {
	if title, ok := positionsMap[p]; ok {
		return title
	}
	return ""
}

func AllPositions() []Position {
	result := make([]Position, 0, len(positionsMap))
	for p := range positionsMap {
		result = append(result, p)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] > result[j]
	})

	return result
}
