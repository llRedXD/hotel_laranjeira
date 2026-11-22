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

func TestCheckInCheckOutPersistsStateAndContractedPrice(t *testing.T) {
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
	schema := pgx.Identifier{fmt.Sprintf("reservation_stay_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE") }()
	if _, err = admin.Exec(ctx, "SET search_path TO "+schema+", public"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"001_create_users.sql", "003_create_guests_and_rooms.sql", "005_reservations.sql", "006_reservation_management.sql", "007_reservation_audit.sql", "008_reservation_stay.sql"} {
		migration, readErr := os.ReadFile("../../../database/migrations/" + name)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, err = admin.Exec(ctx, string(migration)); err != nil {
			t.Fatalf("migration %s: %v", name, err)
		}
	}
	var actorID, roomID string
	if err = admin.QueryRow(ctx, `INSERT INTO users(name,email,password_hash) VALUES ('Recepção','staff@stay.test','test') RETURNING id`).Scan(&actorID); err != nil {
		t.Fatal(err)
	}
	if err = admin.QueryRow(ctx, `WITH category AS (INSERT INTO room_categories(name) VALUES ('Standard') RETURNING id)
		INSERT INTO rooms(number,category_id,capacity,daily_rate) SELECT 'ST-1',id,2,125.50 FROM category RETURNING id`).Scan(&roomID); err != nil {
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
	today := HotelDate(time.Now())
	checkIn, checkOut := today.AddDate(0, 0, 1), today.AddDate(0, 0, 4)
	reservation, err := service.CreateReservation(ctx, ReservationInput{
		ActorID: actorID, RoomID: roomID, CheckIn: checkIn, CheckOut: checkOut, Occupants: 1, ExpectedDailyRate: "125.50",
		Guest: &ReservationGuestInput{Name: "Hóspede Teste", DocumentType: "RG", Document: "STAY-001", Email: "stay@test.example"},
	})
	if err != nil {
		t.Fatal(err)
	}
	started, err := service.CheckInReservation(ctx, ReservationTransitionInput{ReservationID: reservation.ID, ActorID: actorID, Version: reservation.Version})
	if err != nil || started.Status != ReservationInHouse || started.Version != 2 || started.CheckedInAt == nil || started.CheckedInBy == nil || *started.CheckedInBy != actorID {
		t.Fatalf("check-in: reservation=%+v err=%v", started, err)
	}
	room, err := service.GetRoom(ctx, roomID)
	if err != nil || !room.IsOccupied || room.OccupiedByReservationID == nil || *room.OccupiedByReservationID != reservation.ID {
		t.Fatalf("occupied room: room=%+v err=%v", room, err)
	}
	if _, err = service.CheckInReservation(ctx, ReservationTransitionInput{ReservationID: reservation.ID, ActorID: actorID, Version: started.Version}); !errors.Is(err, ErrReservationState) {
		t.Fatalf("duplicate check-in should conflict: %v", err)
	}
	options, err := service.ListAvailability(ctx, AvailabilityQuery{CheckIn: today, CheckOut: today.AddDate(0, 0, 1), Occupants: 1})
	if err != nil || len(options) != 0 {
		t.Fatalf("actual early check-in must block overlapping dates: options=%+v err=%v", options, err)
	}
	if _, err = service.CreateReservation(ctx, ReservationInput{
		RoomID: roomID, CheckIn: today, CheckOut: today.AddDate(0, 0, 1), Occupants: 1, ExpectedDailyRate: "125.50",
		Guest: &ReservationGuestInput{Name: "Conflito", DocumentType: "RG", Document: "STAY-CONFLICT", Email: "conflict@test.example"},
	}); !errors.Is(err, ErrReservationConflict) {
		t.Fatalf("reservation overlapping actual stay should conflict: %v", err)
	}
	futureStart := checkOut.AddDate(0, 0, 1)
	future, err := service.CreateReservation(ctx, ReservationInput{
		RoomID: roomID, CheckIn: futureStart, CheckOut: futureStart.AddDate(0, 0, 1), Occupants: 1, ExpectedDailyRate: "125.50",
		Guest: &ReservationGuestInput{Name: "Reserva Futura", DocumentType: "RG", Document: "STAY-FUTURE", Email: "future@test.example"},
	})
	if err != nil || future.ID == "" {
		t.Fatalf("non-overlapping future reservation while occupied: reservation=%+v err=%v", future, err)
	}
	if _, err = pool.Exec(ctx, `UPDATE reservations SET checked_in_at=(clock_timestamp() AT TIME ZONE 'America/Sao_Paulo' - interval '2 days') AT TIME ZONE 'America/Sao_Paulo' WHERE id=$1`, reservation.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `UPDATE rooms SET daily_rate=900.00 WHERE id=$1`, roomID); err != nil {
		t.Fatal(err)
	}
	service.now = time.Now
	quote, err := service.PreviewCheckOutReservation(ctx, ReservationTransitionInput{ReservationID: reservation.ID, ActorID: actorID, Version: started.Version})
	if err != nil || quote.ActualNights != 2 || quote.DailyRate != "125.50" || quote.FinalAmount != "251.00" {
		t.Fatalf("checkout quote: %+v err=%v", quote, err)
	}
	completed, err := service.CheckOutReservation(ctx, ReservationTransitionInput{ReservationID: reservation.ID, ActorID: actorID, Version: started.Version})
	if err != nil || completed.Status != ReservationCompleted || completed.Version != 3 || completed.ActualNights == nil || *completed.ActualNights != 2 || completed.FinalAmount == nil || *completed.FinalAmount != "251.00" || completed.CheckedOutBy == nil || *completed.CheckedOutBy != actorID {
		t.Fatalf("check-out: reservation=%+v err=%v", completed, err)
	}
	room, err = service.GetRoom(ctx, roomID)
	if err != nil || room.IsOccupied || room.OccupiedByReservationID != nil {
		t.Fatalf("released room: room=%+v err=%v", room, err)
	}
	events, total, err := service.ListReservationEvents(ctx, ReservationEventQuery{ReservationID: reservation.ID, Limit: 20})
	if err != nil || total != 4 || len(events) != 4 || events[0].Action != "checked_out" || events[0].ActorID == nil || *events[0].ActorID != actorID || events[1].Action != "updated" || events[2].Action != "checked_in" {
		t.Fatalf("stay audit: events=%+v total=%d err=%v", events, total, err)
	}
}
