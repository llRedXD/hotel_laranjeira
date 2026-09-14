package hotel

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

type Room struct {
	ID           string    `json:"id"`
	Number       string    `json:"number"`
	CategoryID   string    `json:"categoryId"`
	CategoryName string    `json:"categoryName"`
	Capacity     int       `json:"capacity"`
	DailyRate    string    `json:"dailyRate"`
	IsAvailable  bool      `json:"isAvailable"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Decimal prices are strings to avoid binary floating-point rounding.
type RoomPatch struct {
	Number      *string
	CategoryID  *string
	Capacity    *int
	DailyRate   *string
	IsAvailable *bool
}
type RoomFilter struct {
	Search, CategoryID string
	Available          *bool
	Limit, Offset      int
}

var (
	roomDailyRatePattern = regexp.MustCompile(`^[0-9]{1,10}(\.[0-9]{1,2})?$`)
	roomUUIDPattern      = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

func (s *Service) SaveRoom(ctx context.Context, id string, p RoomPatch) (Room, error) {
	fields := map[string]string{}
	if p.Number == nil && p.CategoryID == nil && p.Capacity == nil && p.DailyRate == nil && p.IsAvailable == nil {
		fields["body"] = "Informe ao menos um campo."
	}
	if id == "" {
		if p.Number == nil {
			fields["number"] = "Informe o número."
		}
		if p.CategoryID == nil {
			fields["categoryId"] = "Selecione uma categoria."
		}
		if p.Capacity == nil {
			fields["capacity"] = "Informe a capacidade."
		}
		if p.DailyRate == nil {
			fields["dailyRate"] = "Informe a diária."
		}
		if p.IsAvailable == nil {
			value := true
			p.IsAvailable = &value
		}
	}
	if p.Number != nil {
		value := strings.TrimSpace(*p.Number)
		p.Number = &value
		if value == "" || utf8.RuneCountInString(value) > 20 {
			fields["number"] = "Use de 1 a 20 caracteres."
		}
	}
	if p.CategoryID != nil && !validRoomUUID(*p.CategoryID) {
		fields["categoryId"] = "Selecione uma categoria válida."
	}
	if p.Capacity != nil && (*p.Capacity < 1 || *p.Capacity > 2147483647) {
		fields["capacity"] = "A capacidade deve ser um inteiro positivo."
	}
	if p.DailyRate != nil && !roomDailyRatePattern.MatchString(*p.DailyRate) {
		fields["dailyRate"] = "Informe um valor não negativo com até duas casas decimais."
	}
	if len(fields) > 0 {
		return Room{}, &ValidationError{Fields: fields}
	}
	return s.repository.SaveRoom(ctx, id, p)
}
func validRoomUUID(value string) bool {
	return roomUUIDPattern.MatchString(value)
}
func (s *Service) GetRoom(ctx context.Context, id string) (Room, error) {
	return s.repository.GetRoom(ctx, id)
}
func (s *Service) ListRooms(ctx context.Context, f RoomFilter) ([]Room, error) {
	return s.repository.ListRooms(ctx, f)
}
