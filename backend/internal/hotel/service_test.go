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
