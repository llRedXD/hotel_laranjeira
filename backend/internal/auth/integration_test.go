package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"os"
	"testing"
	"time"
)

func TestPostgresAuthLifecycle(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL ausente")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	schema := pgx.Identifier{fmt.Sprintf("auth_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err = tx.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, "SET LOCAL search_path TO "+schema+", public"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"001_create_users.sql", "002_create_auth_tokens.sql", "004_auth_version.sql"} {
		sql, err := os.ReadFile("../../../database/migrations/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = tx.Exec(ctx, string(sql)); err != nil {
			t.Fatal(err)
		}
	}
	repository := NewPostgresRepository(tx)
	service := NewService(repository, "integration-secret-with-at-least-32-bytes")
	password := "senha-teste-segura"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	if err = repository.InsertUser(ctx, "Admin", "admin@example.com", string(hash), "admin"); err != nil {
		t.Fatal(err)
	}
	user, access, refresh, err := service.Login(ctx, " ADMIN@example.com ", password)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.UserFromAccessToken(ctx, access); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err = service.Login(ctx, "admin@example.com", "incorreta"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("senha incorreta: %v", err)
	}
	if _, _, _, err = service.Login(ctx, "ausente@example.com", password); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("conta ausente: %v", err)
	}

	// Unique violation after revoking old token must roll back the revocation.
	_, err = repository.RotateRefresh(ctx, tokenHash(refresh), tokenHash(refresh), time.Now().Add(time.Hour))
	if err == nil {
		t.Fatal("rotação com hash duplicado foi aceita")
	}
	_, access, next, err := service.Refresh(ctx, refresh)
	if err != nil {
		t.Fatalf("refresh perdido após rollback: %v", err)
	}
	if _, _, _, err = service.Refresh(ctx, refresh); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("reuso: %v", err)
	}
	if err = service.Logout(ctx, next); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err = service.Refresh(ctx, next); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("logout: %v", err)
	}
	_, _, live, err := service.Login(ctx, "admin@example.com", password)
	if err != nil {
		t.Fatal(err)
	}
	first, err := service.RequestPasswordReset(ctx, "admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	reset, err := service.RequestPasswordReset(ctx, "admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if err = service.ResetPassword(ctx, first, "outra-senha-segura"); !errors.Is(err, ErrInvalidResetToken) {
		t.Fatalf("token substituído: %v", err)
	}
	if err = service.ResetPassword(ctx, reset, "outra-senha-segura"); err != nil {
		t.Fatal(err)
	}
	if err = service.ResetPassword(ctx, reset, password); !errors.Is(err, ErrInvalidResetToken) {
		t.Fatalf("reuso reset: %v", err)
	}
	if _, err = service.UserFromAccessToken(ctx, access); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("access antigo: %v", err)
	}
	if _, _, _, err = service.Refresh(ctx, live); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("sessão antiga: %v", err)
	}
	_, newAccess, activeRefresh, err := service.Login(ctx, "admin@example.com", "outra-senha-segura")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, "UPDATE users SET is_active=false WHERE id=$1", user.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = service.UserFromAccessToken(ctx, newAccess); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("inativo access: %v", err)
	}
	if _, _, _, err = service.Refresh(ctx, activeRefresh); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("inativo refresh: %v", err)
	}
	if token, err := service.RequestPasswordReset(ctx, "ausente@example.com"); err != nil || token != "" {
		t.Fatalf("enumeração: %v", err)
	}
}
