package hotel

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestReservationQueryRejectsInvalidFilters(t *testing.T) {
	service := NewService(nil)
	date := time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)
	for _, query := range []ReservationQuery{
		{GuestID: "invalid", Limit: 20},
		{From: date.AddDate(0, 0, 1), To: date, Limit: 20},
		{From: date.Add(time.Hour), Limit: 20},
		{Limit: 0}, {Limit: 101}, {Limit: 20, Offset: -1}, {Limit: 20, Offset: 100001},
	} {
		_, _, err := service.ListReservations(context.Background(), query)
		var validation *ValidationError
		if !errors.As(err, &validation) {
			t.Errorf("query=%+v err=%v", query, err)
		}
	}
}

func TestReservationNightsAndTotal(t *testing.T) {
	checkIn := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	nights, total, err := calculateReservationTotal(checkIn, checkOut, "125.50")
	if err != nil || nights != 2 || total != "251.00" {
		t.Fatalf("got nights=%d total=%q err=%v", nights, total, err)
	}
}

func TestValidateStayAndAvailability(t *testing.T) {
	checkIn := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	if err := validateAvailabilityQuery(AvailabilityQuery{CheckIn: checkIn, CheckOut: checkOut, Occupants: 1}); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []AvailabilityQuery{
		{CheckIn: checkIn, CheckOut: checkIn, Occupants: 1},
		{CheckIn: checkIn, CheckOut: checkOut, Occupants: 0},
		{CheckIn: checkIn.Add(time.Hour), CheckOut: checkOut, Occupants: 1},
	} {
		if err := validateAvailabilityQuery(invalid); err == nil {
			t.Fatalf("expected validation error for %+v", invalid)
		}
	}
	if _, _, err := calculateReservationTotal(checkIn, checkIn, "50.00"); !errors.Is(err, ErrInvalidReservation) {
		t.Fatalf("same-day checkout accepted: %v", err)
	}
	if _, _, err := calculateReservationTotal(checkIn, checkOut, "9999999999.99"); !errors.Is(err, ErrInvalidReservation) {
		t.Fatalf("total beyond NUMERIC(12,2) accepted: %v", err)
	}
}

func TestNormalizeDocumentByType(t *testing.T) {
	cpf, cpfDocument := "CPF", "123.456.789-01"
	rg, rgDocument := "RG", "ab-123.456-x"
	if actual := normalizeDocument(&cpf, &cpfDocument); actual != "12345678901" {
		t.Fatalf("CPF normalized to %q", actual)
	}
	if actual := normalizeDocument(&rg, &rgDocument); actual != "AB123456X" {
		t.Fatalf("RG normalized to %q", actual)
	}
}

func TestValidateReservationGuest(t *testing.T) {
	valid, fields := validateReservationGuest(ReservationGuestInput{
		Name: " Ana Silva ", DocumentType: "cpf", Document: "123.456.789-01", Email: "ANA@example.com",
	})
	if len(fields) != 0 || valid.Name != "Ana Silva" || valid.DocumentType != "CPF" || valid.Email != "ana@example.com" {
		t.Fatalf("normalized=%+v fields=%v", valid, fields)
	}
	_, fields = validateReservationGuest(ReservationGuestInput{Name: " ", DocumentType: "RGI", Email: "invalido"})
	for _, field := range []string{"name", "documentType", "document", "email"} {
		if fields[field] == "" {
			t.Errorf("expected validation error for %s; got %v", field, fields)
		}
	}
}
