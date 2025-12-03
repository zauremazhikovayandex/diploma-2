package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"diploma-2/pkg/logger/message"
)

// TestNewInitializesGlobals проверяет, что New инициализирует глобальные переменные.
func TestNewInitializesGlobals(t *testing.T) {
	l := New(message.InfoLevel)
	if l == nil {
		t.Fatal("New returned nil logger")
	}
	if Log == nil {
		t.Fatal("global Log is nil after New")
	}
	if Logging == nil {
		t.Fatal("global Logging is nil after New")
	}
}

// TestWriter_WriteToLog проверяет, что метод WriteToLog не паникует.
func TestWriter_WriteToLog(t *testing.T) {
	New(message.InfoLevel)

	w := &Writer{}
	start := time.Now().Add(-time.Second)
	w.WriteToLog(start, "/test", http.MethodGet, http.StatusOK, "OK")
}

// TestRequestLogger проверяет, что middleware вызывает вложенный handler
// и пишет лог без паники.
func TestRequestLogger(t *testing.T) {
	New(message.InfoLevel)

	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})

	logged := RequestLogger(h)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	logged.ServeHTTP(rr, req)

	if !called {
		t.Fatal("inner handler was not called")
	}
	if rr.Code != http.StatusTeapot {
		t.Fatalf("unexpected status code: got %d, want %d", rr.Code, http.StatusTeapot)
	}
}
