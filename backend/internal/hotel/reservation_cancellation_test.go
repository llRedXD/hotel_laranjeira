package hotel

import (
	"context"
	"errors"
	"testing"
)

func TestCancellationAndAuditRejectInvalidInputs(t *testing.T) {
	service := NewService(nil)
	id := "00000000-0000-4000-8000-000000000001"
	for _, input := range []CancelReservationInput{
		{ReservationID: "invalid", ActorID: id, Version: 1},
		{ReservationID: id, ActorID: "invalid", Version: 1},
		{ReservationID: id, ActorID: id, Version: 0},
		{ReservationID: id, ActorID: id, Version: -1},
	} {
		_, err := service.CancelReservation(context.Background(), input)
		var validation *ValidationError
		if !errors.As(err, &validation) {
			t.Fatalf("input=%+v err=%v", input, err)
		}
	}
	for _, query := range []ReservationEventQuery{
		{ReservationID: "invalid", Limit: 20}, {Limit: 0}, {Limit: 101}, {Limit: 20, Offset: -1}, {Limit: 20, Offset: 100001},
	} {
		_, _, err := service.ListReservationEvents(context.Background(), query)
		var validation *ValidationError
		if !errors.As(err, &validation) {
			t.Fatalf("query=%+v err=%v", query, err)
		}
	}
}
