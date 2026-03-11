package combo

import (
	"testing"
)

func TestEndpointUrl(t *testing.T) {
	tests := []struct {
		endpoint Endpoint
		api      string
		want     string
	}{
		{Endpoint_China, "create-order", "https://api.seayoo.com/v1/server/create-order"},
		{Endpoint_Global, "enter-game", "https://api.seayoo.io/v1/server/enter-game"},
	}

	for _, tt := range tests {
		got := tt.endpoint.url(tt.api)
		if got != tt.want {
			t.Errorf("Endpoint(%s).url(%s) = %s, want %s", tt.endpoint, tt.api, got, tt.want)
		}
	}
}

func TestSecretKeyHmacSha256(t *testing.T) {
	sk := SecretKey("sk_test")
	result := sk.hmacSha256([]byte("hello"))
	if len(result) == 0 {
		t.Fatal("expected non-empty HMAC result")
	}

	// Same input should produce same output
	result2 := sk.hmacSha256([]byte("hello"))
	if len(result) != len(result2) {
		t.Fatal("expected same length for same input")
	}
	for i := range result {
		if result[i] != result2[i] {
			t.Fatal("expected same HMAC for same input")
		}
	}

	// Different input should produce different output
	result3 := sk.hmacSha256([]byte("world"))
	same := true
	for i := range result {
		if result[i] != result3[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("expected different HMAC for different input")
	}
}

func TestEndpointConstants(t *testing.T) {
	if Endpoint_China != "https://api.seayoo.com" {
		t.Fatalf("unexpected China endpoint: %s", Endpoint_China)
	}
	if Endpoint_Global != "https://api.seayoo.io" {
		t.Fatalf("unexpected Global endpoint: %s", Endpoint_Global)
	}
}

func TestPlatformConstants(t *testing.T) {
	platforms := map[Platform]string{
		Platform_iOS:       "ios",
		Platform_Android:   "android",
		Platform_Windows:   "windows",
		Platform_macOS:     "macos",
		Platform_WebGL:     "webgl",
		Platform_HarmonyOS: "harmonyos",
	}
	for p, expected := range platforms {
		if string(p) != expected {
			t.Errorf("expected %s, got %s", expected, p)
		}
	}
}
