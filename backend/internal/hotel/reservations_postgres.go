package hotel

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *PostgresRepository) ListAvailability(ctx context.Context, query AvailabilityQuery) ([]AvailabilityOption, error) {
	var excluded any
	if query.ReservationID != "" {
		excluded = query.ReservationID
	}
	rows, err := s.db.Query(ctx, `SELECT r.id, r.number, c.name, r.capacity,
        CASE WHEN r.id = (SELECT room_id FROM reservations WHERE id=$4::uuid)
          THEN (SELECT daily_rate FROM reservations WHERE id=$4::uuid)
          ELSE r.daily_rate END::text
		FROM rooms r
		JOIN room_categories c ON c.id = r.category_id
		WHERE r.is_available = TRUE
		  AND r.capacity >= $1
		  AND NOT EXISTS (
			SELECT 1 FROM reservations current
			WHERE current.cancelled_at IS NULL AND current.room_id = r.id
			  AND ($4::uuid IS NULL OR current.id <> $4::uuid)
			  AND current.check_in < $3
			  AND current.check_out > $2
		  )
		ORDER BY r.number, r.id`, query.Occupants, query.CheckIn, query.CheckOut, excluded)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	options := make([]AvailabilityOption, 0)
	for rows.Next() {
		var option AvailabilityOption
		if err := rows.Scan(&option.RoomID, &option.Number, &option.CategoryName, &option.Capacity, &option.DailyRate); err != nil {
			return nil, err
		}
		option.Nights, option.EstimatedTotal, err = calculateReservationTotal(query.CheckIn, query.CheckOut, option.DailyRate)
		if err != nil {
			return nil, err
		}
		options = append(options, option)
	}
	return options, rows.Err()
}

func (s *PostgresRepository) ListReservations(ctx context.Context, query ReservationQuery) ([]Reservation, int, error) {
	search := strings.TrimSpace(query.Search)
	var from, to, guestID any
	if query.GuestID != "" {
		guestID = query.GuestID
	}
	if !query.From.IsZero() {
		from = query.From
	}
	if !query.To.IsZero() {
		to = query.To
	}
	const where = `($1 = '' OR r.guest_name_snapshot ILIKE '%' || $1 || '%' OR rm.number ILIKE '%' || $1 || '%' OR r.document_snapshot ILIKE '%' || $1 || '%')
		AND ($2::date IS NULL OR r.check_out > $2::date)
		AND ($3::date IS NULL OR r.check_in <= $3::date)
		AND ($4::uuid IS NULL OR r.guest_id = $4::uuid)`
	var total int
	if err := s.db.QueryRow(ctx, `SELECT count(*) FROM reservations r JOIN rooms rm ON rm.id = r.room_id WHERE `+where,
		search, from, to, guestID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.Query(ctx, `SELECT r.id, r.guest_id, r.room_id, r.check_in, r.check_out, r.occupants,
		r.daily_rate::text, r.total_amount::text, r.guest_name_snapshot, r.document_type_snapshot,
		r.document_snapshot, r.contact_email_snapshot, r.created_at, rm.number, c.name, r.cancelled_at, r.cancelled_by, r.version
		FROM reservations r JOIN rooms rm ON rm.id = r.room_id JOIN room_categories c ON c.id = rm.category_id
		WHERE `+where+` ORDER BY CASE
			WHEN r.cancelled_at IS NOT NULL THEN 3
			WHEN r.check_in <= $7::date AND r.check_out > $7::date THEN 0
			WHEN r.check_in > $7::date THEN 1 ELSE 2 END,
		r.check_in ASC, r.created_at DESC, r.id LIMIT $5 OFFSET $6`,
		search, from, to, guestID, query.Limit, query.Offset, query.date)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	reservations := make([]Reservation, 0)
	for rows.Next() {
		var reservation Reservation
		if err := rows.Scan(&reservation.ID, &reservation.GuestID, &reservation.RoomID,
			&reservation.CheckIn, &reservation.CheckOut, &reservation.Occupants, &reservation.DailyRate,
			&reservation.TotalAmount, &reservation.GuestName, &reservation.DocumentType, &reservation.Document,
			&reservation.ContactEmail, &reservation.CreatedAt, &reservation.RoomNumber, &reservation.CategoryName,
			&reservation.CancelledAt, &reservation.CancelledBy, &reservation.Version); err != nil {
			return nil, 0, err
		}
		reservations = append(reservations, reservation)
	}
	return reservations, total, rows.Err()
}

func (s *PostgresRepository) ReservationSummary(ctx context.Context, date time.Time) (ReservationSummary, error) {
	var summary ReservationSummary
	err := s.db.QueryRow(ctx, `SELECT
		COALESCE((SELECT sum(occupants) FROM reservations WHERE cancelled_at IS NULL AND check_in <= $1 AND check_out > $1), 0)::int,
		(SELECT count(*) FROM reservations WHERE cancelled_at IS NULL AND check_in = $1),
		(SELECT count(*) FROM rooms rm WHERE rm.is_available = TRUE AND NOT EXISTS (
			SELECT 1 FROM reservations r WHERE r.cancelled_at IS NULL AND r.room_id = rm.id AND r.check_in <= $1 AND r.check_out > $1
		))`, date).Scan(&summary.GuestsInHouse, &summary.CheckInsToday, &summary.AvailableRooms)
	return summary, err
}

func (s *PostgresRepository) CreateReservation(ctx context.Context, input ReservationInput) (Reservation, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Reservation{}, err
	}
	defer tx.Rollback(ctx)
	source := "public"
	if input.ActorID != "" {
		source = "staff"
	}
	if _, err = tx.Exec(ctx, `SELECT set_config('hotel.actor_id',$1,true), set_config('hotel.event_source',$2,true)`, input.ActorID, source); err != nil {
		return Reservation{}, err
	}

	var capacity int
	var isAvailable bool
	var dailyRate string
	err = tx.QueryRow(ctx, `SELECT capacity, is_available, daily_rate::text FROM rooms WHERE id = $1 FOR UPDATE`, input.RoomID).Scan(&capacity, &isAvailable, &dailyRate)
	if errors.Is(err, pgx.ErrNoRows) {
		return Reservation{}, ErrNotFound
	}
	if err != nil {
		return Reservation{}, err
	}
	if !isAvailable {
		return Reservation{}, ErrReservationConflict
	}
	if input.ExpectedDailyRate != dailyRate {
		return Reservation{}, ErrReservationConflict
	}
	if input.Occupants > capacity {
		return Reservation{}, &ValidationError{Fields: map[string]string{"occupants": "A quantidade excede a capacidade do quarto."}}
	}
	var conflicting bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS (
		SELECT 1 FROM reservations WHERE cancelled_at IS NULL AND room_id = $1 AND check_in < $3 AND check_out > $2
	)`, input.RoomID, input.CheckIn, input.CheckOut).Scan(&conflicting); err != nil {
		return Reservation{}, err
	}
	if conflicting {
		return Reservation{}, ErrReservationConflict
	}

	guestID, guestSnapshot, err := s.resolveReservationGuest(ctx, tx, input)
	if err != nil {
		return Reservation{}, err
	}
	nights, total, err := calculateReservationTotal(input.CheckIn, input.CheckOut, dailyRate)
	if err != nil {
		return Reservation{}, &ValidationError{Fields: map[string]string{"dates": "O valor total previsto excede o limite permitido."}}
	}
	if nights < 1 {
		return Reservation{}, ErrInvalidReservation
	}
	var reservation Reservation
	reservation.GuestID = guestID
	reservation.RoomID = input.RoomID
	reservation.CheckIn = input.CheckIn
	reservation.CheckOut = input.CheckOut
	reservation.Occupants = input.Occupants
	reservation.DailyRate = dailyRate
	reservation.TotalAmount = total
	reservation.GuestName = guestSnapshot.Name
	reservation.DocumentType = guestSnapshot.DocumentType
	reservation.Document = guestSnapshot.Document
	reservation.ContactEmail = guestSnapshot.Email
	err = tx.QueryRow(ctx, `INSERT INTO reservations (
		guest_id, room_id, check_in, check_out, occupants, daily_rate, total_amount,
		guest_name_snapshot, document_type_snapshot, document_snapshot, contact_email_snapshot
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	RETURNING id, created_at, version`, guestID, input.RoomID, input.CheckIn, input.CheckOut,
		input.Occupants, dailyRate, total, guestSnapshot.Name, guestSnapshot.DocumentType,
		guestSnapshot.Document, guestSnapshot.Email).Scan(&reservation.ID, &reservation.CreatedAt, &reservation.Version)
	if err != nil {
		return Reservation{}, repositoryError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Reservation{}, err
	}
	return reservation, nil
}

type reservationGuestSnapshot struct {
	Name, DocumentType, Document, Email string
}

func (s *PostgresRepository) resolveReservationGuest(ctx context.Context, tx pgx.Tx, input ReservationInput) (string, reservationGuestSnapshot, error) {
	if input.GuestID != "" {
		var guest reservationGuestSnapshot
		var documentType *string
		var document, email *string
		err := tx.QueryRow(ctx, `SELECT name, document_type, document, email FROM guests WHERE id = $1 FOR UPDATE`, input.GuestID).Scan(
			&guest.Name, &documentType, &document, &email)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", reservationGuestSnapshot{}, ErrNotFound
		}
		if err != nil {
			return "", reservationGuestSnapshot{}, err
		}
		if documentType != nil {
			guest.DocumentType = *documentType
		}
		if document != nil {
			guest.Document = *document
		}
		if email != nil {
			guest.Email = *email
		}
		if guest.Email == "" || guest.Document == "" || guest.DocumentType == "" {
			return "", reservationGuestSnapshot{}, &ValidationError{Fields: map[string]string{"guestId": "Complete nome, tipo/documento e e-mail do hóspede antes da reserva."}}
		}
		return input.GuestID, guest, nil
	}
	if input.Guest == nil {
		return "", reservationGuestSnapshot{}, ErrInvalidReservation
	}
	inputGuest := *input.Guest
	documentType := inputGuest.DocumentType
	document := inputGuest.Document
	normalized := normalizeDocument(&documentType, &document)
	var guestID string
	var existingType *string
	err := tx.QueryRow(ctx, `SELECT id, document_type FROM guests WHERE document_normalized = $1 FOR UPDATE`, normalized).Scan(&guestID, &existingType)
	if err == nil {
		if existingType != nil && *existingType != inputGuest.DocumentType {
			return "", reservationGuestSnapshot{}, ErrConflict
		}
		return guestID, reservationGuestSnapshot{inputGuest.Name, inputGuest.DocumentType, inputGuest.Document, inputGuest.Email}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", reservationGuestSnapshot{}, err
	}
	var created Guest
	err = tx.QueryRow(ctx, `INSERT INTO guests (name, document_type, document, document_normalized, email)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`, inputGuest.Name, inputGuest.DocumentType,
		inputGuest.Document, normalized, inputGuest.Email).Scan(&created.ID)
	if err != nil {
		return "", reservationGuestSnapshot{}, repositoryError(err)
	}
	return created.ID, reservationGuestSnapshot{inputGuest.Name, inputGuest.DocumentType, inputGuest.Document, inputGuest.Email}, nil
}
