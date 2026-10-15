package hotel

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (s *PostgresRepository) ListAvailability(ctx context.Context, query AvailabilityQuery) ([]AvailabilityOption, error) {
	rows, err := s.db.Query(ctx, `SELECT r.id, r.number, c.name, r.capacity, r.daily_rate::text
		FROM rooms r
		JOIN room_categories c ON c.id = r.category_id
		WHERE r.is_available = TRUE
		  AND r.capacity >= $1
		  AND NOT EXISTS (
			SELECT 1 FROM reservations current
			WHERE current.room_id = r.id
			  AND current.check_in < $3
			  AND current.check_out > $2
		  )
		ORDER BY r.number, r.id`, query.Occupants, query.CheckIn, query.CheckOut)
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

func (s *PostgresRepository) CreateReservation(ctx context.Context, input ReservationInput) (Reservation, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Reservation{}, err
	}
	defer tx.Rollback(ctx)

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
		SELECT 1 FROM reservations WHERE room_id = $1 AND check_in < $3 AND check_out > $2
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
	RETURNING id, created_at`, guestID, input.RoomID, input.CheckIn, input.CheckOut,
		input.Occupants, dailyRate, total, guestSnapshot.Name, guestSnapshot.DocumentType,
		guestSnapshot.Document, guestSnapshot.Email).Scan(&reservation.ID, &reservation.CreatedAt)
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
