package httpapi

import (
	"errors"
	"net/http"

	"github.com/hotel-laranjeira/backend/internal/hotel"
)

func (a *api) cancelReservation(w http.ResponseWriter, r *http.Request) {
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
	reservation, err := a.hotel.CancelReservation(r.Context(), hotel.CancelReservationInput{
		ReservationID: r.PathValue("id"), ActorID: user.ID, Version: body.Version,
	})
	if errors.Is(err, hotel.ErrReservationState) {
		writeError(w, http.StatusConflict, "reservation_state", "A reserva foi alterada ou não pode mais ser cancelada. Atualize a consulta.", nil)
		return
	}
	if errors.Is(err, hotel.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Reserva não encontrada.", nil)
		return
	}
	if writeReservationError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reservation": staffReservationResponse(reservation)})
}

func staffReservationResponse(r hotel.Reservation) reservationDateResponse {
	return reservationDateResponse{
		ID: r.ID, GuestID: r.GuestID, RoomID: r.RoomID, CheckIn: r.CheckIn.Format("2006-01-02"), CheckOut: r.CheckOut.Format("2006-01-02"),
		Occupants: r.Occupants, DailyRate: r.DailyRate, TotalAmount: r.TotalAmount, GuestName: r.GuestName,
		DocumentType: r.DocumentType, Document: r.Document, ContactEmail: r.ContactEmail, RoomNumber: r.RoomNumber, CategoryName: r.CategoryName,
		Status: r.Status, Version: r.Version, CanEdit: &r.CanEdit, CanCancel: &r.CanCancel,
	}
}
