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

type ComboResponse interface {
	ReadHttp(resp *http.Response) error
	StatusCode() int
	TraceId() string
}

type BaseResponse struct {
	// HTTP Status Code, e.g. 200, 400, 500...
	statusCode int

	// Unique ID for this request, used for debugging and tracing.
	traceId string
}

func (b *BaseResponse) ReadHttp(resp *http.Response) error {
	b.statusCode = resp.StatusCode
	b.traceId = resp.Header.Get(traceIdHeader)
	return nil
}

func (b *BaseResponse) StatusCode() int {
	return b.statusCode
}

func (b *BaseResponse) TraceId() string {
	return b.traceId
}

type ErrorResponse struct {
	BaseResponse

	ErrorCode    string `json:"error"`
	ErrorMessage string `json:"message"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf(`status=%d, trace_id=%s, error=%s, message="%s"`,
		e.StatusCode(), e.TraceId(), e.ErrorCode, e.ErrorMessage)
}

func (b *ErrorResponse) ReadHttp(resp *http.Response) error {
	b.BaseResponse.ReadHttp(resp)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return fmt.Errorf(`unexpected response: status=%d, body="%s"`, resp.StatusCode, string(body))
	}
	if err := json.Unmarshal(body, b); err != nil {
		return fmt.Errorf("failed to unmarshal error response: %w", err)
	}
	return nil
}
