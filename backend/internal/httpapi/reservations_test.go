package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hotel-laranjeira/backend/internal/auth"
	"github.com/hotel-laranjeira/backend/internal/hotel"
)

type reservationHTTPRepository struct {
	hotel.Repository
	created hotel.ReservationInput
}

func (r *reservationHTTPRepository) ListAvailability(context.Context, hotel.AvailabilityQuery) ([]hotel.AvailabilityOption, error) {
	return []hotel.AvailabilityOption{{RoomID: "room", Number: "10", CategoryName: "Standard", Capacity: 2, DailyRate: "125.50", Nights: 2, EstimatedTotal: "251.00"}}, nil
}

func (r *reservationHTTPRepository) CreateReservation(_ context.Context, input hotel.ReservationInput) (hotel.Reservation, error) {
	r.created = input
	return hotel.Reservation{ID: "reservation-id", GuestID: "guest-id", RoomID: input.RoomID, CheckIn: input.CheckIn, CheckOut: input.CheckOut, Occupants: input.Occupants, DailyRate: "125.50", TotalAmount: "251.00", GuestName: "Ana Silva", DocumentType: "CPF", Document: "12345678901", ContactEmail: "ana@example.com"}, nil
}

func TestParseReservationDatesUsesCivilDates(t *testing.T) {
	recorder := httptest.NewRecorder()
	checkIn, checkOut, ok := parseReservationDates(recorder, "2026-12-10", "2026-12-12")
	if !ok || checkIn.Location() != time.UTC || checkOut.Sub(checkIn).Hours() != 48 {
		t.Fatalf("dates: %v %v ok=%v", checkIn, checkOut, ok)
	}
	if _, _, ok := parseReservationDates(httptest.NewRecorder(), "2026-12-10T10:00:00Z", "2026-12-12"); ok {
		t.Fatal("timestamp accepted instead of YYYY-MM-DD")
	}
}

func TestPublicReservationConfirmationOmitsGuestData(t *testing.T) {
	body, err := json.Marshal(map[string]any{"reservation": reservationDateResponse{
		ID: "reservation-id", RoomID: "room-id", CheckIn: "2026-12-10", CheckOut: "2026-12-12",
		Occupants: 2, DailyRate: "125.50", TotalAmount: "251.00",
	}})
	if err != nil {
		t.Fatal(err)
	}
	response := string(body)
	for _, sensitive := range []string{"guestId", "guestName", "documentType", "document", "contactEmail", "phone", "address"} {
		if strings.Contains(response, `"`+sensitive+`"`) {
			t.Errorf("public confirmation exposed %s: %s", sensitive, response)
		}
	}
}

func TestPublicReservationHandlerUsesSharedServiceAndHidesGuestData(t *testing.T) {
	repository := &reservationHTTPRepository{}
	server := &api{hotel: hotel.NewService(repository)}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/public/reservations", strings.NewReader(`{"roomId":"12345678-1234-1234-1234-123456789abc","checkIn":"2026-12-10","checkOut":"2026-12-12","occupants":2,"expectedDailyRate":"125.50","guest":{"name":"Ana Silva","documentType":"CPF","document":"12345678901","email":"ana@example.com"}}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	server.createPublicReservation(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if repository.created.Guest == nil || repository.created.Guest.Document != "12345678901" {
		t.Fatalf("guest input did not reach service: %+v", repository.created)
	}
	for _, privateField := range []string{"guestId", "guestName", "documentType", "document", "contactEmail", "ana@example.com"} {
		if strings.Contains(recorder.Body.String(), privateField) {
			t.Errorf("public response leaked %q: %s", privateField, recorder.Body.String())
		}
	}
}

func TestStaffAvailabilityRequiresAuthentication(t *testing.T) {
	server := &api{auth: auth.NewService(nil, "01234567890123456789012345678901")}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/reservations/availability?checkIn=2026-12-10&checkOut=2026-12-12&occupants=1", nil)
	request.Header.Set("Authorization", "Bearer invalid")
	recorder := httptest.NewRecorder()
	server.staffAvailability(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
