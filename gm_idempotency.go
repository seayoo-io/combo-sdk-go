package combo

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// NewIdempotentGmListener 创建一个具有幂等性处理能力的 GmListener。
func NewIdempotentGmListener(cfg IdempotentGmListenerConfig) GmListener {
	if cfg.Store == nil {
		panic("missing required cfg.Store")
	}
	if cfg.Listener == nil {
		panic("missing required cfg.Listener")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	return &idempotentGmListener{
		store:  cfg.Store,
		real:   cfg.Listener,
		logger: cfg.Logger,
	}
}

// IdempotentGmListenerConfig 包含了创建具有幂等性处理能力的 GmListener 时所必需的配置项。
type IdempotentGmListenerConfig struct {
	Store    IdempotencyStore // 幂等性数据存储。实现可以是 Redis 或 Memory，也可以自行实现 IdempotencyStore
	Listener GmListener       // 实际执行业务逻辑的 GmListener
	Logger   *slog.Logger     // 记录日志的 logger，如果不指定，则默认会使用输出到 stderr 的 TextHandler
}

// NewMemoryIdempotencyStore 创建一个基于 Memory 的 IdempotencyStore 实现。
//
// 注意：该实现仅用于开发调试，不适合生产环境。
//
// 数据仅在内存中存储，重启服务后数据会丢失。数据不会过期，不会自动清理。
func NewMemoryIdempotencyStore() IdempotencyStore {
	return &memoryIdempotencyStore{
		kv: make(map[string]string),
	}
}

// NewRedisIdempotencyStore 创建一个基于 Redis 的 IdempotencyStore 实现。
//
// 数据会存储在 Redis 中，可以保证数据的的高可用性和到期自动清理。推荐生产环境使用。
func NewRedisIdempotencyStore(cfg RedisIdempotencyStoreConfig) IdempotencyStore {
	if cfg.Client == nil {
		panic("missing required cfg.Client")
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 24 * time.Hour
	}
	return &redisIdempotencyStore{
		client: cfg.Client,
		ttl:    cfg.TTL,
		prefix: cfg.Prefix,
	}
}

// RedisIdempotencyStoreConfig 包含了创建基于 Redis 的 IdempotencyStore 时所必需的配置项。
type RedisIdempotencyStoreConfig struct {
	Client redis.Cmdable // Redis 客户端。这里不假设 Redis 的运维部署方式。可以是 redis.Client 或者 redis.ClusterClient，由游戏侧自行创建和配置。
	TTL    time.Duration // Idempotency Key 的过期时间，如果不指定，则默认为 24 小时。
	Prefix string        // Idempotency Key 的前缀，如果不指定，则默认为空字符串。
}

// IdempotencyStore 是一个用于存储 GM 命令的幂等记录的接口。
//
// Combo SDK 内置了 Redis 和 Memory 两种实现，可分别通过 NewMemoryIdempotencyStore() 和 NewRedisIdempotencyStore() 创建。
//
// 游戏侧也可以选择自行实现 IdempotencyStore 接口。
type IdempotencyStore interface {
	// SetNX 用于原子性地存储幂等记录并返回旧值。
	// value 仅在 key 不存在时才会被存储 (Only set the key if it does not already exist)。
	// 返回值是 key 存在时的旧值。如果 key 不存在则返回空字符串。
	SetNX(ctx context.Context, key, value string) (string, error)

	// SetXX 用于原子性地更新已存在的幂等记录。
	// value 仅在 key 存在时才会被存储 (Only set the key if it already exists)。
	SetXX(ctx context.Context, key, value string) error
}

type idempotentGmListener struct {
	store  IdempotencyStore
	real   GmListener
	logger *slog.Logger
}

type idempotencyRecord struct {
	Nonce string           `json:"nonce"` // 这里的 nonce 是 idempotencyRecord 的唯一标识，用于解决 Redis 的 SetNX 缺乏幂等性的问题。
	Key   string           `json:"key"`
	Id    string           `json:"id"`
	Cmd   string           `json:"cmd"`
	Args  string           `json:"args"`
	Done  bool             `json:"done"`
	Resp  json.RawMessage  `json:"resp"`
	Error *GmErrorResponse `json:"error"`
}

// HandleGmRequest implements GmListener.
func (i *idempotentGmListener) HandleGmRequest(ctx context.Context, req *GmRequest) (any, *GmErrorResponse) {
	if req.IdempotencyKey == "" {
		return i.real.HandleGmRequest(ctx, req)
	}
	nonce, err := uuid.NewRandom()
	if err != nil {
		return nil, &GmErrorResponse{
			Error:   GmError_InternalError,
			Message: fmt.Sprintf("failed to generate nonce: %v", err),
		}
	}
	record := &idempotencyRecord{
		Nonce: nonce.String(),
		Key:   req.IdempotencyKey,
		Id:    req.Id,
		Cmd:   req.Cmd,
		Args:  string(req.Args),
		Done:  false,
	}
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return nil, &GmErrorResponse{
			Error:   GmError_InternalError,
			Message: fmt.Sprintf("failed to marshal idempotency record: %v", err),
		}
	}
	oldRecordStr, err := i.store.SetNX(ctx, req.IdempotencyKey, string(recordBytes))
	if err != nil {
		return nil, &GmErrorResponse{
			Error:   GmError_InternalError,
			Message: fmt.Sprintf("failed to SetNX idempotency record: %v", err),
		}
	}
	oldRecord, err := i.parseOldRecord(oldRecordStr)
	if err != nil {
		return nil, &GmErrorResponse{
			Error:   GmError_InternalError,
			Message: fmt.Sprintf("failed to unmarshal old idempotency record: %v", err),
		}
	}
	// go-redis 内部具有自动重试机制（默认最大重试 3 次），但 SetNX 缺乏幂等性。
	// 在极端情况下可能会出现 SetNX 先失败（但 Redis 执行成功）再重试成功的情况。
	// 此时返回的 oldRecord 的 Nonce 其实就是当前 goroutine 生成的，所以应当作为首次请求来处理。
	firstTimeRequest := oldRecord == nil || oldRecord.Nonce == record.Nonce
	if firstTimeRequest {
		return i.processRequest(ctx, req, record)
	}
	return i.previousResponse(req, oldRecord)
}

func (i *idempotentGmListener) parseOldRecord(str string) (*idempotencyRecord, error) {
	if str == "" {
		return nil, nil
	}
	record := &idempotencyRecord{}
	if err := json.Unmarshal([]byte(str), record); err != nil {
		return nil, err
	}
	return record, nil
}

func (i *idempotentGmListener) processRequest(ctx context.Context, req *GmRequest, record *idempotencyRecord) (any, *GmErrorResponse) {
	resp, err := i.real.HandleGmRequest(ctx, req)
	respBytes, _ := json.Marshal(resp)
	record.Done = true
	record.Resp = respBytes
	record.Error = err
	i.saveIdempotencyRecord(ctx, record)
	return record.Resp, record.Error
}

func (i *idempotentGmListener) saveIdempotencyRecord(ctx context.Context, record *idempotencyRecord) {
	recordBytes, _ := json.Marshal(record)
	recordStr := string(recordBytes)
	// 这里不做额外重试，而是依赖于 go-redis 内部的重试机制。
	// 游戏侧可通过 redis.Options 来配置最大重试次数、指数退避规则。
	// 如果 err != nil 通常是访问 Redis 出现了可用性问题，则仅记录日志而不改变返回给调用方的 GM response。
	// 相比直接返回给调用方失败，这里在赌调用方成功接收并处理 GM response。如果赌对了则业务完全不受影响，而赌错的概率是比较低的。
	if err := i.store.SetXX(ctx, record.Key, recordStr); err != nil {
		i.logger.ErrorContext(ctx,
			"failed to SetXX idempotency record",
			slog.Any("err", err),
			slog.Group("record",
				slog.String("nonce", record.Nonce),
				slog.String("key", record.Key),
				slog.String("id", record.Id),
				slog.String("cmd", record.Cmd),
				slog.String("args", record.Args),
				slog.Bool("done", record.Done),
				slog.String("resp", string(record.Resp)),
				slog.Any("error", record.Error),
			),
		)
	}
}

func (i *idempotentGmListener) previousResponse(req *GmRequest, record *idempotencyRecord) (any, *GmErrorResponse) {
	if !record.Done {
		return nil, &GmErrorResponse{
			Error:   GmError_IdempotencyConflict,
			Message: "previous request is not completed",
		}
	}
	if record.Cmd != req.Cmd {
		return nil, &GmErrorResponse{
			Error:   GmError_IdempotencyMismatch,
			Message: fmt.Sprintf("cmd mismatch, expecting %s, got %s", record.Cmd, req.Cmd),
		}
	}
	if record.Args != string(req.Args) {
		return nil, &GmErrorResponse{
			Error:   GmError_IdempotencyMismatch,
			Message: fmt.Sprintf("args mismatch, expecting %s, got %s", string(record.Args), string(req.Args)),
		}
	}
	return record.Resp, record.Error
}

type memoryIdempotencyStore struct {
	mu sync.Mutex
	kv map[string]string
}

// SetNX implements IdempotencyStore.
func (s *memoryIdempotencyStore) SetNX(ctx context.Context, key string, value string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	oldValue, ok := s.kv[key]
	if ok {
		return oldValue, nil
	}
	s.kv[key] = value
	return "", nil
}

// SetXX implements IdempotencyStore.
func (s *memoryIdempotencyStore) SetXX(ctx context.Context, key string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.kv[key]
	if ok {
		s.kv[key] = value
	}
	return nil
}

type redisIdempotencyStore struct {
	client redis.Cmdable
	ttl    time.Duration
	prefix string
}

// SetNX implements IdempotencyStore.
func (s *redisIdempotencyStore) SetNX(ctx context.Context, key string, value string) (string, error) {
	oldValue, err := s.client.SetArgs(ctx, s.prefix+key, value, redis.SetArgs{
		Mode: "NX",
		TTL:  s.ttl,
		Get:  true,
	}).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	} else {
		return oldValue, nil
	}
}

// SetXX implements IdempotencyStore.
func (s *redisIdempotencyStore) SetXX(ctx context.Context, key string, value string) error {
	return s.client.SetArgs(ctx, s.prefix+key, value, redis.SetArgs{
		Mode:    "XX",
		KeepTTL: true,
	}).Err()
}
