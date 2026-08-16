package auth

import "testing"

func TestValidateProductionConfigRejectsEscapeHatches(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "bypass", key: "ALLOW_BYPASS", value: "true"},
		{name: "dev token", key: "DEV_AUTH_TOKEN", value: "secret"},
		{name: "registration", key: "REGISTER_ENABLED", value: "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ENV", "production")
			t.Setenv("ALLOW_BYPASS", "false")
			t.Setenv("DEV_AUTH_TOKEN", "")
			t.Setenv("REGISTER_ENABLED", "false")
			t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
			t.Setenv("CORS_ORIGINS", "https://tracker.example.com")
			t.Setenv(tt.key, tt.value)
			if err := ValidateProductionConfig(); err == nil {
				t.Fatal("expected production config error")
			}
		})
	}
}

func TestValidateProductionConfigAllowsSafeProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("ALLOW_BYPASS", "false")
	t.Setenv("DEV_AUTH_TOKEN", "")
	t.Setenv("REGISTER_ENABLED", "false")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("CORS_ORIGINS", "https://tracker.example.com")
	if err := ValidateProductionConfig(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateProductionConfigRejectsWeakSecretsAndInsecureOrigins(t *testing.T) {
	tests := []struct {
		name       string
		jwtSecret  string
		corsOrigin string
	}{
		{name: "weak secret", jwtSecret: "too-short", corsOrigin: "https://tracker.example.com"},
		{name: "HTTP origin", jwtSecret: "0123456789abcdef0123456789abcdef", corsOrigin: "http://tracker.example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ENV", "production")
			t.Setenv("ALLOW_BYPASS", "false")
			t.Setenv("DEV_AUTH_TOKEN", "")
			t.Setenv("REGISTER_ENABLED", "false")
			t.Setenv("JWT_SECRET", tt.jwtSecret)
			t.Setenv("CORS_ORIGINS", tt.corsOrigin)
			if err := ValidateProductionConfig(); err == nil {
				t.Fatal("expected production config error")
			}
		})
	}
}
