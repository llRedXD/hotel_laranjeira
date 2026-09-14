package hotel

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func scanRoom(row pgx.Row) (Room, error) {
	var r Room
	err := row.Scan(&r.ID, &r.Number, &r.CategoryID, &r.CategoryName, &r.Capacity, &r.DailyRate, &r.IsAvailable, &r.CreatedAt, &r.UpdatedAt)
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) && pgerr.Code == "23503" {
		return r, &ValidationError{Fields: map[string]string{"categoryId": "Categoria não encontrada."}}
	}
	return r, repositoryError(err)
}
func (s *PostgresRepository) SaveRoom(ctx context.Context, id string, p RoomPatch) (Room, error) {
	if id == "" {
		return scanRoom(s.db.QueryRow(ctx, `WITH changed AS (
   INSERT INTO rooms(number,category_id,capacity,daily_rate,is_available)
   VALUES($1,$2,$3,$4::numeric,$5) RETURNING *)
   SELECT r.id,r.number,r.category_id,c.name,r.capacity,r.daily_rate::text,r.is_available,r.created_at,r.updated_at
   FROM changed r JOIN room_categories c ON c.id=r.category_id`, p.Number, p.CategoryID, p.Capacity, p.DailyRate, p.IsAvailable))
	}
	return scanRoom(s.db.QueryRow(ctx, `WITH changed AS (
  UPDATE rooms SET number=COALESCE($2,number), category_id=COALESCE($3::uuid,category_id),
  capacity=COALESCE($4,capacity),daily_rate=COALESCE($5::numeric,daily_rate),
  is_available=COALESCE($6,is_available),updated_at=NOW() WHERE id=$1 RETURNING *)
  SELECT r.id,r.number,r.category_id,c.name,r.capacity,r.daily_rate::text,r.is_available,r.created_at,r.updated_at
  FROM changed r JOIN room_categories c ON c.id=r.category_id`, id, p.Number, p.CategoryID, p.Capacity, p.DailyRate, p.IsAvailable))
}
func (s *PostgresRepository) GetRoom(ctx context.Context, id string) (Room, error) {
	return scanRoom(s.db.QueryRow(ctx, `SELECT r.id,r.number,r.category_id,c.name,r.capacity,r.daily_rate::text,r.is_available,r.created_at,r.updated_at
 FROM rooms r JOIN room_categories c ON c.id=r.category_id WHERE r.id=$1`, id))
}
func (s *PostgresRepository) ListRooms(ctx context.Context, f RoomFilter) ([]Room, error) {
	rows, err := s.db.Query(ctx, `SELECT r.id,r.number,r.category_id,c.name,r.capacity,r.daily_rate::text,r.is_available,r.created_at,r.updated_at
 FROM rooms r JOIN room_categories c ON c.id=r.category_id
 WHERE ($1='' OR r.number ILIKE '%'||$1||'%') AND ($2='' OR r.category_id::text=$2)
 AND ($3::boolean IS NULL OR r.is_available=$3)
 ORDER BY r.number,r.id LIMIT $4 OFFSET $5`, f.Search, f.CategoryID, f.Available, f.Limit, f.Offset)
	if err != nil {
		return nil, repositoryError(err)
	}
	defer rows.Close()
	result := []Room{}
	for rows.Next() {
		r, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, repositoryError(rows.Err())
}
