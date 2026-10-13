package excel

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/xuri/excelize/v2"
)

type reportExporter struct{}

var _ ports.ReportExporter = (*reportExporter)(nil)

func NewReportExporter() *reportExporter {
	return &reportExporter{}
}

func (e *reportExporter) Export(_ context.Context, r vacation.Report) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	const sheet = "График отпусков"

	f.MergeCell(sheet, "H1", "I4")
	f.SetCellValue(sheet, "H1",
		fmt.Sprintf("Утверждаю\n%s\n%s\n«___» ____________ %d г.",
			r.Approver.Position, r.Approver.FullName, r.Year))
	// f.SetCellStyle(sheet, "H1", "I4", approverStyle(f))

	// 2. Заголовок по центру: строка 5 — «График отпусков на N год»,
	//    строка 6 — название территориального отдела.
	f.MergeCell(sheet, "A5", "J5")
	f.SetCellValue(sheet, "A5", fmt.Sprintf("График отпусков на %d год", r.Year))
	// f.SetCellStyle(sheet, "A5", "J5", titleStyle(f))

	f.MergeCell(sheet, "A6", "J6")
	f.SetCellValue(sheet, "A6", fmt.Sprintf("Территориальный отдел %s", r.DistrictTitle))
	// f.SetCellStyle(sheet, "A6", "J6", titleStyle(f))

	// 3. Шапка таблицы (строка 7).
	headers := []string{
		"Должность",
		"ФИО",
		"Кол-во дней основного отпуска",
		"Запланированная дата",
		"Кол-во дней ненормированного служебного отпуска",
		"Запланированная дата",
		"Кол-во дней за выслугу лет",
		"Запланированная дата",
		"Период рабочего года",
		"С графиком ознакомлен",
		"О времени начала извещен",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 7)
		f.SetCellValue(sheet, cell, h)
	}
	// f.SetCellStyle(sheet, "A7", "K7", headerStyle(f))

	// 4. Строки данных (с 8-й).
	for i, row := range r.Rows {
		r := 8 + i
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), row.Position)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), row.FullName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", r), row.BaseDays)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", r), joinDates(row.BaseDates))
		f.SetCellValue(sheet, fmt.Sprintf("E%d", r), row.IrregularDays)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", r), joinDates(row.IrregularDates))
		f.SetCellValue(sheet, fmt.Sprintf("G%d", r), row.SeniorityDays)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", r), joinDates(row.SeniorityDates))
		f.SetCellValue(sheet, fmt.Sprintf("I%d", r),
			fmt.Sprintf("%s – %s",
				row.WorkYearFrom.Format("02.01.2006"),
				row.WorkYearTo.Format("02.01.2006")))
		// J и K — пустые, под подпись.
	}
	// f.SetCellStyle(sheet, "A8", fmt.Sprintf("K%d", 7+len(r.Rows)), bodyStyle(f))

	// 5. Границы таблицы.
	// f.SetCellStyle(sheet, "A7", fmt.Sprintf("K%d", 7+len(r.Rows)), borderedStyle(f))

	// 6. Ширины колонок.
	f.SetColWidth(sheet, "A", "A", 30)
	f.SetColWidth(sheet, "B", "B", 30)
	f.SetColWidth(sheet, "C", "C", 18)
	f.SetColWidth(sheet, "D", "D", 20)
	// ...

	// 7. Ориентация — альбомная, чтобы влезло.
	if idx, err := f.GetSheetIndex(sheet); err == nil {
		f.SetPageLayout(sheet, &excelize.PageLayoutOptions{
			Orientation: strPtr("landscape"),
			Size:        intPtr(9), // A4
		})
		_ = idx
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func joinDates(dates []time.Time) string {
	parts := make([]string, 0, len(dates))
	for _, d := range dates {
		parts = append(parts, d.Format("02.01.2006"))
	}
	return strings.Join(parts, ", ")
}
