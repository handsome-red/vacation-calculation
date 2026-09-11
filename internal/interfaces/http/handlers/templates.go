package handlers

import (
	"html/template"
	"net/http"

	"github.com/handsome-red/vacation-calculation/web"
)

type Templates struct {
	tpl *template.Template
}

func NewTemplates() (*Templates, error) {
	tpl, err := template.ParseFS(web.FS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Templates{tpl: tpl}, nil
}

func (t *Templates) Render(w http.ResponseWriter, name string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return t.tpl.ExecuteTemplate(w, name, data)
}
