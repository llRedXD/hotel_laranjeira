package hotel

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *PostgresRepository) CheckInReservation(ctx context.Context, input ReservationTransitionInput) (Reservation, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Reservation{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT set_config('hotel.actor_id',$1,true),set_config('hotel.event_source','staff',true)`, input.ActorID); err != nil {
		return Reservation{}, err
	}
	var roomID string
	err = tx.QueryRow(ctx, `SELECT room_id FROM reservations WHERE id=$1`, input.ReservationID).Scan(&roomID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Reservation{}, ErrNotFound
	}
	if err != nil {
		return Reservation{}, err
	}
	var occupiedBy *string
	err = tx.QueryRow(ctx, `SELECT occupied_by_reservation_id FROM rooms WHERE id=$1 FOR UPDATE`, roomID).Scan(&occupiedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return Reservation{}, ErrNotFound
	}
	if err != nil {
		return Reservation{}, err
	}
	r, err := scanReservation(tx.QueryRow(ctx, reservationSelect+` WHERE r.id=$1 FOR UPDATE OF r`, input.ReservationID))
	if err != nil {
		return Reservation{}, err
	}
	var now time.Time
	if err = tx.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return Reservation{}, err
	}
	today := HotelDate(now)
	r.setOperationalState(today)
	if r.RoomID != roomID || r.Version != input.Version || r.CancelledAt != nil || r.CheckedInAt != nil || r.CheckedOutAt != nil || !today.Before(r.CheckOut) || occupiedBy != nil {
		return Reservation{}, ErrReservationState
	}
	var conflict bool
	err = tx.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM reservations other
		WHERE other.room_id=$1 AND other.id<>$2 AND other.cancelled_at IS NULL
		AND COALESCE((other.checked_in_at AT TIME ZONE 'America/Sao_Paulo')::date, other.check_in) <= $3
		AND COALESCE((other.checked_out_at AT TIME ZONE 'America/Sao_Paulo')::date, other.check_out) > $3
	)`, roomID, r.ID, today).Scan(&conflict)
	if err != nil {
		return Reservation{}, err
	}
	if conflict {
		return Reservation{}, ErrReservationConflict
	}
	if _, err = tx.Exec(ctx, `UPDATE rooms SET occupied_by_reservation_id=$2,updated_at=clock_timestamp() WHERE id=$1`, roomID, r.ID); err != nil {
		return Reservation{}, err
	}
	err = tx.QueryRow(ctx, `UPDATE reservations SET checked_in_at=$2,checked_in_by=$3,version=version+1 WHERE id=$1 RETURNING version`, r.ID, now, input.ActorID).Scan(&r.Version)
	if err != nil {
		return Reservation{}, err
	}
	r.CheckedInAt = &now
	r.CheckedInBy = &input.ActorID
	r.setOperationalState(today)
	if err = tx.Commit(ctx); err != nil {
		return Reservation{}, err
	}
	return r, nil
}

func (s *PostgresRepository) CheckOutReservation(ctx context.Context, input ReservationTransitionInput) (Reservation, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Reservation{}, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT set_config('hotel.actor_id',$1,true),set_config('hotel.event_source','staff',true)`, input.ActorID); err != nil {
		return Reservation{}, err
	}
	var roomID string
	err = tx.QueryRow(ctx, `SELECT room_id FROM reservations WHERE id=$1`, input.ReservationID).Scan(&roomID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Reservation{}, ErrNotFound
	}
	if err != nil {
		return Reservation{}, err
	}
	var occupiedBy *string
	err = tx.QueryRow(ctx, `SELECT occupied_by_reservation_id FROM rooms WHERE id=$1 FOR UPDATE`, roomID).Scan(&occupiedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return Reservation{}, ErrNotFound
	}
	if err != nil {
		return Reservation{}, err
	}
	r, err := scanReservation(tx.QueryRow(ctx, reservationSelect+` WHERE r.id=$1 FOR UPDATE OF r`, input.ReservationID))
	if err != nil {
		return Reservation{}, err
	}
	var now time.Time
	if err = tx.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return Reservation{}, err
	}
	today := HotelDate(now)
	if r.RoomID != roomID || r.Version != input.Version || r.CancelledAt != nil || r.CheckedInAt == nil || r.CheckedOutAt != nil || occupiedBy == nil || *occupiedBy != r.ID {
		return Reservation{}, ErrReservationState
	}
	nights, total, err := calculateActualStayTotal(
		HotelDate(*r.CheckedInAt).Format("2006-01-02"),
		today.Format("2006-01-02"),
		r.DailyRate,
	)
	if err != nil {
		return Reservation{}, &ValidationError{Fields: map[string]string{"checkoutAt": "O checkout não pode ser anterior ao check-in realizado."}}
	}
	err = tx.QueryRow(ctx, `UPDATE reservations SET checked_out_at=$2,checked_out_by=$3,actual_nights=$4,final_amount=$5,version=version+1 WHERE id=$1 RETURNING version`, r.ID, now, input.ActorID, nights, total).Scan(&r.Version)
	if err != nil {
		return Reservation{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE rooms SET occupied_by_reservation_id=NULL,updated_at=clock_timestamp() WHERE id=$1 AND occupied_by_reservation_id=$2`, roomID, r.ID); err != nil {
		return Reservation{}, err
	}
	r.CheckedOutAt = &now
	r.CheckedOutBy = &input.ActorID
	r.ActualNights = &nights
	r.FinalAmount = &total
	r.setOperationalState(today)
	if err = tx.Commit(ctx); err != nil {
		return Reservation{}, err
	}
	return r, nil
}
