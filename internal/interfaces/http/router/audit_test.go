package routes

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
)

// Изменения попадают в журнал; руководитель видит только записи своего теротдела.
func TestAuditLogRecordsChangesWithinCallerScope(t *testing.T) {
	c, srv := newTestApp(t)
	adminID := addUser(t, c, "admin@example.com", "lead_specialist_expert", "information_technology")
	makeAdmin(t, c, adminID)
	kazan := addUser(t, c, "kazan@example.com", "lead_specialist_expert", "information_technology")
	inNurlat := func(email, position string) string {
		t.Helper()
		created, err := c.RegisterUserUseCase.Handle(context.Background(), register_user.Command{
			Email: email, Password: testPassword, FirstName: "Гузэль", LastName: "Бадыгина",
			MiddleName: "Ринатовна", BirthDate: "1980-01-01", HiredAt: "2020-01-01", Position: position,
			DistrictCode: "nurlat", DepartmentCode: "nurlat_to", WorkdayDuration: 8,
		})
		if err != nil {
			t.Fatal(err)
		}
		return created.UserID
	}
	inNurlat("head@example.com", "head_of_territorial_department")
	nurlat := inNurlat("nurlat@example.com", "lead_specialist")

	admin := login(t, srv, "admin@example.com")
	postForm(t, admin, srv.URL+"/users/"+kazan+"/work-years/1",
		url.Values{"start": {"2020-01-01"}, "end": {"2021-03-31"}}, http.StatusOK)

	head := login(t, srv, "head@example.com")
	body := `{"userId":"` + nurlat + `","periods":[{"kind":"base","startDate":"2030-06-03","endDate":"2030-06-05"}]}`
	res, err := head.Post(srv.URL+"/api/v1/users/"+nurlat+"/vacations", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("create vacation: status %d", res.StatusCode)
	}
	// Отклонённое действие в журнал не попадает.
	get(t, head, srv.URL+"/users/"+kazan+"/edit", http.StatusForbidden)

	all := get(t, admin, srv.URL+"/admin/audit", http.StatusOK)
	mustContain(t, all,
		"Изменены границы рабочего года", "год 1: 01.01.2020–31.03.2021",
		"Оформлен отпуск", "Основной 03.06.2030–05.06.2030", "В обход правил")

	own := get(t, head, srv.URL+"/admin/audit", http.StatusOK)
	mustContain(t, own, "Оформлен отпуск", "/users/"+nurlat)
	if strings.Contains(own, "Изменены границы рабочего года") {
		t.Error("head sees an action from another territorial department")
	}
	mustContain(t, get(t, head, srv.URL+"/users/"+nurlat+"/history", http.StatusOK), "Оформлен отпуск")
	get(t, head, srv.URL+"/users/"+kazan+"/history", http.StatusForbidden)

	// Расхождений нет: страница доступна руководителю, но не сотруднику.
	mustContain(t, get(t, head, srv.URL+"/admin/discrepancies", http.StatusOK), "Расхождений нет")

	employee := login(t, srv, "nurlat@example.com")
	get(t, employee, srv.URL+"/admin/discrepancies", http.StatusForbidden)
	get(t, employee, srv.URL+"/admin/audit", http.StatusForbidden)
	get(t, employee, srv.URL+"/users/"+nurlat+"/history", http.StatusForbidden)
}
