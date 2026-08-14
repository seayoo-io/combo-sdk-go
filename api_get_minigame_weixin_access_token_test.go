package combo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientGetMiniGameWeixinAccessToken(t *testing.T) {
	const appId = "wx1234567890abcdef"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/v1/server/minigame-weixin-access-token") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		signer := httpSigner{game: testGameId, signingKey: SecretKey(testSecretKey)}
		if err := signer.AuthHttp(r, time.Now()); err != nil {
			t.Errorf("signature verification failed: %v", err)
		}
		var input GetMiniGameWeixinAccessTokenInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if input.AppId != appId {
			t.Errorf("expected app_id %s, got %s", appId, input.AppId)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-trace-id", "trace_ak")
		json.NewEncoder(w).Encode(map[string]any{
			"app_id":       appId,
			"access_token": "STABLE_AK",
		})
	}))
	defer server.Close()

	cfg := Config{
		Endpoint:  Endpoint(server.URL),
		GameId:    testGameId,
		SecretKey: SecretKey(testSecretKey),
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	output, err := client.GetMiniGameWeixinAccessToken(context.Background(), &GetMiniGameWeixinAccessTokenInput{
		AppId: appId,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.AppId != appId {
		t.Fatalf("expected app_id %s, got %s", appId, output.AppId)
	}
	if output.AccessToken != "STABLE_AK" {
		t.Fatalf("expected access_token STABLE_AK, got %s", output.AccessToken)
	}
	if output.TraceId() != "trace_ak" {
		t.Fatalf("expected trace_id trace_ak, got %s", output.TraceId())
	}
}
