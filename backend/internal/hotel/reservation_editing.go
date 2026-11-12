package hotel

import (
	"context"
	"fmt"
	"time"
)

type ReservationUpdate struct {
	ID                string
	ActorID           string
	RoomID            string
	CheckIn           time.Time
	CheckOut          time.Time
	Occupants         int
	ExpectedDailyRate string
	Version           int
}

type reservationEditor interface {
	GetReservation(context.Context, string) (Reservation, error)
	UpdateReservation(context.Context, ReservationUpdate) (Reservation, error)
}

func (s *Service) GetReservation(ctx context.Context, id string) (Reservation, error) {
	if !validRoomUUID(id) {
		return Reservation{}, &ValidationError{Fields: map[string]string{"reservation": "Informe uma reserva válida."}}
	}
	repository, ok := s.repository.(reservationEditor)
	if !ok {
		return Reservation{}, fmt.Errorf("repositório não oferece edição de reservas")
	}
	r, err := repository.GetReservation(ctx, id)
	if err != nil {
		return Reservation{}, err
	}
	r.setOperationalState(HotelDate(s.now()))
	return r, nil
}

func (s *Service) UpdateReservation(ctx context.Context, input ReservationUpdate) (Reservation, error) {
	fields := map[string]string{}
	if !validRoomUUID(input.ID) || !validRoomUUID(input.ActorID) || input.Version < 1 {
		fields["reservation"] = "Informe uma reserva e uma versão válidas."
	}
	if !validRoomUUID(input.RoomID) {
		fields["roomId"] = "Selecione um quarto válido."
	}
	if validateStay(input.CheckIn, input.CheckOut) != nil {
		fields["dates"] = "Informe check-out posterior ao check-in."
	}
	if input.Occupants < 1 {
		fields["occupants"] = "Informe pelo menos um ocupante."
	}
	if !moneyPattern.MatchString(input.ExpectedDailyRate) {
		fields["expectedDailyRate"] = "Consulte novamente a diária antes de confirmar."
	}
	if len(fields) > 0 {
		return Reservation{}, &ValidationError{Fields: fields}
	}
	repository, ok := s.repository.(reservationEditor)
	if !ok {
		return Reservation{}, fmt.Errorf("repositório não oferece edição de reservas")
	}
	return repository.UpdateReservation(ctx, input)
}
