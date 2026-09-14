package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/hotel-laranjeira/backend/internal/hotel"
)

// Required room attributes cannot be explicitly cleared.
type roomValue[T any] struct{ Value *T }

func (v *roomValue[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("campo não aceita null")
	}
	return json.Unmarshal(data, &v.Value)
}

type roomRequest struct {
	Number      roomValue[string] `json:"number"`
	CategoryID  roomValue[string] `json:"categoryId"`
	Capacity    roomValue[int]    `json:"capacity"`
	DailyRate   roomValue[string] `json:"dailyRate"`
	IsAvailable roomValue[bool]   `json:"isAvailable"`
}

func (a *api) saveRoom(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	id := r.PathValue("id")
	if id != "" && !isUUID(id) {
		writeValidationError(w, map[string]string{"id": "Identificador inválido."})
		return
	}
	var in roomRequest
	if !decodeJSON(w, r, &in) {
		return
	}
	room, err := a.hotel.SaveRoom(r.Context(), id, hotel.RoomPatch{Number: in.Number.Value, CategoryID: in.CategoryID.Value, Capacity: in.Capacity.Value, DailyRate: in.DailyRate.Value, IsAvailable: in.IsAvailable.Value})
	status := http.StatusOK
	if id == "" {
		status = http.StatusCreated
	}
	a.roomResult(w, room, err, status)
}
func (a *api) roomAvailability(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	id := r.PathValue("id")
	if !isUUID(id) {
		writeValidationError(w, map[string]string{"id": "Identificador inválido."})
		return
	}
	var in struct {
		IsAvailable roomValue[bool] `json:"isAvailable"`
	}
	if !decodeJSON(w, r, &in) {
		return
	}
	room, err := a.hotel.SaveRoom(r.Context(), id, hotel.RoomPatch{IsAvailable: in.IsAvailable.Value})
	a.roomResult(w, room, err, http.StatusOK)
}
func (a *api) getRoom(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	id := r.PathValue("id")
	if !isUUID(id) {
		writeValidationError(w, map[string]string{"id": "Identificador inválido."})
		return
	}
	room, err := a.hotel.GetRoom(r.Context(), id)
	a.roomResult(w, room, err, http.StatusOK)
}
func (a *api) listRooms(w http.ResponseWriter, r *http.Request) {
	if !a.requireAuth(w, r) {
		return
	}
	limit, offset, ok := pagination(w, r)
	if !ok {
		return
	}
	f := hotel.RoomFilter{Search: r.URL.Query().Get("search"), CategoryID: r.URL.Query().Get("categoryId"), Limit: limit, Offset: offset}
	if len(f.Search) > 255 || (f.CategoryID != "" && !isUUID(f.CategoryID)) {
		writeValidationError(w, map[string]string{"filters": "Filtros inválidos."})
		return
	}
	if raw := r.URL.Query().Get("isAvailable"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			writeValidationError(w, map[string]string{"isAvailable": "Disponibilidade inválida."})
			return
		}
		f.Available = &value
	}
	rooms, err := a.hotel.ListRooms(r.Context(), f)
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rooms": rooms})
}
func (a *api) roomResult(w http.ResponseWriter, room hotel.Room, err error, status int) {
	var validation *hotel.ValidationError
	if errors.As(err, &validation) {
		writeValidationError(w, validation.Fields)
		return
	}
	if errors.Is(err, hotel.ErrConflict) {
		writeError(w, 409, "conflict", "Já existe um quarto com esse número.", nil)
		return
	}
	if errors.Is(err, hotel.ErrNotFound) {
		writeError(w, 404, "not_found", "Quarto não encontrado.", nil)
		return
	}
	if err != nil {
		internalError(w, err)
		return
	}
	writeJSON(w, status, map[string]any{"room": room})
}
