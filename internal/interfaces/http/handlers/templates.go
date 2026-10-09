package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/handsome-red/vacation-calculation/web"
)

type Templates struct {
	registry map[string]*template.Template
}

var pages = map[string]string{
	"calendar.html":         "templates/calendar.html",
	"get_holiday_form.html": "templates/get_holiday_form.html",
	"home.html":             "templates/home.html",
	"register_user.html":    "templates/register_user.html",
	"user.html":             "templates/user.html",
	"users.html":            "templates/users.html",
	"vacations.html":        "templates/vacations.html",
	"vacation_form.html":    "templates/vacation_form.html",
	"shift_form.html":       "templates/shift_form.html",
}

func NewTemplates() (*Templates, error) {
	reg := make(map[string]*template.Template, len(pages))

	for name, path := range pages {
		// Для каждой страницы — свой набор: base + её контент
		tpl, err := template.ParseFS(web.FS,
			"templates/base.html",
			"templates/partials/*.html",
			path,
		)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		reg[name] = tpl
	}

	return &Templates{registry: reg}, nil
}

func (t *Templates) Render(w http.ResponseWriter, name string, data any) error {
	tpl, ok := t.registry[name]
	if !ok {
		return fmt.Errorf("unknown template: %q", name)
	}

	// Рендерим в буфер, чтобы при ошибке не отправить клиенту полусломанный HTML
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "base", data); err != nil {
		log.Printf("TEMPLATE ERROR: name=%q err=%v", name, err)
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := buf.WriteTo(w)
	return err
}
