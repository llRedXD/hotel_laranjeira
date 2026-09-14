package httpapi

import (
	"encoding/json"
	"github.com/hotel-laranjeira/backend/internal/hotel"
)

type patchString struct{ hotel.Field }

func (f *patchString) UnmarshalJSON(data []byte) error {
	f.Set = true
	return json.Unmarshal(data, &f.Value)
}

type guestPatchRequest struct {
	Name     patchString `json:"name"`
	Document patchString `json:"document"`
	Email    patchString `json:"email"`
	Phone    patchString `json:"phone"`
	Address  patchString `json:"address"`
}

func (p guestPatchRequest) domain() hotel.GuestPatch {
	return hotel.GuestPatch{Name: p.Name.Field, Document: p.Document.Field, Email: p.Email.Field, Phone: p.Phone.Field, Address: p.Address.Field}
}

type categoryPatchRequest struct {
	Name        patchString `json:"name"`
	Description patchString `json:"description"`
}

func (p categoryPatchRequest) domain() hotel.CategoryPatch {
	return hotel.CategoryPatch{Name: p.Name.Field, Description: p.Description.Field}
}
