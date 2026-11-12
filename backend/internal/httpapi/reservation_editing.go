package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/hotel-laranjeira/backend/internal/hotel"
)

func (a *api) getReservation(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	reservation, err := a.hotel.GetReservation(r.Context(), r.PathValue("id"))
	if errors.Is(err, hotel.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Reserva não encontrada.", nil)
		return
	}
	if writeReservationError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reservation": staffReservationResponse(reservation), "hotelDate": hotel.HotelDate(time.Now()).Format("2006-01-02")})
}

func (a *api) updateReservation(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireUser(w, r)
	if !ok {
		return
	}
	var body struct {
		RoomID            string `json:"roomId"`
		CheckIn           string `json:"checkIn"`
		CheckOut          string `json:"checkOut"`
		Occupants         int    `json:"occupants"`
		ExpectedDailyRate string `json:"expectedDailyRate"`
		Version           int    `json:"version"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	checkIn, checkOut, ok := parseReservationDates(w, body.CheckIn, body.CheckOut)
	if !ok {
		return
	}
	reservation, err := a.hotel.UpdateReservation(r.Context(), hotel.ReservationUpdate{
		ID: r.PathValue("id"), ActorID: user.ID, RoomID: body.RoomID, CheckIn: checkIn, CheckOut: checkOut,
		Occupants: body.Occupants, ExpectedDailyRate: body.ExpectedDailyRate, Version: body.Version,
	})
	if errors.Is(err, hotel.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Reserva ou quarto não encontrado.", nil)
		return
	}
	if errors.Is(err, hotel.ErrReservationConflict) {
		writeError(w, http.StatusConflict, "reservation_conflict", "O quarto ou a diária mudou. Consulte a disponibilidade novamente antes de salvar.", nil)
		return
	}
	if writeReservationError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reservation": staffReservationResponse(reservation)})
}
