package handlers

import (
	"net/http"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/shift/create_shift"
	"github.com/handsome-red/vacation-calculation/internal/application/queries/shift/shift_form"
)

type ShiftHandler struct {
	createShiftUseCase *create_shift.Handler
	shiftFormUseCase   *shift_form.Handler
	templates          *Templates
}

func NewShiftHandler(
	createShiftUseCase *create_shift.Handler,
	shiftFormUseCase *shift_form.Handler,
	templates *Templates,
) *ShiftHandler {
	return &ShiftHandler{
		createShiftUseCase: createShiftUseCase,
		shiftFormUseCase:   shiftFormUseCase,
		templates:          templates,
	}
}

func (h *ShiftHandler) CreateShift(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "cannot parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.FormValue("user_id")
	kind := r.FormValue("kind")
	fromStr := r.FormValue("date_from")
	toStr := r.FormValue("date_to")

	if userID == "" || kind == "" || fromStr == "" || toStr == "" {
		http.Error(w, "user_id, kind, date_from, date_to are required", http.StatusBadRequest)
		return
	}

	from, err := time.Parse(time.DateOnly, fromStr)
	if err != nil {
		http.Error(w, "invalid date_from, want YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	to, err := time.Parse(time.DateOnly, toStr)
	if err != nil {
		http.Error(w, "invalid date_to, want YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	cmd := create_shift.Command{
		UserID: userID,
		Kind:   kind,
		From:   from,
		To:     to,
	}

	if _, err := h.createShiftUseCase.Handle(r.Context(), cmd); err != nil {
		http.Error(w, "cannot create shift: "+err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/users/"+userID, http.StatusSeeOther)
}

func (h *ShiftHandler) NewShiftForm(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userId")
	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	result, err := h.shiftFormUseCase.Handle(r.Context(), shift_form.Query{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Cmd": create_shift.Command{
			UserID: userID,
			Kind:   "",
			From:   time.Time{},
			To:     time.Time{},
		},
		"Kinds": result.Kinds,
	}

	if err := h.templates.Render(w, "shift_form.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
