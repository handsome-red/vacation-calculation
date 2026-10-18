// internal/interfaces/http/handlers/registration.go
package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/user/register_form"
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
		log.Printf("register form: %v", err)
		http.Error(w, "form unavailable", http.StatusInternalServerError)
		return
	}
	h.templates.Render(w, "register_user.html", view)
}

// Register обрабатывает POST регистрации.
func (h *RegistrationHandler) Register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	experience, err :=  strconv.Atoi(r.FormValue("experience"))
	if err != nil {
		experience = 0
	}

	workday, err := strconv.Atoi(r.FormValue("workday_duration"))
	if err != nil {
		workday = 0 // валидация ниже, в use case
	}

	cmd := register_user.Command{
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		FirstName:       r.FormValue("first_name"),
		LastName:        r.FormValue("last_name"),
		MiddleName:      r.FormValue("middle_name"),
		DepartmentCode:  r.FormValue("department"),
		DistrictCode:    r.FormValue("district"),
		Position:        r.FormValue("position"),
		BirthDate:       r.FormValue("birth_date"),
		Experience: 	experience,
		HiredAt:         r.FormValue("hired_at"),
		WorkdayDuration: workday,
		IsInvalid:       r.FormValue("is_invalid") == "true",
	}

	if _, err := h.register.Handle(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}