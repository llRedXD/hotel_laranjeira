package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/hotel-laranjeira/backend/internal/auth"
	"github.com/hotel-laranjeira/backend/internal/config"
	"github.com/hotel-laranjeira/backend/internal/hotel"
	"github.com/jackc/pgx/v5/pgxpool"
)

const refreshCookieName = "hotel_laranjeira_refresh"

type api struct {
	auth           *auth.Service
	hotel          *hotel.Service
	frontendOrigin string
	isDevelopment  bool
}

func NewRouter(db *pgxpool.Pool, cfg config.Config) http.Handler {
	api := &api{auth: auth.NewService(auth.NewPostgresRepository(db), cfg.JWTSecret), hotel: hotel.NewService(hotel.NewPostgresRepository(db)), frontendOrigin: cfg.FrontendOrigin, isDevelopment: cfg.IsDevelopment()}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", healthHandler(db))
	mux.HandleFunc("GET /api/v1/public/availability", api.publicAvailability)
	mux.HandleFunc("POST /api/v1/public/reservations", api.createPublicReservation)
	mux.HandleFunc("GET /api/v1/reservations/availability", api.staffAvailability)
	mux.HandleFunc("POST /api/v1/reservations", api.createStaffReservation)
	mux.HandleFunc("POST /api/v1/auth/login", api.login)
	mux.HandleFunc("POST /api/v1/auth/refresh", api.refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", api.logout)
	mux.HandleFunc("GET /api/v1/auth/me", api.me)
	mux.HandleFunc("POST /api/v1/auth/password-recovery", api.passwordRecovery)
	mux.HandleFunc("POST /api/v1/auth/password-reset", api.passwordReset)
	mux.HandleFunc("POST /api/v1/guests", api.createGuest)
	mux.HandleFunc("GET /api/v1/guests", api.listGuests)
	mux.HandleFunc("GET /api/v1/guests/{id}", api.getGuest)
	mux.HandleFunc("PATCH /api/v1/guests/{id}", api.updateGuest)
	mux.HandleFunc("POST /api/v1/room-categories", api.createCategory)
	mux.HandleFunc("GET /api/v1/room-categories", api.listCategories)
	mux.HandleFunc("PATCH /api/v1/room-categories/{id}", api.updateCategory)
	mux.HandleFunc("GET /api/v1/rooms", api.listRooms)
	mux.HandleFunc("GET /api/v1/rooms/{id}", api.getRoom)
	mux.HandleFunc("POST /api/v1/rooms", api.saveRoom)
	mux.HandleFunc("PATCH /api/v1/rooms/{id}", api.saveRoom)
	mux.HandleFunc("PATCH /api/v1/rooms/{id}/availability", api.roomAvailability)
	return cors(security(mux), cfg.FrontendOrigin)
}

func (a *api) requireAuth(w http.ResponseWriter, r *http.Request) bool {
	if _, err := a.auth.UserFromAccessToken(r.Context(), bearerToken(r)); err != nil {
		if !errors.Is(err, auth.ErrUnauthorized) {
			internalError(w, err)
			return false
		}
		writeError(w, http.StatusUnauthorized, "unauthorized", "Token de acesso inválido ou expirado.", nil)
		return false
	}
	return true
}

func isUUID(value string) bool {
	if len(value) != 36 {
		return false
	}
	for index, char := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			if char != '-' {
				return false
			}
			continue
		}
		if !(char >= '0' && char <= '9') && !(char >= 'a' && char <= 'f') && !(char >= 'A' && char <= 'F') {
			return false
		}
	}
	return true
}

func pagination(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	limit, offset := 50, 0
	var err error
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			writeValidationError(w, map[string]string{"limit": "Use um inteiro entre 1 e 100."})
			return 0, 0, false
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 || offset > 100000 {
			writeValidationError(w, map[string]string{"offset": "Use um inteiro entre 0 e 100000."})
			return 0, 0, false
		}
	}
	return limit, offset, true
}

func healthHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := db.Exec(r.Context(), "SELECT 1"); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "indisponível", "database": "indisponível"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "conectado"})
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *api) login(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if fields := validateCredentials(input); len(fields) > 0 {
		writeValidationError(w, fields)
		return
	}
	user, accessToken, refreshToken, err := a.auth.Login(r.Context(), input.Email, input.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "E-mail ou senha inválidos.", nil)
		return
	}
	if err != nil {
		internalError(w, err)
		return
	}
	a.setRefreshCookie(w, refreshToken)
	writeJSON(w, http.StatusOK, map[string]any{"user": user, "accessToken": accessToken})
}

func (a *api) refresh(w http.ResponseWriter, r *http.Request) {
	user, accessToken, refreshToken, err := a.auth.Refresh(r.Context(), a.refreshCookie(r))
	if errors.Is(err, auth.ErrUnauthorized) {
		a.clearRefreshCookie(w)
		writeError(w, http.StatusUnauthorized, "unauthorized", "Sessão inválida ou expirada.", nil)
		return
	}
	if err != nil {
		internalError(w, err)
		return
	}
	a.setRefreshCookie(w, refreshToken)
	writeJSON(w, http.StatusOK, map[string]any{"user": user, "accessToken": accessToken})
}

func (a *api) logout(w http.ResponseWriter, r *http.Request) {
	if err := a.auth.Logout(r.Context(), a.refreshCookie(r)); err != nil {
		internalError(w, err)
		return
	}
	a.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) me(w http.ResponseWriter, r *http.Request) {
	user, err := a.auth.UserFromAccessToken(r.Context(), bearerToken(r))
	if err != nil {
		if !errors.Is(err, auth.ErrUnauthorized) {
			internalError(w, err)
			return
		}
		writeError(w, http.StatusUnauthorized, "unauthorized", "Token de acesso inválido ou expirado.", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

type recoveryRequest struct {
	Email string `json:"email"`
}

func (a *api) passwordRecovery(w http.ResponseWriter, r *http.Request) {
	var input recoveryRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if !isValidEmail(input.Email) {
		writeValidationError(w, map[string]string{"email": "Informe um e-mail válido."})
		return
	}
	token, err := a.auth.RequestPasswordReset(r.Context(), input.Email)
	if err != nil {
		internalError(w, err)
		return
	}
	if token != "" && a.isDevelopment {
		log.Printf("recuperação de senha solicitada: %s/redefinir-senha?token=%s", a.frontendOrigin, token)
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"message": "Se houver uma conta para este endereço, você receberá instruções para redefinir a senha."})
}

type resetRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (a *api) passwordReset(w http.ResponseWriter, r *http.Request) {
	var input resetRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	fields := map[string]string{}
	if strings.TrimSpace(input.Token) == "" {
		fields["token"] = "Token de redefinição é obrigatório."
	}
	if len(input.Password) < 12 {
		fields["password"] = "Use uma senha com pelo menos 12 caracteres."
	} else if len(input.Password) > 72 {
		fields["password"] = "Use uma senha com no máximo 72 caracteres."
	}
	if len(fields) > 0 {
		writeValidationError(w, fields)
		return
	}
	err := a.auth.ResetPassword(r.Context(), input.Token, input.Password)
	if errors.Is(err, auth.ErrInvalidResetToken) {
		writeError(w, http.StatusBadRequest, "invalid_reset_token", "Token de redefinição inválido ou expirado.", nil)
		return
	}
	if err != nil {
		internalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validateCredentials(input loginRequest) map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(input.Email) == "" {
		fields["email"] = "Informe seu e-mail."
	} else if !isValidEmail(input.Email) {
		fields["email"] = "Informe um e-mail válido."
	}
	if input.Password == "" {
		fields["password"] = "Informe sua senha."
	}
	if len(input.Password) > 72 {
		fields["password"] = "Senha excede 72 bytes."
	}
	return fields
}

func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return false
	}
	if len(email) < 3 || len(email) > 255 || strings.Count(email, "@") != 1 {
		return false
	}
	parts := strings.Split(email, "@")
	return parts[0] != "" && strings.Contains(parts[1], ".") && !strings.HasPrefix(parts[1], ".") && !strings.HasSuffix(parts[1], ".")
}

func (a *api) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{Name: refreshCookieName, Value: token, Path: "/api/v1/auth", MaxAge: int((30 * 24 * time.Hour).Seconds()), HttpOnly: true, Secure: !a.isDevelopment, SameSite: http.SameSiteLaxMode})
}

func (a *api) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: refreshCookieName, Value: "", Path: "/api/v1/auth", MaxAge: -1, HttpOnly: true, Secure: !a.isDevelopment, SameSite: http.SameSiteLaxMode})
}

func (a *api) refreshCookie(r *http.Request) string {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func bearerToken(r *http.Request) string {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

func cors(next http.Handler, frontendOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Add("Vary", "Origin")
		if origin := r.Header.Get("Origin"); origin != "" && origin != frontendOrigin {
			writeError(w, http.StatusForbidden, "origin_forbidden", "Origem não permitida.", nil)
			return
		}
		if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			writeError(w, http.StatusForbidden, "origin_forbidden", "Origem não permitida.", nil)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Use Content-Type: application/json.", nil)
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Corpo da requisição inválido.", nil)
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid_json", "Corpo da requisição inválido.", nil)
		return false
	}
	return true
}

func writeValidationError(w http.ResponseWriter, fields map[string]string) {
	writeError(w, http.StatusUnprocessableEntity, "validation_error", "Revise os campos informados.", fields)
}

func writeError(w http.ResponseWriter, status int, code, message string, fields map[string]string) {
	errorPayload := map[string]any{"code": code, "message": message}
	if fields != nil {
		errorPayload["fields"] = fields
	}
	writeJSON(w, status, map[string]any{"error": errorPayload})
}

func internalError(w http.ResponseWriter, err error) {
	log.Printf("erro interno: %v", err)
	writeError(w, http.StatusInternalServerError, "internal_error", "Não foi possível concluir sua solicitação.", nil)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("erro ao escrever resposta JSON: %v", err)
	}
}
