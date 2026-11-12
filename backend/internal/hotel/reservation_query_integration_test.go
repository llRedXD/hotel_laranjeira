package hotel

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestReservationFiltersAndGuestHistory(t *testing.T) {
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
	schema := pgx.Identifier{fmt.Sprintf("reservation_query_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err = tx.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, "SET LOCAL search_path TO "+schema+", public"); err != nil {
		t.Fatal(err)
	}
	if err = applyReservationTestSchema(ctx, tx); err != nil {
		t.Fatal(err)
	}
	var guestA, guestB, actor, room string
	for _, target := range []*string{&guestA, &guestB} {
		if err = tx.QueryRow(ctx, `INSERT INTO guests (name) VALUES ('Ana Silva') RETURNING id`).Scan(target); err != nil {
			t.Fatal(err)
		}
	}
	if err = tx.QueryRow(ctx, `INSERT INTO users(name,email,password_hash) VALUES ('Staff','staff@test.example','test') RETURNING id`).Scan(&actor); err != nil {
		t.Fatal(err)
	}
	if err = tx.QueryRow(ctx, `WITH category AS (INSERT INTO room_categories(name) VALUES ('Standard') RETURNING id)
		INSERT INTO rooms(number,category_id,capacity,daily_rate) SELECT '101',id,2,100 FROM category RETURNING id`).Scan(&room); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct {
		guest, in, out, name string
		cancelled            bool
	}{
		{guestA, "2026-11-10", "2026-11-14", "Nome antigo", false},
		{guestA, "2026-11-01", "2026-11-03", "Ana Silva", false},
		{guestA, "2026-11-20", "2026-11-22", "Ana Silva", true},
		{guestB, "2026-11-15", "2026-11-17", "Ana Silva", false},
	} {
		_, err = tx.Exec(ctx, `INSERT INTO reservations(guest_id,room_id,check_in,check_out,occupants,daily_rate,total_amount,
		guest_name_snapshot,document_type_snapshot,document_snapshot,contact_email_snapshot,cancelled_at,cancelled_by)
		VALUES($1,$2,$3::date,$4::date,1,100,200,$5,'RG','123','stay@test.example',
		CASE WHEN $6 THEN now() END, CASE WHEN $6 THEN $7::uuid END)`,
			fixture.guest, room, fixture.in, fixture.out, fixture.name, fixture.cancelled, actor)
		if err != nil {
			t.Fatal(err)
		}
	}
	service := NewService(NewPostgresRepository(tx))
	// UTC has advanced to Nov 12, but the hotel is still on Nov 11.
	service.now = func() time.Time { return time.Date(2026, 11, 12, 1, 0, 0, 0, time.UTC) }
	date := func(day int) time.Time { return time.Date(2026, 11, day, 0, 0, 0, 0, time.UTC) }
	for _, test := range []struct {
		name          string
		query         ReservationQuery
		total, length int
	}{
		{"all", ReservationQuery{Limit: 20}, 4, 4},
		{"exact guest with historical name", ReservationQuery{GuestID: guestA, Limit: 20}, 3, 3},
		{"homonym", ReservationQuery{GuestID: guestB, Limit: 20}, 1, 1},
		{"night within earlier stay", ReservationQuery{From: date(12), To: date(12), Limit: 20}, 1, 1},
		{"checkout excluded", ReservationQuery{From: date(14), To: date(14), Limit: 20}, 0, 0},
		{"entry included", ReservationQuery{From: date(15), To: date(15), Limit: 20}, 1, 1},
		{"from only includes cancelled", ReservationQuery{From: date(20), Limit: 20}, 1, 1},
		{"to only", ReservationQuery{To: date(10), Limit: 20}, 2, 2},
		{"combined", ReservationQuery{GuestID: guestA, Search: "101", From: date(12), To: date(12), Limit: 20}, 1, 1},
		{"combined excludes homonym", ReservationQuery{GuestID: guestB, From: date(12), To: date(12), Limit: 20}, 0, 0},
		{"snapshot search", ReservationQuery{Search: "Nome antigo", Limit: 20}, 1, 1},
		{"page one", ReservationQuery{GuestID: guestA, Limit: 1}, 3, 1},
		{"page two", ReservationQuery{GuestID: guestA, Limit: 1, Offset: 1}, 3, 1},
		{"beyond last page", ReservationQuery{GuestID: guestA, Limit: 1, Offset: 3}, 3, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			items, total, err := service.ListReservations(ctx, test.query)
			if err != nil || total != test.total || len(items) != test.length {
				t.Fatalf("items=%+v total=%d err=%v", items, total, err)
			}
			for _, item := range items {
				if test.query.GuestID != "" && item.GuestID != test.query.GuestID {
					t.Fatalf("mixed guests: %+v", item)
				}
			}
		})
	}
	items, _, err := service.ListReservations(ctx, ReservationQuery{GuestID: guestA, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if items[0].Status != ReservationInProgress || items[0].CanEdit || items[0].CanCancel || items[0].Version != 1 {
		t.Fatalf("active metadata: %+v", items[0])
	}
	if items[2].Status != ReservationCancelled || items[2].CanEdit || items[2].CanCancel {
		t.Fatalf("cancelled metadata: %+v", items[2])
	}
	summary, err := service.ReservationSummary(ctx, HotelDate(service.now()))
	if err != nil || summary.Date != "2026-11-11" {
		t.Fatalf("summary reference: %+v err=%v", summary, err)
	}
}
