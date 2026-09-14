package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Begin(context.Context) (pgx.Tx, error)
}
type PostgresRepository struct{ db DB }

func NewPostgresRepository(db DB) *PostgresRepository { return &PostgresRepository{db: db} }

func persistenceError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return fmt.Errorf("persistência de autenticação: %w", err)
}
func (s *PostgresRepository) UserByEmail(ctx context.Context, email string) (User, string, error) {
	var user User
	var passwordHash string
	err := s.db.QueryRow(ctx, `SELECT id, name, email, role, is_active, password_hash, auth_version FROM users WHERE email = $1`, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.IsActive, &passwordHash, &user.Version)
	return user, passwordHash, persistenceError(err)
}

func (s *PostgresRepository) InsertUser(ctx context.Context, name, email, passwordHash, role string) error {
	_, err := s.db.Exec(ctx, `INSERT INTO users (name,email,password_hash,role) VALUES ($1,$2,$3,$4) ON CONFLICT (email) DO NOTHING`, name, email, passwordHash, role)
	return persistenceError(err)
}
func (s *PostgresRepository) InsertRefresh(ctx context.Context, userID, hash string, expires time.Time) error {
	_, err := s.db.Exec(ctx, "INSERT INTO refresh_tokens (user_id,token_hash,expires_at) VALUES ($1,$2,$3)", userID, hash, expires)
	return persistenceError(err)
}
func (s *PostgresRepository) RevokeRefresh(ctx context.Context, hash string) error {
	_, err := s.db.Exec(ctx, "UPDATE refresh_tokens SET revoked_at=NOW() WHERE token_hash=$1 AND revoked_at IS NULL", hash)
	return persistenceError(err)
}
func (s *PostgresRepository) ActiveUser(ctx context.Context, id string, version int64) (User, error) {
	var user User
	err := s.db.QueryRow(ctx, "SELECT id,name,email,role,is_active FROM users WHERE id::text=$1 AND is_active AND auth_version=$2", id, version).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.IsActive)
	return user, persistenceError(err)
}
func (s *PostgresRepository) RotateRefresh(ctx context.Context, oldHash, newHash string, expires time.Time) (User, error) {
	var user User
	var tokenID string
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `SELECT rt.id, u.id, u.name, u.email, u.role, u.is_active, u.auth_version
		FROM refresh_tokens rt JOIN users u ON u.id = rt.user_id
		WHERE rt.token_hash = $1 AND rt.revoked_at IS NULL AND rt.expires_at > NOW() FOR UPDATE OF rt`, oldHash).Scan(
		&tokenID, &user.ID, &user.Name, &user.Email, &user.Role, &user.IsActive, &user.Version)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !user.IsActive) {
		return User{}, ErrUnauthorized
	}
	if err != nil {
		return User{}, err
	}
	command, err := tx.Exec(ctx, "UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL", tokenID)
	if err != nil {
		return User{}, err
	}
	if command.RowsAffected() != 1 {
		return User{}, ErrUnauthorized
	}
	if _, err = tx.Exec(ctx, "INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)", user.ID, newHash, expires); err != nil {
		return User{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return User{}, err
	}
	return user, nil
}
func (s *PostgresRepository) ConsumeReset(ctx context.Context, hash, passwordHash string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var userID string
	var tokenID string
	err = tx.QueryRow(ctx, `SELECT id, user_id FROM password_reset_tokens
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW() FOR UPDATE`, hash).Scan(&tokenID, &userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidResetToken
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "UPDATE users SET password_hash = $1, auth_version = auth_version + 1, updated_at = NOW() WHERE id = $2", passwordHash, userID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1", tokenID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL", userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *PostgresRepository) ReplaceReset(ctx context.Context, userID, hash string, expires time.Time) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer tx.Rollback(ctx)
	// Serialize replacement requests for the same account.
	var id string
	if err = tx.QueryRow(ctx, "SELECT id FROM users WHERE id=$1 FOR UPDATE", userID).Scan(&id); err != nil {
		return persistenceError(err)
	}
	if _, err = tx.Exec(ctx, "UPDATE password_reset_tokens SET used_at=NOW() WHERE user_id=$1 AND used_at IS NULL", userID); err != nil {
		return persistenceError(err)
	}
	if _, err = tx.Exec(ctx, "INSERT INTO password_reset_tokens (user_id,token_hash,expires_at) VALUES ($1,$2,$3)", userID, hash, expires); err != nil {
		return persistenceError(err)
	}
	return persistenceError(tx.Commit(ctx))
}
