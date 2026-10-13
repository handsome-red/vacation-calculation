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
	PositionDeputyDirector Position = "deputy_director"

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
	PositionDirector:              "Начальник территориального отдела", // +
	PositionDeputyDirector:        "Заместитель начальника",            // +
	PositionChiefSpecialistExpert: "Главный специалист-эксперт",        // +
	PositionSpecialistExpert:      "Специалист-эксперт",                // +
	PositionSeniorSpecialistFirst: "Старший специалист 1 разряда",
	// Помощник +
	// Рук  +
	// Зам рук +
	// Ведущий +
	// Старший спец 2 раз
	// Специалист 1 раз

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
		PositionDeputyDirector,
		PositionChiefSpecialistExpert,
		PositionSpecialistExpert,
		PositionSeniorSpecialistFirst:
		return true
	default:
		return false
	}
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
