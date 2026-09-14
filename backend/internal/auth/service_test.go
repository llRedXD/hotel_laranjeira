package auth

import (
	"testing"
	"time"
)

func TestJWTSignatureAndExpiration(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Unix(1_700_000_000, 0)
	token, err := signJWT(jwtClaims{
		Subject: "9c9b961b-6996-4a3d-8070-4e31a7efb9c7",
		Email:   "admin@hotel.com",
		Role:    "admin",
		Expires: now.Add(time.Minute).Unix(),
	}, secret)
	if err != nil {
		t.Fatalf("signJWT retornou erro: %v", err)
	}
	claims, err := verifyJWT(token, secret, now)
	if err != nil {
		t.Fatalf("verifyJWT retornou erro: %v", err)
	}
	if claims.Subject != "9c9b961b-6996-4a3d-8070-4e31a7efb9c7" {
		t.Fatalf("subject inesperado: %q", claims.Subject)
	}
	if _, err := verifyJWT(token, []byte("outro-segredo"), now); err == nil {
		t.Fatal("token com assinatura inválida foi aceito")
	}
	if _, err := verifyJWT(token, secret, now.Add(2*time.Minute)); err == nil {
		t.Fatal("token expirado foi aceito")
	}
}

func TestPasswordPolicy(t *testing.T) {
	if err := validatePassword("curta"); err == nil {
		t.Fatal("senha curta foi aceita")
	}
	if err := validatePassword("senha-valida-12"); err != nil {
		t.Fatalf("senha válida foi rejeitada: %v", err)
	}
}

func TestOpaqueTokensAreUniqueAndHashable(t *testing.T) {
	first, err := newOpaqueToken()
	if err != nil {
		t.Fatalf("primeiro token: %v", err)
	}
	second, err := newOpaqueToken()
	if err != nil {
		t.Fatalf("segundo token: %v", err)
	}
	if first == second {
		t.Fatal("tokens aleatórios coincidiram")
	}
	if len(tokenHash(first)) != 64 {
		t.Fatal("hash do token não tem 64 caracteres")
	}
}
