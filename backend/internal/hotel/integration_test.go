package hotel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestPostgresPartialUpdates(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL ausente: integração PostgreSQL não executada")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
	schema := pgx.Identifier{fmt.Sprintf("hotel_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err = tx.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, "SET LOCAL search_path TO "+schema+", public"); err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile("../../../database/migrations/003_create_guests_and_rooms.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, string(migration)); err != nil {
		t.Fatal(err)
	}
	service := NewService(NewPostgresRepository(tx))
	name, email, phone := "Ana", "ANA@example.com", "11999999999"
	guest, err := service.CreateGuest(ctx, GuestInput{Name: name, Email: &email, Phone: &phone})
	if err != nil {
		t.Fatal(err)
	}
	next := "Ana Silva"
	updated, err := service.UpdateGuest(ctx, guest.ID, GuestPatch{Name: Field{true, &next}})
	if err != nil || updated.Email == nil || *updated.Email != "ana@example.com" || updated.Phone == nil || *updated.Phone != phone {
		t.Fatalf("campos omitidos não preservados: %+v, %v", updated, err)
	}
	updated, err = service.UpdateGuest(ctx, guest.ID, GuestPatch{Email: Field{true, nil}})
	if err != nil || updated.Email != nil || updated.Name != next || updated.Phone == nil {
		t.Fatalf("null não aplicado: %+v %v", updated, err)
	}
	_, err = service.UpdateGuest(ctx, guest.ID, GuestPatch{Name: Field{true, nil}})
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("nome null aceito: %v", err)
	}
	_, err = service.UpdateGuest(ctx, guest.ID, GuestPatch{})
	if !errors.As(err, &validation) {
		t.Fatalf("PATCH vazio aceito: %v", err)
	}
	_, err = service.UpdateGuest(ctx, "00000000-0000-0000-0000-000000000001", GuestPatch{Name: Field{true, &next}})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("404: %v", err)
	}
	description := "Categoria original"
	category, err := service.CreateCategory(ctx, CategoryInput{Name: "Standard", Description: &description})
	if err != nil {
		t.Fatal(err)
	}
	category, err = service.UpdateCategory(ctx, category.ID, CategoryPatch{Name: Field{true, &next}})
	if err != nil || category.Description == nil || *category.Description != description {
		t.Fatalf("descrição perdida: %+v %v", category, err)
	}
	category, err = service.UpdateCategory(ctx, category.ID, CategoryPatch{Description: Field{true, nil}})
	if err != nil || category.Description != nil || category.Name != next {
		t.Fatalf("null categoria: %+v %v", category, err)
	}
	if _, err = tx.Exec(ctx, "SAVEPOINT conflict_test"); err != nil {
		t.Fatal(err)
	}
	_, err = service.CreateCategory(ctx, CategoryInput{Name: next})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("conflito: %v", err)
	}
	if _, err = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT conflict_test"); err != nil {
		t.Fatal(err)
	}
	found, err := service.GetGuest(ctx, guest.ID)
	if err != nil || found.Name != next || found.Phone == nil {
		t.Fatalf("persistência: %+v %v", found, err)
	}
	number, rate, capacity := "T-101", "199.90", 2
	room, err := service.SaveRoom(ctx, "", RoomPatch{Number: &number, CategoryID: &category.ID, Capacity: &capacity, DailyRate: &rate})
	if err != nil || room.DailyRate != "199.90" || !room.IsAvailable {
		t.Fatalf("cadastro quarto: %+v %v", room, err)
	}
	available := false
	room, err = service.SaveRoom(ctx, room.ID, RoomPatch{IsAvailable: &available})
	if err != nil || room.IsAvailable || room.Capacity != 2 || room.Number != number {
		t.Fatalf("disponibilidade: %+v %v", room, err)
	}
	filtered, err := service.ListRooms(ctx, RoomFilter{CategoryID: category.ID, Available: &available, Limit: 20})
	if err != nil || len(filtered) != 1 {
		t.Fatalf("filtro: %+v %v", filtered, err)
	}
	invalidRate := "1.999"
	if _, err = service.SaveRoom(ctx, room.ID, RoomPatch{DailyRate: &invalidRate}); !errors.As(err, &validation) {
		t.Fatalf("precisão: %v", err)
	}
	if _, err = tx.Exec(ctx, "SAVEPOINT room_conflict"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.SaveRoom(ctx, "", RoomPatch{Number: &number, CategoryID: &category.ID, Capacity: &capacity, DailyRate: &rate}); !errors.Is(err, ErrConflict) {
		t.Fatalf("número duplicado: %v", err)
	}
	if _, err = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT room_conflict"); err != nil {
		t.Fatal(err)
	}
}
