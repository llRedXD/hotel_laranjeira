package hotel

import (
	"context"
	"errors"
	"testing"
)

func TestReservationTransitionRejectsInvalidInput(t *testing.T) {
	service := NewService(nil)
	id := "00000000-0000-4000-8000-000000000001"
	for _, input := range []ReservationTransitionInput{
		{ReservationID: "invalid", ActorID: id, Version: 1},
		{ReservationID: id, ActorID: "invalid", Version: 1},
		{ReservationID: id, ActorID: id, Version: 0},
	} {
		if _, err := service.CheckInReservation(context.Background(), input); !isValidationError(err) {
			t.Errorf("check-in input=%+v err=%v", input, err)
		}
		if _, err := service.CheckOutReservation(context.Background(), input); !isValidationError(err) {
			t.Errorf("check-out input=%+v err=%v", input, err)
		}
		if _, err := service.PreviewCheckOutReservation(context.Background(), input); !isValidationError(err) {
			t.Errorf("quote input=%+v err=%v", input, err)
		}
	}
}

func isValidationError(err error) bool {
	var validation *ValidationError
	return errors.As(err, &validation)
}
