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

type httpResponseReader interface {
	ReadResponse(resp *http.Response) error
}

type baseResponse struct {
	// HTTP Status Code, e.g. 200, 400, 500...
	statusCode int

	// Unique ID for this request, used for debugging and tracing.
	traceId string
}

func (b *baseResponse) ReadResponse(resp *http.Response) error {
	b.statusCode = resp.StatusCode
	b.traceId = resp.Header.Get(traceIdHeader)
	return nil
}

func (b *baseResponse) StatusCode() int {
	return b.statusCode
}

func (b *baseResponse) TraceId() string {
	return b.traceId
}

type ErrorResponse struct {
	baseResponse

	ErrorCode    string `json:"error"`
	ErrorMessage string `json:"message"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf(`status=%d, trace_id=%s, error=%s, message="%s"`,
		e.StatusCode(), e.TraceId(), e.ErrorCode, e.ErrorMessage)
}

func (b *ErrorResponse) ReadResponse(resp *http.Response) error {
	if err := b.baseResponse.ReadResponse(resp); err != nil {
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
	if err := json.Unmarshal(body, b); err != nil {
		return fmt.Errorf("failed to unmarshal error response: %w", err)
	}
	return nil
}
