package combo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockNotificationListener struct {
	shipOrderCalled bool
	refundCalled    bool
	shipOrderErr    error
	refundErr       error
	lastShipOrder   *ShipOrderNotification
	lastRefund      *RefundNotification
	lastId          NotificationId
}

func (m *mockNotificationListener) HandleShipOrder(_ context.Context, id NotificationId, payload *ShipOrderNotification) error {
	m.shipOrderCalled = true
	m.lastShipOrder = payload
	m.lastId = id
	return m.shipOrderErr
}

func (m *mockNotificationListener) HandleRefund(_ context.Context, id NotificationId, payload *RefundNotification) error {
	m.refundCalled = true
	m.lastRefund = payload
	m.lastId = id
	return m.refundErr
}

func newTestNotificationHandler(t *testing.T, listener NotificationListener) (http.Handler, *httpSigner) {
	t.Helper()
	cfg := newTestConfig()
	handler, err := NewNotificationHandler(cfg, listener)
	if err != nil {
		t.Fatalf("failed to create notification handler: %v", err)
	}
	signer := &httpSigner{
		game:       cfg.GameId,
		signingKey: cfg.SecretKey,
	}
	return handler, signer
}

func signedNotificationRequest(t *testing.T, signer *httpSigner, body []byte) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if err := signer.SignHttp(req, time.Now()); err != nil {
		t.Fatalf("failed to sign request: %v", err)
	}
	return req
}

func TestNewNotificationHandlerInvalidConfig(t *testing.T) {
	_, err := NewNotificationHandler(Config{}, &mockNotificationListener{})
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

func TestNewNotificationHandlerNilListener(t *testing.T) {
	_, err := NewNotificationHandler(newTestConfig(), nil)
	if err == nil {
		t.Fatal("expected error for nil listener")
	}
}

func TestNotificationHandlerMethodNotAllowed(t *testing.T) {
	listener := &mockNotificationListener{}
	handler, _ := newTestNotificationHandler(t, listener)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestNotificationHandlerWrongContentType(t *testing.T) {
	listener := &mockNotificationListener{}
	handler, _ := newTestNotificationHandler(t, listener)

	req := httptest.NewRequest(http.MethodPost, "/notifications", nil)
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected %d, got %d", http.StatusUnsupportedMediaType, rec.Code)
	}
}

func TestNotificationHandlerUnauthorized(t *testing.T) {
	listener := &mockNotificationListener{}
	handler, _ := newTestNotificationHandler(t, listener)

	req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestNotificationHandlerShipOrder(t *testing.T) {
	listener := &mockNotificationListener{}
	handler, signer := newTestNotificationHandler(t, listener)

	shipData := ShipOrderNotification{
		OrderId:     "order_001",
		ReferenceId: "ref_001",
		ComboId:     "combo_001",
		ProductId:   "product_001",
		Quantity:    1,
		Currency:    "CNY",
		Amount:      648,
		Context:     "test_context",
		IsSandbox:   false,
	}
	dataBytes, _ := json.Marshal(shipData)
	body := notificationRequestBody{
		Version: "1.0",
		Id:      "notif_001",
		Type:    "ship_order",
		Data:    dataBytes,
	}
	bodyBytes, _ := json.Marshal(body)

	req := signedNotificationRequest(t, signer, bodyBytes)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if !listener.shipOrderCalled {
		t.Fatal("HandleShipOrder was not called")
	}
	if listener.lastId != "notif_001" {
		t.Fatalf("expected notification id notif_001, got %s", listener.lastId)
	}
	if listener.lastShipOrder.OrderId != "order_001" {
		t.Fatalf("expected order_id order_001, got %s", listener.lastShipOrder.OrderId)
	}
	if listener.lastShipOrder.Amount != 648 {
		t.Fatalf("expected amount 648, got %d", listener.lastShipOrder.Amount)
	}
}

func TestNotificationHandlerRefund(t *testing.T) {
	listener := &mockNotificationListener{}
	handler, signer := newTestNotificationHandler(t, listener)

	refundData := RefundNotification{
		OrderId:     "order_002",
		ReferenceId: "ref_002",
		ComboId:     "combo_002",
		ProductId:   "product_002",
		Quantity:    1,
		Currency:    "USD",
		Amount:      999,
		Context:     "refund_context",
	}
	dataBytes, _ := json.Marshal(refundData)
	body := notificationRequestBody{
		Version: "1.0",
		Id:      "notif_002",
		Type:    "refund",
		Data:    dataBytes,
	}
	bodyBytes, _ := json.Marshal(body)

	req := signedNotificationRequest(t, signer, bodyBytes)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if !listener.refundCalled {
		t.Fatal("HandleRefund was not called")
	}
	if listener.lastRefund.OrderId != "order_002" {
		t.Fatalf("expected order_id order_002, got %s", listener.lastRefund.OrderId)
	}
}

func TestNotificationHandlerShipOrderError(t *testing.T) {
	listener := &mockNotificationListener{
		shipOrderErr: errors.New("delivery failed"),
	}
	handler, signer := newTestNotificationHandler(t, listener)

	dataBytes, _ := json.Marshal(ShipOrderNotification{OrderId: "order_003"})
	body := notificationRequestBody{
		Version: "1.0",
		Id:      "notif_003",
		Type:    "ship_order",
		Data:    dataBytes,
	}
	bodyBytes, _ := json.Marshal(body)

	req := signedNotificationRequest(t, signer, bodyBytes)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestNotificationHandlerUnknownType(t *testing.T) {
	listener := &mockNotificationListener{}
	handler, signer := newTestNotificationHandler(t, listener)

	body := notificationRequestBody{
		Version: "1.0",
		Id:      "notif_004",
		Type:    "unknown_type",
		Data:    json.RawMessage(`{}`),
	}
	bodyBytes, _ := json.Marshal(body)

	req := signedNotificationRequest(t, signer, bodyBytes)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestNotificationHandlerInvalidJSON(t *testing.T) {
	listener := &mockNotificationListener{}
	handler, signer := newTestNotificationHandler(t, listener)

	req := signedNotificationRequest(t, signer, []byte(`not json`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
