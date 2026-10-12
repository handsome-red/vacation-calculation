// internal/interfaces/http/handlers/registration.go
package handlers

import (
	"net/http"

	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/register_form"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
)

type RegistrationHandler struct {
	register     *register_user.Handler
	registerForm *register_form.Handler
	templates    *Templates
}

func NewRegistrationHandler(
	register *register_user.Handler,
	registerForm *register_form.Handler,
	templates *Templates,
) *RegistrationHandler {
	return &RegistrationHandler{
		register:     register,
		registerForm: registerForm,
		templates:    templates,
	}
}

// Form отдаёт HTML-форму регистрации.
func (h *RegistrationHandler) Form(w http.ResponseWriter, r *http.Request) {
	view, err := h.registerForm.Handle(r.Context(), register_form.Query{})
	if err != nil {
		http.Error(w, "form unavailable", http.StatusInternalServerError)
		return
	}
	h.templates.Render(w, "register.html", view)
}

// Register обрабатывает POST регистрации.
func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	_, err := h.register.Handle(r.Context(), register_user.Command{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	})
	if err != nil {
		// Здесь нужна нормальная обработка доменных ошибок → 400/409.
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}