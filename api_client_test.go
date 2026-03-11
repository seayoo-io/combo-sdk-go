package combo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClientInvalidConfig(t *testing.T) {
	_, err := NewClient(Config{})
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestNewClientValid(t *testing.T) {
	client, err := NewClient(newTestConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.client != http.DefaultClient {
		t.Fatal("expected default http client when no option provided")
	}
}

func TestNewClientWithHttpClient(t *testing.T) {
	customClient := &http.Client{}
	client, err := NewClient(newTestConfig(), WithHttpClient(customClient))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.client != customClient {
		t.Fatal("expected custom http client")
	}
}

func TestClientUserAgent(t *testing.T) {
	client, err := NewClient(newTestConfig())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(client.userAgent, SdkName) {
		t.Fatalf("user agent should contain SDK name, got %s", client.userAgent)
	}
	if !strings.Contains(client.userAgent, SdkVersion) {
		t.Fatalf("user agent should contain SDK version, got %s", client.userAgent)
	}
	if !strings.Contains(client.userAgent, string(testGameId)) {
		t.Fatalf("user agent should contain game ID, got %s", client.userAgent)
	}
}

func TestClientCreateOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/v1/server/create-order") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-trace-id", "trace_001")
		json.NewEncoder(w).Encode(map[string]any{
			"order_id":    "order_123",
			"order_token": "token_abc",
			"expires_at":  1700000000,
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

	output, err := client.CreateOrder(context.Background(), &CreateOrderInput{
		Platform:    Platform_iOS,
		ReferenceId: "ref_001",
		ComboId:     "combo_001",
		ProductId:   "product_001",
		Quantity:    1,
		NotifyUrl:   "https://example.com/notify",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.OrderId != "order_123" {
		t.Fatalf("expected order_id order_123, got %s", output.OrderId)
	}
	if output.OrderToken != "token_abc" {
		t.Fatalf("expected order_token token_abc, got %s", output.OrderToken)
	}
	if output.TraceId() != "trace_001" {
		t.Fatalf("expected trace_id trace_001, got %s", output.TraceId())
	}
}

func TestClientCreateOrderDefaultQuantity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input CreateOrderInput
		json.NewDecoder(r.Body).Decode(&input)
		if input.Quantity != 1 {
			t.Errorf("expected quantity 1 (default), got %d", input.Quantity)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"order_id":    "order_123",
			"order_token": "token_abc",
			"expires_at":  1700000000,
		})
	}))
	defer server.Close()

	cfg := Config{
		Endpoint:  Endpoint(server.URL),
		GameId:    testGameId,
		SecretKey: SecretKey(testSecretKey),
	}
	client, _ := NewClient(cfg)
	client.CreateOrder(context.Background(), &CreateOrderInput{
		Platform:    Platform_iOS,
		ReferenceId: "ref_001",
		ComboId:     "combo_001",
		ProductId:   "product_001",
		Quantity:    0, // should default to 1
		NotifyUrl:   "https://example.com/notify",
	})
}

func TestClientErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("x-trace-id", "trace_err")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_request",
			"message": "missing combo_id",
		})
	}))
	defer server.Close()

	cfg := Config{
		Endpoint:  Endpoint(server.URL),
		GameId:    testGameId,
		SecretKey: SecretKey(testSecretKey),
	}
	client, _ := NewClient(cfg)

	_, err := client.CreateOrder(context.Background(), &CreateOrderInput{})
	if err == nil {
		t.Fatal("expected error")
	}

	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected ErrorResponse, got %T: %v", err, err)
	}
	if errResp.ErrorCode != "invalid_request" {
		t.Fatalf("expected error code invalid_request, got %s", errResp.ErrorCode)
	}
	if errResp.StatusCode() != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", errResp.StatusCode())
	}
	if errResp.TraceId() != "trace_err" {
		t.Fatalf("expected trace_id trace_err, got %s", errResp.TraceId())
	}
}

func TestClientEnterGame(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v1/server/enter-game") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer server.Close()

	cfg := Config{
		Endpoint:  Endpoint(server.URL),
		GameId:    testGameId,
		SecretKey: SecretKey(testSecretKey),
	}
	client, _ := NewClient(cfg)

	output, err := client.EnterGame(context.Background(), &EnterGameInput{
		ComboId:   "combo_001",
		SessionId: "session_001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output == nil {
		t.Fatal("expected non-nil output")
	}
}

func TestClientLeaveGame(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v1/server/leave-game") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{})
	}))
	defer server.Close()

	cfg := Config{
		Endpoint:  Endpoint(server.URL),
		GameId:    testGameId,
		SecretKey: SecretKey(testSecretKey),
	}
	client, _ := NewClient(cfg)

	output, err := client.LeaveGame(context.Background(), &LeaveGameInput{
		ComboId:   "combo_001",
		SessionId: "session_001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output == nil {
		t.Fatal("expected non-nil output")
	}
}
