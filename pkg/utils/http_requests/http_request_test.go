package http_requests

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestPropagatesContextCancellation(t *testing.T) {
	requestCanceled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
		close(requestCanceled)
	}))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithCancel(context.Background())
	response, err := Request(server.Client(), server.URL, http.MethodGet, HttpContext(ctx))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	t.Cleanup(func() {
		response.Body.Close()
	})

	cancel()
	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("server request did not observe context cancellation")
	}
}

func TestRequestPropagatesContextCancellationBeforeResponseHeaders(t *testing.T) {
	requestStarted := make(chan struct{})
	requestCanceled := make(chan struct{})
	releaseHandler := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		close(requestStarted)
		select {
		case <-r.Context().Done():
			close(requestCanceled)
		case <-releaseHandler:
		}
	}))
	t.Cleanup(func() {
		close(releaseHandler)
		server.Close()
	})

	ctx, cancel := context.WithCancel(context.Background())
	requestDone := make(chan error, 1)
	go func() {
		_, err := Request(
			server.Client(),
			server.URL,
			http.MethodPost,
			HttpPayloadReader(io.NopCloser(bytes.NewReader([]byte("{}")))),
			HttpContext(ctx),
		)
		requestDone <- err
	}()

	<-requestStarted
	cancel()

	select {
	case <-requestCanceled:
	case <-time.After(time.Second):
		t.Fatal("server request did not observe context cancellation")
	}
	select {
	case err := <-requestDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context cancellation, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("request did not return after context cancellation")
	}
}
