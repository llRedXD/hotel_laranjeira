package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hotel-laranjeira/backend/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenLifetime  = 15 * time.Minute
	refreshTokenLifetime = 30 * 24 * time.Hour
	resetTokenLifetime   = time.Hour
)

var (
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrUnauthorized       = errors.New("não autorizado")
	ErrInvalidResetToken  = errors.New("token de redefinição inválido ou expirado")
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"-"`
}

type Service struct {
	db        *pgxpool.Pool
	jwtSecret []byte
	now       func() time.Time
}

func NewService(db *pgxpool.Pool, jwtSecret string) *Service {
	return &Service{db: db, jwtSecret: []byte(jwtSecret), now: time.Now}
}

func BootstrapUsers(ctx context.Context, db *pgxpool.Pool, cfg config.Config) error {
	if !cfg.IsDevelopment() {
		return nil
	}
	if err := bootstrapUser(ctx, db, cfg.InitialAdminName, cfg.InitialAdminEmail, cfg.InitialAdminPassword, "admin", "INITIAL_ADMIN"); err != nil {
		return err
	}
	return bootstrapUser(ctx, db, cfg.InitialReceptionName, cfg.InitialReceptionEmail, cfg.InitialReceptionPassword, "reception", "INITIAL_RECEPTION")
}

func bootstrapUser(ctx context.Context, db *pgxpool.Pool, name, email, password, role, variablePrefix string) error {
	if name == "" && email == "" && password == "" {
		return nil
	}
	if name == "" || email == "" || password == "" {
		return fmt.Errorf("%s_NAME, %s_EMAIL e %s_PASSWORD devem ser configurados juntos", variablePrefix, variablePrefix, variablePrefix)
	}
	if err := validateEmail(email); err != nil {
		return fmt.Errorf("e-mail de %s: %w", strings.ToLower(role), err)
	}
	if err := validatePassword(password); err != nil {
		return fmt.Errorf("senha de %s: %w", strings.ToLower(role), err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("gerar hash da senha: %w", err)
	}
	_, err = db.Exec(ctx, `INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4) ON CONFLICT (email) DO NOTHING`, strings.TrimSpace(name), normalizeEmail(email), string(hash), role)
	if err != nil {
		return fmt.Errorf("criar usuário inicial %s: %w", role, err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, email, password string) (User, string, string, error) {
	user, passwordHash, err := s.userByEmail(ctx, email)
	if err != nil || !user.IsActive || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		return User{}, "", "", ErrInvalidCredentials
	}
	accessToken, err := s.newAccessToken(user)
	if err != nil {
		return User{}, "", "", err
	}
	refreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return User{}, "", "", err
	}
	return user, accessToken, refreshToken, nil
}

func (s *Service) Refresh(ctx context.Context, rawToken string) (User, string, string, error) {
	if rawToken == "" {
		return User{}, "", "", ErrUnauthorized
	}
	var user User
	var tokenID string
	err := s.db.QueryRow(ctx, `SELECT rt.id, u.id, u.name, u.email, u.role, u.is_active
		FROM refresh_tokens rt JOIN users u ON u.id = rt.user_id
		WHERE rt.token_hash = $1 AND rt.revoked_at IS NULL AND rt.expires_at > NOW()`, tokenHash(rawToken)).Scan(
		&tokenID, &user.ID, &user.Name, &user.Email, &user.Role, &user.IsActive)
	if err != nil || !user.IsActive {
		return User{}, "", "", ErrUnauthorized
	}
	command, err := s.db.Exec(ctx, "UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL", tokenID)
	if err != nil || command.RowsAffected() != 1 {
		return User{}, "", "", ErrUnauthorized
	}
	accessToken, err := s.newAccessToken(user)
	if err != nil {
		return User{}, "", "", err
	}
	newRefreshToken, err := s.createRefreshToken(ctx, user.ID)
	if err != nil {
		return User{}, "", "", err
	}
	return user, accessToken, newRefreshToken, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	_, err := s.db.Exec(ctx, "UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL", tokenHash(rawToken))
	return err
}

func (s *Service) UserFromAccessToken(token string) (User, error) {
	claims, err := verifyJWT(token, s.jwtSecret, s.now())
	if err != nil {
		return User{}, ErrUnauthorized
	}
	return User{ID: claims.Subject, Name: claims.Name, Email: claims.Email, Role: claims.Role, IsActive: true}, nil
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	user, _, err := s.userByEmail(ctx, email)
	if err != nil || !user.IsActive {
		return "", nil
	}
	_, err = s.db.Exec(ctx, "UPDATE password_reset_tokens SET used_at = NOW() WHERE user_id = $1 AND used_at IS NULL", user.ID)
	if err != nil {
		return "", err
	}
	token, err := newOpaqueToken()
	if err != nil {
		return "", err
	}
	_, err = s.db.Exec(ctx, `INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`, user.ID, tokenHash(token), s.now().Add(resetTokenLifetime))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *Service) ResetPassword(ctx context.Context, rawToken, password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var userID string
	var tokenID string
	err = tx.QueryRow(ctx, `SELECT id, user_id FROM password_reset_tokens
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW() FOR UPDATE`, tokenHash(rawToken)).Scan(&tokenID, &userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidResetToken
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2", string(passwordHash), userID); err != nil {
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

func (s *Service) userByEmail(ctx context.Context, email string) (User, string, error) {
	var user User
	var passwordHash string
	err := s.db.QueryRow(ctx, `SELECT id, name, email, role, is_active, password_hash FROM users WHERE email = $1`, normalizeEmail(email)).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.IsActive, &passwordHash)
	return user, passwordHash, err
}

func (s *Service) createRefreshToken(ctx context.Context, userID string) (string, error) {
	token, err := newOpaqueToken()
	if err != nil {
		return "", err
	}
	_, err = s.db.Exec(ctx, `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`, userID, tokenHash(token), s.now().Add(refreshTokenLifetime))
	return token, err
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if len(email) < 3 || len(email) > 255 || strings.Count(email, "@") != 1 {
		return errors.New("informe um e-mail válido")
	}
	parts := strings.Split(email, "@")
	if parts[0] == "" || !strings.Contains(parts[1], ".") || strings.HasPrefix(parts[1], ".") || strings.HasSuffix(parts[1], ".") {
		return errors.New("informe um e-mail válido")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 12 {
		return errors.New("a senha deve ter pelo menos 12 caracteres")
	}
	if len(password) > 72 {
		return errors.New("a senha deve ter no máximo 72 caracteres")
	}
	return nil
}

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func newOpaqueToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

type jwtClaims struct {
	Subject string `json:"sub"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	Issued  int64  `json:"iat"`
	Expires int64  `json:"exp"`
}

func (s *Service) newAccessToken(user User) (string, error) {
	now := s.now()
	return signJWT(jwtClaims{Subject: user.ID, Name: user.Name, Email: user.Email, Role: user.Role, Issued: now.Unix(), Expires: now.Add(accessTokenLifetime).Unix()}, s.jwtSecret)
}

func signJWT(claims jwtClaims, secret []byte) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	unsigned := header + "." + encodedPayload
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func verifyJWT(token string, secret []byte, now time.Time) (jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return jwtClaims{}, errors.New("jwt inválido")
	}
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature, mac.Sum(nil)) {
		return jwtClaims{}, errors.New("assinatura inválida")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return jwtClaims{}, err
	}
	var claims jwtClaims
	if err = json.Unmarshal(payload, &claims); err != nil || claims.Subject == "" || claims.Expires <= now.Unix() {
		return jwtClaims{}, errors.New("jwt expirado ou inválido")
	}
	return claims, nil
}
