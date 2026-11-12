package hotel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrReservationState = errors.New("a reserva foi alterada ou não permite esta operação")

type CancelReservationInput struct {
	ReservationID string
	ActorID       string
	Version       int
}

type reservationCanceller interface {
	CancelReservation(context.Context, CancelReservationInput) (Reservation, error)
}

func (s *Service) CancelReservation(ctx context.Context, input CancelReservationInput) (Reservation, error) {
	if !validRoomUUID(input.ReservationID) || !validRoomUUID(input.ActorID) || input.Version < 1 {
		return Reservation{}, &ValidationError{Fields: map[string]string{"reservation": "Informe uma reserva e uma versão válidas."}}
	}
	repository, ok := s.repository.(reservationCanceller)
	if !ok {
		return Reservation{}, fmt.Errorf("repositório não oferece cancelamento de reservas")
	}
	return repository.CancelReservation(ctx, input)
}

const reservationSelect = `SELECT r.id, r.guest_id, r.room_id, r.check_in, r.check_out, r.occupants,
 r.daily_rate::text, r.total_amount::text, r.guest_name_snapshot, r.document_type_snapshot,
 r.document_snapshot, r.contact_email_snapshot, r.created_at, rm.number, c.name,
 r.cancelled_at, r.cancelled_by, r.version
 FROM reservations r JOIN rooms rm ON rm.id = r.room_id JOIN room_categories c ON c.id = rm.category_id`

func scanReservation(row pgx.Row) (Reservation, error) {
	var r Reservation
	err := row.Scan(&r.ID, &r.GuestID, &r.RoomID, &r.CheckIn, &r.CheckOut, &r.Occupants,
		&r.DailyRate, &r.TotalAmount, &r.GuestName, &r.DocumentType, &r.Document, &r.ContactEmail,
		&r.CreatedAt, &r.RoomNumber, &r.CategoryName, &r.CancelledAt, &r.CancelledBy, &r.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return Reservation{}, ErrNotFound
	}
	return r, err
}

func (s *PostgresRepository) CancelReservation(ctx context.Context, input CancelReservationInput) (Reservation, error) {
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
	// Match creation's room lock order; a concurrent room change requires a refresh.
	var lockedRoom string
	if err = tx.QueryRow(ctx, `SELECT id FROM rooms WHERE id=$1 FOR UPDATE`, roomID).Scan(&lockedRoom); err != nil {
		return Reservation{}, err
	}
	reservation, err := scanReservation(tx.QueryRow(ctx, reservationSelect+` WHERE r.id=$1 FOR UPDATE OF r`, input.ReservationID))
	if err != nil {
		return Reservation{}, err
	}
	var now time.Time
	if err = tx.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return Reservation{}, err
	}
	reservation.setOperationalState(HotelDate(now))
	if reservation.RoomID != roomID || reservation.Version != input.Version || !reservation.CanCancel {
		return Reservation{}, ErrReservationState
	}
	err = tx.QueryRow(ctx, `UPDATE reservations SET cancelled_at=$2,cancelled_by=$3,version=version+1
	 WHERE id=$1 RETURNING version`, reservation.ID, now, input.ActorID).Scan(&reservation.Version)
	if err != nil {
		return Reservation{}, err
	}
	reservation.CancelledAt = &now
	reservation.CancelledBy = &input.ActorID
	reservation.setOperationalState(HotelDate(now))
	if err = tx.Commit(ctx); err != nil {
		return Reservation{}, err
	}
	return reservation, nil
}
