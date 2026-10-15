package hotel

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound = errors.New("recurso não encontrado")
	ErrConflict = errors.New("dados já cadastrados")
)

type Guest struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	DocumentType *string   `json:"documentType"`
	Document     *string   `json:"document"`
	Email        *string   `json:"email"`
	Phone        *string   `json:"phone"`
	Address      *string   `json:"address"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type GuestInput struct {
	Name         string
	DocumentType *string
	Document     *string
	Email        *string
	Phone        *string
	Address      *string
}

type PostgresRepository struct{ db DB }

// DB is satisfied by both pgxpool.Pool and pgx.Tx.
type DB interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Begin(context.Context) (pgx.Tx, error)
}

func NewPostgresRepository(db DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (s *PostgresRepository) CreateGuest(ctx context.Context, input GuestInput) (Guest, error) {
	return s.saveGuest(ctx, input)
}

func (s *PostgresRepository) UpdateGuest(ctx context.Context, id string, p GuestPatch) (Guest, error) {
	var g Guest
	err := s.db.QueryRow(ctx, `UPDATE guests SET
 name = CASE WHEN $2 THEN $3 ELSE name END,
	document_type = CASE WHEN $4 THEN $5 ELSE document_type END,
	document = CASE WHEN $6 THEN $7 ELSE document END,
	email = CASE WHEN $8 THEN $9 ELSE email END,
	phone = CASE WHEN $10 THEN $11 ELSE phone END,
	address = CASE WHEN $12 THEN $13 ELSE address END,
 document_normalized = CASE WHEN $4 OR $6 THEN
   CASE WHEN coalesce(CASE WHEN $4 THEN $5 ELSE document_type END, '') = 'CPF'
     THEN regexp_replace(CASE WHEN $6 THEN $7 ELSE document END, '[^0-9]', '', 'g')
     ELSE regexp_replace(upper(CASE WHEN $6 THEN $7 ELSE document END), '[^[:alnum:]]', '', 'g')
   END
   ELSE document_normalized END,
 updated_at = NOW() WHERE id = $1
 RETURNING id, name, document_type, document, email, phone, address, created_at, updated_at`,
		id, p.Name.Set, p.Name.Value, p.DocumentType.Set, p.DocumentType.Value, p.Document.Set, p.Document.Value,
		p.Email.Set, p.Email.Value, p.Phone.Set, p.Phone.Value, p.Address.Set, p.Address.Value).Scan(
		&g.ID, &g.Name, &g.DocumentType, &g.Document, &g.Email, &g.Phone, &g.Address, &g.CreatedAt, &g.UpdatedAt)
	return g, repositoryError(err)
}

func (s *PostgresRepository) saveGuest(ctx context.Context, input GuestInput) (Guest, error) {
	var guest Guest
	var err error
	err = s.db.QueryRow(ctx, `INSERT INTO guests (name, document_type, document, document_normalized, email, phone, address)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, name, document_type, document, email, phone, address, created_at, updated_at`,
		input.Name, input.DocumentType, input.Document, normalizeDocument(input.DocumentType, input.Document), input.Email, input.Phone, input.Address).Scan(
		&guest.ID, &guest.Name, &guest.DocumentType, &guest.Document, &guest.Email, &guest.Phone, &guest.Address, &guest.CreatedAt, &guest.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return Guest{}, ErrNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Guest{}, ErrConflict
		}
		return Guest{}, err
	}
	return guest, nil
}

func (s *PostgresRepository) GetGuest(ctx context.Context, id string) (Guest, error) {
	var guest Guest
	err := s.db.QueryRow(ctx, `SELECT id, name, document_type, document, email, phone, address, created_at, updated_at FROM guests WHERE id = $1`, id).Scan(
		&guest.ID, &guest.Name, &guest.DocumentType, &guest.Document, &guest.Email, &guest.Phone, &guest.Address, &guest.CreatedAt, &guest.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Guest{}, ErrNotFound
	}
	return guest, err
}

func (s *PostgresRepository) ListGuests(ctx context.Context, search string, limit, offset int) ([]Guest, error) {
	rows, err := s.db.Query(ctx, `SELECT id, name, document_type, document, email, phone, address, created_at, updated_at
		FROM guests WHERE $1 = '' OR name ILIKE '%' || $1 || '%' OR document ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%'
		ORDER BY name, id LIMIT $2 OFFSET $3`, strings.TrimSpace(search), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	guests := make([]Guest, 0)
	for rows.Next() {
		var guest Guest
		if err = rows.Scan(&guest.ID, &guest.Name, &guest.DocumentType, &guest.Document, &guest.Email, &guest.Phone, &guest.Address, &guest.CreatedAt, &guest.UpdatedAt); err != nil {
			return nil, err
		}
		guests = append(guests, guest)
	}
	return guests, rows.Err()
}

func repositoryError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return fmt.Errorf("persistência de hotel: %w", err)
}
