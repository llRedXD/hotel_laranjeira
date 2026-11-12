package hotel

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestReservationEditingRejectsInvalidInputs(t *testing.T) {
	service := NewService(nil)
	id := "00000000-0000-4000-8000-000000000001"
	today := HotelDate(time.Now())
	for _, mutate := range []func(*ReservationUpdate){
		func(v *ReservationUpdate) { v.ID = "invalid" }, func(v *ReservationUpdate) { v.ActorID = "invalid" },
		func(v *ReservationUpdate) { v.RoomID = "invalid" }, func(v *ReservationUpdate) { v.Version = 0 },
		func(v *ReservationUpdate) { v.Occupants = 0 }, func(v *ReservationUpdate) { v.CheckOut = v.CheckIn },
		func(v *ReservationUpdate) { v.ExpectedDailyRate = "1.999" },
	} {
		input := ReservationUpdate{ID: id, ActorID: id, RoomID: id, Version: 1, Occupants: 1, CheckIn: today.AddDate(0, 0, 3), CheckOut: today.AddDate(0, 0, 5), ExpectedDailyRate: "100.00"}
		mutate(&input)
		_, err := service.UpdateReservation(context.Background(), input)
		var validation *ValidationError
		if !errors.As(err, &validation) {
			t.Fatalf("input=%+v err=%v", input, err)
		}
	}
	_, err := service.GetReservation(context.Background(), "invalid")
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("invalid lookup: %v", err)
	}
}
