package hotel

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ReservationAuditState struct {
	Status      string `json:"status"`
	RoomID      string `json:"roomId"`
	RoomNumber  string `json:"roomNumber"`
	CheckIn     string `json:"checkIn"`
	CheckOut    string `json:"checkOut"`
	Occupants   int    `json:"occupants"`
	DailyRate   string `json:"dailyRate"`
	TotalAmount string `json:"totalAmount"`
	Version     int    `json:"version"`
}

type ReservationEvent struct {
	ID            string                 `json:"id"`
	ReservationID string                 `json:"reservationId"`
	Action        string                 `json:"action"`
	Source        string                 `json:"source"`
	ActorID       *string                `json:"actorId"`
	ActorName     *string                `json:"actorName"`
	GuestName     string                 `json:"guestName"`
	RoomNumber    string                 `json:"roomNumber"`
	OccurredAt    time.Time              `json:"occurredAt"`
	Before        *ReservationAuditState `json:"beforeState"`
	After         ReservationAuditState  `json:"afterState"`
}

type ReservationEventQuery struct {
	ReservationID string
	Search        string
	Limit         int
	Offset        int
}

type reservationEventReader interface {
	ListReservationEvents(context.Context, ReservationEventQuery) ([]ReservationEvent, int, error)
}

func (s *Service) ListReservationEvents(ctx context.Context, query ReservationEventQuery) ([]ReservationEvent, int, error) {
	if (query.ReservationID != "" && !validRoomUUID(query.ReservationID)) || len(query.Search) > 255 || query.Limit < 1 || query.Limit > 100 || query.Offset < 0 || query.Offset > 100000 {
		return nil, 0, &ValidationError{Fields: map[string]string{"filters": "Informe filtros de auditoria válidos."}}
	}
	repository, ok := s.repository.(reservationEventReader)
	if !ok {
		return nil, 0, fmt.Errorf("repositório não oferece auditoria de reservas")
	}
	return repository.ListReservationEvents(ctx, query)
}

func (s *PostgresRepository) ListReservationEvents(ctx context.Context, query ReservationEventQuery) ([]ReservationEvent, int, error) {
	var reservationID any
	if query.ReservationID != "" {
		reservationID = query.ReservationID
	}
	search := strings.TrimSpace(query.Search)
	const where = `($1::uuid IS NULL OR reservation_id=$1::uuid)
	 AND ($2='' OR guest_name ILIKE '%'||$2||'%' OR room_number ILIKE '%'||$2||'%' OR reservation_id::text ILIKE '%'||$2||'%')`
	var total int
	if err := s.db.QueryRow(ctx, `SELECT count(*) FROM reservation_events WHERE `+where, reservationID, search).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.db.Query(ctx, `SELECT id,reservation_id,action,source,actor_id,actor_name,guest_name,room_number,occurred_at,before_state,after_state
	 FROM reservation_events WHERE `+where+` ORDER BY occurred_at DESC,id DESC LIMIT $3 OFFSET $4`, reservationID, search, query.Limit, query.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	events := make([]ReservationEvent, 0)
	for rows.Next() {
		var event ReservationEvent
		if err = rows.Scan(&event.ID, &event.ReservationID, &event.Action, &event.Source, &event.ActorID, &event.ActorName, &event.GuestName, &event.RoomNumber, &event.OccurredAt, &event.Before, &event.After); err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}
	return events, total, rows.Err()
}
