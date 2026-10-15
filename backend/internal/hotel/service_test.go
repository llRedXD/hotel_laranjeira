package hotel

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestServiceRejectsInvalidInputWithoutRepository(t *testing.T) {
	service := NewService(nil)
	for _, input := range []GuestInput{
		{Name: " "}, {Name: strings.Repeat("á", 121)},
	} {
		_, err := service.CreateGuest(context.Background(), input)
		var validation *ValidationError
		if !errors.As(err, &validation) {
			t.Fatalf("validação ausente: %v", err)
		}
	}
	for _, patch := range []CategoryPatch{{}, {Name: Field{Set: true}}} {
		_, err := service.UpdateCategory(context.Background(), "unused", patch)
		var validation *ValidationError
		if !errors.As(err, &validation) {
			t.Fatalf("validação ausente: %v", err)
		}
	}
}

func TestNormalizeUnicodeAndBlankOptional(t *testing.T) {
	name, phone := strings.Repeat("á", 120), "  "
	p, err := validateGuestPatch(GuestPatch{Name: Field{true, &name}, Phone: Field{true, &phone}})
	if err != nil || p.Name.Value == nil || p.Phone.Value != nil || !p.Phone.Set {
		t.Fatalf("%+v %v", p, err)
	}
}

func TestGuestDocumentTypeValidation(t *testing.T) {
	documentType, document := " cpf ", "123.456.789-01"
	validated, err := validateGuestPatch(GuestPatch{
		Name:         Field{Set: true, Value: stringPointer("Ana")},
		DocumentType: Field{Set: true, Value: &documentType},
		Document:     Field{Set: true, Value: &document},
	})
	if err != nil {
		t.Fatal(err)
	}
	if validated.DocumentType.Value == nil || *validated.DocumentType.Value != "CPF" {
		t.Fatalf("document type was not normalized: %+v", validated.DocumentType.Value)
	}
	invalid := "Passaporte"
	_, err = validateGuestPatch(GuestPatch{Name: Field{Set: true, Value: stringPointer("Ana")}, DocumentType: Field{Set: true, Value: &invalid}})
	var validation *ValidationError
	if !errors.As(err, &validation) || validation.Fields["documentType"] == "" {
		t.Fatalf("invalid document type accepted: %v", err)
	}
}

func stringPointer(value string) *string { return &value }
