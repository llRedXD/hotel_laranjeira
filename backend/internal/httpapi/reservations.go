package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/hotel-laranjeira/backend/internal/hotel"
)

type availabilityQuery struct {
	CheckIn   time.Time
	CheckOut  time.Time
	Occupants int
}

type reservationRequest struct {
	RoomID            string                       `json:"roomId"`
	GuestID           string                       `json:"guestId"`
	Guest             *hotel.ReservationGuestInput `json:"guest"`
	CheckIn           string                       `json:"checkIn"`
	CheckOut          string                       `json:"checkOut"`
	Occupants         int                          `json:"occupants"`
	ExpectedDailyRate string                       `json:"expectedDailyRate"`
}

type reservationDateResponse struct {
	ID           string `json:"id"`
	GuestID      string `json:"guestId,omitempty"`
	RoomID       string `json:"roomId"`
	CheckIn      string `json:"checkIn"`
	CheckOut     string `json:"checkOut"`
	Occupants    int    `json:"occupants"`
	DailyRate    string `json:"dailyRate"`
	TotalAmount  string `json:"totalAmount"`
	GuestName    string `json:"guestName,omitempty"`
	DocumentType string `json:"documentType,omitempty"`
	Document     string `json:"document,omitempty"`
	ContactEmail string `json:"contactEmail,omitempty"`
	RoomNumber   string `json:"roomNumber,omitempty"`
	CategoryName string `json:"categoryName,omitempty"`
}

func (a *api) publicAvailability(w http.ResponseWriter, r *http.Request) {
	query, ok := parseAvailabilityQuery(r)
	if !ok {
		writeValidationError(w, map[string]string{"filters": "Informe um período e uma quantidade válidos."})
		return
	}
	rooms, err := a.hotel.ListAvailability(r.Context(), hotel.AvailabilityQuery{CheckIn: query.CheckIn, CheckOut: query.CheckOut, Occupants: query.Occupants})
	if writeReservationError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rooms": rooms})
}

func (a *api) staffAvailability(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	query, ok := parseAvailabilityQuery(r)
	if !ok {
		writeValidationError(w, map[string]string{"filters": "Informe um período e uma quantidade válidos."})
		return
	}
	rooms, err := a.hotel.ListAvailability(r.Context(), hotel.AvailabilityQuery{CheckIn: query.CheckIn, CheckOut: query.CheckOut, Occupants: query.Occupants})
	if writeReservationError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rooms": rooms})
}

func (a *api) listReservations(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	search := r.URL.Query().Get("search")
	if len(search) > 255 {
		writeValidationError(w, map[string]string{"search": "Busca excede 255 bytes."})
		return
	}
	from, err := parseOptionalDate(r.URL.Query().Get("from"))
	if err != nil {
		writeValidationError(w, map[string]string{"from": "Use uma data válida no formato AAAA-MM-DD."})
		return
	}
	to, err := parseOptionalDate(r.URL.Query().Get("to"))
	if err != nil {
		writeValidationError(w, map[string]string{"to": "Use uma data válida no formato AAAA-MM-DD."})
		return
	}
	reservations, total, err := a.hotel.ListReservations(r.Context(), hotel.ReservationQuery{
		Search: search, From: from, To: to, Limit: limit, Offset: offset,
	})
	if writeReservationError(w, err) {
		return
	}
	items := make([]reservationDateResponse, 0, len(reservations))
	for _, reservation := range reservations {
		items = append(items, reservationDateResponse{
			ID: reservation.ID, GuestID: reservation.GuestID, RoomID: reservation.RoomID,
			CheckIn: reservation.CheckIn.Format("2006-01-02"), CheckOut: reservation.CheckOut.Format("2006-01-02"),
			Occupants: reservation.Occupants, DailyRate: reservation.DailyRate, TotalAmount: reservation.TotalAmount,
			GuestName: reservation.GuestName, DocumentType: reservation.DocumentType, Document: reservation.Document,
			ContactEmail: reservation.ContactEmail, RoomNumber: reservation.RoomNumber, CategoryName: reservation.CategoryName,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"reservations": items, "total": total})
}

func (a *api) reservationSummary(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	date, err := parseOptionalDate(r.URL.Query().Get("date"))
	if err != nil || date.IsZero() {
		writeValidationError(w, map[string]string{"date": "Informe uma data no formato AAAA-MM-DD."})
		return
	}
	summary, err := a.hotel.ReservationSummary(r.Context(), date)
	if writeReservationError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"summary": summary})
}

func parseOptionalDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", value)
}

func (a *api) createPublicReservation(w http.ResponseWriter, r *http.Request) {
	var request reservationRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	checkIn, checkOut, ok := parseReservationDates(w, request.CheckIn, request.CheckOut)
	if !ok {
		return
	}
	reservation, err := a.hotel.CreateReservation(r.Context(), hotel.ReservationInput{
		RoomID: request.RoomID, Guest: request.Guest, CheckIn: checkIn, CheckOut: checkOut, Occupants: request.Occupants, ExpectedDailyRate: request.ExpectedDailyRate,
	})
	if writeReservationError(w, err) {
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"reservation": reservationDateResponse{
		ID: reservation.ID, RoomID: reservation.RoomID, CheckIn: reservation.CheckIn.Format("2006-01-02"),
		CheckOut: reservation.CheckOut.Format("2006-01-02"), Occupants: reservation.Occupants,
		DailyRate: reservation.DailyRate, TotalAmount: reservation.TotalAmount,
	}})
}

func (a *api) createStaffReservation(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	var request reservationRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	checkIn, checkOut, ok := parseReservationDates(w, request.CheckIn, request.CheckOut)
	if !ok {
		return
	}
	reservation, err := a.hotel.CreateReservation(r.Context(), hotel.ReservationInput{
		RoomID: request.RoomID, GuestID: request.GuestID, CheckIn: checkIn, CheckOut: checkOut, Occupants: request.Occupants, ExpectedDailyRate: request.ExpectedDailyRate,
	})
	if writeReservationError(w, err) {
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"reservation": reservationDateResponse{
		ID: reservation.ID, GuestID: reservation.GuestID, RoomID: reservation.RoomID,
		CheckIn: reservation.CheckIn.Format("2006-01-02"), CheckOut: reservation.CheckOut.Format("2006-01-02"),
		Occupants: reservation.Occupants, DailyRate: reservation.DailyRate, TotalAmount: reservation.TotalAmount,
		GuestName: reservation.GuestName, DocumentType: reservation.DocumentType,
		Document: reservation.Document, ContactEmail: reservation.ContactEmail,
	}})
}

func parseAvailabilityQuery(r *http.Request) (availabilityQuery, bool) {
	checkIn, errIn := time.Parse("2006-01-02", r.URL.Query().Get("checkIn"))
	checkOut, errOut := time.Parse("2006-01-02", r.URL.Query().Get("checkOut"))
	occupants, errOccupants := strconv.Atoi(r.URL.Query().Get("occupants"))
	if errIn != nil || errOut != nil || errOccupants != nil {
		return availabilityQuery{}, false
	}
	return availabilityQuery{CheckIn: checkIn, CheckOut: checkOut, Occupants: occupants}, true
}

func parseReservationDates(w http.ResponseWriter, checkInValue, checkOutValue string) (time.Time, time.Time, bool) {
	checkIn, errIn := time.Parse("2006-01-02", checkInValue)
	checkOut, errOut := time.Parse("2006-01-02", checkOutValue)
	if errIn != nil || errOut != nil {
		writeValidationError(w, map[string]string{"dates": "Use datas no formato AAAA-MM-DD."})
		return time.Time{}, time.Time{}, false
	}
	return checkIn, checkOut, true
}

func writeReservationError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	var validation *hotel.ValidationError
	if errors.As(err, &validation) {
		writeValidationError(w, validation.Fields)
		return true
	}
	if errors.Is(err, hotel.ErrReservationConflict) || errors.Is(err, hotel.ErrConflict) {
		writeError(w, http.StatusConflict, "reservation_conflict", "Não foi possível confirmar a reserva para estes dados e período.", nil)
		return true
	}
	if errors.Is(err, hotel.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Quarto ou hóspede não encontrado.", nil)
		return true
	}
	internalError(w, err)
	return true
}
