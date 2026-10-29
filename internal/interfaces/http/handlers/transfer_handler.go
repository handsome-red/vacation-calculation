package handlers

import (
	"net/http"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/transfer_vacation"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/vacation/get_transfer_form"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/middleware"
)

// TransferHandler — перенос отпуска администратором с сохранением истории.
type TransferHandler struct {
	transfer  *transfer_vacation.Handler
	form      *get_transfer_form.Handler
	templates *Templates
	logger    ports.Logger
}

func NewTransferHandler(
	transfer *transfer_vacation.Handler,
	form *get_transfer_form.Handler,
	templates *Templates,
	logger ports.Logger,
) *TransferHandler {
	return &TransferHandler{transfer: transfer, form: form, templates: templates, logger: logger}
}

type transferFormData struct {
	Vacation *get_transfer_form.Result
	NewStart string
	NewEnd   string
	Reason   string
	Error    string
}

func (h *TransferHandler) Form(w http.ResponseWriter, r *http.Request) {
	v, err := h.form.Handle(r.Context(), get_transfer_form.Query{
		UserID:     r.PathValue("userId"),
		VacationID: r.PathValue("id"),
	})
	if err != nil {
		http.Error(w, "Отпуск не найден", http.StatusNotFound)
		return
	}

	data := transferFormData{Vacation: v, NewStart: v.StartDate, NewEnd: v.EndDate}
	if !v.Transferable {
		data.Error = "Отпуск со статусом «" + v.Status + "» перенести нельзя."
	}
	if err := h.templates.Render(w, "transfer_form.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TransferHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	caller, ok := middleware.CallerFromContext(r.Context())
	if !ok {
		http.Error(w, "Необходим вход", http.StatusUnauthorized)
		return
	}

	cmd := transfer_vacation.Command{
		UserID:     r.PathValue("userId"),
		VacationID: r.PathValue("id"),
		NewStart:   r.FormValue("new_start"),
		NewEnd:     r.FormValue("new_end"),
		Reason:     r.FormValue("reason"),
	}

	result, err := h.transfer.Handle(r.Context(), cmd, caller)
	if err != nil {
		status := vacationErrorStatus(err)
		if status == http.StatusInternalServerError {
			h.logger.Error(r.Context(), "transfer vacation failed", "error", err)
		}
		v, ferr := h.form.Handle(r.Context(), get_transfer_form.Query{UserID: cmd.UserID, VacationID: cmd.VacationID})
		if ferr != nil {
			http.Error(w, err.Error(), status)
			return
		}
		w.WriteHeader(status)
		_ = h.templates.Render(w, "transfer_form.html", transferFormData{
			Vacation: v, NewStart: cmd.NewStart, NewEnd: cmd.NewEnd, Reason: cmd.Reason, Error: err.Error(),
		})
		return
	}

	h.logger.Info(r.Context(), "vacation transferred",
		"user_id", cmd.UserID, "vacation_id", cmd.VacationID, "transfer_id", result.TransferID,
		"admin_id", caller.UserID.String(), "warnings", result.Warnings)
	http.Redirect(w, r, "/api/v1/users/"+cmd.UserID+"/vacations", http.StatusSeeOther)
}
