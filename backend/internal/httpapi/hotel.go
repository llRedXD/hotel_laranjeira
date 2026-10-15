package httpapi

import (
	"errors"
	"github.com/hotel-laranjeira/backend/internal/hotel"
	"net/http"
)

type categoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (a *api) createCategory(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	var input categoryRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	category, err := a.hotel.CreateCategory(r.Context(), hotel.CategoryInput{Name: input.Name, Description: input.Description})
	a.writeCategoryResult(w, category, err, http.StatusCreated)
}

func (a *api) listCategories(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	categories, err := a.hotel.ListCategories(r.Context(), limit, offset)
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"categories": categories})
}

func (a *api) updateCategory(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	id := r.PathValue("id")
	if !isUUID(id) {
		writeValidationError(w, map[string]string{"id": "Identificador inválido."})
		return
	}
	var input categoryPatchRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	category, err := a.hotel.UpdateCategory(r.Context(), id, input.domain())
	a.writeCategoryResult(w, category, err, http.StatusOK)
}

func (a *api) writeCategoryResult(w http.ResponseWriter, category hotel.RoomCategory, err error, status int) {
	var validation *hotel.ValidationError
	if errors.As(err, &validation) {
		writeValidationError(w, validation.Fields)
		return
	}
	if errors.Is(err, hotel.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Categoria não encontrada.", nil)
		return
	}
	if errors.Is(err, hotel.ErrConflict) {
		writeError(w, http.StatusConflict, "conflict", "Já existe uma categoria com este nome.", nil)
		return
	}
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, status, map[string]any{"category": category})
}

type guestRequest struct {
	Name         string  `json:"name"`
	DocumentType *string `json:"documentType"`
	Document     *string `json:"document"`
	Email        *string `json:"email"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
}

func (a *api) createGuest(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	var input guestRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	guest, err := a.hotel.CreateGuest(r.Context(), hotel.GuestInput{Name: input.Name, DocumentType: input.DocumentType, Document: input.Document, Email: input.Email, Phone: input.Phone, Address: input.Address})
	a.writeGuestResult(w, guest, err, http.StatusCreated)
}

func (a *api) listGuests(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	if len(r.URL.Query().Get("search")) > 255 {
		writeValidationError(w, map[string]string{"search": "Busca excede 255 bytes."})
		return
	}
	guests, err := a.hotel.ListGuests(r.Context(), r.URL.Query().Get("search"), limit, offset)
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"guests": guests})
}

func (a *api) getGuest(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	id := r.PathValue("id")
	if !isUUID(id) {
		writeValidationError(w, map[string]string{"id": "Identificador inválido."})
		return
	}
	guest, err := a.hotel.GetGuest(r.Context(), id)
	a.writeGuestResult(w, guest, err, http.StatusOK)
}

func (a *api) updateGuest(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	id := r.PathValue("id")
	if !isUUID(id) {
		writeValidationError(w, map[string]string{"id": "Identificador inválido."})
		return
	}
	var input guestPatchRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	guest, err := a.hotel.UpdateGuest(r.Context(), id, input.domain())
	a.writeGuestResult(w, guest, err, http.StatusOK)
}

func (a *api) writeGuestResult(w http.ResponseWriter, guest hotel.Guest, err error, successStatus int) {
	var validation *hotel.ValidationError
	if errors.As(err, &validation) {
		writeValidationError(w, validation.Fields)
		return
	}
	if errors.Is(err, hotel.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Hóspede não encontrado.", nil)
		return
	}
	if errors.Is(err, hotel.ErrConflict) {
		writeError(w, http.StatusConflict, "conflict", "Documento ou e-mail já cadastrado.", nil)
		return
	}
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, successStatus, map[string]any{"guest": guest})
}
