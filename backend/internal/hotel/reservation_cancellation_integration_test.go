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

func TestCancellationAndAuditTransactions(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL ausente: integração PostgreSQL não executada")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(context.Background())
	schema := pgx.Identifier{fmt.Sprintf("cancellation_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE") }()
	tx, err := admin.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(ctx, "SET LOCAL search_path TO "+schema+", public"); err != nil {
		t.Fatal(err)
	}
	if err = applyReservationTestSchema(ctx, tx); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	configuration, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	configuration.ConnConfig.RuntimeParams["search_path"] = schema + ",public"
	pool, err := pgxpool.NewWithConfig(ctx, configuration)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	service := NewService(NewPostgresRepository(pool))
	var actor, category string
	if err = pool.QueryRow(ctx, `INSERT INTO users(name,email,password_hash) VALUES ('Funcionário original','actor@test.example','test') RETURNING id`).Scan(&actor); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO room_categories(name) VALUES ('Teste') RETURNING id`).Scan(&category); err != nil {
		t.Fatal(err)
	}
	today := HotelDate(time.Now())
	fixture := func(number string, offset int, staff bool) Reservation {
		t.Helper()
		var room string
		if err := pool.QueryRow(ctx, `INSERT INTO rooms(number,category_id,capacity,daily_rate) VALUES ($1,$2,2,100) RETURNING id`, number, category).Scan(&room); err != nil {
			t.Fatal(err)
		}
		input := ReservationInput{RoomID: room, CheckIn: today.AddDate(0, 0, offset), CheckOut: today.AddDate(0, 0, offset+2), Occupants: 2, ExpectedDailyRate: "100.00",
			Guest: &ReservationGuestInput{Name: "Teste " + number, DocumentType: "RG", Document: "TEST" + number, Email: number + "@test.example"}}
		if staff {
			input.ActorID = actor
		}
		r, err := service.CreateReservation(ctx, input)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	events := func(id string) []ReservationEvent {
		t.Helper()
		items, total, err := service.ListReservationEvents(ctx, ReservationEventQuery{ReservationID: id, Limit: 20})
		if err != nil || total != len(items) {
			t.Fatalf("audit: %+v total=%d err=%v", items, total, err)
		}
		return items
	}
	r := fixture("future", 3, true)
	created := events(r.ID)
	if len(created) != 1 || created[0].Action != "created" || created[0].Source != "staff" || created[0].Before != nil || created[0].ActorID == nil || *created[0].ActorID != actor {
		t.Fatalf("staff creation: %+v", created)
	}
	_, err = service.CancelReservation(ctx, CancelReservationInput{ReservationID: r.ID, ActorID: actor, Version: 2})
	if !errors.Is(err, ErrReservationState) || len(events(r.ID)) != 1 {
		t.Fatalf("stale cancellation changed audit: %v", err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.CancelReservation(ctx, CancelReservationInput{ReservationID: r.ID, ActorID: actor, Version: 1})
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	successes, conflicts := 0, 0
	for result := range results {
		if result == nil {
			successes++
		} else if errors.Is(result, ErrReservationState) {
			conflicts++
		} else {
			t.Fatal(result)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
	trail := events(r.ID)
	if len(trail) != 2 || trail[0].Action != "cancelled" || trail[0].Before == nil || trail[0].Before.Status != ReservationConfirmed || trail[0].After.Status != ReservationCancelled || trail[0].After.Version != 2 || trail[0].ActorName == nil || *trail[0].ActorName != "Funcionário original" {
		t.Fatalf("cancellation event: %+v", trail)
	}
	if _, err = pool.Exec(ctx, `UPDATE users SET name='Outro nome' WHERE id=$1`, actor); err != nil {
		t.Fatal(err)
	}
	if *events(r.ID)[0].ActorName != "Funcionário original" {
		t.Fatal("audit author name was rewritten")
	}
	availability, err := service.ListAvailability(ctx, AvailabilityQuery{CheckIn: r.CheckIn, CheckOut: r.CheckOut, Occupants: 2})
	if err != nil || len(availability) != 1 || availability[0].RoomID != r.RoomID {
		t.Fatalf("room not released: %+v err=%v", availability, err)
	}
	_, err = service.CreateReservation(ctx, ReservationInput{RoomID: r.RoomID, GuestID: r.GuestID, CheckIn: r.CheckIn, CheckOut: r.CheckOut, Occupants: 1, ExpectedDailyRate: "100.00", ActorID: actor})
	if err != nil {
		t.Fatalf("rebook cancelled period: %v", err)
	}
	for _, offset := range []int{0, -1, -4} {
		past := fixture(fmt.Sprintf("past%d", offset), offset, false)
		_, err = service.CancelReservation(ctx, CancelReservationInput{ReservationID: past.ID, ActorID: actor, Version: 1})
		if !errors.Is(err, ErrReservationState) || len(events(past.ID)) != 1 {
			t.Fatalf("past/current cancellation: offset=%d err=%v", offset, err)
		}
		if events(past.ID)[0].ActorID != nil || events(past.ID)[0].Source != "public" {
			t.Fatal("public booking attributed to an employee")
		}
	}
	_, err = service.CancelReservation(ctx, CancelReservationInput{ReservationID: "00000000-0000-4000-8000-000000000001", ActorID: actor, Version: 1})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing: %v", err)
	}
	rollback := fixture("rollback", 5, false)
	if _, err = pool.Exec(ctx, `ALTER TABLE reservation_events ADD CONSTRAINT test_audit_failure CHECK(action <> 'cancelled') NOT VALID`); err != nil {
		t.Fatal(err)
	}
	_, err = service.CancelReservation(ctx, CancelReservationInput{ReservationID: rollback.ID, ActorID: actor, Version: 1})
	if err == nil {
		t.Fatal("audit failure did not abort cancellation")
	}
	var cancelled *time.Time
	var version int
	if err = pool.QueryRow(ctx, `SELECT cancelled_at,version FROM reservations WHERE id=$1`, rollback.ID).Scan(&cancelled, &version); err != nil {
		t.Fatal(err)
	}
	if cancelled != nil || version != 1 || len(events(rollback.ID)) != 1 {
		t.Fatal("audit failure left a partially cancelled reservation")
	}
	if _, err = pool.Exec(ctx, `ALTER TABLE reservation_events DROP CONSTRAINT test_audit_failure`); err != nil {
		t.Fatal(err)
	}
	items, total, err := service.ListReservationEvents(ctx, ReservationEventQuery{Search: "future", Limit: 1})
	if err != nil || total != 3 || len(items) != 1 {
		t.Fatalf("audit filter/pagination: %+v total=%d err=%v", items, total, err)
	}
}
