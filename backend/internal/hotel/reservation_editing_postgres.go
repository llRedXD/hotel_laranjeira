package hotel

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *PostgresRepository) GetReservation(ctx context.Context, id string) (Reservation, error) {
	return scanReservation(s.db.QueryRow(ctx, reservationSelect+` WHERE r.id=$1`, id))
}

func (s *PostgresRepository) UpdateReservation(ctx context.Context, input ReservationUpdate) (Reservation, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Reservation{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT set_config('hotel.actor_id',$1,true),set_config('hotel.event_source','staff',true)`, input.ActorID); err != nil {
		return Reservation{}, err
	}
	var originalRoom string
	err = tx.QueryRow(ctx, `SELECT room_id FROM reservations WHERE id=$1`, input.ID).Scan(&originalRoom)
	if errors.Is(err, pgx.ErrNoRows) {
		return Reservation{}, ErrNotFound
	}
	if err != nil {
		return Reservation{}, err
	}
	// All writers lock rooms before reservations. UUID order also covers opposite moves.
	rows, err := tx.Query(ctx, `SELECT id FROM rooms WHERE id=ANY($1::uuid[]) ORDER BY id FOR UPDATE`, []string{originalRoom, input.RoomID})
	if err != nil {
		return Reservation{}, err
	}
	count := 0
	for rows.Next() {
		count++
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return Reservation{}, err
	}
	want := 2
	if originalRoom == input.RoomID {
		want = 1
	}
	if count != want {
		return Reservation{}, ErrNotFound
	}
	r, err := scanReservation(tx.QueryRow(ctx, reservationSelect+` WHERE r.id=$1 FOR UPDATE OF r`, input.ID))
	if err != nil {
		return Reservation{}, err
	}
	var now time.Time
	if err = tx.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return Reservation{}, err
	}
	today := HotelDate(now)
	r.setOperationalState(today)
	if r.RoomID != originalRoom || r.Version != input.Version || !r.CanEdit {
		return Reservation{}, ErrReservationState
	}
	if !input.CheckIn.After(today) {
		return Reservation{}, &ValidationError{Fields: map[string]string{"checkIn": "A nova entrada deve ser posterior à data atual do hotel."}}
	}
	var capacity int
	var enabled bool
	var rate string
	if err = tx.QueryRow(ctx, `SELECT capacity,is_available,daily_rate::text FROM rooms WHERE id=$1`, input.RoomID).Scan(&capacity, &enabled, &rate); err != nil {
		return Reservation{}, err
	}
	if !enabled {
		return Reservation{}, ErrReservationConflict
	}
	if input.Occupants > capacity {
		return Reservation{}, &ValidationError{Fields: map[string]string{"occupants": "A quantidade excede a capacidade do quarto."}}
	}
	if input.RoomID == r.RoomID {
		rate = r.DailyRate
	}
	if input.ExpectedDailyRate != rate {
		return Reservation{}, ErrReservationConflict
	}
	var conflicting bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reservations WHERE room_id=$1 AND id<>$2 AND cancelled_at IS NULL AND check_in<$4 AND check_out>$3)`, input.RoomID, input.ID, input.CheckIn, input.CheckOut).Scan(&conflicting); err != nil {
		return Reservation{}, err
	}
	if conflicting {
		return Reservation{}, ErrReservationConflict
	}
	_, total, err := calculateReservationTotal(input.CheckIn, input.CheckOut, rate)
	if err != nil {
		return Reservation{}, &ValidationError{Fields: map[string]string{"dates": "O total previsto excede o limite permitido."}}
	}
	if r.RoomID == input.RoomID && r.CheckIn.Equal(input.CheckIn) && r.CheckOut.Equal(input.CheckOut) && r.Occupants == input.Occupants {
		return r, nil // No mutation or audit event when nothing changed.
	}
	_, err = tx.Exec(ctx, `UPDATE reservations SET room_id=$2,check_in=$3,check_out=$4,occupants=$5,daily_rate=$6,total_amount=$7,version=version+1 WHERE id=$1`, input.ID, input.RoomID, input.CheckIn, input.CheckOut, input.Occupants, rate, total)
	if err != nil {
		return Reservation{}, err
	}
	updated, err := scanReservation(tx.QueryRow(ctx, reservationSelect+` WHERE r.id=$1`, input.ID))
	if err != nil {
		return Reservation{}, err
	}
	updated.setOperationalState(today)
	if err = tx.Commit(ctx); err != nil {
		return Reservation{}, err
	}
	return updated, nil
}
