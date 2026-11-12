package hotel

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestReservationManagementMigrationAndCancellation(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL ausente: integração PostgreSQL não executada")
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
	schema := pgx.Identifier{fmt.Sprintf("reservation_management_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err = tx.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, "SET LOCAL search_path TO "+schema+", public"); err != nil {
		t.Fatal(err)
	}
	apply := func(name string) {
		t.Helper()
		migration, err := os.ReadFile("../../../database/migrations/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = tx.Exec(ctx, string(migration)); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"001_create_users.sql", "003_create_guests_and_rooms.sql", "005_reservations.sql"} {
		apply(name)
	}
	var roomID, guestID, reservationID, userID string
	if err = tx.QueryRow(ctx, `INSERT INTO users (name, email, password_hash) VALUES ('Recepção', 'staff@example.com', 'test') RETURNING id`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	if err = tx.QueryRow(ctx, `INSERT INTO guests (name, document_type, document, document_normalized, email)
		VALUES ('Ana', 'RG', '123', '123', 'ana@example.com') RETURNING id`).Scan(&guestID); err != nil {
		t.Fatal(err)
	}
	if err = tx.QueryRow(ctx, `WITH category AS (INSERT INTO room_categories (name) VALUES ('Standard') RETURNING id)
		INSERT INTO rooms (number, category_id, capacity, daily_rate) SELECT '101', id, 2, 100 FROM category RETURNING id`).Scan(&roomID); err != nil {
		t.Fatal(err)
	}
	if err = tx.QueryRow(ctx, `INSERT INTO reservations (guest_id, room_id, check_in, check_out, occupants, daily_rate,
		total_amount, guest_name_snapshot, document_type_snapshot, document_snapshot, contact_email_snapshot)
		VALUES ($1, $2, '2026-12-10', '2026-12-12', 2, 100, 200, 'Ana original', 'RG', '123', 'stay@example.com') RETURNING id`, guestID, roomID).Scan(&reservationID); err != nil {
		t.Fatal(err)
	}
	apply("006_reservation_management.sql")
	apply("007_reservation_audit.sql")
	service := NewService(NewPostgresRepository(tx))
	baseline, totalEvents, eventErr := service.ListReservationEvents(ctx, ReservationEventQuery{ReservationID: reservationID, Limit: 20})
	if eventErr != nil || totalEvents != 1 || len(baseline) != 1 || baseline[0].Action != "imported" || baseline[0].ActorID != nil || baseline[0].Before != nil || baseline[0].After.Version != 1 {
		t.Fatalf("migration baseline: %+v total=%d err=%v", baseline, totalEvents, eventErr)
	}
	service.now = func() time.Time { return time.Date(2026, 12, 10, 2, 59, 0, 0, time.UTC) }
	listed, total, err := service.ListReservations(ctx, ReservationQuery{Limit: 20})
	if err != nil || total != 1 || len(listed) != 1 {
		t.Fatalf("list after migration: %+v total=%d err=%v", listed, total, err)
	}
	old := listed[0]
	if old.ID != reservationID || old.Version != 1 || old.CancelledAt != nil || old.CancelledBy != nil ||
		old.Status != ReservationConfirmed || old.GuestName != "Ana original" || old.ContactEmail != "stay@example.com" ||
		old.TotalAmount != "200.00" || old.DailyRate != "100.00" || old.GuestID != guestID || old.RoomID != roomID {
		t.Fatalf("migration changed existing reservation: %+v", old)
	}
	date := func(day int) time.Time { return time.Date(2026, 12, day, 0, 0, 0, 0, time.UTC) }
	summary, err := service.ReservationSummary(ctx, date(10))
	if err != nil || summary.GuestsInHouse != 2 || summary.CheckInsToday != 1 || summary.AvailableRooms != 0 {
		t.Fatalf("active summary: %+v err=%v", summary, err)
	}
	// Checkpoint 1 exercises persisted cancellation; the HTTP action comes in checkpoint 3.
	if _, err = tx.Exec(ctx, `UPDATE reservations SET cancelled_at=now(), cancelled_by=$2, version=version+1 WHERE id=$1`, reservationID, userID); err != nil {
		t.Fatal(err)
	}
	listed, total, err = service.ListReservations(ctx, ReservationQuery{Limit: 20})
	if err != nil || total != 1 || len(listed) != 1 {
		t.Fatalf("cancelled history: %+v total=%d err=%v", listed, total, err)
	}
	if listed[0].Status != ReservationCancelled || listed[0].Version != 2 || listed[0].CancelledAt == nil || listed[0].CancelledBy == nil || *listed[0].CancelledBy != userID {
		t.Fatalf("cancelled metadata: %+v", listed[0])
	}
	for _, day := range []int{10, 11} {
		summary, err = service.ReservationSummary(ctx, date(day))
		if err != nil || summary.GuestsInHouse != 0 || summary.CheckInsToday != 0 || summary.AvailableRooms != 1 {
			t.Fatalf("cancelled summary day %d: %+v err=%v", day, summary, err)
		}
	}
	options, err := service.ListAvailability(ctx, AvailabilityQuery{CheckIn: date(10), CheckOut: date(12), Occupants: 2})
	if err != nil || len(options) != 1 {
		t.Fatalf("cancelled room not released: %+v err=%v", options, err)
	}
	created, err := service.CreateReservation(ctx, ReservationInput{RoomID: roomID, GuestID: guestID, CheckIn: date(10), CheckOut: date(12), Occupants: 1, ExpectedDailyRate: "100.00"})
	if err != nil || created.Version != 1 || created.Status != ReservationConfirmed {
		t.Fatalf("rebooking: %+v err=%v", created, err)
	}
	listed, total, err = service.ListReservations(ctx, ReservationQuery{Limit: 20})
	if err != nil || total != 2 || len(listed) != 2 || listed[0].ID != created.ID || listed[1].ID != reservationID {
		t.Fatalf("active and cancelled history/order: %+v total=%d err=%v", listed, total, err)
	}
	for _, statement := range []string{
		`UPDATE reservations SET version=0 WHERE id=$1`,
		`UPDATE reservations SET cancelled_at=now() WHERE id=$1`,
		`UPDATE reservations SET cancelled_by='00000000-0000-0000-0000-000000000001', cancelled_at=now() WHERE id=$1`,
	} {
		if _, err = tx.Exec(ctx, "SAVEPOINT invalid_change"); err != nil {
			t.Fatal(err)
		}
		if _, err = tx.Exec(ctx, statement, created.ID); err == nil {
			t.Fatalf("invalid metadata accepted: %s", statement)
		}
		if _, err = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT invalid_change"); err != nil {
			t.Fatal(err)
		}
	}
}
