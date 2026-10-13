package excel

import "github.com/xuri/excelize/v2"

// approverStyle — шапка «Утверждаю»: жирный, выравнивание по левому краю, перенос строк.
func approverStyle(f *excelize.File) *excelize.Style {
	return &excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 11,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "top",
			WrapText:   true,
		},
	}
}

// titleStyle — заголовок по центру: жирный, по центру.
func titleStyle(f *excelize.File) *excelize.Style {
	return &excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
	}
}

// headerStyle — шапка таблицы: жирный, по центру, с границами, с переносом.
func headerStyle(f *excelize.File) *excelize.Style {
	return &excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 10,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	}
}

// bodyStyle — строки данных: обычный шрифт, выравнивание, перенос.
func bodyStyle(f *excelize.File) *excelize.Style {
	return &excelize.Style{
		Font: &excelize.Font{
			Size: 10,
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
		},
	}
}

// borderedStyle — границы для всей таблицы.
func borderedStyle(f *excelize.File) *excelize.Style {
	return &excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	}
}

// strPtr / intPtr — утилиты для указателей (нужны для PageLayoutOptions).
func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
