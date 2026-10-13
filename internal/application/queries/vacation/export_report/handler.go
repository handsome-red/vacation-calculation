package export_report

import (
	"context"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

type Handler struct {
	users       ports.UserRepository
	vacations   ports.VacationRepository
	departments ports.DepartmentRepository
	districts   ports.DistrictRepository
	calc        vacation.WorkYearCalculator
	exporter    ports.ReportExporter
}

func NewHandler(
	users ports.UserRepository,
	vacations ports.VacationRepository,
	departments ports.DepartmentRepository,
	districts ports.DistrictRepository,
	calc vacation.WorkYearCalculator,
	exporter ports.ReportExporter,
) *Handler {
	return &Handler{
		users:       users,
		vacations:   vacations,
		departments: departments,
		districts:   districts,
		calc:        calc,
		exporter:    exporter,
	}
}

func (h *Handler) Handle(ctx context.Context, q Query) ([]byte, error) {
	// from := vacation.NewDate(time.Date(q.Year, 1, 1, 0, 0, 0, 0, time.UTC))
	// to   := vacation.NewDate(time.Date(q.Year, 12, 31, 0, 0, 0, 0, time.UTC))

	// users, err := h.listUsers(ctx, q.DistrictCode)
	// if err != nil { return nil, err }

	// // 3. Для каждого пользователя собрать строку.
	// rows := make([]vacation.ScheduleRow, 0, len(users))
	// for _, u := range users {
	// 	vacs, err := h.vacations.FindByUserIDInRange(ctx, u.ID(), from, to)
	// 	if err != nil { return nil, err }

	// 	row := vacation.ScheduleRow{
	// 		Position: u.Position().Title(),
	// 		FullName: u.FullName(),
	// 	}

	// 	// Разложить отпуска по типам: base / irregular / seniority.
	// 	// Здесь нужно знать, как у вас кодируются типы отпусков.
	// 	for _, v := range vacs {
	// 		switch v.Kind() {
	// 		case vacation.KindBase:
	// 			row.BaseDays    += vacation.DaysInclusive(v.StartDate().Time(), v.EndDate().Time())
	// 			row.BaseDates    = append(row.BaseDates, v.StartDate().Time())
	// 		case vacation.KindIrregular:
	// 			row.IrregularDays += vacation.DaysInclusive(v.StartDate().Time(), v.EndDate().Time())
	// 			row.IrregularDates = append(row.IrregularDates, v.StartDate().Time())
	// 		case vacation.KindSeniority:
	// 			row.SeniorityDays += vacation.DaysInclusive(v.StartDate().Time(), v.EndDate().Time())
	// 			row.SeniorityDates = append(row.SeniorityDates, v.StartDate().Time())
	// 		}
	// 	}

	// 	// Рабочий год — из YearStat или из чего-то похожего.
	// 	// Возможно, у вас есть отдельный запрос/калькулятор.
	// 	row.WorkYearFrom = _
	// 	row.WorkYearTo   = _

	// 	rows = append(rows, row)
	// }

	// // 4. Собрать шапку и название района.
	// districtTitle := ""
	// if q.DistrictCode != "" {
	// 	d, err := h.districts.FindByCode(ctx, q.DistrictCode)
	// 	if err != nil { return nil, err }
	// 	districtTitle = d.Title
	// }

	// return h.exporter.Export(ctx, vacation.Schedule{
	// 	Year:          q.Year,
	// 	DistrictTitle: districtTitle,
	// 	Approver:      vacation.Approver{ /* из конфига/настроек */ },
	// 	Rows:          rows,
	// })
	return nil, nil
}
