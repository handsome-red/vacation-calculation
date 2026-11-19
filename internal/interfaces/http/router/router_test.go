package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/approve_vacation"
	change_status "github.com/handsome-red/vacation-calculation/internal/application/commands/vacation/change_status"
	"github.com/handsome-red/vacation-calculation/internal/config"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/di"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/persistence/sqlite"
)

type noopLogger struct{}

func (noopLogger) Debug(context.Context, string, ...any) {}
func (noopLogger) Info(context.Context, string, ...any)  {}
func (noopLogger) Warn(context.Context, string, ...any)  {}
func (noopLogger) Error(context.Context, string, ...any) {}

const testPassword = "test-password-123"

func newTestApp(t *testing.T) (*di.Container, *httptest.Server) {
	t.Helper()
	container, err := di.NewContainer(context.Background(), config.Config{DatabasePath: filepath.Join(t.TempDir(), "app.db")}, noopLogger{})
	if err != nil {
		t.Fatalf("container: %v", err)
	}
	t.Cleanup(func() { _ = container.Close() })

	router, err := NewRouter(container, noopLogger{}, nil)
	if err != nil {
		t.Fatalf("router: %v", err)
	}
	handler, err := router.Build()
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return container, srv
}

func addUser(t *testing.T, c *di.Container, email, position, department string) string {
	t.Helper()
	return addUserHired(t, c, email, position, department, "2020-01-01")
}

func addUserHired(t *testing.T, c *di.Container, email, position, department, hiredAt string) string {
	t.Helper()
	created, err := c.RegisterUserUseCase.Handle(context.Background(), register_user.Command{
		Email: email, Password: testPassword, FirstName: "Иван", LastName: "Иванов",
		MiddleName: "Иванович", BirthDate: "1980-01-01", HiredAt: hiredAt, Position: position,
		DistrictCode: "kazan", DepartmentCode: department, WorkdayDuration: 8,
	})
	if err != nil {
		t.Fatalf("register %s: %v", email, err)
	}
	return created.UserID
}

func makeAdmin(t *testing.T, c *di.Container, id string) {
	t.Helper()
	ctx := context.Background()
	uid, err := user.ParseUserID(id)
	if err != nil {
		t.Fatal(err)
	}
	u, err := c.UserRepo.FindByID(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	u.GrantAdmin()
	if err := c.UserRepo.Save(ctx, u); err != nil {
		t.Fatal(err)
	}
}

func login(t *testing.T, srv *httptest.Server, email string) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar}
	res, err := client.PostForm(srv.URL+"/login", url.Values{"email": {email}, "password": {testPassword}})
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("login %s: status %d", email, res.StatusCode)
	}
	return client
}

func get(t *testing.T, client *http.Client, target string, wantStatus int) string {
	t.Helper()
	res, err := client.Get(target)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != wantStatus {
		t.Fatalf("GET %s: status %d, want %d: %s", target, res.StatusCode, wantStatus, body)
	}
	return string(body)
}

func postForm(t *testing.T, client *http.Client, target string, form url.Values, wantStatus int) string {
	t.Helper()
	res, err := client.PostForm(target, form)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != wantStatus {
		t.Fatalf("POST %s: status %d, want %d: %s", target, res.StatusCode, wantStatus, body)
	}
	return string(body)
}

func mustContain(t *testing.T, body string, parts ...string) {
	t.Helper()
	for _, p := range parts {
		if !strings.Contains(body, p) {
			t.Errorf("response does not contain %q", p)
		}
	}
}

// Сотрудник сам выбирает вид отпуска для каждого периода; в списке видно число дней.
func TestEmployeeChoosesVacationKindPerPeriod(t *testing.T) {
	c, srv := newTestApp(t)
	id := addUser(t, c, "employee@example.com", "lead_specialist", "epidemiological_supervision")
	client := login(t, srv, "employee@example.com")

	mustContain(t, get(t, client, srv.URL+"/api/v1/users/"+id+"/vacations/new", http.StatusOK),
		`value="seniority"`, `value="irregular"`, `value="base"`, `value="auto"`)

	start := time.Date(time.Now().Year()+1, time.July, 1, 0, 0, 0, 0, time.UTC) // без праздников
	iso := func(d int) string { return start.AddDate(0, 0, d).Format("2006-01-02") }
	payload := fmt.Sprintf(`{"userId":%q,"periods":[
		{"kind":"seniority","startDate":%q,"endDate":%q},
		{"kind":"base","startDate":%q,"endDate":%q}
	]}`, id, iso(0), iso(0), iso(1), iso(14))

	res, err := client.Post(srv.URL+"/api/v1/users/"+id+"/vacations/preview", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var preview struct {
		Allocations []struct {
			Kind string `json:"kind"`
			Days int    `json:"days"`
		} `json:"allocations"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&preview); err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("preview: status %d: %s", res.StatusCode, preview.Error)
	}
	if len(preview.Allocations) != 2 || preview.Allocations[0].Kind != "seniority" || preview.Allocations[1].Kind != "base" {
		t.Fatalf("allocations = %+v", preview.Allocations)
	}

	created, err := client.Post(srv.URL+"/api/v1/users/"+id+"/vacations", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	created.Body.Close()
	if created.StatusCode != http.StatusOK {
		t.Fatalf("create: status %d", created.StatusCode)
	}

	mustContain(t, get(t, client, srv.URL+"/api/v1/users/"+id+"/vacations", http.StatusOK),
		"<th>Всего дней</th>", "<td>1</td>", "<td>14</td>", "За выслугу лет", "Основной")
}

// Начальник территориального отдела (как Бадыгина в Нурлатском ТО) получает права
// администратора в своём теротделе автоматически, по должности: видит его сотрудников
// и оформляет им отпуска, но не трогает другие теротделы, администраторов и праздники.
func TestHeadOfTerritorialDepartmentActsAsAdminWithinOwnDistrict(t *testing.T) {
	c, srv := newTestApp(t)
	ctx := context.Background()
	adminID := addUser(t, c, "admin@example.com", "software_engineer", "epidemiological_supervision")
	makeAdmin(t, c, adminID)
	outside := addUser(t, c, "outside@example.com", "software_engineer", "epidemiological_supervision") // Казань
	inNurlat := func(email, position string) string {
		t.Helper()
		created, err := c.RegisterUserUseCase.Handle(ctx, register_user.Command{
			Email: email, Password: testPassword, FirstName: "Гузэль", LastName: "Бадыгина",
			MiddleName: "Ринатовна", BirthDate: "1980-01-01", HiredAt: "2020-01-01", Position: position,
			DistrictCode: "nurlat", DepartmentCode: "nurlat_to", WorkdayDuration: 8,
		})
		if err != nil {
			t.Fatal(err)
		}
		return created.UserID
	}
	headID := inNurlat("head@example.com", "head_of_territorial_department")
	inside := inNurlat("inside@example.com", "lead_specialist")

	profile := func(district, department, email string) url.Values {
		return url.Values{
			"last_name": {"Иванов"}, "first_name": {"Иван"}, "middle_name": {"Иванович"}, "email": {email},
			"birth_date": {"1980-01-01"}, "position": {"lead_specialist"}, "experience": {"0"},
			"hired_at": {"2020-01-01"}, "district": {district}, "department": {department}, "workday_duration": {"8"},
		}
	}

	admin := login(t, srv, "admin@example.com")
	mustContain(t, get(t, admin, srv.URL+"/users/"+headID, http.StatusOK), "Руководитель территориального отдела")

	head := login(t, srv, "head@example.com")
	list := get(t, head, srv.URL+"/users", http.StatusOK)
	mustContain(t, list, "/users/"+inside)
	for _, hidden := range []string{outside, adminID} {
		if strings.Contains(list, "/users/"+hidden) {
			t.Errorf("head's list contains %s", hidden)
		}
	}

	// Свой теротдел — как администратор, чужой и администратор — запрещено.
	mustContain(t, get(t, head, srv.URL+"/users/"+inside, http.StatusOK), `/users/`+inside+`/edit`)
	get(t, head, srv.URL+"/users/"+outside, http.StatusForbidden)
	get(t, head, srv.URL+"/users/"+adminID, http.StatusForbidden)
	get(t, head, srv.URL+"/users/"+outside+"/edit", http.StatusForbidden)
	get(t, head, srv.URL+"/api/v1/users/"+outside+"/vacations", http.StatusForbidden)
	mustContain(t, get(t, head, srv.URL+"/api/v1/users/"+inside+"/vacations", http.StatusOK), "Изменить статус")
	get(t, head, srv.URL+"/users/"+inside+"/work-years", http.StatusOK)

	// Отпуск сотруднику своего ТО — как администратор: нарушение правил (3 дня) лишь предупреждение.
	short := `{"userId":"` + inside + `","periods":[{"kind":"base","startDate":"2030-06-03","endDate":"2030-06-05"}]}`
	res, err := head.Post(srv.URL+"/api/v1/users/"+inside+"/vacations/preview", "application/json", strings.NewReader(short))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(body), `"warnings":["`) {
		t.Fatalf("head preview for own employee: %d %s", res.StatusCode, body)
	}
	other := strings.ReplaceAll(short, inside, outside)
	res, err = head.Post(srv.URL+"/api/v1/users/"+outside+"/vacations/preview", "application/json", strings.NewReader(other))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("head preview for other district: %d", res.StatusCode)
	}

	// Перевести сотрудника или добавить нового в чужой теротдел нельзя.
	if edit := get(t, head, srv.URL+"/users/"+inside+"/edit", http.StatusOK); strings.Contains(edit, `value="kazan"`) {
		t.Error("head's edit form offers other districts")
	}
	mustContain(t, postForm(t, head, srv.URL+"/users/"+inside+"/edit", profile("kazan", "epidemiological_supervision", "inside@example.com"), http.StatusBadRequest),
		"которым вы руководите")
	postForm(t, head, srv.URL+"/users/"+inside+"/edit", profile("nurlat", "nurlat_to", "inside@example.com"), http.StatusOK)
	register := profile("kazan", "epidemiological_supervision", "new@example.com")
	register.Set("password", testPassword)
	postForm(t, head, srv.URL+"/user/register", register, http.StatusBadRequest)

	// График — только своего ТО, праздники — только администратор.
	get(t, head, srv.URL+"/admin/schedule", http.StatusOK)
	get(t, head, srv.URL+"/admin/schedule?district=kazan", http.StatusForbidden)
	get(t, head, srv.URL+"/admin/holidays", http.StatusForbidden)

	// Обычный сотрудник того же ТО прав руководителя не имеет.
	employee := login(t, srv, "inside@example.com")
	get(t, employee, srv.URL+"/users", http.StatusForbidden)
	get(t, employee, srv.URL+"/users/"+headID, http.StatusForbidden)
}

func TestAdminEditsEmployeeProfile(t *testing.T) {
	c, srv := newTestApp(t)
	adminID := addUser(t, c, "admin@example.com", "software_engineer", "epidemiological_supervision")
	makeAdmin(t, c, adminID)
	id := addUser(t, c, "employee@example.com", "software_engineer", "epidemiological_supervision")
	editURL := srv.URL + "/users/" + id + "/edit"

	employee := login(t, srv, "employee@example.com")
	get(t, employee, editURL, http.StatusForbidden)
	postForm(t, employee, editURL, url.Values{"last_name": {"Взлом"}}, http.StatusForbidden)
	if body := get(t, employee, srv.URL+"/users/"+id, http.StatusOK); strings.Contains(body, "/edit") {
		t.Error("employee must not see the edit link")
	}

	admin := login(t, srv, "admin@example.com")
	mustContain(t, get(t, admin, srv.URL+"/users/"+id, http.StatusOK), `href="/users/`+id+`/edit"`)
	mustContain(t, get(t, admin, editURL, http.StatusOK),
		`value="employee@example.com"`, `value="1980-01-01"`, `value="2020-01-01"`,
		`<option value="kazan" selected>`, `<option value="software_engineer" selected>`)

	form := url.Values{
		"last_name": {"Петров"}, "first_name": {"Пётр"}, "middle_name": {"Петрович"},
		"email": {"bad-email"}, "birth_date": {"1981-02-03"}, "position": {"software_engineer"},
		"experience": {"10"}, "hired_at": {"2020-01-01"}, "district": {"nurlat"},
		"department": {"nurlat_to"}, "workday_duration": {"7"}, "is_invalid": {"true"},
	}
	mustContain(t, postForm(t, admin, editURL, form, http.StatusBadRequest),
		"некорректный e-mail", `value="Петров"`, `value="bad-email"`, `<option value="nurlat" selected>`)

	form.Set("email", "petrov@example.com")
	mustContain(t, postForm(t, admin, editURL, form, http.StatusOK),
		"Петров", "petrov@example.com", "Нурлатский ТО", "7 ч.")

	// Новый e-mail работает для входа, пароль прежний.
	login(t, srv, "petrov@example.com")

	form.Set("password", "short")
	mustContain(t, postForm(t, admin, editURL, form, http.StatusBadRequest), "пароль — от 8 символов")

	form.Set("password", "new-password-456")
	postForm(t, admin, editURL, form, http.StatusOK)
	res, err := (&http.Client{}).PostForm(srv.URL+"/login", url.Values{"email": {"petrov@example.com"}, "password": {"new-password-456"}})
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("login with new password: status %d", res.StatusCode)
	}
}

func TestScheduleFiltersByDepartment(t *testing.T) {
	c, srv := newTestApp(t)
	adminID := addUser(t, c, "admin@example.com", "software_engineer", "food_hygiene_supervision")
	makeAdmin(t, c, adminID)
	id := addUser(t, c, "employee@example.com", "lead_specialist", "epidemiological_supervision")
	employee := login(t, srv, "employee@example.com")

	start := time.Date(time.Now().Year()+1, time.July, 1, 0, 0, 0, 0, time.UTC) // без праздников
	payload := fmt.Sprintf(`{"userId":%q,"periods":[{"kind":"base","startDate":%q,"endDate":%q}]}`,
		id, start.Format("2006-01-02"), start.AddDate(0, 0, 13).Format("2006-01-02"))
	created, err := employee.Post(srv.URL+"/api/v1/users/"+id+"/vacations", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	created.Body.Close()
	if created.StatusCode != http.StatusOK {
		t.Fatalf("create: status %d", created.StatusCode)
	}

	admin := login(t, srv, "admin@example.com")
	year := start.Year()
	query := fmt.Sprintf("year=%d&district=kazan&department=", year)

	body := get(t, admin, srv.URL+"/admin/schedule?"+query+"epidemiological_supervision", http.StatusOK)
	mustContain(t, body, `/users/`+id, `<option value="epidemiological_supervision" selected>`)
	if strings.Contains(body, `/users/`+adminID) {
		t.Error("employee of another department in the schedule")
	}
	if strings.Contains(body, `value="buinsk_to"`) {
		t.Error("department of another district offered")
	}

	body = get(t, admin, srv.URL+"/admin/schedule?"+query+"food_hygiene_supervision", http.StatusOK)
	if strings.Contains(body, `/users/`+id) {
		t.Error("filtered schedule contains employee of another department")
	}

	// Утверждение по чужому отделу не затрагивает отпуск, по своему — утверждает.
	approveForm := func(department string) url.Values {
		return url.Values{"year": {fmt.Sprint(year)}, "district": {"kazan"}, "department": {department}}
	}
	mustContain(t, postForm(t, admin, srv.URL+"/admin/schedule/approve", approveForm("food_hygiene_supervision"), http.StatusOK),
		"Утверждено записей отпусков: 0")
	mustContain(t, postForm(t, admin, srv.URL+"/admin/schedule/approve", approveForm("epidemiological_supervision"), http.StatusOK),
		"Утверждено записей отпусков: 1")

	res, err := admin.Get(srv.URL + "/admin/reports/schedule?" + query + "epidemiological_supervision")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	doc, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK || len(doc) == 0 {
		t.Fatalf("report: status %d, %d bytes", res.StatusCode, len(doc))
	}
	if cd := res.Header.Get("Content-Disposition"); !strings.Contains(cd, "epidemiological_supervision") {
		t.Errorf("Content-Disposition = %q", cd)
	}
}

// Администратор задаёт границы рабочего года: следующие годы сдвигаются, а использованные
// дни остаются в годе с тем же номером и по-прежнему уменьшают его квоту.
func TestAdminSetsWorkYearBounds(t *testing.T) {
	c, srv := newTestApp(t)
	adminID := addUser(t, c, "admin@example.com", "lead_specialist_expert", "information_technology")
	makeAdmin(t, c, adminID)
	id := addUser(t, c, "employee@example.com", "lead_specialist_expert", "information_technology")
	uid, err := user.ParseUserID(id)
	if err != nil {
		t.Fatal(err)
	}
	opening := sqlite.NewOpeningUsageRepository(c.DB)
	err = opening.Replace(context.Background(), uid, []vacation.OpeningUsage{
		{WorkYearStart: date.New(2021, time.January, 1), Days: vacation.OpeningDays{Base: 20}},
	})
	if err != nil {
		t.Fatal(err)
	}

	admin := login(t, srv, "admin@example.com")
	page := srv.URL + "/users/" + id + "/work-years"
	set := func(number string, form url.Values, status int) string {
		return postForm(t, admin, page+"/"+number, form, status)
	}

	// Год 1 продлён до 31.03.2021: год 2 начинается 01.04.2021, год 3 — 01.04.2022.
	body := set("1", url.Values{"start": {"2020-01-01"}, "end": {"2021-03-31"}}, http.StatusOK)
	mustContain(t, body, `value="2021-03-31"`, `value="2021-04-01"`, `value="2022-04-01"`, "вручную")

	usages, err := opening.ListByUser(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(usages) != 1 || usages[0].WorkYearStart != date.New(2021, time.April, 1) {
		t.Fatalf("opening usage not moved: %+v", usages)
	}

	// Изменение начала года 3 задаёт конец года 2.
	body = set("3", url.Values{"start": {"2022-06-01"}, "end": {"2023-05-31"}}, http.StatusOK)
	mustContain(t, body, `value="2022-05-31"`, `value="2022-06-01"`, `value="2023-05-31"`)

	// Начало первого года, конец раньше начала и наложение лет не принимаются.
	mustContain(t, set("1", url.Values{"start": {"2019-01-01"}, "end": {"2021-03-31"}}, http.StatusBadRequest), "дата приёма")
	mustContain(t, set("2", url.Values{"start": {"2021-04-01"}, "end": {"2021-03-01"}}, http.StatusBadRequest), "раньше его начала")
	mustContain(t, set("1", url.Values{"start": {"2020-01-01"}, "end": {"2023-01-01"}}, http.StatusBadRequest), "конец раньше начала")

	// Сброс возвращает расчётный конец года 3.
	body = set("3", url.Values{"action": {"reset"}}, http.StatusOK)
	mustContain(t, body, `value="2023-05-31"`)
	usages, err = opening.ListByUser(context.Background(), uid)
	if err != nil {
		t.Fatal(err)
	}
	if usages[0].WorkYearStart != date.New(2021, time.April, 1) {
		t.Fatalf("opening usage moved on reset: %+v", usages)
	}

	// Сотруднику страница недоступна.
	get(t, login(t, srv, "employee@example.com"), page, http.StatusForbidden)
}

// Сдвиг и утверждение административного отпуска продлевают текущий год: отпуска и
// использованные дни следующего года переходят на его новое начало, пересечение сдвигов
// отклоняется.
func TestShiftsMoveVacationsAndOpeningUsage(t *testing.T) {
	c, srv := newTestApp(t)
	ctx := context.Background()
	adminID := addUser(t, c, "admin@example.com", "lead_specialist_expert", "information_technology")
	makeAdmin(t, c, adminID)
	// Принят полгода назад: первый год ещё идёт и не сохранён.
	hired := vacation.DateOf(time.Now().UTC()).Time().AddDate(0, -6, 0)
	day := func(months, days int) date.Date { return vacation.DateOf(hired.AddDate(0, months, days)) }
	iso := func(d date.Date) string { return d.ISO() }
	id := addUserHired(t, c, "employee@example.com", "lead_specialist_expert", "information_technology", iso(day(0, 0)))
	uid, err := user.ParseUserID(id)
	if err != nil {
		t.Fatal(err)
	}

	year2 := day(12, 0)
	opening := sqlite.NewOpeningUsageRepository(c.DB)
	if err := opening.Replace(ctx, uid, []vacation.OpeningUsage{
		{WorkYearStart: year2, Days: vacation.OpeningDays{Base: 20}},
	}); err != nil {
		t.Fatal(err)
	}
	vacations := sqlite.NewVacationRepository(c.DB)
	annual, err := vacation.NewVacation(vacation.NewVacationID(), uid, vacation.KindBase,
		day(14, 0), day(14, 9), year2, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	unpaid, err := vacation.NewVacation(vacation.NewVacationID(), uid, vacation.KindUnpaid,
		day(2, 0), day(2, 19), day(0, 0), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []*vacation.Vacation{annual, unpaid} {
		if err := vacations.Save(ctx, v); err != nil {
			t.Fatal(err)
		}
	}

	check := func(want date.Date) {
		t.Helper()
		usages, err := opening.ListByUser(ctx, uid)
		if err != nil {
			t.Fatal(err)
		}
		if len(usages) != 1 || usages[0].WorkYearStart != want {
			t.Fatalf("opening usage = %+v, want year %v", usages, want)
		}
		got, err := vacations.FindByID(ctx, uid, annual.ID())
		if err != nil {
			t.Fatal(err)
		}
		if got.WorkYearStart() != want {
			t.Fatalf("vacation year = %v, want %v", got.WorkYearStart(), want)
		}
	}

	// Прогул на 10 дней: год 2 начинается на 10 дней позже.
	admin := login(t, srv, "admin@example.com")
	target := srv.URL + "/api/v1/users/" + id + "/shift"
	shift := func(from, to date.Date, status int) {
		t.Helper()
		postForm(t, admin, target, url.Values{"user_id": {id}, "kind": {"absenteeism"}, "date_from": {iso(from)}, "date_to": {iso(to)}}, status)
	}
	shift(day(1, 0), day(1, 9), http.StatusOK)
	check(day(12, 10))

	// Пересечение отклонено, границы не изменились.
	shift(day(1, 5), day(1, 14), http.StatusBadRequest)
	check(day(12, 10))

	// 20 дней административного отпуска — ещё 6 дней сдвига.
	if _, err := c.ApproveVacationUseCase.Handle(ctx, approve_vacation.Command{UserID: id, VacationID: unpaid.ID().String()}); err != nil {
		t.Fatal(err)
	}
	check(day(12, 16))

	// Отмена возвращает прежнюю границу.
	if err := c.ChangeVacationStatusUseCase.Handle(ctx, change_status.Command{UserID: id, VacationID: unpaid.ID().String(), Status: "cancelled"}); err != nil {
		t.Fatal(err)
	}
	check(day(12, 10))
}

// Закончившиеся рабочие годы сохраняются: сдвиг задним числом их границы не меняет,
// отпуска остаются в прежних годах.
func TestShiftDoesNotMoveEndedWorkYears(t *testing.T) {
	c, srv := newTestApp(t)
	ctx := context.Background()
	adminID := addUser(t, c, "admin@example.com", "lead_specialist_expert", "information_technology")
	makeAdmin(t, c, adminID)
	id := addUser(t, c, "employee@example.com", "lead_specialist_expert", "information_technology")
	uid, err := user.ParseUserID(id)
	if err != nil {
		t.Fatal(err)
	}
	vacations := sqlite.NewVacationRepository(c.DB)
	annual, err := vacation.NewVacation(vacation.NewVacationID(), uid, vacation.KindBase,
		date.New(2021, time.March, 1), date.New(2021, time.March, 10), date.New(2021, time.January, 1), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := vacations.Save(ctx, annual); err != nil {
		t.Fatal(err)
	}

	admin := login(t, srv, "admin@example.com")
	postForm(t, admin, srv.URL+"/api/v1/users/"+id+"/shift",
		url.Values{"user_id": {id}, "kind": {"absenteeism"}, "date_from": {"2020-03-01"}, "date_to": {"2020-03-10"}}, http.StatusOK)

	got, err := vacations.FindByID(ctx, uid, annual.ID())
	if err != nil {
		t.Fatal(err)
	}
	if got.WorkYearStart() != date.New(2021, time.January, 1) {
		t.Fatalf("vacation moved to %v", got.WorkYearStart())
	}

	stored, err := sqlite.NewWorkYearRepository(c.DB).ListByUser(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) < 2 || stored[0].End != date.New(2020, time.December, 31) || stored[1].Start != date.New(2021, time.January, 1) || stored[0].Manual {
		t.Fatalf("stored work years = %+v", stored)
	}

	body := get(t, admin, srv.URL+"/users/"+id+"/work-years", http.StatusOK)
	mustContain(t, body, `value="2020-12-31"`, `value="2021-01-01"`, "сохранены")
}

// Выбранный вид списывается из нескольких лет, но не больше остатка: сервер подсказывает
// последний доступный день, а сверх нормы не оформляет даже администратор.
func TestVacationLimitAndNoOverdraft(t *testing.T) {
	c, srv := newTestApp(t)
	ctx := context.Background()
	adminID := addUser(t, c, "admin@example.com", "lead_specialist_expert", "information_technology")
	makeAdmin(t, c, adminID)
	id := addUserHired(t, c, "employee@example.com", "lead_specialist_expert", "information_technology", "2025-01-01")
	uid, err := user.ParseUserID(id)
	if err != nil {
		t.Fatal(err)
	}
	// В 2025 году использовано 4 дня: осталось 26 основного.
	used, err := vacation.NewVacation(vacation.NewVacationID(), uid, vacation.KindBase,
		date.New(2025, time.March, 3), date.New(2025, time.March, 6), date.New(2025, time.January, 1), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlite.NewVacationRepository(c.DB).Save(ctx, used); err != nil {
		t.Fatal(err)
	}

	postJSON := func(client *http.Client, path, body string, status int) string {
		t.Helper()
		res, err := client.Post(srv.URL+"/api/v1/users/"+id+path, "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		out, _ := io.ReadAll(res.Body)
		if res.StatusCode != status {
			t.Fatalf("POST %s: status %d, want %d: %s", path, res.StatusCode, status, out)
		}
		return string(out)
	}

	// 26 дней 2025 года + 30 дней 2026 года (заработаны к 23.11.2026); праздники 1–8 января не списываются.
	employee := login(t, srv, "employee@example.com")
	body := postJSON(employee, "/vacations/limit", `{"userId":"`+id+`","kind":"base","start":"2026-11-23","periods":[]}`, http.StatusOK)
	mustContain(t, body, `"days":56`, `"lastDay":"2027-01-25"`)

	body = postJSON(employee, "/vacations/limit", `{"userId":"`+id+`","kind":"unpaid","start":"2026-11-23","periods":[]}`, http.StatusOK)
	mustContain(t, body, `"unlimited":true`)

	// На день больше — отказ и администратору.
	admin := login(t, srv, "admin@example.com")
	over := `{"userId":"` + id + `","periods":[{"kind":"base","startDate":"2026-11-23","endDate":"2027-01-26"}]}`
	mustContain(t, postJSON(admin, "/vacations/preview", over, http.StatusBadRequest), "доступно 56 дн.")

	fits := `{"userId":"` + id + `","periods":[{"kind":"base","startDate":"2026-11-23","endDate":"2027-01-25"}]}`
	mustContain(t, postJSON(admin, "/vacations/preview", fits, http.StatusOK), `"workYearStart":"2025-01-01"`, `"workYearStart":"2026-01-01"`, `"years":[`)
}

// Дни, использованные до начала учёта, уменьшают остаток и видны в списке отпусков.
func TestOpeningUsageReducesBalanceAndIsShown(t *testing.T) {
	c, srv := newTestApp(t)
	id := addUser(t, c, "employee@example.com", "lead_specialist_expert", "information_technology")
	uid, err := user.ParseUserID(id)
	if err != nil {
		t.Fatal(err)
	}
	err = sqlite.NewOpeningUsageRepository(c.DB).Replace(context.Background(), uid, []vacation.OpeningUsage{
		{WorkYearStart: date.New(2020, time.January, 1), Days: vacation.OpeningDays{Base: 17, Irregular: 3, Seniority: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	client := login(t, srv, "employee@example.com")

	mustContain(t, get(t, client, srv.URL+"/api/v1/users/"+id+"/vacations", http.StatusOK),
		"Использовано до начала учёта", "01.01.2020", "<td>17</td>")

	body := get(t, client, srv.URL+"/api/v1/users/"+id+"/vacations/new", http.StatusOK)
	mustContain(t, body, `"base":{"total":30,"blocked":17}`, `"irregular":{"total":3,"blocked":3}`)
}
