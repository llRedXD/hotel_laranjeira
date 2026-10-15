package hotel

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"
)

// Field distinguishes omitted fields from explicit null, independently of HTTP.
type Field struct {
	Set   bool
	Value *string
}
type GuestPatch struct{ Name, DocumentType, Document, Email, Phone, Address Field }
type CategoryPatch struct{ Name, Description Field }
type ValidationError struct{ Fields map[string]string }

func (e *ValidationError) Error() string { return "dados inválidos" }

type Repository interface {
	SaveRoom(context.Context, string, RoomPatch) (Room, error)
	GetRoom(context.Context, string) (Room, error)
	ListRooms(context.Context, RoomFilter) ([]Room, error)
	CreateGuest(context.Context, GuestInput) (Guest, error)
	UpdateGuest(context.Context, string, GuestPatch) (Guest, error)
	GetGuest(context.Context, string) (Guest, error)
	ListGuests(context.Context, string, int, int) ([]Guest, error)
	CreateCategory(context.Context, CategoryInput) (RoomCategory, error)
	UpdateCategory(context.Context, string, CategoryPatch) (RoomCategory, error)
	ListCategories(context.Context, int, int) ([]RoomCategory, error)
	ListAvailability(context.Context, AvailabilityQuery) ([]AvailabilityOption, error)
	CreateReservation(context.Context, ReservationInput) (Reservation, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func normalize(f Field, key string, max int, required bool, fields map[string]string) Field {
	if !f.Set {
		return f
	}
	if f.Value != nil {
		value := strings.TrimSpace(*f.Value)
		f.Value = &value
		if value == "" {
			f.Value = nil
		}
	}
	if f.Value == nil {
		if required {
			fields[key] = "Campo obrigatório."
		}
	} else if !utf8.ValidString(*f.Value) || utf8.RuneCountInString(*f.Value) > max {
		fields[key] = fmt.Sprintf("Use até %d caracteres válidos.", max)
	}
	return f
}

func validateGuestPatch(p GuestPatch) (GuestPatch, error) {
	fields := map[string]string{}
	if !p.Name.Set && !p.DocumentType.Set && !p.Document.Set && !p.Email.Set && !p.Phone.Set && !p.Address.Set {
		fields["body"] = "Informe ao menos um campo."
	}
	p.Name = normalize(p.Name, "name", 120, true, fields)
	p.DocumentType = normalize(p.DocumentType, "documentType", 3, false, fields)
	p.Document = normalize(p.Document, "document", 30, false, fields)
	p.Email = normalize(p.Email, "email", 255, false, fields)
	p.Phone = normalize(p.Phone, "phone", 30, false, fields)
	p.Address = normalize(p.Address, "address", 2000, false, fields)
	if p.Email.Value != nil {
		email := strings.ToLower(*p.Email.Value)
		address, err := mail.ParseAddress(email)
		parts := strings.Split(email, "@")
		if len(parts) != 2 || err != nil || address.Address != email || !strings.Contains(parts[1], ".") {
			fields["email"] = "Informe um e-mail válido."
		}
		p.Email.Value = &email
	}
	if p.DocumentType.Value != nil {
		value := strings.ToUpper(*p.DocumentType.Value)
		if value != "RG" && value != "CPF" {
			fields["documentType"] = "Selecione RG ou CPF."
		}
		p.DocumentType.Value = &value
	}
	if len(fields) > 0 {
		return p, &ValidationError{Fields: fields}
	}
	return p, nil
}

func validateCategoryPatch(p CategoryPatch) (CategoryPatch, error) {
	fields := map[string]string{}
	if !p.Name.Set && !p.Description.Set {
		fields["body"] = "Informe ao menos um campo."
	}
	p.Name = normalize(p.Name, "name", 80, true, fields)
	p.Description = normalize(p.Description, "description", 2000, false, fields)
	if len(fields) > 0 {
		return p, &ValidationError{Fields: fields}
	}
	return p, nil
}

func (s *Service) CreateGuest(ctx context.Context, in GuestInput) (Guest, error) {
	p, err := validateGuestPatch(GuestPatch{
		Name: Field{Set: true, Value: &in.Name}, DocumentType: Field{Set: true, Value: in.DocumentType},
		Document: Field{Set: true, Value: in.Document}, Email: Field{Set: true, Value: in.Email},
		Phone: Field{Set: true, Value: in.Phone}, Address: Field{Set: true, Value: in.Address},
	})
	if err != nil {
		return Guest{}, err
	}
	return s.repository.CreateGuest(ctx, GuestInput{Name: *p.Name.Value, DocumentType: p.DocumentType.Value, Document: p.Document.Value, Email: p.Email.Value, Phone: p.Phone.Value, Address: p.Address.Value})
}
func (s *Service) UpdateGuest(ctx context.Context, id string, p GuestPatch) (Guest, error) {
	p, err := validateGuestPatch(p)
	if err != nil {
		return Guest{}, err
	}
	return s.repository.UpdateGuest(ctx, id, p)
}
func (s *Service) CreateCategory(ctx context.Context, in CategoryInput) (RoomCategory, error) {
	p, err := validateCategoryPatch(CategoryPatch{Field{true, &in.Name}, Field{true, in.Description}})
	if err != nil {
		return RoomCategory{}, err
	}
	return s.repository.CreateCategory(ctx, CategoryInput{*p.Name.Value, p.Description.Value})
}
func (s *Service) UpdateCategory(ctx context.Context, id string, p CategoryPatch) (RoomCategory, error) {
	p, err := validateCategoryPatch(p)
	if err != nil {
		return RoomCategory{}, err
	}
	return s.repository.UpdateCategory(ctx, id, p)
}
func (s *Service) GetGuest(ctx context.Context, id string) (Guest, error) {
	return s.repository.GetGuest(ctx, id)
}
func (s *Service) ListGuests(ctx context.Context, search string, limit, offset int) ([]Guest, error) {
	return s.repository.ListGuests(ctx, search, limit, offset)
}
func (s *Service) ListCategories(ctx context.Context, limit, offset int) ([]RoomCategory, error) {
	return s.repository.ListCategories(ctx, limit, offset)
}
