package httpapi

import (
	"github.com/hotel-laranjeira/backend/internal/hotel"
	"net/http"
)

func (a *api) listReservationEvents(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	events, total, err := a.hotel.ListReservationEvents(r.Context(), hotel.ReservationEventQuery{
		ReservationID: r.URL.Query().Get("reservationId"), Search: r.URL.Query().Get("search"), Limit: limit, Offset: offset,
	})
	if writeReservationError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events, "total": total})
}
