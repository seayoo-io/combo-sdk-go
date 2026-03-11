package combo

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func newTestSigner() *httpSigner {
	return &httpSigner{
		game:       "test_game",
		signingKey: SecretKey("sk_test_secret"),
	}
}

func TestSignAndAuthHttp(t *testing.T) {
	signer := newTestSigner()
	body := []byte(`{"key":"value"}`)
	req, err := http.NewRequest(http.MethodPost, "https://example.com/v1/server/test-api", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	signingTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	if err := signer.SignHttp(req, signingTime); err != nil {
		t.Fatalf("SignHttp failed: %v", err)
	}

	authHeader := req.Header.Get(authorizationHeader)
	if authHeader == "" {
		t.Fatal("Authorization header not set")
	}
	if !strings.HasPrefix(authHeader, signingAlgorithm) {
		t.Fatalf("Authorization header should start with %s, got %s", signingAlgorithm, authHeader)
	}
	if !strings.Contains(authHeader, "Game=test_game") {
		t.Fatalf("Authorization header should contain Game=test_game, got %s", authHeader)
	}

	// AuthHttp should succeed within time window
	if err := signer.AuthHttp(req, signingTime); err != nil {
		t.Fatalf("AuthHttp failed: %v", err)
	}
}

func TestAuthHttpTimeDiffExceeded(t *testing.T) {
	signer := newTestSigner()
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/v1/server/test", bytes.NewBuffer([]byte(`{}`)))
	signingTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	_ = signer.SignHttp(req, signingTime)

	// Verify 6 minutes later should fail
	err := signer.AuthHttp(req, signingTime.Add(6*time.Minute))
	if err == nil {
		t.Fatal("expected error for time difference exceeding maximum")
	}
	if !strings.Contains(err.Error(), "time difference exceeds maximum") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthHttpWrongGame(t *testing.T) {
	signer := newTestSigner()
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/v1/server/test", bytes.NewBuffer([]byte(`{}`)))
	signingTime := time.Now()
	_ = signer.SignHttp(req, signingTime)

	otherSigner := &httpSigner{
		game:       "other_game",
		signingKey: SecretKey("sk_test_secret"),
	}
	err := otherSigner.AuthHttp(req, signingTime)
	if err == nil {
		t.Fatal("expected error for wrong game")
	}
	if !strings.Contains(err.Error(), "invalid game") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthHttpWrongSignature(t *testing.T) {
	signer := newTestSigner()
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/v1/server/test", bytes.NewBuffer([]byte(`{}`)))
	signingTime := time.Now()
	_ = signer.SignHttp(req, signingTime)

	otherSigner := &httpSigner{
		game:       "test_game",
		signingKey: SecretKey("sk_different_key"),
	}
	err := otherSigner.AuthHttp(req, signingTime)
	if err == nil {
		t.Fatal("expected error for wrong signature")
	}
	if !strings.Contains(err.Error(), "invalid signature") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthHttpMissingHeader(t *testing.T) {
	signer := newTestSigner()
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/test", nil)
	err := signer.AuthHttp(req, time.Now())
	if err == nil {
		t.Fatal("expected error for missing authorization header")
	}
}

func TestParseAuthorizationHeader(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		wantErr bool
	}{
		{
			name:    "empty header",
			header:  "",
			wantErr: true,
		},
		{
			name:    "no space separator",
			header:  "InvalidHeader",
			wantErr: true,
		},
		{
			name:   "valid header",
			header: "SEAYOO-HMAC-SHA256 Game=test,Timestamp=20240115T120000Z,Signature=abc123",
		},
		{
			name:    "invalid timestamp",
			header:  "SEAYOO-HMAC-SHA256 Game=test,Timestamp=not-a-timestamp,Signature=abc123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := parseAuthorizationHeader(tt.header)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if auth.scheme != signingAlgorithm {
				t.Fatalf("expected scheme %s, got %s", signingAlgorithm, auth.scheme)
			}
			if auth.game != "test" {
				t.Fatalf("expected game 'test', got %s", auth.game)
			}
		})
	}
}

func TestComputePayloadHashNilBody(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/test", nil)
	hash, err := computePayloadHash(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != emptyStringSha256 {
		t.Fatalf("expected empty string sha256, got %s", hash)
	}
}

func TestComputePayloadHashWithBody(t *testing.T) {
	body := []byte(`{"test":"data"}`)
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/test", bytes.NewBuffer(body))
	hash, err := computePayloadHash(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == emptyStringSha256 {
		t.Fatal("expected non-empty hash for body with content")
	}
	// Verify body is still readable after hash computation
	readBody, _ := io.ReadAll(req.Body)
	if !bytes.Equal(readBody, body) {
		t.Fatal("body should still be readable after hash computation")
	}
}

func TestGetTimestamp(t *testing.T) {
	ts := time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC)
	result := getTimestamp(ts)
	if result != "20240115T123045Z" {
		t.Fatalf("expected 20240115T123045Z, got %s", result)
	}
}

func TestBuildStringToSign(t *testing.T) {
	body := []byte(`{"test":"data"}`)
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/v1/server/test-api", bytes.NewBuffer(body))
	timestamp := "20240115T120000Z"
	result, err := buildStringToSign(req, timestamp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parts := strings.Split(result, "\n")
	if len(parts) != 5 {
		t.Fatalf("expected 5 parts, got %d", len(parts))
	}
	if parts[0] != signingAlgorithm {
		t.Fatalf("expected first part to be %s, got %s", signingAlgorithm, parts[0])
	}
	if parts[1] != http.MethodPost {
		t.Fatalf("expected second part to be POST, got %s", parts[1])
	}
	if parts[3] != timestamp {
		t.Fatalf("expected fourth part to be %s, got %s", timestamp, parts[3])
	}
}
