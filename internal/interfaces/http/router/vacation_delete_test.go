package routes

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
)

// Администратор удаляет отпуск из базы, в том числе перенесённый: запись, созданная
// переносом, остаётся. Сотрудник удалить отпуск не может. Удаление попадает в журнал.
func TestManagerDeletesVacationFromDatabase(t *testing.T) {
	c, srv := newTestApp(t)
	ctx := context.Background()
	adminID := addUser(t, c, "admin@example.com", "lead_specialist_expert", "information_technology")
	makeAdmin(t, c, adminID)
	id := addUser(t, c, "employee@example.com", "lead_specialist_expert", "information_technology")
	uid, err := user.ParseUserID(id)
	if err != nil {
		t.Fatal(err)
	}

	admin := login(t, srv, "admin@example.com")
	create := func(start, end string) {
		t.Helper()
		body := `{"userId":"` + id + `","periods":[{"kind":"base","startDate":"` + start + `","endDate":"` + end + `"}]}`
		res, err := admin.Post(srv.URL+"/api/v1/users/"+id+"/vacations", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("create vacation %s: status %d", start, res.StatusCode)
		}
	}
	find := func(start string) *vacation.Vacation {
		t.Helper()
		all, err := c.VacationRepo.FindByUserID(ctx, uid)
		if err != nil {
			t.Fatal(err)
		}
		for _, v := range all {
			if v.StartDate().ISO() == start {
				return v
			}
		}
		return nil
	}
	vacationsURL := srv.URL + "/api/v1/users/" + id + "/vacations/"

	create("2030-06-03", "2030-06-16")
	create("2030-08-05", "2030-08-18")
	plain, moved := find("2030-06-03"), find("2030-08-05")
	if plain == nil || moved == nil {
		t.Fatal("vacations not created")
	}
	mustContain(t, get(t, admin, srv.URL+"/api/v1/users/"+id+"/vacations", http.StatusOK), "/vacations/"+plain.ID().String()+"/delete")

	// Сотрудник свой отпуск не удаляет.
	employee := login(t, srv, "employee@example.com")
	postForm(t, employee, vacationsURL+plain.ID().String()+"/delete", url.Values{}, http.StatusForbidden)

	postForm(t, admin, vacationsURL+plain.ID().String()+"/delete", url.Values{}, http.StatusOK)
	if find("2030-06-03") != nil {
		t.Fatal("vacation still in database")
	}
	postForm(t, admin, vacationsURL+plain.ID().String()+"/delete", url.Values{}, http.StatusNotFound)

	// Перенос создаёт новую запись со ссылкой на исходную; исходная удаляется без ошибки
	// внешнего ключа, новая остаётся.
	postForm(t, admin, vacationsURL+moved.ID().String()+"/transfer",
		url.Values{"new_start": {"2030-09-02"}, "new_end": {"2030-09-15"}, "reason": {"по заявлению"}}, http.StatusOK)
	postForm(t, admin, vacationsURL+moved.ID().String()+"/delete", url.Values{}, http.StatusOK)
	child := find("2030-09-02")
	if child == nil {
		t.Fatal("vacation created by transfer was deleted too")
	}
	if !child.TransferredFrom().IsZero() {
		t.Error("child still references the deleted vacation")
	}

	mustContain(t, get(t, admin, srv.URL+"/users/"+id+"/history", http.StatusOK),
		"Отпуск удалён из базы", "Основной, 03.06.2030–16.06.2030")
}
