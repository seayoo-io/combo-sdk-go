package combo

import (
	"net/http"
	"strings"
	"testing"
)

func TestErrorResponseError(t *testing.T) {
	er := &ErrorResponse{
		baseResponse: baseResponse{
			statusCode: 400,
			traceId:    "trace_123",
		},
		ErrorCode:    "invalid_request",
		ErrorMessage: "missing field",
	}

	s := er.Error()
	if !strings.Contains(s, "status=400") {
		t.Errorf("error string should contain status code, got %s", s)
	}
	if !strings.Contains(s, "trace_id=trace_123") {
		t.Errorf("error string should contain trace id, got %s", s)
	}
	if !strings.Contains(s, "error=invalid_request") {
		t.Errorf("error string should contain error code, got %s", s)
	}
	if !strings.Contains(s, `message="missing field"`) {
		t.Errorf("error string should contain error message, got %s", s)
	}
}

func TestBaseResponseReadResponse(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"X-Trace-Id": []string{"trace_abc"}},
	}

	br := &baseResponse{}
	err := br.readResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if br.StatusCode() != 200 {
		t.Fatalf("expected status 200, got %d", br.StatusCode())
	}
	if br.TraceId() != "trace_abc" {
		t.Fatalf("expected trace_id trace_abc, got %s", br.TraceId())
	}
}

func TestErrorResponseImplementsError(t *testing.T) {
	var err error = &ErrorResponse{
		ErrorCode:    "test",
		ErrorMessage: "test message",
	}
	if err == nil {
		t.Fatal("ErrorResponse should implement error interface")
	}
}
