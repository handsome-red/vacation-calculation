package sqlite_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/schedule/approve_schedule"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/update_user"
	"github.com/handsome-red/vacation-calculation/internal/domain/auth"
	"github.com/handsome-red/vacation-calculation/internal/domain/date"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/domain/vacation"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/hasher"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/persistence/sqlite"
	"github.com/jmoiron/sqlx"
)

func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	db, err := sqlite.NewDB(context.Background(), sqlite.Config{Path: filepath.Join(t.TempDir(), "test.db")}, testLogger{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func registerEmployee(t *testing.T, db *sqlx.DB, email, department string) user.UserID {
	t.Helper()
	register := register_user.NewHandler(sqlite.NewUserRepository(db), sqlite.NewDistrictRepository(db), hasher.NewBcryptHasher(4), testLogger{})
	created, err := register.Handle(context.Background(), register_user.Command{
		Email: email, Password: "test-password-123", FirstName: "Иван", LastName: "Иванов",
		MiddleName: "Иванович", BirthDate: "1990-01-01", HiredAt: "2020-01-01",
		Position: "software_engineer", DistrictCode: "kazan", DepartmentCode: department,
		WorkdayDuration: 8,
	})
	if err != nil {
		t.Fatalf("register user: %v", err)
	}
	id, err := user.ParseUserID(created.UserID)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// admin — права администратора для вызовов use case без HTTP.
var admin = auth.Caller{IsAdmin: true}

func updateCommand(id user.UserID) update_user.Command {
	return update_user.Command{
		Caller: admin,
		UserID: id.String(), LastName: "Петров", FirstName: "Пётр", MiddleName: "Петрович",
		BirthDate: "1991-02-03", Position: "software_engineer", HiredAt: "2020-01-01", Experience: 100,
		DepartmentCode: "nurlat_to", DistrictCode: "nurlat", WorkdayDuration: 7,
		Email: "new.mail@example.com", IsInvalid: true,
	}
}

func addVacation(t *testing.T, db *sqlx.DB, id user.UserID, start, end date.Date) vacation.VacationID {
	t.Helper()
	v, err := vacation.NewVacation(vacation.NewVacationID(), id, vacation.KindBase, start, end, date.New(2020, time.January, 1), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlite.NewVacationRepository(db).Save(context.Background(), v); err != nil {
		t.Fatalf("save vacation: %v", err)
	}
	return v.ID()
}

func TestUpdateUserChangesProfile(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	repo := sqlite.NewUserRepository(db)
	update := update_user.NewHandler(repo, sqlite.NewDistrictRepository(db), sqlite.NewDepartmentRepository(db), sqlite.NewVacationRepository(db), sqlite.NewOpeningUsageRepository(db), sqlite.NewWorkYearRepository(db), hasher.NewBcryptHasher(4))
	id := registerEmployee(t, db, "old@example.com", "epidemiological_supervision")
	before, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if err := update.Handle(ctx, updateCommand(id)); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.LastName() != "Петров" || got.Email().Value() != "new.mail@example.com" || got.Department().Code != "nurlat_to" ||
		got.District().Code() != "nurlat" || got.ExperienceDays() != 100 || got.WorkdayDuration().Int() != 7 || !got.IsInvalid() {
		t.Errorf("profile not updated: %+v", got)
	}
	if got.Password() != before.Password() || got.IsAdmin() || !got.IsActive() {
		t.Error("password, rights and status must not change")
	}
}

func TestUpdateUserRejectsInvalidData(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	update := update_user.NewHandler(sqlite.NewUserRepository(db), sqlite.NewDistrictRepository(db), sqlite.NewDepartmentRepository(db), sqlite.NewVacationRepository(db), sqlite.NewOpeningUsageRepository(db), sqlite.NewWorkYearRepository(db), hasher.NewBcryptHasher(4))
	id := registerEmployee(t, db, "first@example.com", "epidemiological_supervision")
	registerEmployee(t, db, "taken@example.com", "epidemiological_supervision")

	tests := map[string]func(*update_user.Command){
		"чужой e-mail":            func(c *update_user.Command) { c.Email = "Taken@example.com" },
		"отдел другого теротдела": func(c *update_user.Command) { c.DepartmentCode = "buinsk_to" },
		"неизвестный отдел":       func(c *update_user.Command) { c.DepartmentCode = "missing" },
		"пустое отчество":         func(c *update_user.Command) { c.MiddleName = "" },
		"отрицательный стаж":      func(c *update_user.Command) { c.Experience = -1 },
		"рабочий день":            func(c *update_user.Command) { c.WorkdayDuration = 9 },
		"дата приёма":             func(c *update_user.Command) { c.HiredAt = "не дата" },
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := updateCommand(id)
			change(&cmd)
			if err := update.Handle(ctx, cmd); !errors.Is(err, update_user.ErrInvalidInput) {
				t.Errorf("err = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestUpdateUserKeepsOwnEmailInAnyCase(t *testing.T) {
	db := openTestDB(t)
	update := update_user.NewHandler(sqlite.NewUserRepository(db), sqlite.NewDistrictRepository(db), sqlite.NewDepartmentRepository(db), sqlite.NewVacationRepository(db), sqlite.NewOpeningUsageRepository(db), sqlite.NewWorkYearRepository(db), hasher.NewBcryptHasher(4))
	id := registerEmployee(t, db, "same@example.com", "epidemiological_supervision")
	cmd := updateCommand(id)
	cmd.Email = "SAME@example.com"
	if err := update.Handle(context.Background(), cmd); err != nil {
		t.Fatalf("update: %v", err)
	}
}

func TestUpdateUserHiredDateLockedByActiveVacation(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	update := update_user.NewHandler(sqlite.NewUserRepository(db), sqlite.NewDistrictRepository(db), sqlite.NewDepartmentRepository(db), sqlite.NewVacationRepository(db), sqlite.NewOpeningUsageRepository(db), sqlite.NewWorkYearRepository(db), hasher.NewBcryptHasher(4))
	id := registerEmployee(t, db, "hired@example.com", "epidemiological_supervision")

	cmd := updateCommand(id)
	cmd.HiredAt = "2020-02-01"
	if err := update.Handle(ctx, cmd); err != nil {
		t.Fatalf("no vacations: %v", err)
	}

	cmd.HiredAt = "2020-03-01"
	addVacation(t, db, id, date.New(2021, time.June, 1), date.New(2021, time.June, 14))
	if err := update.Handle(ctx, cmd); !errors.Is(err, update_user.ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}

	// Другие данные при неизменной дате приёма менять можно.
	cmd.HiredAt = "2020-02-01"
	cmd.LastName = "Сидоров"
	if err := update.Handle(ctx, cmd); err != nil {
		t.Fatalf("same hired date: %v", err)
	}
}

func TestApproveScheduleFiltersByDepartment(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	vacations := sqlite.NewVacationRepository(db)
	approve := approve_schedule.NewHandler(sqlite.NewUserRepository(db), vacations)

	inside := registerEmployee(t, db, "inside@example.com", "epidemiological_supervision")
	outside := registerEmployee(t, db, "outside@example.com", "food_hygiene_supervision")
	start, end := date.New(2027, time.June, 1), date.New(2027, time.June, 14)
	insideVacation := addVacation(t, db, inside, start, end)
	outsideVacation := addVacation(t, db, outside, start, end)

	res, err := approve.Handle(ctx, approve_schedule.Command{Year: 2027, DistrictCode: "kazan", DepartmentCode: "epidemiological_supervision", Caller: admin})
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if res.Approved != 1 {
		t.Errorf("approved = %d, want 1", res.Approved)
	}
	if v, err := vacations.FindByID(ctx, inside, insideVacation); err != nil || v.Status() != vacation.StatusApproved {
		t.Errorf("vacation in department: %v, %v", v, err)
	}
	if v, err := vacations.FindByID(ctx, outside, outsideVacation); err != nil || v.Status() != vacation.StatusDraft {
		t.Errorf("vacation in other department must stay draft: %v, %v", v, err)
	}
}
