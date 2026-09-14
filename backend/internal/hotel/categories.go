package hotel

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type RoomCategory struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CategoryInput struct {
	Name        string
	Description *string
}

func (s *PostgresRepository) CreateCategory(ctx context.Context, input CategoryInput) (RoomCategory, error) {
	return s.saveCategory(ctx, input)
}

func (s *PostgresRepository) UpdateCategory(ctx context.Context, id string, p CategoryPatch) (RoomCategory, error) {
	var c RoomCategory
	err := s.db.QueryRow(ctx, `UPDATE room_categories SET
 name = CASE WHEN $2 THEN $3 ELSE name END,
 description = CASE WHEN $4 THEN $5 ELSE description END,
 updated_at = NOW() WHERE id = $1
 RETURNING id, name, description, created_at, updated_at`,
		id, p.Name.Set, p.Name.Value, p.Description.Set, p.Description.Value).Scan(
		&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	return c, repositoryError(err)
}

func (s *PostgresRepository) saveCategory(ctx context.Context, input CategoryInput) (RoomCategory, error) {
	var category RoomCategory
	var err error
	err = s.db.QueryRow(ctx, `INSERT INTO room_categories (name, description) VALUES ($1, $2)
			RETURNING id, name, description, created_at, updated_at`, input.Name, input.Description).Scan(
		&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return RoomCategory{}, ErrNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return RoomCategory{}, ErrConflict
		}
		return RoomCategory{}, err
	}
	return category, nil
}

func (s *PostgresRepository) ListCategories(ctx context.Context, limit, offset int) ([]RoomCategory, error) {
	rows, err := s.db.Query(ctx, `SELECT id, name, description, created_at, updated_at FROM room_categories ORDER BY name, id LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	categories := make([]RoomCategory, 0)
	for rows.Next() {
		var category RoomCategory
		if err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, rows.Err()
}
