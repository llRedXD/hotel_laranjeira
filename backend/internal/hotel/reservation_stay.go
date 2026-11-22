package hotel

import (
	"context"
	"fmt"
	"time"
)

type ReservationTransitionInput struct {
	ReservationID string
	ActorID       string
	Version       int
}

type ReservationCheckoutQuote struct {
	CheckoutAt   time.Time `json:"checkoutAt"`
	ActualNights int       `json:"actualNights"`
	DailyRate    string    `json:"dailyRate"`
	FinalAmount  string    `json:"finalAmount"`
	Version      int       `json:"version"`
}

type reservationStayOperator interface {
	CheckInReservation(context.Context, ReservationTransitionInput) (Reservation, error)
	CheckOutReservation(context.Context, ReservationTransitionInput) (Reservation, error)
}

func validateReservationTransition(input ReservationTransitionInput) error {
	if !validRoomUUID(input.ReservationID) || !validRoomUUID(input.ActorID) || input.Version < 1 {
		return &ValidationError{Fields: map[string]string{"reservation": "Informe uma reserva e uma versão válidas."}}
	}
	return nil
}

func (s *Service) CheckInReservation(ctx context.Context, input ReservationTransitionInput) (Reservation, error) {
	if err := validateReservationTransition(input); err != nil {
		return Reservation{}, err
	}
	repository, ok := s.repository.(reservationStayOperator)
	if !ok {
		return Reservation{}, fmt.Errorf("repositório não oferece check-in/check-out")
	}
	return repository.CheckInReservation(ctx, input)
}

func (s *Service) CheckOutReservation(ctx context.Context, input ReservationTransitionInput) (Reservation, error) {
	if err := validateReservationTransition(input); err != nil {
		return Reservation{}, err
	}
	repository, ok := s.repository.(reservationStayOperator)
	if !ok {
		return Reservation{}, fmt.Errorf("repositório não oferece check-in/check-out")
	}
	return repository.CheckOutReservation(ctx, input)
}

func (s *Service) PreviewCheckOutReservation(ctx context.Context, input ReservationTransitionInput) (ReservationCheckoutQuote, error) {
	if err := validateReservationTransition(input); err != nil {
		return ReservationCheckoutQuote{}, err
	}
	repository, ok := s.repository.(reservationEditor)
	if !ok {
		return ReservationCheckoutQuote{}, fmt.Errorf("repositório não oferece consulta de reservas")
	}
	reservation, err := repository.GetReservation(ctx, input.ReservationID)
	if err != nil {
		return ReservationCheckoutQuote{}, err
	}
	if reservation.Version != input.Version || reservation.CancelledAt != nil || reservation.CheckedInAt == nil || reservation.CheckedOutAt != nil {
		return ReservationCheckoutQuote{}, ErrReservationState
	}
	room, err := s.repository.GetRoom(ctx, reservation.RoomID)
	if err != nil {
		return ReservationCheckoutQuote{}, err
	}
	if !room.IsOccupied || room.OccupiedByReservationID == nil || *room.OccupiedByReservationID != reservation.ID {
		return ReservationCheckoutQuote{}, ErrReservationState
	}
	checkoutAt := s.now()
	if checkoutAt.Before(*reservation.CheckedInAt) {
		return ReservationCheckoutQuote{}, &ValidationError{Fields: map[string]string{"checkoutAt": "O checkout não pode ser anterior ao check-in realizado."}}
	}
	nights, total, err := calculateActualStayTotal(
		HotelDate(*reservation.CheckedInAt).Format("2006-01-02"),
		HotelDate(checkoutAt).Format("2006-01-02"),
		reservation.DailyRate,
	)
	if err != nil {
		return ReservationCheckoutQuote{}, &ValidationError{Fields: map[string]string{"dates": "Não foi possível calcular o valor final da estadia."}}
	}
	return ReservationCheckoutQuote{CheckoutAt: checkoutAt, ActualNights: nights, DailyRate: reservation.DailyRate, FinalAmount: total, Version: reservation.Version}, nil
}
