package sqlite_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/auth/login"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/hasher"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/persistence/sqlite"
	"github.com/handsome-red/vacation-calculation/internal/interfaces/http/handlers"
)

type testLogger struct{}

func (testLogger) Debug(context.Context, string, ...any) {}
func (testLogger) Info(context.Context, string, ...any)  {}
func (testLogger) Warn(context.Context, string, ...any)  {}
func (testLogger) Error(context.Context, string, ...any) {}

func TestRegisterUserCanLogIn(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.NewDB(ctx, sqlite.Config{Path: filepath.Join(t.TempDir(), "registration.db")}, testLogger{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := sqlite.NewUserRepository(db)
	passwordHasher := hasher.NewBcryptHasher(4)
	register := register_user.NewHandler(repo, sqlite.NewDistrictRepository(db), passwordHasher, testLogger{})
	const password = "test-password-123"
	const email = "new.employee@example.com"
	created, err := register.Handle(ctx, register_user.Command{
		Email: email, Password: password, FirstName: "Иван", LastName: "Иванов",
		MiddleName: "Иванович", BirthDate: "1990-01-01", HiredAt: "2020-01-01",
		Position: "software_engineer", DistrictCode: "kazan", DepartmentCode: "epidemiological_supervision",
		WorkdayDuration: 8,
	})
	if err != nil {
		t.Fatalf("register user: %v", err)
	}

	id, err := user.ParseUserID(created.UserID)
	if err != nil {
		t.Fatalf("parse created ID: %v", err)
	}
	stored, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("load created user by ID: %v", err)
	}
	if got := stored.District().Code(); got != "kazan" {
		t.Errorf("district code = %q, want kazan", got)
	}
	if stored.IsAdmin() {
		t.Error("new user must not be a superadmin")
	}
	if stored.Password().String() == password {
		t.Error("password saved in plaintext")
	}

	loginHandler := login.NewHandler(repo, passwordHasher)
	loggedIn, err := loginHandler.Handle(ctx, login.Command{Email: email, Password: password})
	if err != nil {
		t.Fatalf("log in with registered password: %v", err)
	}
	if loggedIn.UserID != created.UserID {
		t.Errorf("logged-in ID = %q, want %q", loggedIn.UserID, created.UserID)
	}
	if _, err := loginHandler.Handle(ctx, login.Command{Email: email, Password: "wrong-password"}); !errors.Is(err, login.ErrInvalidCredentials) {
		t.Errorf("wrong password: got %v, want ErrInvalidCredentials", err)
	}
	if _, err := register.Handle(ctx, register_user.Command{Email: email}); !errors.Is(err, user.ErrEmailAlreadyExists) {
		t.Errorf("duplicate email: got %v, want ErrEmailAlreadyExists", err)
	}
}

func TestRegistrationFormCreatesUserWhoCanLogIn(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.NewDB(ctx, sqlite.Config{Path: filepath.Join(t.TempDir(), "form.db")}, testLogger{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := sqlite.NewUserRepository(db)
	passwordHasher := hasher.NewBcryptHasher(4)
	registration := handlers.NewRegistrationHandler(
		register_user.NewHandler(repo, sqlite.NewDistrictRepository(db), passwordHasher, testLogger{}), nil, nil,
	)
	const password = "пароль-для-входа-123"
	const email = "from.form@example.com"
	form := url.Values{
		"email": {email}, "password": {password}, "first_name": {"Иван"},
		"last_name": {"Иванов"}, "middle_name": {"Иванович"},
		"birth_date": {"1990-01-01"}, "hired_at": {"2020-01-01"},
		"position": {"software_engineer"}, "district": {"kazan"},
		"department": {"epidemiological_supervision"}, "workday_duration": {"8"},
	}
	req := httptest.NewRequest(http.MethodPost, "/user/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	registration.Register(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("registration status = %d, want 303; response: %s", w.Code, w.Body.String())
	}

	storedEmail, err := user.NewEmail(email)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := repo.FindByEmail(ctx, storedEmail)
	if err != nil {
		t.Fatalf("load registered user: %v", err)
	}
	if stored.Department().Code != "epidemiological_supervision" || stored.District().Code() != "kazan" {
		t.Errorf("wrong department/district: %q/%q", stored.Department().Code, stored.District().Code())
	}
	if _, err := login.NewHandler(repo, passwordHasher).Handle(ctx, login.Command{Email: email, Password: password}); err != nil {
		t.Fatalf("log in after submitting registration form: %v", err)
	}
}
