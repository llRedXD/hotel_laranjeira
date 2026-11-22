package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/hotel-laranjeira/backend/internal/hotel"
)

func (a *api) checkInReservation(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Version int `json:"version"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	reservation, err := a.hotel.CheckInReservation(r.Context(), hotel.ReservationTransitionInput{
		ReservationID: r.PathValue("id"), ActorID: user.ID, Version: body.Version,
	})
	if errors.Is(err, hotel.ErrReservationConflict) {
		writeError(w, http.StatusConflict, "reservation_conflict", "O quarto está ocupado por outra estadia. Atualize a consulta.", nil)
		return
	}
	if writeReservationError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reservation": staffReservationResponse(reservation)})
}

func (a *api) checkOutReservation(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Version int `json:"version"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	reservation, err := a.hotel.CheckOutReservation(r.Context(), hotel.ReservationTransitionInput{
		ReservationID: r.PathValue("id"), ActorID: user.ID, Version: body.Version,
	})
	if writeReservationError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reservation": staffReservationResponse(reservation)})
}

func (a *api) quoteCheckOutReservation(w http.ResponseWriter, r *http.Request) {
	user, ok := a.requireUser(w, r)
	if !ok {
		return
	}
	version, err := strconv.Atoi(r.URL.Query().Get("version"))
	if err != nil {
		writeValidationError(w, map[string]string{"version": "Informe a versão atual da reserva."})
		return
	}
	quote, err := a.hotel.PreviewCheckOutReservation(r.Context(), hotel.ReservationTransitionInput{
		ReservationID: r.PathValue("id"), ActorID: user.ID, Version: version,
	})
	if writeReservationError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"quote": quote})
}
