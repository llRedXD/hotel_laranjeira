package hotel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func editingTestPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL ausente: integração PostgreSQL não executada")
	}
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := pgx.Identifier{fmt.Sprintf("editing_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		admin.Close(context.Background())
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
		admin.Close(context.Background())
	})
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
	t.Cleanup(pool.Close)
	return pool
}

func TestReservationEditingPriceRulesAndAudit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	pool := editingTestPool(t, ctx)
	service := NewService(NewPostgresRepository(pool))
	var actor, category string
	if err := pool.QueryRow(ctx, `INSERT INTO users(name,email,password_hash) VALUES ('Recepção','staff@test.example','test') RETURNING id`).Scan(&actor); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO room_categories(name) VALUES ('Teste') RETURNING id`).Scan(&category); err != nil {
		t.Fatal(err)
	}
	room := func(number, rate string) string {
		t.Helper()
		var id string
		if err := pool.QueryRow(ctx, `INSERT INTO rooms(number,category_id,capacity,daily_rate) VALUES ($1,$2,4,$3) RETURNING id`, number, category, rate).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	a, b, c := room("A", "100.00"), room("B", "200.00"), room("C", "100.00")
	today := HotelDate(time.Now())
	date := func(offset int) time.Time { return today.AddDate(0, 0, offset) }
	create := func(id string, start int, document string) Reservation {
		t.Helper()
		r, err := service.CreateReservation(ctx, ReservationInput{RoomID: id, CheckIn: date(start), CheckOut: date(start + 2), Occupants: 2, ExpectedDailyRate: "100.00", Guest: &ReservationGuestInput{Name: "Nome original", DocumentType: "RG", Document: document, Email: document + "@test.example"}})
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	r := create(a, 3, "original")
	if _, err := pool.Exec(ctx, `UPDATE rooms SET daily_rate=9999999999.99 WHERE id=$1`, a); err != nil {
		t.Fatal(err)
	}
	options, err := service.ListAvailability(ctx, AvailabilityQuery{ReservationID: r.ID, CheckIn: r.CheckIn, CheckOut: r.CheckOut, Occupants: 2})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, option := range options {
		if option.RoomID == a {
			found = true
			if option.DailyRate != "100.00" || option.EstimatedTotal != "200.00" {
				t.Fatalf("contract quote: %+v", option)
			}
		}
	}
	if !found {
		t.Fatal("editing excluded its own room")
	}
	public, err := service.ListAvailability(ctx, AvailabilityQuery{CheckIn: r.CheckIn, CheckOut: r.CheckOut, Occupants: 2})
	if err != nil {
		t.Fatal(err)
	}
	for _, option := range public {
		if option.RoomID == a {
			t.Fatal("public query ignored active reservation")
		}
	}
	if _, err = pool.Exec(ctx, `UPDATE guests SET name='Nome novo' WHERE id=$1`, r.GuestID); err != nil {
		t.Fatal(err)
	}
	input := ReservationUpdate{ID: r.ID, ActorID: actor, RoomID: a, CheckIn: date(4), CheckOut: date(7), Occupants: 3, ExpectedDailyRate: "100.00", Version: 1}
	updated, err := service.UpdateReservation(ctx, input)
	if err != nil || updated.DailyRate != "100.00" || updated.TotalAmount != "300.00" || updated.Version != 2 || updated.GuestID != r.GuestID || updated.GuestName != r.GuestName || updated.Document != r.Document || updated.ContactEmail != r.ContactEmail {
		t.Fatalf("same room: %+v err=%v", updated, err)
	}
	events, total, err := service.ListReservationEvents(ctx, ReservationEventQuery{ReservationID: r.ID, Limit: 20})
	if err != nil || total != 2 || events[0].Action != "updated" || events[0].ActorID == nil || *events[0].ActorID != actor || events[0].Before.TotalAmount != "200.00" || events[0].After.TotalAmount != "300.00" {
		t.Fatalf("update audit: %+v total=%d err=%v", events, total, err)
	}
	if _, err = service.UpdateReservation(ctx, input); !errors.Is(err, ErrReservationState) {
		t.Fatalf("stale version: %v", err)
	}
	input.Version = 2
	noChange, err := service.UpdateReservation(ctx, input)
	if err != nil || noChange.Version != 2 {
		t.Fatalf("no-op: %+v err=%v", noChange, err)
	}
	_, total, err = service.ListReservationEvents(ctx, ReservationEventQuery{ReservationID: r.ID, Limit: 20})
	if err != nil || total != 2 {
		t.Fatalf("no-op audit count=%d err=%v", total, err)
	}
	for _, mutate := range []func(*ReservationUpdate){
		func(v *ReservationUpdate) { v.CheckIn = today }, func(v *ReservationUpdate) { v.Occupants = 5 },
	} {
		invalid := input
		mutate(&invalid)
		_, err = service.UpdateReservation(ctx, invalid)
		var validation *ValidationError
		if !errors.As(err, &validation) {
			t.Fatalf("invalid update: %+v err=%v", invalid, err)
		}
	}
	input.RoomID = b
	input.ExpectedDailyRate = "200.00"
	if _, err = pool.Exec(ctx, `UPDATE rooms SET is_available=false WHERE id=$1`, b); err != nil {
		t.Fatal(err)
	}
	if _, err = service.UpdateReservation(ctx, input); !errors.Is(err, ErrReservationConflict) {
		t.Fatalf("disabled room: %v", err)
	}
	if _, err = pool.Exec(ctx, `UPDATE rooms SET is_available=true,daily_rate=250.25 WHERE id=$1`, b); err != nil {
		t.Fatal(err)
	}
	if _, err = service.UpdateReservation(ctx, input); !errors.Is(err, ErrReservationConflict) {
		t.Fatalf("stale quote: %v", err)
	}
	input.ExpectedDailyRate = "250.25"
	updated, err = service.UpdateReservation(ctx, input)
	if err != nil || updated.RoomID != b || updated.DailyRate != "250.25" || updated.TotalAmount != "750.75" || updated.Version != 3 {
		t.Fatalf("new room: %+v err=%v", updated, err)
	}
	if _, err = pool.Exec(ctx, `UPDATE rooms SET daily_rate=100 WHERE id=$1`, a); err != nil {
		t.Fatal(err)
	}
	public, err = service.ListAvailability(ctx, AvailabilityQuery{CheckIn: date(4), CheckOut: date(7), Occupants: 2})
	if err != nil {
		t.Fatal(err)
	}
	released := false
	for _, option := range public {
		if option.RoomID == a {
			released = true
		}
		if option.RoomID == b {
			t.Fatal("destination not blocked")
		}
	}
	if !released {
		t.Fatal("old room not released")
	}
	create(c, 4, "conflict")
	conflict := input
	conflict.Version = 3
	conflict.RoomID = c
	conflict.ExpectedDailyRate = "100.00"
	if _, err = service.UpdateReservation(ctx, conflict); !errors.Is(err, ErrReservationConflict) {
		t.Fatalf("overlapping room: %v", err)
	}
	if _, err = pool.Exec(ctx, `ALTER TABLE reservation_events ADD CONSTRAINT editing_audit_failure CHECK(action <> 'updated') NOT VALID`); err != nil {
		t.Fatal(err)
	}
	input.Version = 3
	input.CheckOut = date(8)
	if _, err = service.UpdateReservation(ctx, input); err == nil {
		t.Fatal("audit failure did not abort edit")
	}
	stored, err := service.GetReservation(ctx, r.ID)
	if err != nil || stored.Version != 3 || stored.TotalAmount != "750.75" || !stored.CheckOut.Equal(date(7)) {
		t.Fatalf("partial edit after audit failure: %+v err=%v", stored, err)
	}
	if _, err = pool.Exec(ctx, `ALTER TABLE reservation_events DROP CONSTRAINT editing_audit_failure`); err != nil {
		t.Fatal(err)
	}
	for _, offset := range []int{0, -1, -4} {
		old := create(room(fmt.Sprintf("past-room%d", offset), "100.00"), offset, fmt.Sprintf("past%d", offset))
		attempt := input
		attempt.ID = old.ID
		attempt.Version = 1
		attempt.RoomID = a
		attempt.ExpectedDailyRate = "100.00"
		if _, err = service.UpdateReservation(ctx, attempt); !errors.Is(err, ErrReservationState) {
			t.Fatalf("started reservation offset=%d: %v", offset, err)
		}
	}
	cancelled, err := service.CancelReservation(ctx, CancelReservationInput{ReservationID: r.ID, ActorID: actor, Version: 3})
	if err != nil {
		t.Fatal(err)
	}
	input.Version = cancelled.Version
	if _, err = service.UpdateReservation(ctx, input); !errors.Is(err, ErrReservationState) {
		t.Fatalf("cancelled edit: %v", err)
	}
	if _, err = service.ListAvailability(ctx, AvailabilityQuery{ReservationID: r.ID, CheckIn: date(4), CheckOut: date(7), Occupants: 2}); !errors.Is(err, ErrReservationState) {
		t.Fatalf("cancelled editing quote: %v", err)
	}
}

func TestReservationEditingConcurrency(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	pool := editingTestPool(t, ctx)
	service := NewService(NewPostgresRepository(pool))
	var actor, category string
	if err := pool.QueryRow(ctx, `INSERT INTO users(name,email,password_hash) VALUES ('Staff','staff@test.example','test') RETURNING id`).Scan(&actor); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO room_categories(name) VALUES ('Teste') RETURNING id`).Scan(&category); err != nil {
		t.Fatal(err)
	}
	today := HotelDate(time.Now())
	room := func(number string) string {
		t.Helper()
		var id string
		if err := pool.QueryRow(ctx, `INSERT INTO rooms(number,category_id,capacity,daily_rate) VALUES ($1,$2,2,100) RETURNING id`, number, category).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	create := func(roomID, document string) (Reservation, error) {
		return service.CreateReservation(ctx, ReservationInput{RoomID: roomID, CheckIn: today.AddDate(0, 0, 5), CheckOut: today.AddDate(0, 0, 7), Occupants: 1, ExpectedDailyRate: "100.00", Guest: &ReservationGuestInput{Name: document, DocumentType: "RG", Document: document, Email: document + "@test.example"}})
	}
	race := func(first, second func() error) []error {
		start := make(chan struct{})
		results := make(chan error, 2)
		go func() { <-start; results <- first() }()
		go func() { <-start; results <- second() }()
		close(start)
		return []error{<-results, <-results}
	}
	input := func(r Reservation, destination string) ReservationUpdate {
		return ReservationUpdate{ID: r.ID, ActorID: actor, RoomID: destination, CheckIn: r.CheckIn, CheckOut: r.CheckOut, Occupants: 2, ExpectedDailyRate: "100.00", Version: r.Version}
	}
	t.Run("edit versus creation", func(t *testing.T) {
		a, b := room("createA"), room("createB")
		r, err := create(a, "createRace")
		if err != nil {
			t.Fatal(err)
		}
		results := race(func() error { _, err := service.UpdateReservation(ctx, input(r, b)); return err }, func() error { _, err := create(b, "newRace"); return err })
		success := 0
		for _, err := range results {
			if err == nil {
				success++
			} else if !errors.Is(err, ErrReservationConflict) {
				t.Fatal(err)
			}
		}
		if success != 1 {
			t.Fatalf("results=%v", results)
		}
	})
	t.Run("two edits same version", func(t *testing.T) {
		a := room("editA")
		r, err := create(a, "editRace")
		if err != nil {
			t.Fatal(err)
		}
		update := input(r, a)
		results := race(func() error { _, err := service.UpdateReservation(ctx, update); return err }, func() error { _, err := service.UpdateReservation(ctx, update); return err })
		success := 0
		for _, err := range results {
			if err == nil {
				success++
			} else if !errors.Is(err, ErrReservationState) {
				t.Fatal(err)
			}
		}
		if success != 1 {
			t.Fatalf("results=%v", results)
		}
	})
	t.Run("edit versus cancellation", func(t *testing.T) {
		a, b := room("cancelA"), room("cancelB")
		r, err := create(a, "cancelRace")
		if err != nil {
			t.Fatal(err)
		}
		results := race(func() error { _, err := service.UpdateReservation(ctx, input(r, b)); return err }, func() error {
			_, err := service.CancelReservation(ctx, CancelReservationInput{ReservationID: r.ID, ActorID: actor, Version: r.Version})
			return err
		})
		success := 0
		for _, err := range results {
			if err == nil {
				success++
			} else if !errors.Is(err, ErrReservationState) {
				t.Fatal(err)
			}
		}
		if success != 1 {
			t.Fatalf("results=%v", results)
		}
	})
	t.Run("opposite room moves", func(t *testing.T) {
		a, b := room("oppositeA"), room("oppositeB")
		first, err := create(a, "oppositeFirst")
		if err != nil {
			t.Fatal(err)
		}
		second, err := create(b, "oppositeSecond")
		if err != nil {
			t.Fatal(err)
		}
		results := race(func() error { _, err := service.UpdateReservation(ctx, input(first, b)); return err }, func() error { _, err := service.UpdateReservation(ctx, input(second, a)); return err })
		for _, err := range results {
			if !errors.Is(err, ErrReservationConflict) {
				t.Fatalf("opposite moves: %v", results)
			}
		}
	})
	var conflicts int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reservations a JOIN reservations b ON a.id<b.id AND a.room_id=b.room_id AND a.check_in<b.check_out AND a.check_out>b.check_in WHERE a.cancelled_at IS NULL AND b.cancelled_at IS NULL`).Scan(&conflicts); err != nil || conflicts != 0 {
		t.Fatalf("conflicts=%d err=%v", conflicts, err)
	}
}
