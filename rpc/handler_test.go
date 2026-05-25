package rpc_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/snicol/rf/rpc"
)

type testReq struct {
	Name string `json:"name"`
}

type testRes struct {
	Greeting string `json:"greeting"`
}

func newEchoHandler() *rpc.Handler[testReq, testRes] {
	return rpc.NewHandler(func(_ context.Context, req testReq) (testRes, error) {
		return testRes{Greeting: "hello " + req.Name}, nil
	}, nil)
}

func TestHandle_WithBody(t *testing.T) {
	h := newEchoHandler()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"world"}`))
	w := httptest.NewRecorder()

	if err := h.Handle()(w, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if body := w.Body.String(); !strings.Contains(body, "hello world") {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestHandle_EmptyBody(t *testing.T) {
	h := newEchoHandler()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	w := httptest.NewRecorder()

	// Empty body must not return EOF — zero-value Req should be used.
	if err := h.Handle()(w, req); err != nil {
		t.Fatalf("unexpected error for empty body: %v", err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandle_InvalidJSON(t *testing.T) {
	h := newEchoHandler()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{bad json`))
	w := httptest.NewRecorder()

	if err := h.Handle()(w, req); err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestHandle_MultipartJSONIsRejected(t *testing.T) {
	h := newEchoHandler()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"foo"}{"name":"bar"}`))
	w := httptest.NewRecorder()

	if err := h.Handle()(w, req); err == nil {
		t.Fatal("expected error for multipart JSON body, got nil")
	}
}

func TestError_NonYaelErrorIs500(t *testing.T) {
	h := newEchoHandler()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	h.Error()(w, req, errors.New("some internal error"))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if body := w.Body.String(); !strings.Contains(body, "unknown") {
		t.Fatalf("expected 'unknown' in body, got: %s", body)
	}
}

func TestHandle_ZeroResponseIs204(t *testing.T) {
	h := rpc.NewHandler(func(_ context.Context, _ testReq) (testRes, error) {
		return testRes{}, nil
	}, nil)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	if err := h.Handle()(w, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}
