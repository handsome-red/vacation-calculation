// Package docx формирует документы Word (Office Open XML) без внешних зависимостей.
package docx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

type ScheduleRenderer struct{}

var _ ports.ScheduleRenderer = ScheduleRenderer{}

func NewScheduleRenderer() ScheduleRenderer { return ScheduleRenderer{} }

// Ширины колонок таблицы графика (twips) — альбомный A4 с полями 1,5 см: 15138 доступно.
var colWidths = []int{2000, 2300, 900, 1500, 900, 1500, 900, 1500, 1500, 1100, 1100}

// RenderSchedule — «График отпусков за ____ год» для территориального отдела.
func (ScheduleRenderer) RenderSchedule(r ports.ScheduleReport) ([]byte, error) {
	var body strings.Builder

	// Гриф утверждения — справа вверху.
	for _, line := range []string{
		"УТВЕРЖДАЮ",
		"Руководитель Управления Роспотребнадзора по _____",
		"М.А. Патяшина",
		"«___» ____________ 20__ г.",
	} {
		body.WriteString(paragraph(line, "right", false, 24, 0))
	}
	body.WriteString(paragraph("", "left", false, 24, 0))

	// Заголовок.
	body.WriteString(paragraph(fmt.Sprintf("График отпусков за %d год", r.Year), "center", true, 28, 0))
	departments := strings.Join(r.Departments, ", ")
	if departments == "" {
		departments = "____"
	}
	body.WriteString(paragraph("Территориальный отдел Управления Федеральной службы по надзору в сфере защиты прав "+
		"потребителей и благополучия человека по Республике Татарстан (Татарстан) в "+departments, "center", false, 24, 120))

	body.WriteString(scheduleTable(r.Rows))
	body.WriteString(paragraph("", "left", false, 24, 0))

	// Подписи — по центру, под чертой «(подпись)» и «(Ф.И.О.)».
	body.WriteString(signatureTable([]string{
		"Начальник отдела/территориального отдела",
		"Начальник отдела государственной службы и кадров",
	}))
	// Word ожидает абзац между последней таблицей и параметрами раздела.
	body.WriteString(paragraph("", "left", false, 24, 0))

	body.WriteString(`<w:sectPr><w:pgSz w:w="16838" w:h="11906" w:orient="landscape"/>` +
		`<w:pgMar w:top="850" w:right="850" w:bottom="850" w:left="850" w:header="567" w:footer="567" w:gutter="0"/></w:sectPr>`)

	document := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>` +
		body.String() + `</w:body></w:document>`

	return pack(document)
}

func scheduleTable(rows []ports.ScheduleReportRow) string {
	var b strings.Builder
	b.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/><w:jc w:val="center"/>` + borders("single") +
		`<w:tblLayout w:type="fixed"/><w:tblCellMar><w:left w:w="57" w:type="dxa"/><w:right w:w="57" w:type="dxa"/></w:tblCellMar></w:tblPr><w:tblGrid>`)
	for _, w := range colWidths {
		b.WriteString(`<w:gridCol w:w="` + strconv.Itoa(w) + `"/>`)
	}
	b.WriteString(`</w:tblGrid>`)

	// Шапка в две строки; повторяется на каждой странице.
	const hdr = `<w:trPr><w:tblHeader/></w:trPr>`
	b.WriteString(`<w:tr>` + hdr)
	b.WriteString(cell(colWidths[0], []string{"Должность (специальность, профессия) по штатному расписанию"}, "restart", 1, true))
	b.WriteString(cell(colWidths[1], []string{"Фамилия, имя, отчество"}, "restart", 1, true))
	b.WriteString(cell(colWidths[2]+colWidths[3], []string{"Основной отпуск"}, "", 2, true))
	b.WriteString(cell(colWidths[4]+colWidths[5], []string{"Дополнительный отпуск за ненормированный служебный день"}, "", 2, true))
	b.WriteString(cell(colWidths[6]+colWidths[7], []string{"Дополнительный отпуск за выслугу лет"}, "", 2, true))
	b.WriteString(cell(colWidths[8], []string{"Период работы, за который предоставляется отпуск"}, "restart", 1, true))
	b.WriteString(cell(colWidths[9], []string{"С графиком ознакомлен (дата, подпись)"}, "restart", 1, true))
	b.WriteString(cell(colWidths[10], []string{"О времени начала отпуска извещен (дата, подпись)"}, "restart", 1, true))
	b.WriteString(`</w:tr><w:tr>` + hdr)
	b.WriteString(cell(colWidths[0], nil, "continue", 1, true))
	b.WriteString(cell(colWidths[1], nil, "continue", 1, true))
	for i := 0; i < 3; i++ {
		b.WriteString(cell(colWidths[2+2*i], []string{"Количество календарных дней"}, "", 1, true))
		b.WriteString(cell(colWidths[3+2*i], []string{"Запланированная дата"}, "", 1, true))
	}
	b.WriteString(cell(colWidths[8], nil, "continue", 1, true))
	b.WriteString(cell(colWidths[9], nil, "continue", 1, true))
	b.WriteString(cell(colWidths[10], nil, "continue", 1, true))
	b.WriteString(`</w:tr>`)

	days := func(n int) []string {
		if n == 0 {
			return nil
		}
		return []string{strconv.Itoa(n)}
	}
	for _, r := range rows {
		b.WriteString(`<w:tr><w:trPr><w:cantSplit/></w:trPr>`)
		b.WriteString(cell(colWidths[0], []string{r.Position}, "", 1, false))
		b.WriteString(cell(colWidths[1], []string{r.FullName}, "", 1, false))
		b.WriteString(cell(colWidths[2], days(r.BaseDays), "", 1, false))
		b.WriteString(cell(colWidths[3], r.BaseDates, "", 1, false))
		b.WriteString(cell(colWidths[4], days(r.IrregularDays), "", 1, false))
		b.WriteString(cell(colWidths[5], r.IrregularDates, "", 1, false))
		b.WriteString(cell(colWidths[6], days(r.SeniorityDays), "", 1, false))
		b.WriteString(cell(colWidths[7], r.SeniorityDates, "", 1, false))
		b.WriteString(cell(colWidths[8], []string{r.WorkPeriod}, "", 1, false))
		b.WriteString(cell(colWidths[9], nil, "", 1, false))
		b.WriteString(cell(colWidths[10], nil, "", 1, false))
		b.WriteString(`</w:tr>`)
	}
	if len(rows) == 0 {
		total := 0
		for _, w := range colWidths {
			total += w
		}
		b.WriteString(`<w:tr>` + cell(total, []string{"Отпуска на этот год не запланированы"}, "", len(colWidths), false) + `</w:tr>`)
	}
	b.WriteString(`</w:tbl>`)
	return b.String()
}

// signatureTable — «Должность ________ ________» с подписями «(подпись)» и «(Ф.И.О.)» под чертой.
func signatureTable(titles []string) string {
	widths := []int{6200, 2800, 600, 3600}
	var b strings.Builder
	b.WriteString(`<w:tbl><w:tblPr><w:tblW w:w="0" w:type="auto"/><w:jc w:val="center"/>` + borders("nil") + `</w:tblPr><w:tblGrid>`)
	for _, w := range widths {
		b.WriteString(`<w:gridCol w:w="` + strconv.Itoa(w) + `"/>`)
	}
	b.WriteString(`</w:tblGrid>`)
	underline := `<w:tcBorders><w:bottom w:val="single" w:sz="4" w:space="0" w:color="000000"/></w:tcBorders>`
	for _, title := range titles {
		b.WriteString(`<w:tr>`)
		b.WriteString(plainCell(widths[0], title, "left", 24, ""))
		b.WriteString(plainCell(widths[1], "", "center", 24, underline))
		b.WriteString(plainCell(widths[2], "", "center", 24, ""))
		b.WriteString(plainCell(widths[3], "", "center", 24, underline))
		b.WriteString(`</w:tr><w:tr>`)
		b.WriteString(plainCell(widths[0], "", "left", 18, ""))
		b.WriteString(plainCell(widths[1], "(подпись)", "center", 18, ""))
		b.WriteString(plainCell(widths[2], "", "center", 18, ""))
		b.WriteString(plainCell(widths[3], "(Ф.И.О.)", "center", 18, ""))
		b.WriteString(`</w:tr>`)
	}
	b.WriteString(`</w:tbl>`)
	return b.String()
}

func borders(val string) string {
	attrs := `w:val="` + val + `"`
	if val != "nil" {
		attrs += ` w:sz="4" w:space="0" w:color="000000"`
	}
	var b strings.Builder
	b.WriteString(`<w:tblBorders>`)
	for _, side := range []string{"top", "left", "bottom", "right", "insideH", "insideV"} {
		b.WriteString(`<w:` + side + ` ` + attrs + `/>`)
	}
	b.WriteString(`</w:tblBorders>`)
	return b.String()
}

// cell — ячейка таблицы графика: строки текста шрифтом 10 pt, по центру.
func cell(width int, lines []string, vMerge string, span int, bold bool) string {
	var b strings.Builder
	b.WriteString(`<w:tc><w:tcPr><w:tcW w:w="` + strconv.Itoa(width) + `" w:type="dxa"/>`)
	if span > 1 {
		b.WriteString(`<w:gridSpan w:val="` + strconv.Itoa(span) + `"/>`)
	}
	switch vMerge {
	case "restart":
		b.WriteString(`<w:vMerge w:val="restart"/>`)
	case "continue":
		b.WriteString(`<w:vMerge/>`)
	}
	b.WriteString(`<w:vAlign w:val="center"/></w:tcPr><w:p><w:pPr><w:jc w:val="center"/><w:spacing w:before="0" w:after="0"/></w:pPr>`)
	for i, line := range lines {
		if i > 0 {
			b.WriteString(`<w:r><w:br/></w:r>`)
		}
		b.WriteString(run(line, bold, 20))
	}
	b.WriteString(`</w:p></w:tc>`)
	return b.String()
}

func plainCell(width int, text, align string, size int, extraPr string) string {
	return `<w:tc><w:tcPr><w:tcW w:w="` + strconv.Itoa(width) + `" w:type="dxa"/>` + extraPr + `<w:vAlign w:val="bottom"/></w:tcPr>` +
		`<w:p><w:pPr><w:jc w:val="` + align + `"/><w:spacing w:before="0" w:after="0"/></w:pPr>` + run(text, false, size) + `</w:p></w:tc>`
}

func paragraph(text, align string, bold bool, size, after int) string {
	return `<w:p><w:pPr><w:jc w:val="` + align + `"/><w:spacing w:before="0" w:after="` + strconv.Itoa(after) + `"/></w:pPr>` +
		run(text, bold, size) + `</w:p>`
}

// run — фрагмент текста Times New Roman; size — в половинах пункта.
func run(text string, bold bool, size int) string {
	if text == "" {
		return ""
	}
	var rPr strings.Builder
	rPr.WriteString(`<w:rPr><w:rFonts w:ascii="Times New Roman" w:hAnsi="Times New Roman" w:cs="Times New Roman"/>`)
	if bold {
		rPr.WriteString(`<w:b/>`)
	}
	rPr.WriteString(`<w:sz w:val="` + strconv.Itoa(size) + `"/><w:szCs w:val="` + strconv.Itoa(size) + `"/></w:rPr>`)
	return `<w:r>` + rPr.String() + `<w:t xml:space="preserve">` + escape(text) + `</w:t></w:r>`
}

func escape(s string) string {
	var b bytes.Buffer
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// pack собирает минимальный пакет .docx.
func pack(document string) ([]byte, error) {
	files := []struct{ name, body string }{
		{"[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
			`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
			`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
			`<Default Extension="xml" ContentType="application/xml"/>` +
			`<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>` +
			`</Types>`},
		{"_rels/.rels", `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
			`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
			`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>` +
			`</Relationships>`},
		{"word/document.xml", document},
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range files {
		w, err := zw.Create(f.name)
		if err != nil {
			return nil, fmt.Errorf("docx: %w", err)
		}
		if _, err := w.Write([]byte(f.body)); err != nil {
			return nil, fmt.Errorf("docx: %w", err)
		}
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("docx: %w", err)
	}
	return buf.Bytes(), nil
}
