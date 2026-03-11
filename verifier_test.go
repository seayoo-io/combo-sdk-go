package combo

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testEndpoint  = Endpoint("https://api.test.com")
	testGameId    = GameId("test_game")
	testSecretKey = "sk_test_secret_key_12345"
)

func newTestConfig() Config {
	return Config{
		Endpoint:  testEndpoint,
		GameId:    testGameId,
		SecretKey: SecretKey(testSecretKey),
	}
}

func newTestVerifier(t *testing.T) *TokenVerifier {
	t.Helper()
	v, err := NewTokenVerifier(newTestConfig())
	if err != nil {
		t.Fatalf("failed to create verifier: %v", err)
	}
	return v
}

func signToken(t *testing.T, claims jwt.Claims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(testSecretKey))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenString
}

func TestNewTokenVerifierInvalidConfig(t *testing.T) {
	_, err := NewTokenVerifier(Config{})
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestVerifyIdentityToken(t *testing.T) {
	v := newTestVerifier(t)

	claims := &identityClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    string(testEndpoint),
			Subject:   "combo_123",
			Audience:  jwt.ClaimStrings{string(testGameId)},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Scope:         "auth",
		IdP:           "google",
		ExternalId:    "ext_123",
		ExternalName:  "TestUser",
		WeixinUnionid: "",
		DeviceId:      "device_abc",
		Distro:        "default",
		Variant:       "v1",
		Age:           25,
		RegTime:       1700000000,
	}

	tokenString := signToken(t, claims)
	payload, err := v.VerifyIdentityToken(tokenString)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.ComboId != "combo_123" {
		t.Errorf("expected ComboId combo_123, got %s", payload.ComboId)
	}
	if payload.IdP != "google" {
		t.Errorf("expected IdP google, got %s", payload.IdP)
	}
	if payload.ExternalId != "ext_123" {
		t.Errorf("expected ExternalId ext_123, got %s", payload.ExternalId)
	}
	if payload.ExternalName != "TestUser" {
		t.Errorf("expected ExternalName TestUser, got %s", payload.ExternalName)
	}
	if payload.DeviceId != "device_abc" {
		t.Errorf("expected DeviceId device_abc, got %s", payload.DeviceId)
	}
	if payload.Distro != "default" {
		t.Errorf("expected Distro default, got %s", payload.Distro)
	}
	if payload.Variant != "v1" {
		t.Errorf("expected Variant v1, got %s", payload.Variant)
	}
	if payload.Age != 25 {
		t.Errorf("expected Age 25, got %d", payload.Age)
	}
	if payload.RegTime != 1700000000 {
		t.Errorf("expected RegTime 1700000000, got %d", payload.RegTime)
	}
}

func TestVerifyIdentityTokenWrongScope(t *testing.T) {
	v := newTestVerifier(t)

	claims := &identityClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    string(testEndpoint),
			Subject:   "combo_123",
			Audience:  jwt.ClaimStrings{string(testGameId)},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Scope: "wrong_scope",
	}

	tokenString := signToken(t, claims)
	_, err := v.VerifyIdentityToken(tokenString)
	if err == nil {
		t.Fatal("expected error for wrong scope")
	}
	if !strings.Contains(err.Error(), "invalid scope") {
		t.Fatalf("expected scope error, got: %v", err)
	}
}

func TestVerifyIdentityTokenExpired(t *testing.T) {
	v := newTestVerifier(t)

	claims := &identityClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    string(testEndpoint),
			Subject:   "combo_123",
			Audience:  jwt.ClaimStrings{string(testGameId)},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
		Scope: "auth",
	}

	tokenString := signToken(t, claims)
	_, err := v.VerifyIdentityToken(tokenString)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestVerifyIdentityTokenWrongIssuer(t *testing.T) {
	v := newTestVerifier(t)

	claims := &identityClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "https://wrong.issuer.com",
			Subject:   "combo_123",
			Audience:  jwt.ClaimStrings{string(testGameId)},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Scope: "auth",
	}

	tokenString := signToken(t, claims)
	_, err := v.VerifyIdentityToken(tokenString)
	if err == nil {
		t.Fatal("expected error for wrong issuer")
	}
}

func TestVerifyIdentityTokenWrongAudience(t *testing.T) {
	v := newTestVerifier(t)

	claims := &identityClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    string(testEndpoint),
			Subject:   "combo_123",
			Audience:  jwt.ClaimStrings{"wrong_game"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Scope: "auth",
	}

	tokenString := signToken(t, claims)
	_, err := v.VerifyIdentityToken(tokenString)
	if err == nil {
		t.Fatal("expected error for wrong audience")
	}
}

func TestVerifyIdentityTokenWrongSigningKey(t *testing.T) {
	v := newTestVerifier(t)

	claims := &identityClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    string(testEndpoint),
			Subject:   "combo_123",
			Audience:  jwt.ClaimStrings{string(testGameId)},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Scope: "auth",
	}

	// Sign with a different key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("sk_wrong_key"))

	_, err := v.VerifyIdentityToken(tokenString)
	if err == nil {
		t.Fatal("expected error for wrong signing key")
	}
}

func TestVerifyIdentityTokenInvalidFormat(t *testing.T) {
	v := newTestVerifier(t)
	_, err := v.VerifyIdentityToken("not.a.valid.jwt")
	if err == nil {
		t.Fatal("expected error for invalid token format")
	}
}

func TestVerifyAdToken(t *testing.T) {
	v := newTestVerifier(t)

	claims := &adClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    string(testEndpoint),
			Subject:   "combo_456",
			Audience:  jwt.ClaimStrings{string(testGameId)},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Scope:        "ads",
		PlacementId:  "placement_001",
		ImpressionId: "impression_001",
	}

	tokenString := signToken(t, claims)
	payload, err := v.VerifyAdToken(tokenString)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payload.ComboId != "combo_456" {
		t.Errorf("expected ComboId combo_456, got %s", payload.ComboId)
	}
	if payload.PlacementId != "placement_001" {
		t.Errorf("expected PlacementId placement_001, got %s", payload.PlacementId)
	}
	if payload.ImpressionId != "impression_001" {
		t.Errorf("expected ImpressionId impression_001, got %s", payload.ImpressionId)
	}
}

func TestVerifyAdTokenWrongScope(t *testing.T) {
	v := newTestVerifier(t)

	claims := &adClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    string(testEndpoint),
			Subject:   "combo_456",
			Audience:  jwt.ClaimStrings{string(testGameId)},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Scope: "wrong",
	}

	tokenString := signToken(t, claims)
	_, err := v.VerifyAdToken(tokenString)
	if err == nil {
		t.Fatal("expected error for wrong scope")
	}
}
