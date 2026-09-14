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
	"net/mail"
	"strings"
	"time"

	"github.com/hotel-laranjeira/backend/internal/config"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenLifetime  = 15 * time.Minute
	refreshTokenLifetime = 30 * 24 * time.Hour
	resetTokenLifetime   = time.Hour
)

var (
	ErrNotFound           = errors.New("usuário não encontrado")
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrUnauthorized       = errors.New("não autorizado")
	ErrInvalidResetToken  = errors.New("token de redefinição inválido ou expirado")
)

type User struct {
	Version  int64  `json:"-"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"-"`
}

type Service struct {
	repository Repository
	jwtSecret  []byte
	now        func() time.Time
}

func NewService(repository Repository, jwtSecret string) *Service {
	return &Service{repository: repository, jwtSecret: []byte(jwtSecret), now: time.Now}
}

func BootstrapUsers(ctx context.Context, repository Repository, cfg config.Config) error {
	if !cfg.IsDevelopment() {
		return nil
	}
	if err := bootstrapUser(ctx, repository, cfg.InitialAdminName, cfg.InitialAdminEmail, cfg.InitialAdminPassword, "admin", "INITIAL_ADMIN"); err != nil {
		return err
	}
	return bootstrapUser(ctx, repository, cfg.InitialReceptionName, cfg.InitialReceptionEmail, cfg.InitialReceptionPassword, "reception", "INITIAL_RECEPTION")
}

func bootstrapUser(ctx context.Context, repository Repository, name, email, password, role, variablePrefix string) error {
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
	err = repository.InsertUser(ctx, strings.TrimSpace(name), normalizeEmail(email), string(hash), role)
	if err != nil {
		return fmt.Errorf("criar usuário inicial %s: %w", role, err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, email, password string) (User, string, string, error) {
	user, passwordHash, err := s.userByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return User{}, "", "", err
	}
	if errors.Is(err, ErrNotFound) {
		// Valid cost-10 dummy hash keeps unknown accounts on the password verification path.
		passwordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	}
	passwordErr := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil || !user.IsActive || passwordErr != nil {
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
	next, err := newOpaqueToken()
	if err != nil {
		return User{}, "", "", err
	}
	user, err := s.repository.RotateRefresh(ctx, tokenHash(rawToken), tokenHash(next), s.now().Add(refreshTokenLifetime))
	if err != nil {
		return User{}, "", "", err
	}
	access, err := s.newAccessToken(user)
	if err != nil {
		return User{}, "", "", err
	}
	return user, access, next, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	return s.repository.RevokeRefresh(ctx, tokenHash(rawToken))
}

func (s *Service) UserFromAccessToken(ctx context.Context, token string) (User, error) {
	claims, err := verifyJWT(token, s.jwtSecret, s.now())
	if err != nil {
		return User{}, ErrUnauthorized
	}
	user, err := s.repository.ActiveUser(ctx, claims.Subject, claims.Version)
	if errors.Is(err, ErrNotFound) {
		return User{}, ErrUnauthorized
	}
	return user, err
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	user, _, err := s.repository.UserByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, ErrNotFound) || (err == nil && !user.IsActive) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	token, err := newOpaqueToken()
	if err != nil {
		return "", err
	}
	if err = s.repository.ReplaceReset(ctx, user.ID, tokenHash(token), s.now().Add(resetTokenLifetime)); err != nil {
		return "", err
	}
	return token, nil
}

func (s *Service) ResetPassword(ctx context.Context, rawToken, password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repository.ConsumeReset(ctx, tokenHash(rawToken), string(hash))
}

func (s *Service) userByEmail(ctx context.Context, email string) (User, string, error) {
	return s.repository.UserByEmail(ctx, normalizeEmail(email))
}

func (s *Service) createRefreshToken(ctx context.Context, userID string) (string, error) {
	token, err := newOpaqueToken()
	if err != nil {
		return "", err
	}
	return token, s.repository.InsertRefresh(ctx, userID, tokenHash(token), s.now().Add(refreshTokenLifetime))
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return errors.New("informe um e-mail válido")
	}
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
	Version int64  `json:"ver"`
	Subject string `json:"sub"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	Issued  int64  `json:"iat"`
	Expires int64  `json:"exp"`
}

func (s *Service) newAccessToken(user User) (string, error) {
	now := s.now()
	return signJWT(jwtClaims{Version: user.Version, Subject: user.ID, Name: user.Name, Email: user.Email, Role: user.Role, Issued: now.Unix(), Expires: now.Add(accessTokenLifetime).Unix()}, s.jwtSecret)
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
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || string(header) != `{"alg":"HS256","typ":"JWT"}` {
		return jwtClaims{}, ErrUnauthorized
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
