package hotel

import (
	"testing"
	"time"
)

func TestReservationStatusUsesHotelDate(t *testing.T) {
	date := func(day int) time.Time { return time.Date(2026, 11, day, 0, 0, 0, 0, time.UTC) }
	reservation := Reservation{CheckIn: date(10), CheckOut: date(12)}
	for _, test := range []struct {
		day  int
		want string
	}{
		{9, ReservationConfirmed},
		{10, ReservationInProgress},
		{11, ReservationInProgress},
		{12, ReservationDepartureToday},
		{13, ReservationEnded},
	} {
		if got := reservation.StatusOn(date(test.day)); got != test.want {
			t.Errorf("day %d: got %s, want %s", test.day, got, test.want)
		}
		cancelled := reservation
		now := date(8)
		cancelled.CancelledAt = &now
		if got := cancelled.StatusOn(date(test.day)); got != ReservationCancelled {
			t.Errorf("cancelled on day %d: got %s", test.day, got)
		}
	}
}

func TestHotelDateAtTimezoneBoundary(t *testing.T) {
	for _, test := range []struct{ instant, want string }{
		{"2026-11-11T02:59:59Z", "2026-11-10"},
		{"2026-11-11T03:00:00Z", "2026-11-11"},
		{"2026-11-11T23:59:59-08:00", "2026-11-12"},
	} {
		now, err := time.Parse(time.RFC3339, test.instant)
		if err != nil {
			t.Fatal(err)
		}
		got := HotelDate(now)
		if got.Format("2006-01-02") != test.want || got.Location() != time.UTC || !isCivilDate(got) {
			t.Errorf("HotelDate(%s) = %v, want %s at UTC midnight", test.instant, got, test.want)
		}
	}
}

func TestReservationOperationPermissions(t *testing.T) {
	today := time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)
	for _, days := range []int{-3, -1, 0, 1} {
		for _, cancelled := range []bool{false, true} {
			r := Reservation{CheckIn: today.AddDate(0, 0, days), CheckOut: today.AddDate(0, 0, days+2)}
			if cancelled {
				r.CancelledAt = &today
			}
			r.setOperationalState(today)
			want := days > 0 && !cancelled
			if r.CanEdit != want || r.CanCancel != want {
				t.Errorf("days=%d cancelled=%v: %+v", days, cancelled, r)
			}
		}
	}
}
