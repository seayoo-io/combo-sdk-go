package combo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestNewIdempotentGmListenerPanicsWithoutStore(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for missing store")
		}
	}()
	NewIdempotentGmListener(IdempotentGmListenerConfig{
		Listener: &mockGmListener{},
	})
}

func TestNewIdempotentGmListenerPanicsWithoutListener(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for missing listener")
		}
	}()
	NewIdempotentGmListener(IdempotentGmListenerConfig{
		Store: NewMemoryIdempotencyStore(),
	})
}

func TestIdempotentGmListenerNoKey(t *testing.T) {
	inner := &mockGmListener{resp: map[string]string{"result": "ok"}}
	listener := NewIdempotentGmListener(IdempotentGmListenerConfig{
		Store:    NewMemoryIdempotencyStore(),
		Listener: inner,
	})

	req := &GmRequest{
		Version: "2.0",
		Id:      "req_001",
		Cmd:     "TestCmd",
		Args:    json.RawMessage(`{}`),
		// No IdempotencyKey
	}

	resp, err := listener.HandleGmRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIdempotentGmListenerFirstRequest(t *testing.T) {
	inner := &mockGmListener{resp: map[string]string{"result": "ok"}}
	listener := NewIdempotentGmListener(IdempotentGmListenerConfig{
		Store:    NewMemoryIdempotencyStore(),
		Listener: inner,
	})

	req := &GmRequest{
		Version:        "2.0",
		Id:             "req_001",
		IdempotencyKey: "idem_001",
		Cmd:            "TestCmd",
		Args:           json.RawMessage(`{"key":"value"}`),
	}

	resp, err := listener.HandleGmRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestIdempotentGmListenerReplay(t *testing.T) {
	inner := &mockGmListener{resp: map[string]string{"result": "ok"}}
	store := NewMemoryIdempotencyStore()
	listener := NewIdempotentGmListener(IdempotentGmListenerConfig{
		Store:    store,
		Listener: inner,
	})

	req := &GmRequest{
		Version:        "2.0",
		Id:             "req_001",
		IdempotencyKey: "idem_001",
		Cmd:            "TestCmd",
		Args:           json.RawMessage(`{"key":"value"}`),
	}

	// First request
	resp1, err1 := listener.HandleGmRequest(context.Background(), req)
	if err1 != nil {
		t.Fatalf("first request error: %v", err1)
	}

	// Replay with same key, cmd, args
	req2 := &GmRequest{
		Version:        "2.0",
		Id:             "req_002",
		IdempotencyKey: "idem_001",
		Cmd:            "TestCmd",
		Args:           json.RawMessage(`{"key":"value"}`),
	}
	resp2, err2 := listener.HandleGmRequest(context.Background(), req2)
	if err2 != nil {
		t.Fatalf("replay request error: %v", err2)
	}

	// Both responses should be the same JSON
	r1, _ := json.Marshal(resp1)
	r2, _ := json.Marshal(resp2)
	if string(r1) != string(r2) {
		t.Fatalf("expected same response for replay, got %s vs %s", string(r1), string(r2))
	}
}

func TestIdempotentGmListenerMismatchCmd(t *testing.T) {
	inner := &mockGmListener{resp: map[string]string{"result": "ok"}}
	store := NewMemoryIdempotencyStore()
	listener := NewIdempotentGmListener(IdempotentGmListenerConfig{
		Store:    store,
		Listener: inner,
	})

	req := &GmRequest{
		Version:        "2.0",
		Id:             "req_001",
		IdempotencyKey: "idem_002",
		Cmd:            "CmdA",
		Args:           json.RawMessage(`{}`),
	}
	listener.HandleGmRequest(context.Background(), req)

	// Same key but different command
	req2 := &GmRequest{
		Version:        "2.0",
		Id:             "req_002",
		IdempotencyKey: "idem_002",
		Cmd:            "CmdB",
		Args:           json.RawMessage(`{}`),
	}
	_, err := listener.HandleGmRequest(context.Background(), req2)
	if err == nil {
		t.Fatal("expected error for cmd mismatch")
	}
	if err.Error != GmError_IdempotencyMismatch {
		t.Fatalf("expected IdempotencyMismatch, got %s", err.Error)
	}
}

func TestIdempotentGmListenerMismatchArgs(t *testing.T) {
	inner := &mockGmListener{resp: map[string]string{"result": "ok"}}
	store := NewMemoryIdempotencyStore()
	listener := NewIdempotentGmListener(IdempotentGmListenerConfig{
		Store:    store,
		Listener: inner,
	})

	req := &GmRequest{
		Version:        "2.0",
		Id:             "req_001",
		IdempotencyKey: "idem_003",
		Cmd:            "TestCmd",
		Args:           json.RawMessage(`{"key":"value1"}`),
	}
	listener.HandleGmRequest(context.Background(), req)

	// Same key and cmd but different args
	req2 := &GmRequest{
		Version:        "2.0",
		Id:             "req_002",
		IdempotencyKey: "idem_003",
		Cmd:            "TestCmd",
		Args:           json.RawMessage(`{"key":"value2"}`),
	}
	_, err := listener.HandleGmRequest(context.Background(), req2)
	if err == nil {
		t.Fatal("expected error for args mismatch")
	}
	if err.Error != GmError_IdempotencyMismatch {
		t.Fatalf("expected IdempotencyMismatch, got %s", err.Error)
	}
}

func TestIdempotentGmListenerErrorReplay(t *testing.T) {
	inner := &mockGmListener{
		err: &GmErrorResponse{
			Error:   GmError_InternalError,
			Message: "something broke",
		},
	}
	store := NewMemoryIdempotencyStore()
	listener := NewIdempotentGmListener(IdempotentGmListenerConfig{
		Store:    store,
		Listener: inner,
	})

	req := &GmRequest{
		Version:        "2.0",
		Id:             "req_001",
		IdempotencyKey: "idem_004",
		Cmd:            "TestCmd",
		Args:           json.RawMessage(`{}`),
	}

	// First request returns error
	_, err1 := listener.HandleGmRequest(context.Background(), req)
	if err1 == nil {
		t.Fatal("expected error from first request")
	}

	// Replay should return same error
	req2 := &GmRequest{
		Version:        "2.0",
		Id:             "req_002",
		IdempotencyKey: "idem_004",
		Cmd:            "TestCmd",
		Args:           json.RawMessage(`{}`),
	}
	_, err2 := listener.HandleGmRequest(context.Background(), req2)
	if err2 == nil {
		t.Fatal("expected error from replay")
	}
	if err2.Error != GmError_InternalError {
		t.Fatalf("expected InternalError, got %s", err2.Error)
	}
}

func TestMemoryIdempotencyStoreSetNX(t *testing.T) {
	store := NewMemoryIdempotencyStore()
	ctx := context.Background()

	// First SetNX should succeed with empty old value
	old, err := store.SetNX(ctx, "key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if old != "" {
		t.Fatalf("expected empty old value for first SetNX, got %s", old)
	}

	// Second SetNX with same key should return old value
	old, err = store.SetNX(ctx, "key1", "value2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if old != "value1" {
		t.Fatalf("expected old value 'value1', got %s", old)
	}
}

func TestMemoryIdempotencyStoreSetXX(t *testing.T) {
	store := NewMemoryIdempotencyStore()
	ctx := context.Background()

	// SetXX on non-existent key should not store
	err := store.SetXX(ctx, "key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify key was not set
	old, _ := store.SetNX(ctx, "key1", "value2")
	if old != "" {
		t.Fatalf("expected key not to be set by SetXX, but got old value %s", old)
	}

	// Now SetXX should update existing key
	err = store.SetXX(ctx, "key1", "value3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it was updated
	old, _ = store.SetNX(ctx, "key1", "value4")
	if old != "value3" {
		t.Fatalf("expected old value 'value3' after SetXX, got %s", old)
	}
}

// --- Redis-backed IdempotencyStore tests ---

func newTestRedisStore(t *testing.T) (IdempotencyStore, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := NewRedisIdempotencyStore(RedisIdempotencyStoreConfig{
		Client: client,
		TTL:    10 * time.Minute,
	})
	return store, mr
}

func mustGet(t *testing.T, mr *miniredis.Miniredis, key string) string {
	t.Helper()
	val, err := mr.Get(key)
	if err != nil {
		t.Fatalf("failed to get key %q: %v", key, err)
	}
	return val
}

func TestNewRedisIdempotencyStorePanicsWithoutClient(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for missing client")
		}
	}()
	NewRedisIdempotencyStore(RedisIdempotencyStoreConfig{})
}

func TestNewRedisIdempotencyStoreDefaultTTL(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := NewRedisIdempotencyStore(RedisIdempotencyStoreConfig{
		Client: client,
	})
	s := store.(*redisIdempotencyStore)
	if s.ttl != 24*time.Hour {
		t.Fatalf("expected default TTL 24h, got %v", s.ttl)
	}
}

func TestRedisIdempotencyStoreSetNXFirstTime(t *testing.T) {
	store, _ := newTestRedisStore(t)
	ctx := context.Background()

	old, err := store.SetNX(ctx, "key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if old != "" {
		t.Fatalf("expected empty old value for first SetNX, got %q", old)
	}
}

func TestRedisIdempotencyStoreSetNXExistingKey(t *testing.T) {
	store, _ := newTestRedisStore(t)
	ctx := context.Background()

	// First set
	_, err := store.SetNX(ctx, "key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second SetNX with same key should return old value
	old, err := store.SetNX(ctx, "key1", "value2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if old != "value1" {
		t.Fatalf("expected old value 'value1', got %q", old)
	}
}

func TestRedisIdempotencyStoreSetNXDoesNotOverwrite(t *testing.T) {
	store, mr := newTestRedisStore(t)
	ctx := context.Background()

	store.SetNX(ctx, "key1", "value1")
	store.SetNX(ctx, "key1", "value2")

	// Verify original value is preserved
	got := mustGet(t, mr, "key1")
	if got != "value1" {
		t.Fatalf("SetNX should not overwrite, expected 'value1', got %q", got)
	}
}

func TestRedisIdempotencyStoreSetXXUpdatesExisting(t *testing.T) {
	store, mr := newTestRedisStore(t)
	ctx := context.Background()

	// Create key first
	store.SetNX(ctx, "key1", "value1")

	// SetXX should update
	err := store.SetXX(ctx, "key1", "value2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := mustGet(t, mr, "key1")
	if got != "value2" {
		t.Fatalf("expected 'value2' after SetXX, got %q", got)
	}
}

func TestRedisIdempotencyStoreSetXXNonExistentKey(t *testing.T) {
	store, mr := newTestRedisStore(t)
	ctx := context.Background()

	// SetXX on non-existent key returns redis.Nil error and does not create the key
	err := store.SetXX(ctx, "key1", "value1")
	if err == nil {
		t.Fatal("expected error for SetXX on non-existent key")
	}

	if mr.Exists("key1") {
		t.Fatal("SetXX should not create a non-existent key")
	}
}

func TestRedisIdempotencyStoreWithPrefix(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := NewRedisIdempotencyStore(RedisIdempotencyStoreConfig{
		Client: client,
		Prefix: "gm:",
		TTL:    10 * time.Minute,
	})
	ctx := context.Background()

	store.SetNX(ctx, "key1", "value1")

	// Key should be stored with prefix
	if !mr.Exists("gm:key1") {
		t.Fatal("expected key to be stored with prefix 'gm:'")
	}
	if mr.Exists("key1") {
		t.Fatal("key should not exist without prefix")
	}
}

func TestRedisIdempotencyStoreTTL(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := NewRedisIdempotencyStore(RedisIdempotencyStoreConfig{
		Client: client,
		TTL:    5 * time.Minute,
	})
	ctx := context.Background()

	store.SetNX(ctx, "key1", "value1")

	// Key should exist before TTL
	if !mr.Exists("key1") {
		t.Fatal("key should exist before TTL expires")
	}

	// Fast-forward past TTL
	mr.FastForward(6 * time.Minute)

	if mr.Exists("key1") {
		t.Fatal("key should expire after TTL")
	}
}

func TestRedisIdempotencyStoreSetXXKeepsTTL(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := NewRedisIdempotencyStore(RedisIdempotencyStoreConfig{
		Client: client,
		TTL:    10 * time.Minute,
	})
	ctx := context.Background()

	store.SetNX(ctx, "key1", "value1")

	// Advance 3 minutes
	mr.FastForward(3 * time.Minute)

	// Update with SetXX (should keep original TTL)
	store.SetXX(ctx, "key1", "value2")

	// Advance another 8 minutes (total 11 minutes, past original 10-min TTL)
	mr.FastForward(8 * time.Minute)

	if mr.Exists("key1") {
		t.Fatal("key should expire based on original TTL")
	}
}

func TestRedisIdempotencyStoreConnectionError(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := NewRedisIdempotencyStore(RedisIdempotencyStoreConfig{
		Client: client,
		TTL:    10 * time.Minute,
	})
	ctx := context.Background()

	// Close Redis to simulate connection error
	mr.Close()

	_, err := store.SetNX(ctx, "key1", "value1")
	if err == nil {
		t.Fatal("expected error when Redis is down")
	}

	err = store.SetXX(ctx, "key1", "value1")
	if err == nil {
		t.Fatal("expected error when Redis is down")
	}
}

func TestIdempotentGmListenerWithRedisStore(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := NewRedisIdempotencyStore(RedisIdempotencyStoreConfig{
		Client: client,
		TTL:    10 * time.Minute,
		Prefix: "gm:idem:",
	})

	inner := &mockGmListener{resp: map[string]string{"status": "delivered"}}
	listener := NewIdempotentGmListener(IdempotentGmListenerConfig{
		Store:    store,
		Listener: inner,
	})
	ctx := context.Background()

	req := &GmRequest{
		Version:        "2.0",
		Id:             "req_001",
		IdempotencyKey: "idem_redis_001",
		Cmd:            "DeliverItem",
		Args:           json.RawMessage(`{"item_id":"sword_001"}`),
	}

	// First request
	resp1, err1 := listener.HandleGmRequest(ctx, req)
	if err1 != nil {
		t.Fatalf("first request error: %v", err1)
	}

	// Replay with same idempotency key
	req2 := &GmRequest{
		Version:        "2.0",
		Id:             "req_002",
		IdempotencyKey: "idem_redis_001",
		Cmd:            "DeliverItem",
		Args:           json.RawMessage(`{"item_id":"sword_001"}`),
	}
	resp2, err2 := listener.HandleGmRequest(ctx, req2)
	if err2 != nil {
		t.Fatalf("replay error: %v", err2)
	}

	r1, _ := json.Marshal(resp1)
	r2, _ := json.Marshal(resp2)
	if string(r1) != string(r2) {
		t.Fatalf("expected same response on replay, got %s vs %s", string(r1), string(r2))
	}

	// Replay with different cmd should fail
	req3 := &GmRequest{
		Version:        "2.0",
		Id:             "req_003",
		IdempotencyKey: "idem_redis_001",
		Cmd:            "DifferentCmd",
		Args:           json.RawMessage(`{"item_id":"sword_001"}`),
	}
	_, err3 := listener.HandleGmRequest(ctx, req3)
	if err3 == nil {
		t.Fatal("expected IdempotencyMismatch for different cmd")
	}
	if err3.Error != GmError_IdempotencyMismatch {
		t.Fatalf("expected IdempotencyMismatch, got %s", err3.Error)
	}
}
