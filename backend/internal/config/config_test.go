package config

import "testing"

func TestIsDevelopment(t *testing.T) {
	for _, environment := range []string{"development", "DEV", " dev "} {
		if !(&Config{AppEnv: environment}).IsDevelopment() {
			t.Fatalf("%q deveria ser reconhecido como desenvolvimento", environment)
		}
	}
	if (&Config{AppEnv: "production"}).IsDevelopment() {
		t.Fatal("produção foi reconhecida como desenvolvimento")
	}
}
