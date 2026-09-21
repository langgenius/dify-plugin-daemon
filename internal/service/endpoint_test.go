package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/endpoint_entities"
)

func TestEndpointSessionRequestContextSurvivesHTTPCancel(t *testing.T) {
	httpCtx, cancelHTTP := context.WithCancel(context.Background())
	sessionCtx, cancelSession := endpointSessionRequestContext(httpCtx, time.Minute)
	defer cancelSession()

	cancelHTTP()
	if err := httpCtx.Err(); err == nil {
		t.Fatal("expected HTTP context to be cancelled")
	}
	if err := sessionCtx.Err(); err != nil {
		t.Fatalf("session context should outlive HTTP disconnect: %v", err)
	}
}

func TestEndpointSessionRequestContextRespectsMaxExecutionTime(t *testing.T) {
	httpCtx := context.Background()
	sessionCtx, cancelSession := endpointSessionRequestContext(httpCtx, 20*time.Millisecond)
	defer cancelSession()

	deadline, ok := sessionCtx.Deadline()
	if !ok {
		t.Fatal("expected session context to have a deadline")
	}
	if time.Until(deadline) <= 0 {
		t.Fatal("expected session deadline to be in the future")
	}

	<-sessionCtx.Done()
	if err := sessionCtx.Err(); err != context.DeadlineExceeded {
		t.Fatalf("expected session context to time out: %v", err)
	}
}

func TestCopyRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8080/test?test=123", nil)
	req.Body = io.NopCloser(bytes.NewReader([]byte("test")))
	if err != nil {
		t.Fatal(err)
	}

	buffer, err := copyRequest(req, "123", "/test")
	if err != nil {
		t.Fatal(err)
	}

	str := buffer.String()
	if str != "GET /test?test=123 HTTP/1.1\r\nHost: localhost:8080\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 4\r\nDify-Hook-Id: 123\r\nDify-Hook-Url: http://localhost:8080/e/123/test\r\n\r\ntest" {
		t.Fatal("request body is not equal, ", str)
	}
}

func TestEndpointWithOriginalHost(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8080/test?test=123", nil)
	req.Header.Set(endpoint_entities.HeaderXOriginalHost, "example.com")
	if err != nil {
		t.Fatal(err)
	}
	payload := "test"
	req.Body = io.NopCloser(bytes.NewReader([]byte(payload)))

	buffer, err := copyRequest(req, "123", "/test")
	if err != nil {
		t.Fatal(err)
	}

	str := buffer.String()
	if str != "GET /test?test=123 HTTP/1.1\r\nHost: example.com\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 4\r\nDify-Hook-Id: 123\r\nDify-Hook-Url: http://example.com/e/123/test\r\n\r\ntest" {
		t.Fatal("request body is not equal, ", str)
	}
}

func TestEndpointWithHeaderDifyHookURLEmpty(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/test?test=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(endpoint_entities.HeaderDifyHookURL, "")
	payload := "test"
	req.Body = io.NopCloser(bytes.NewReader([]byte(payload)))

	buffer, err := copyRequest(req, "123", "/test")
	if err != nil {
		t.Fatal(err)
	}

	str := buffer.String()
	if str != "GET /test?test=123 HTTP/1.1\r\nHost: example.com\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 4\r\nDify-Hook-Id: 123\r\nDify-Hook-Url: http://example.com/e/123/test\r\n\r\ntest" {
		t.Fatal("request body is not equal, ", str)
	}
}

func TestEndpointWithHeaderDifyHookURLEmptyAndTLS(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.com/test?test=123", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.TLS = &tls.ConnectionState{}
	req.Header.Set(endpoint_entities.HeaderDifyHookURL, "")
	payload := "test"
	req.Body = io.NopCloser(bytes.NewReader([]byte(payload)))

	buffer, err := copyRequest(req, "123", "/test")
	if err != nil {
		t.Fatal(err)
	}

	str := buffer.String()
	if str != "GET /test?test=123 HTTP/1.1\r\nHost: example.com\r\nUser-Agent: Go-http-client/1.1\r\nContent-Length: 4\r\nDify-Hook-Id: 123\r\nDify-Hook-Url: https://example.com/e/123/test\r\n\r\ntest" {
		t.Fatal("request body is not equal, ", str)
	}
}
