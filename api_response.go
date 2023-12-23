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

type ErrorResponse struct {
	baseResponse

	ErrorCode    string `json:"error"`
	ErrorMessage string `json:"message"`
}

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
