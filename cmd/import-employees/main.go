// Команда import-employees создаёт или обновляет сотрудников из JSON-файла и задаёт
// дни, использованные до начала учёта. Повторный запуск безопасен. База — DB_PATH.
// Пустой password: новому сотруднику генерируется случайный, у существующего не меняется.
// clearShifts: true удаляет все сдвиги сотрудника.
//
//	import-employees -in employees.json -out report.txt
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/application/leavecalc"
	"github.com/handsome-red/vacation-calculation/internal/config"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/ports"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/hasher"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/persistence/sqlite"
	"github.com/jmoiron/sqlx"
)

type openingJSON struct {
	WorkYearStart string `json:"workYearStart"`
	Base          int    `json:"base"`
	Irregular     int    `json:"irregular"`
	Seniority     int    `json:"seniority"`
}

type employeeJSON struct {
	LastName        string        `json:"lastName"`
	FirstName       string        `json:"firstName"`
	MiddleName      string        `json:"middleName"`
	Email           string        `json:"email"`
	BirthDate       string        `json:"birthDate"`
	Position        string        `json:"position"`
	HiredAt         string        `json:"hiredAt"`
	Experience      int           `json:"experience"`
	District        string        `json:"district"`
	Department      string        `json:"department"`
	WorkdayDuration int           `json:"workdayDuration"`
	Opening         []openingJSON `json:"opening"`
	Password        string        `json:"password"`
	ClearShifts     bool          `json:"clearShifts"`
}

type noopLogger struct{}

func (noopLogger) Debug(context.Context, string, ...any) {}
func (noopLogger) Info(context.Context, string, ...any)  {}
func (noopLogger) Warn(context.Context, string, ...any)  {}
func (noopLogger) Error(context.Context, string, ...any) {}

type importer struct {
	db          *sqlx.DB
	users       ports.UserRepository
	hasher      ports.PasswordHasher
	vacations   ports.VacationRepository
	departments ports.DepartmentRepository
	districts   ports.DistrictRepository
	opening     ports.OpeningUsageRepository
	register    *register_user.Handler
	leave       *leavecalc.Service
	report      io.Writer
}

func main() {
	in := flag.String("in", "", "JSON-файл с сотрудниками")
	out := flag.String("out", "", "файл отчёта: паролей новых сотрудников и остатков")
	flag.Parse()
	if *in == "" || *out == "" {
		log.Fatal("укажите -in и -out")
	}

	raw, err := os.ReadFile(*in)
	if err != nil {
		log.Fatal(err)
	}
	var employees []employeeJSON
	if err := json.Unmarshal(raw, &employees); err != nil {
		log.Fatalf("разбор %s: %v", *in, err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	db, err := sqlite.NewDB(ctx, sqlite.Config{Path: cfg.DatabasePath}, noopLogger{})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	report, err := os.OpenFile(*out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		log.Fatal(err)
	}
	defer report.Close()

	users := sqlite.NewUserRepository(db)
	vacations := sqlite.NewVacationRepository(db)
	opening := sqlite.NewOpeningUsageRepository(db)
	districts := sqlite.NewDistrictRepository(db)
	passwordHasher := hasher.NewBcryptHasher(10)
	imp := &importer{
		db:          db,
		hasher:      passwordHasher,
		users:       users,
		vacations:   vacations,
		departments: sqlite.NewDepartmentRepository(db),
		districts:   districts,
		opening:     opening,
		register:    register_user.NewHandler(users, districts, passwordHasher, noopLogger{}),
		leave:       leavecalc.NewService(users, vacations, sqlite.NewShiftRepository(db), sqlite.NewCalendarDayRepository(db), opening, sqlite.NewWorkYearRepository(db), vacation.NewWorkYearCalculator()),
		report:      report,
	}
	for _, e := range employees {
		if err := imp.run(ctx, e); err != nil {
			log.Fatalf("%s %s: %v", e.LastName, e.FirstName, err)
		}
	}
}

func (i *importer) run(ctx context.Context, e employeeJSON) error {
	u, err := i.find(ctx, e)
	if err != nil {
		return err
	}
	if u == nil {
		if u, err = i.create(ctx, e); err != nil {
			return err
		}
	} else if err := i.update(ctx, u, e); err != nil {
		return err
	}

	if e.ClearShifts {
		if err := i.clearShifts(ctx, u.ID()); err != nil {
			return err
		}
	}

	usages := make([]vacation.OpeningUsage, 0, len(e.Opening))
	for _, o := range e.Opening {
		start, err := date.ParseDate(o.WorkYearStart)
		if err != nil {
			return fmt.Errorf("рабочий год %q: %w", o.WorkYearStart, err)
		}
		usages = append(usages, vacation.OpeningUsage{
			WorkYearStart: start,
			Days:          vacation.OpeningDays{Base: o.Base, Irregular: o.Irregular, Seniority: o.Seniority},
		})
	}
	if err := i.opening.Replace(ctx, u.ID(), usages); err != nil {
		return err
	}
	return i.printBalances(ctx, u.ID())
}

// find ищет сотрудника по e-mail, затем по точному совпадению ФИО.
func (i *importer) find(ctx context.Context, e employeeJSON) (*user.User, error) {
	if e.Email != "" {
		email, err := user.NewEmail(e.Email)
		if err != nil {
			return nil, fmt.Errorf("e-mail %q: %w", e.Email, err)
		}
		u, err := i.users.FindByEmail(ctx, email)
		if err == nil {
			return u, nil
		}
		if !errors.Is(err, user.ErrUserNotFound) {
			return nil, err
		}
	}
	found, _, err := i.users.FindUsers(ctx, ports.Query{Page: 1, Size: 100000, Filters: ports.UserFilter{LastName: e.LastName}})
	if err != nil {
		return nil, err
	}
	var match *user.User
	for _, u := range found {
		if u.LastName() == e.LastName && u.FirstName() == e.FirstName && u.MiddleName() == e.MiddleName {
			if match != nil {
				return nil, errors.New("найдено несколько сотрудников с таким ФИО")
			}
			match = u
		}
	}
	return match, nil
}

func (i *importer) create(ctx context.Context, e employeeJSON) (*user.User, error) {
	if e.Email == "" {
		return nil, errors.New("для нового сотрудника нужен e-mail")
	}
	password := e.Password
	if password == "" {
		buf := make([]byte, 12)
		if _, err := rand.Read(buf); err != nil {
			return nil, err
		}
		password = base64.RawURLEncoding.EncodeToString(buf)
	}

	created, err := i.register.Handle(ctx, register_user.Command{
		Email: e.Email, Password: password, FirstName: e.FirstName, LastName: e.LastName, MiddleName: e.MiddleName,
		DepartmentCode: e.Department, DistrictCode: e.District, BirthDate: e.BirthDate, Position: e.Position,
		Experience: e.Experience, HiredAt: e.HiredAt, WorkdayDuration: e.WorkdayDuration,
	})
	if err != nil {
		return nil, err
	}
	id, err := user.ParseUserID(created.UserID)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(i.report, "СОЗДАН %s %s %s | %s | пароль: %s\n", e.LastName, e.FirstName, e.MiddleName, e.Email, password)
	return i.users.FindByID(ctx, id)
}

// update обновляет найденного сотрудника; пароль и права не меняются. Дату приёма
// нельзя сменить при действующих отпусках: они привязаны к началу рабочего года.
func (i *importer) update(ctx context.Context, u *user.User, e employeeJSON) error {
	hired, err := user.NewHiredDate(e.HiredAt)
	if err != nil {
		return err
	}
	if !hired.Time().Equal(u.HiredAt().Time()) {
		vacations, err := i.vacations.FindByUserID(ctx, u.ID())
		if err != nil {
			return err
		}
		for _, v := range vacations {
			if v.Status().IsActive() {
				return fmt.Errorf("дата приёма меняется с %s на %s, но есть действующий отпуск %s: отмените его или исправьте данные",
					u.HiredAt(), hired, v.StartDate().Time().Format("02.01.2006"))
			}
		}
	}
	birth, err := user.NewBirthDate(e.BirthDate)
	if err != nil {
		return err
	}
	position, err := user.NewPosition(e.Position)
	if err != nil {
		return err
	}
	workday, err := user.NewWorkdayDuration(e.WorkdayDuration)
	if err != nil {
		return err
	}
	district, err := i.districts.FindByCode(ctx, e.District)
	if err != nil {
		return err
	}
	departments, err := i.departments.List(ctx)
	if err != nil {
		return err
	}
	department := user.Department{}
	for _, d := range departments {
		if d.Code == e.Department {
			department = d
		}
	}
	if department.Code == "" {
		return fmt.Errorf("неизвестный отдел %q", e.Department)
	}
	email := u.Email()
	if e.Email != "" {
		if email, err = user.NewEmail(e.Email); err != nil {
			return err
		}
	}
	if err := u.UpdateProfile(user.Profile{
		LastName: e.LastName, FirstName: e.FirstName, MiddleName: e.MiddleName, BirthDate: birth,
		Position: position, HiredAt: hired, ExperienceDays: e.Experience, Department: department,
		District: district, WorkdayDuration: workday, Email: email, IsInvalid: u.IsInvalid(),
	}); err != nil {
		return err
	}
	passwordNote := "пароль без изменений"
	if e.Password != "" {
		hash, err := i.hasher.HashPassword(ctx, e.Password)
		if err != nil {
			return err
		}
		if err := u.ChangePassword(user.NewPasswordFromHash(hash)); err != nil {
			return err
		}
		passwordNote = "пароль заменён"
	}
	fmt.Fprintf(i.report, "ОБНОВЛЁН %s %s %s | %s | %s, права без изменений\n", e.LastName, e.FirstName, e.MiddleName, email.Value(), passwordNote)
	return i.users.Save(ctx, u)
}

// clearShifts удаляет сдвиги сотрудника и записывает удалённые в отчёт.
func (i *importer) clearShifts(ctx context.Context, id user.UserID) error {
	var rows []struct {
		Kind string `db:"kind"`
		From string `db:"date_from"`
		To   string `db:"date_to"`
	}
	if err := i.db.SelectContext(ctx, &rows, `SELECT kind, date_from, date_to FROM shifts WHERE user_id = ?`, id.String()); err != nil {
		return err
	}
	for _, r := range rows {
		fmt.Fprintf(i.report, "  удалён сдвиг %s %s — %s\n", r.Kind, r.From, r.To)
	}
	_, err := i.db.ExecContext(ctx, `DELETE FROM shifts WHERE user_id = ?`, id.String())
	return err
}

// printBalances выводит полные остатки по рабочим годам — для сверки с исходной таблицей.
func (i *importer) printBalances(ctx context.Context, id user.UserID) error {
	u, err := i.users.FindByID(ctx, id)
	if err != nil {
		return err
	}
	today := vacation.DateOf(time.Now())
	snap, err := i.leave.Load(ctx, u, today, "")
	if err != nil {
		return err
	}
	balances, err := snap.Balances(today, false)
	if err != nil {
		return err
	}
	var b strings.Builder
	for _, y := range balances {
		fmt.Fprintf(&b, "  %s–%s: основной %d, ненорм. %d, выслуга %d\n",
			y.Year.From.Format("02.01.2006"), y.Year.To.AddDate(0, 0, -1).Format("02.01.2006"),
			y.Remaining[vacation.KindBase], y.Remaining[vacation.KindIrregular], y.Remaining[vacation.KindSeniority])
	}
	_, err = fmt.Fprint(i.report, b.String())
	return err
}
