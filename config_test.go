package combo

import (
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name:    "missing endpoint",
			cfg:     Config{GameId: "test", SecretKey: SecretKey("sk_test")},
			wantErr: "missing required Endpoint",
		},
		{
			name:    "missing game id",
			cfg:     Config{Endpoint: Endpoint_China, SecretKey: SecretKey("sk_test")},
			wantErr: "missing required GameId",
		},
		{
			name:    "missing secret key",
			cfg:     Config{Endpoint: Endpoint_China, GameId: "test"},
			wantErr: "missing required SecretKey",
		},
		{
			name:    "invalid secret key prefix",
			cfg:     Config{Endpoint: Endpoint_China, GameId: "test", SecretKey: SecretKey("bad_key")},
			wantErr: "invalid SecretKey: must start with sk_",
		},
		{
			name: "valid config",
			cfg:  Config{Endpoint: Endpoint_China, GameId: "test", SecretKey: SecretKey("sk_test")},
		},
		{
			name: "trims trailing slash from endpoint",
			cfg:  Config{Endpoint: "https://api.seayoo.com/", GameId: "test", SecretKey: SecretKey("sk_test")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestConfigValidateTrimsTrailingSlash(t *testing.T) {
	cfg := Config{
		Endpoint:  "https://api.seayoo.com/",
		GameId:    "test",
		SecretKey: SecretKey("sk_test"),
	}
	if err := cfg.validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Endpoint != "https://api.seayoo.com" {
		t.Fatalf("expected endpoint without trailing slash, got %q", cfg.Endpoint)
	}
}
