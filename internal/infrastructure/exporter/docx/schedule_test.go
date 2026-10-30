package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
)

func TestRenderSchedule(t *testing.T) {
	data, err := NewScheduleRenderer().RenderSchedule(ports.ScheduleReport{
		Year:        2027,
		Departments: []string{"Отдел эпидемиологического надзора"},
		Rows: []ports.ScheduleReportRow{{
			Position: "Специалист-эксперт", FullName: "Иванов Пётр Александрович",
			BaseDays: 30, BaseDates: []string{"01.06.2027 — 14.06.2027", "01.08.2027 — 16.08.2027"},
			IrregularDays: 3, IrregularDates: []string{"17.08.2027 — 19.08.2027"},
			WorkPeriod: "10.03.2026-09.03.2027",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("not a zip: %v", err)
	}
	var doc string
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, _ := f.Open()
			b, _ := io.ReadAll(rc)
			doc = string(b)
		}
	}
	if doc == "" {
		t.Fatal("word/document.xml missing")
	}
	// корректный XML
	if err := xml.Unmarshal([]byte(doc), new(struct{ XMLName xml.Name })); err != nil {
		t.Fatalf("document.xml is not valid XML: %v", err)
	}
	for _, want := range []string{
		"УТВЕРЖДАЮ", "М.А. Патяшина", "График отпусков за 2027 год",
		"по Республике Татарстан (Татарстан) в Отдел эпидемиологического надзора",
		"Иванов Пётр Александрович", "01.08.2027 — 16.08.2027", "10.03.2026-09.03.2027",
		"С графиком ознакомлен (дата, подпись)", "Начальник отдела государственной службы и кадров",
		"(подпись)", "(Ф.И.О.)", `w:orient="landscape"`,
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("document does not contain %q", want)
		}
	}
}
