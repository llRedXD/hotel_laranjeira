package hotel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestReservationAvailabilityCreationAndConcurrency(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL ausente: integração PostgreSQL não executada")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(context.Background())
	schema := pgx.Identifier{fmt.Sprintf("hotel_reservation_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE") }()
	if _, err := admin.Exec(ctx, "SET search_path TO "+schema+", public"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"003_create_guests_and_rooms.sql", "005_reservations.sql"} {
		migration, readErr := os.ReadFile("../../../database/migrations/" + name)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, err = admin.Exec(ctx, string(migration)); err != nil {
			t.Fatal(err)
		}
	}
	var categoryID, roomID string
	if err = admin.QueryRow(ctx, `INSERT INTO room_categories (name) VALUES ('Teste') RETURNING id`).Scan(&categoryID); err != nil {
		t.Fatal(err)
	}
	if err = admin.QueryRow(ctx, `INSERT INTO rooms (number, category_id, capacity, daily_rate) VALUES ('T-1', $1, 3, 125.50) RETURNING id`, categoryID).Scan(&roomID); err != nil {
		t.Fatal(err)
	}
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	service := NewService(NewPostgresRepository(pool))
	date := func(day int) time.Time { return time.Date(2026, time.December, day, 0, 0, 0, 0, time.UTC) }
	options, err := service.ListAvailability(ctx, AvailabilityQuery{CheckIn: date(10), CheckOut: date(12), Occupants: 2})
	if err != nil || len(options) != 1 || options[0].Nights != 2 || options[0].EstimatedTotal != "251.00" {
		t.Fatalf("availability: options=%+v err=%v", options, err)
	}
	if _, err = pool.Exec(ctx, `UPDATE rooms SET daily_rate=130.00 WHERE id=$1`, roomID); err != nil {
		t.Fatal(err)
	}
	if _, err = service.CreateReservation(ctx, ReservationInput{RoomID: roomID, Guest: &ReservationGuestInput{Name: "Stale", DocumentType: "RG", Document: "RG-10", Email: "stale@example.com"}, CheckIn: date(10), CheckOut: date(12), Occupants: 2, ExpectedDailyRate: "125.50"}); !errors.Is(err, ErrReservationConflict) {
		t.Fatalf("changed rate should require a new confirmation: %v", err)
	}
	if _, err = pool.Exec(ctx, `UPDATE rooms SET daily_rate=125.50 WHERE id=$1`, roomID); err != nil {
		t.Fatal(err)
	}
	guest := &ReservationGuestInput{Name: "Ana Silva", DocumentType: "CPF", Document: "123.456.789-01", Email: "ana@example.com"}
	reservation, err := service.CreateReservation(ctx, ReservationInput{RoomID: roomID, Guest: guest, CheckIn: date(10), CheckOut: date(12), Occupants: 2, ExpectedDailyRate: "125.50"})
	if err != nil || reservation.TotalAmount != "251.00" || reservation.DailyRate != "125.50" {
		t.Fatalf("reservation: %+v err=%v", reservation, err)
	}
	options, err = service.ListAvailability(ctx, AvailabilityQuery{CheckIn: date(11), CheckOut: date(13), Occupants: 1})
	if err != nil || len(options) != 0 {
		t.Fatalf("overlap must be unavailable: options=%+v err=%v", options, err)
	}
	options, err = service.ListAvailability(ctx, AvailabilityQuery{CheckIn: date(12), CheckOut: date(13), Occupants: 1})
	if err != nil || len(options) != 1 {
		t.Fatalf("checkout day must be reusable: options=%+v err=%v", options, err)
	}
	legacyDocument, primaryEmail := "987.654-X", "main@example.com"
	legacy, err := service.CreateGuest(ctx, GuestInput{Name: "Hóspede legado", Document: &legacyDocument, Email: &primaryEmail})
	if err != nil {
		t.Fatal(err)
	}
	legacyReservation, err := service.CreateReservation(ctx, ReservationInput{
		RoomID: roomID, CheckIn: date(14), CheckOut: date(15), Occupants: 1,
		ExpectedDailyRate: "125.50",
		Guest:             &ReservationGuestInput{Name: "Nome informado", DocumentType: "RG", Document: "987654X", Email: "stay@example.com"},
	})
	if err != nil {
		t.Fatalf("legacy document reservation: %v", err)
	}
	if legacyReservation.GuestID != legacy.ID || legacyReservation.GuestName != "Nome informado" || legacyReservation.ContactEmail != "stay@example.com" {
		t.Fatalf("reservation snapshot did not preserve submitted values: %+v", legacyReservation)
	}
	var preservedType, preservedEmail *string
	if err = pool.QueryRow(ctx, `SELECT document_type, email FROM guests WHERE id=$1`, legacy.ID).Scan(&preservedType, &preservedEmail); err != nil {
		t.Fatal(err)
	}
	if preservedType != nil || preservedEmail == nil || *preservedEmail != primaryEmail {
		t.Fatalf("legacy guest was changed: document_type=%v email=%v", preservedType, preservedEmail)
	}

	input := func(document, email string) ReservationInput {
		return ReservationInput{RoomID: roomID, Guest: &ReservationGuestInput{Name: "Concurrent", DocumentType: "RG", Document: document, Email: email}, CheckIn: date(20), CheckOut: date(21), Occupants: 1, ExpectedDailyRate: "125.50"}
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, request := range []ReservationInput{input("RG-20", "one@example.com"), input("RG-21", "two@example.com")} {
		wg.Add(1)
		go func(request ReservationInput) {
			defer wg.Done()
			_, createErr := service.CreateReservation(ctx, request)
			results <- createErr
		}(request)
	}
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for result := range results {
		if result == nil {
			successes++
		} else if errors.Is(result, ErrReservationConflict) {
			conflicts++
		} else {
			t.Errorf("unexpected concurrent reservation error: %v", result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent result: successes=%d conflicts=%d", successes, conflicts)
	}
}
