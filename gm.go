package combo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

// NewGmHandler 创建一个用于处理世游服务端发送的 GM 命令的 http.Handler。
//
// 游戏侧需要将此 Handler 注册到游戏的 HTTP 服务中。
//
// 注意：注册 Handler 时，应当使用 HTTP POST。
func NewGmHandler(cfg Config, listener GmListener) (http.Handler, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if listener == nil {
		return nil, errors.New("missing required listener")
	}
	return &gmHandler{
		signer: httpSigner{
			game:       cfg.GameId,
			signingKey: cfg.SecretKey,
		},
		listener: listener,
	}, nil
}

// GmListener 是一个用于接收并处理世游服务端发送的 GM 命令的接口。
// 游戏侧需要实现此接口并根据 GM 命令执行对应的业务逻辑。
type GmListener interface {
	// HandleGmRequest 负责执行 GM 命令。
	// resp 是 GM 命令执行成功时的返回值，结构需要 GM 协议文件中 rpc 的 Response 一致。
	// Combo SDK 会对 resp 做 JSON 序列化操作。
	// - 如果游戏侧希望自己控制 JSON 序列化的过程，resp 实现 interface json.Marshaler 即可。
	// - 如果游戏侧希望直接返回 JSON 序列化的结果，确保 resp 的类型是 json.RawMessage 即可。
	// err 是 GM 命令执行失败时的错误信息。
	// - 如果 GM 命令执行成功，应当返回 resp, nil。
	// - 如果 GM 命令执行失败，应当返回 nil, err。
	HandleGmRequest(ctx context.Context, req *GmRequest) (resp any, err *GmErrorResponse)
}

type GmRequest struct {
	// Version 是 GM 请求的的版本号。当前版本固定为 2.0。
	Version string
	// Id 是本次 GM 请求的唯一 ID。游戏后端可用此值来对通知进行去重处理。
	Id string
	// IdempotencyKey 是本次 GM 请求的 Idempotency Key。如果有值则应当执行幂等处理逻辑。
	IdempotencyKey string
	// Cmd 是 GM 命令标识。取值和 GM 协议文件中的 rpc 名称对应。
	Cmd string
	// Args 是和 Cmd 对应的 GM 命令参数。这是一个异构的 JSON Object，结构和 GM 协议文件中 rpc 的 Request 一致。
	// 游戏侧需要根据 Cmd 的值，对 Args 进行 JSON 反序列化。
	Args json.RawMessage
}

// GmErrorResponse 是 GM 命令处理失败时返回的响应。
type GmErrorResponse struct {
	// 错误类型。建议游戏侧优先使用 Combo SDK 预定义好的 GmError 值。
	Error GmError `json:"error"`
	// 错误描述信息。预期是给人看的，便于联调和问题排查。
	Message string `json:"message"`
	// 如果为 true，表示无法确定 GM 命令是否被执行。默认语义是 GM 命令未被执行（否则应当返回成功响应）。
	Uncertain bool `json:"uncertain,omitempty"`
}

// GmError 是 GM 命令处理失败时返回的错误类型。
type GmError string

// 客户端错误。
// 这些错误通常是由于世游侧发送的请求不正确导致的。
const (
	// 请求中的 HTTP method 不正确，没有按照预期使用 POST。
	GmError_InvalidHttpMethod GmError = "invalid_http_method"

	// 请求中的 Content-Type 不是 application/json。
	GmError_InvalidContentType GmError = "invalid_content_type"

	// 对 HTTP 请求的签名验证不通过。这意味着 HTTP 请求不可信。
	GmError_InvalidSignature GmError = "invalid_signature"

	// 请求的结构不正确。例如，缺少必要的字段，或字段类型不正确。
	GmError_InvalidRequest GmError = "invalid_request"

	// 游戏侧不认识请求中的 GM 命令。
	GmError_InvalidCommand GmError = "invalid_command"

	// GM 命令的参数不正确。例如，参数缺少必要的字段，或参数的字段类型不正确。
	GmError_InvalidArgs GmError = "invalid_args"

	// GM 命令发送频率过高，被游戏侧限流，命令未被处理。
	GmError_ThrottlingError GmError = "throttling_error"

	// 幂等处理重试请求时，idempotency_key 所对应的原始请求尚未处理完毕。
	GmError_IdempotencyConflict GmError = "idempotency_conflict"

	// 幂等处理重试请求时，请求内容和 idempotency_key 所对应的原始请求内容不一致。
	GmError_IdempotencyMismatch GmError = "idempotency_mismatch"
)

// 服务端错误。
// 这些错误通常是游戏侧处理 GM 命令时出现问题导致的。
const (
	// 游戏当前处于停服维护状态，无法处理收到的 GM 命令。
	GmError_MaintenanceError GmError = "maintenance_error"

	// 网络通信错误导致 GM 命令执行失败。
	GmError_NetworkError GmError = "network_error"

	// 数据库操作异常导致 GM 命令执行失败。
	GmError_DatabaseError GmError = "database_error"

	// GM 命令处理超时。
	GmError_TimeoutError GmError = "timeout_error"

	// 处理 GM 命令时内部出错。可作为兜底的通用错误类型。
	GmError_InternalError GmError = "internal_error"
)

var jsonContentType = []string{"application/json"}

var gmError2HttpStatus = map[GmError]int{
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

type gmRequestBody struct {
	Version        string          `json:"version"`
	RequestId      string          `json:"request_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	Command        string          `json:"command"`
	Args           json.RawMessage `json:"args"`
}

type gmHandler struct {
	signer   httpSigner
	listener GmListener
}

func (h *gmHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.json(w, http.StatusMethodNotAllowed, GmErrorResponse{
			Error:   GmError_InvalidHttpMethod,
			Message: "Expecting POST, got " + r.Method,
		})
		return
	}
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		h.json(w, http.StatusUnsupportedMediaType, GmErrorResponse{
			Error:   GmError_InvalidContentType,
			Message: "Expecting application/json, got " + contentType,
		})
		return
	}
	if err := h.signer.AuthHttp(r, time.Now()); err != nil {
		h.json(w, http.StatusUnauthorized, GmErrorResponse{
			Error:   GmError_InvalidSignature,
			Message: err.Error(),
		})
		return
	}
	var body gmRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.json(w, http.StatusBadRequest, GmErrorResponse{
			Error:   GmError_InvalidRequest,
			Message: err.Error(),
		})
		return
	}
	if err := h.validateRequestBody(&body); err != nil {
		h.json(w, http.StatusBadRequest, err)
		return
	}
	resp, err := h.listener.HandleGmRequest(r.Context(), &GmRequest{
		Version:        body.Version,
		Id:             body.RequestId,
		IdempotencyKey: body.IdempotencyKey,
		Cmd:            body.Command,
		Args:           body.Args,
	})
	if err != nil {
		status, ok := gmError2HttpStatus[err.Error]
		if !ok {
			status = http.StatusInternalServerError
		}
		h.json(w, status, err)
		return
	}
	h.json(w, http.StatusOK, resp)
}

func (h *gmHandler) validateRequestBody(body *gmRequestBody) *GmErrorResponse {
	if body.Version == "" {
		return &GmErrorResponse{
			Error:   GmError_InvalidRequest,
			Message: "Missing required field: version",
		}
	}
	if body.RequestId == "" {
		return &GmErrorResponse{
			Error:   GmError_InvalidRequest,
			Message: "Missing required field: request_id",
		}
	}
	if body.Command == "" {
		return &GmErrorResponse{
			Error:   GmError_InvalidRequest,
			Message: "Missing required field: command",
		}
	}
	if body.Args == nil {
		return &GmErrorResponse{
			Error:   GmError_InvalidRequest,
			Message: "Missing required field: args",
		}
	}
	return nil
}

func (h *gmHandler) json(w http.ResponseWriter, code int, obj any) {
	header := w.Header()
	header["Content-Type"] = jsonContentType
	w.WriteHeader(code)
	jsonBytes, _ := json.Marshal(obj)
	if jsonBytes != nil {
		_, _ = w.Write(jsonBytes)
	}
}
