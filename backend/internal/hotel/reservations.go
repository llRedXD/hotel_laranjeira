package hotel

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	ErrReservationConflict = errors.New("quarto indisponível para o período")
	ErrInvalidReservation  = errors.New("reserva inválida")
	moneyPattern           = regexp.MustCompile(`^[0-9]{1,10}(\.[0-9]{1,2})?$`)
)

type AvailabilityQuery struct {
	CheckIn   time.Time
	CheckOut  time.Time
	Occupants int
}

type AvailabilityOption struct {
	RoomID         string `json:"roomId"`
	Number         string `json:"number"`
	CategoryName   string `json:"categoryName"`
	Capacity       int    `json:"capacity"`
	DailyRate      string `json:"dailyRate"`
	Nights         int    `json:"nights"`
	EstimatedTotal string `json:"estimatedTotal"`
}

type ReservationGuestInput struct {
	Name         string `json:"name"`
	DocumentType string `json:"documentType"`
	Document     string `json:"document"`
	Email        string `json:"email"`
}

type ReservationInput struct {
	RoomID            string
	GuestID           string
	Guest             *ReservationGuestInput
	CheckIn           time.Time
	CheckOut          time.Time
	Occupants         int
	ExpectedDailyRate string
}

type Reservation struct {
	ID           string    `json:"id"`
	GuestID      string    `json:"guestId"`
	RoomID       string    `json:"roomId"`
	CheckIn      time.Time `json:"checkIn"`
	CheckOut     time.Time `json:"checkOut"`
	Occupants    int       `json:"occupants"`
	DailyRate    string    `json:"dailyRate"`
	TotalAmount  string    `json:"totalAmount"`
	GuestName    string    `json:"guestName"`
	DocumentType string    `json:"documentType"`
	Document     string    `json:"document"`
	ContactEmail string    `json:"contactEmail"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (s *Service) ListAvailability(ctx context.Context, query AvailabilityQuery) ([]AvailabilityOption, error) {
	if err := validateAvailabilityQuery(query); err != nil {
		return nil, err
	}
	return s.repository.ListAvailability(ctx, query)
}

func (s *Service) CreateReservation(ctx context.Context, input ReservationInput) (Reservation, error) {
	fields := map[string]string{}
	if !validRoomUUID(input.RoomID) {
		fields["roomId"] = "Selecione um quarto válido."
	}
	if err := validateStay(input.CheckIn, input.CheckOut); err != nil {
		fields["dates"] = "Informe check-out posterior ao check-in."
	}
	if input.Occupants < 1 {
		fields["occupants"] = "Informe pelo menos um ocupante."
	}
	if !moneyPattern.MatchString(input.ExpectedDailyRate) {
		fields["expectedDailyRate"] = "Consulte novamente a diária antes de confirmar."
	}
	if (input.GuestID == "") == (input.Guest == nil) {
		fields["guest"] = "Informe um hóspede existente ou os dados do responsável."
	}
	if input.GuestID != "" && !validRoomUUID(input.GuestID) {
		fields["guestId"] = "Selecione um hóspede válido."
	}
	if input.Guest != nil {
		guest, validation := validateReservationGuest(*input.Guest)
		input.Guest = &guest
		for key, value := range validation {
			fields["guest."+key] = value
		}
	}
	if len(fields) > 0 {
		return Reservation{}, &ValidationError{Fields: fields}
	}
	return s.repository.CreateReservation(ctx, input)
}

func validateAvailabilityQuery(query AvailabilityQuery) error {
	fields := map[string]string{}
	if err := validateStay(query.CheckIn, query.CheckOut); err != nil {
		fields["dates"] = "Informe check-out posterior ao check-in."
	}
	if query.Occupants < 1 {
		fields["occupants"] = "Informe pelo menos um ocupante."
	}
	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}

func validateStay(checkIn, checkOut time.Time) error {
	if checkIn.IsZero() || checkOut.IsZero() || !isCivilDate(checkIn) || !isCivilDate(checkOut) || !checkOut.After(checkIn) {
		return ErrInvalidReservation
	}
	return nil
}

func isCivilDate(value time.Time) bool {
	return value.Hour() == 0 && value.Minute() == 0 && value.Second() == 0 && value.Nanosecond() == 0
}

func validateReservationGuest(input ReservationGuestInput) (ReservationGuestInput, map[string]string) {
	fields := map[string]string{}
	input.Name = strings.TrimSpace(input.Name)
	input.DocumentType = strings.ToUpper(strings.TrimSpace(input.DocumentType))
	input.Document = strings.TrimSpace(input.Document)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Name == "" || len([]rune(input.Name)) > 120 {
		fields["name"] = "Informe um nome com até 120 caracteres."
	}
	if input.DocumentType != "RG" && input.DocumentType != "CPF" {
		fields["documentType"] = "Selecione RG ou CPF."
	}
	if input.Document == "" || len([]rune(input.Document)) > 30 || normalizeDocument(&input.DocumentType, &input.Document) == "" {
		fields["document"] = "Informe um documento com até 30 caracteres."
	}
	address, err := mail.ParseAddress(input.Email)
	parts := strings.Split(input.Email, "@")
	if input.Email == "" || len(input.Email) > 255 || len(parts) != 2 || err != nil || address.Address != input.Email || !strings.Contains(parts[1], ".") {
		fields["email"] = "Informe um e-mail válido."
	}
	return input, fields
}

func normalizeDocument(documentType, document *string) string {
	if document == nil {
		return ""
	}
	typeValue := ""
	if documentType != nil {
		typeValue = strings.ToUpper(strings.TrimSpace(*documentType))
	}
	var normalized strings.Builder
	for _, char := range strings.ToUpper(strings.TrimSpace(*document)) {
		if typeValue == "CPF" {
			if unicode.IsDigit(char) {
				normalized.WriteRune(char)
			}
		} else if unicode.IsLetter(char) || unicode.IsDigit(char) {
			normalized.WriteRune(char)
		}
	}
	return normalized.String()
}

func calculateReservationTotal(checkIn, checkOut time.Time, dailyRate string) (int, string, error) {
	if err := validateStay(checkIn, checkOut); err != nil || !moneyPattern.MatchString(dailyRate) {
		return 0, "", ErrInvalidReservation
	}
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	parts := strings.SplitN(dailyRate, ".", 2)
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("diária inválida: %w", err)
	}
	fraction := "00"
	if len(parts) == 2 {
		fraction = parts[1]
		if len(fraction) == 1 {
			fraction += "0"
		}
	}
	cents, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("centavos inválidos: %w", err)
	}
	dailyCents := whole*100 + cents
	totalCents := dailyCents * int64(nights)
	if totalCents > 999_999_999_999 {
		return 0, "", ErrInvalidReservation
	}
	return nights, fmt.Sprintf("%d.%02d", totalCents/100, totalCents%100), nil
}
