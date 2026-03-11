package combo

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockGmListener struct {
	resp any
	err  *GmErrorResponse
}

func (m *mockGmListener) HandleGmRequest(_ context.Context, req *GmRequest) (any, *GmErrorResponse) {
	return m.resp, m.err
}

func newTestGmHandler(t *testing.T, listener GmListener) (http.Handler, *httpSigner) {
	t.Helper()
	cfg := newTestConfig()
	handler, err := NewGmHandler(cfg, listener)
	if err != nil {
		t.Fatalf("failed to create gm handler: %v", err)
	}
	signer := &httpSigner{
		game:       cfg.GameId,
		signingKey: cfg.SecretKey,
	}
	return handler, signer
}

func signedGmRequest(t *testing.T, signer *httpSigner, body []byte) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/gm", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if err := signer.SignHttp(req, time.Now()); err != nil {
		t.Fatalf("failed to sign request: %v", err)
	}
	return req
}

func validGmBody(t *testing.T) []byte {
	t.Helper()
	body := gmRequestBody{
		Version:   "2.0",
		Origin:    "test",
		RequestId: "req_001",
		Command:   "ListRoles",
		Args:      json.RawMessage(`{"combo_id":"123"}`),
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestNewGmHandlerInvalidConfig(t *testing.T) {
	_, err := NewGmHandler(Config{}, &mockGmListener{})
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestNewGmHandlerNilListener(t *testing.T) {
	_, err := NewGmHandler(newTestConfig(), nil)
	if err == nil {
		t.Fatal("expected error for nil listener")
	}
}

func TestGmHandlerMethodNotAllowed(t *testing.T) {
	handler, _ := newTestGmHandler(t, &mockGmListener{resp: map[string]string{"ok": "true"}})

	req := httptest.NewRequest(http.MethodGet, "/gm", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
	var resp GmErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Error != GmError_InvalidHttpMethod {
		t.Fatalf("expected error %s, got %s", GmError_InvalidHttpMethod, resp.Error)
	}
}

func TestGmHandlerWrongContentType(t *testing.T) {
	handler, _ := newTestGmHandler(t, &mockGmListener{resp: map[string]string{"ok": "true"}})

	req := httptest.NewRequest(http.MethodPost, "/gm", nil)
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected %d, got %d", http.StatusUnsupportedMediaType, rec.Code)
	}
}

func TestGmHandlerUnauthorized(t *testing.T) {
	handler, _ := newTestGmHandler(t, &mockGmListener{resp: map[string]string{"ok": "true"}})

	req := httptest.NewRequest(http.MethodPost, "/gm", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestGmHandlerInvalidJSON(t *testing.T) {
	handler, signer := newTestGmHandler(t, &mockGmListener{resp: map[string]string{"ok": "true"}})

	req := signedGmRequest(t, signer, []byte(`not json`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGmHandlerMissingRequiredFields(t *testing.T) {
	handler, signer := newTestGmHandler(t, &mockGmListener{resp: map[string]string{"ok": "true"}})

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing version",
			body: `{"request_id":"req_001","command":"Test","args":{}}`,
		},
		{
			name: "missing request_id",
			body: `{"version":"2.0","command":"Test","args":{}}`,
		},
		{
			name: "missing command",
			body: `{"version":"2.0","request_id":"req_001","args":{}}`,
		},
		{
			name: "missing args",
			body: `{"version":"2.0","request_id":"req_001","command":"Test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := signedGmRequest(t, signer, []byte(tt.body))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})
	}
}

func TestGmHandlerSuccess(t *testing.T) {
	listener := &mockGmListener{
		resp: map[string]any{"roles": []string{"role1", "role2"}},
	}
	handler, signer := newTestGmHandler(t, listener)

	req := signedGmRequest(t, signer, validGmBody(t))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", contentType)
	}
}

func TestGmHandlerListenerError(t *testing.T) {
	listener := &mockGmListener{
		err: &GmErrorResponse{
			Error:   GmError_InvalidCommand,
			Message: "unknown command",
		},
	}
	handler, signer := newTestGmHandler(t, listener)

	req := signedGmRequest(t, signer, validGmBody(t))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGmHandlerInternalError(t *testing.T) {
	listener := &mockGmListener{
		err: &GmErrorResponse{
			Error:   GmError_InternalError,
			Message: "something broke",
		},
	}
	handler, signer := newTestGmHandler(t, listener)

	req := signedGmRequest(t, signer, validGmBody(t))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestGmHandlerUnknownGmError(t *testing.T) {
	listener := &mockGmListener{
		err: &GmErrorResponse{
			Error:   GmError("custom_error"),
			Message: "custom",
		},
	}
	handler, signer := newTestGmHandler(t, listener)

	req := signedGmRequest(t, signer, validGmBody(t))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Unknown error should fall back to 500
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestGmError2HttpStatusMapping(t *testing.T) {
	expected := map[GmError]int{
		GmError_InvalidHttpMethod:   http.StatusMethodNotAllowed,
		GmError_InvalidContentType:  http.StatusUnsupportedMediaType,
		GmError_InvalidSignature:    http.StatusUnauthorized,
		GmError_InvalidRequest:      http.StatusBadRequest,
		GmError_InvalidCommand:      http.StatusBadRequest,
		GmError_InvalidArgs:         http.StatusBadRequest,
		GmError_ThrottlingError:     http.StatusTooManyRequests,
		GmError_IdempotencyConflict: http.StatusConflict,
		GmError_IdempotencyMismatch: http.StatusUnprocessableEntity,
		GmError_MaintenanceError:    http.StatusServiceUnavailable,
		GmError_NetworkError:        http.StatusInternalServerError,
		GmError_DatabaseError:       http.StatusInternalServerError,
		GmError_TimeoutError:        http.StatusInternalServerError,
		GmError_InternalError:       http.StatusInternalServerError,
	}

	for gmErr, httpStatus := range expected {
		got, ok := gmError2HttpStatus[gmErr]
		if !ok {
			t.Errorf("missing mapping for %s", gmErr)
			continue
		}
		if got != httpStatus {
			t.Errorf("expected %s -> %d, got %d", gmErr, httpStatus, got)
		}
	}
}
