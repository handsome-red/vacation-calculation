package sqlite_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/handsome-red/vacation-calculation/internal/application/commands/auth/login"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/bootstrap_admin"
	"github.com/handsome-red/vacation-calculation/internal/application/commands/user/register_user"
	"github.com/handsome-red/vacation-calculation/internal/domain/user"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/hasher"
	"github.com/handsome-red/vacation-calculation/internal/infrastructure/persistence/sqlite"
)

func TestBootstrapAdminCreatesOnlyFirstAdmin(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.NewDB(ctx, sqlite.Config{Path: filepath.Join(t.TempDir(), "bootstrap.db")}, testLogger{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := sqlite.NewUserRepository(db)
	passwordHasher := hasher.NewBcryptHasher(4)
	register := register_user.NewHandler(repo, sqlite.NewDistrictRepository(db), passwordHasher, testLogger{})
	bootstrap := bootstrap_admin.NewHandler(repo, register)

	cmd := bootstrap_admin.Command{
		Email: "root@example.com", Password: "root-password-123",
		LastName: "Администратор", FirstName: "Системный",
		BirthDate: "1970-01-01", Position: "software_engineer", HiredAt: "2020-01-01",
		DistrictCode: "kazan", DepartmentCode: "epidemiological_supervision",
	}

	first, err := bootstrap.Handle(ctx, cmd)
	if err != nil {
		t.Fatalf("first bootstrap: %v", err)
	}
	if !first.Created {
		t.Fatalf("first bootstrap result = %+v, want Created", first)
	}

	id, err := user.ParseUserID(first.UserID)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("load admin: %v", err)
	}
	if !stored.IsAdmin() {
		t.Fatal("bootstrapped user is not an admin")
	}
	if _, err := login.NewHandler(repo, passwordHasher).Handle(ctx, login.Command{Email: cmd.Email, Password: cmd.Password}); err != nil {
		t.Fatalf("admin cannot log in: %v", err)
	}

	cmd.Email = "second@example.com"
	second, err := bootstrap.Handle(ctx, cmd)
	if err != nil {
		t.Fatalf("second bootstrap: %v", err)
	}
	if second.Created || second.Promoted {
		t.Fatalf("second bootstrap result = %+v, want no changes when an admin exists", second)
	}
}
