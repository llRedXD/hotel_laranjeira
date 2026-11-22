package hotel

import (
	"time"
	_ "time/tzdata" // Keep the hotel timezone available in minimal containers.
)

const (
	ReservationConfirmed      = "confirmed"
	ReservationInProgress     = "in_progress"
	ReservationDepartureToday = "departure_today"
	ReservationEnded          = "ended"
	ReservationInHouse        = "in_house"
	ReservationCompleted      = "completed"
	ReservationCancelled      = "cancelled"
)

var hotelLocation = func() *time.Location {
	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		panic(err)
	}
	return location
}()

// HotelDate represents the hotel's calendar date at UTC midnight, like SQL DATE.
// It does not represent the instant of midnight in the hotel's timezone.
func HotelDate(now time.Time) time.Time {
	local := now.In(hotelLocation)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

func (r Reservation) StatusOn(date time.Time) string {
	if r.CancelledAt != nil {
		return ReservationCancelled
	}
	if r.CheckedOutAt != nil {
		return ReservationCompleted
	}
	if r.CheckedInAt != nil {
		return ReservationInHouse
	}
	if r.CheckOut.Before(date) {
		return ReservationEnded
	}
	if r.CheckOut.Equal(date) {
		return ReservationDepartureToday
	}
	if !r.CheckIn.After(date) {
		return ReservationInProgress
	}
	return ReservationConfirmed
}

func (r *Reservation) setOperationalState(date time.Time) {
	r.Status = r.StatusOn(date)
	r.CanEdit = r.CancelledAt == nil && r.CheckedInAt == nil && r.CheckIn.After(date)
	r.CanCancel = r.CanEdit
	r.CanCheckIn = r.CancelledAt == nil && r.CheckedInAt == nil && r.CheckedOutAt == nil && date.Before(r.CheckOut)
	r.CanCheckOut = r.CancelledAt == nil && r.CheckedInAt != nil && r.CheckedOutAt == nil
}
