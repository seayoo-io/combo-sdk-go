package combo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	traceIdHeader = "x-trace-id"
)

type responseReader interface {
	readResponse(resp *http.Response) error
}

type baseResponse struct {
	statusCode int
	traceId    string
}

func (r *baseResponse) readResponse(resp *http.Response) error {
	r.statusCode = resp.StatusCode
	r.traceId = resp.Header.Get(traceIdHeader)
	return nil
}

// HTTP 状态码，例如 200, 400, 500...
func (r *baseResponse) StatusCode() int {
	return r.statusCode
}

// 世游服务端生成的，用于追踪本次请求的唯一 ID。
//
// 建议游戏侧将 TraceId 记录到日志中，以便于调试和问题排查。
func (r *baseResponse) TraceId() string {
	return r.traceId
}

// ErrorResponse 对应 Combo Server API 返回的的错误响应。
//
// 游戏侧可使用 errors.As 来获取错误详细信息，示例如下：
//
//	import "log"
//	import "github.com/seayoo-io/combo-sdk-go"
//
//	// an API call to Client.CreateOrder() returns nil, err
//	if err != nil {
//	    var er *combo.ErrorResponse
//	    if errors.As(err, &er) {
//	        log.Printf(`failed to call API: error=%s, message="%s"\n`, er.ErrorCode, er.ErrorMessage)
//	    }
//	    return
//	}
type ErrorResponse struct {
	baseResponse

	// 业务错误码，示例 invalid_request, internal_error。
	ErrorCode string `json:"error"`

	// 错误的描述信息。
	ErrorMessage string `json:"message"`
}

// ErrorResponse 实现了 error 接口。
//
// 如果游戏侧需要将错误信息记录到日志中，可以直接将 ErrorResponse 作为 error 类型输出。实例如下：
//
//	import "log"
//
//	// ...
//
//	if err != nil {
//	    log.Printf("failed to call API: %v\n", err)
//	}
func (r *ErrorResponse) Error() string {
	return fmt.Sprintf(`status=%d, trace_id=%s, error=%s, message="%s"`,
		r.StatusCode(), r.TraceId(), r.ErrorCode, r.ErrorMessage)
}

func (r *ErrorResponse) readResponse(resp *http.Response) error {
	if err := r.baseResponse.readResponse(resp); err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return fmt.Errorf(`unexpected response: status=%d, body="%s"`, resp.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, r); err != nil {
		return fmt.Errorf("failed to unmarshal error response: %w", err)
	}
	return nil
}
