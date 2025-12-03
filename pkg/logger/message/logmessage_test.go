package message

import "testing"

// TestLogMessageBasic проверяет, что структура LogMessage можно создать
// и уровни логирования заданы.
func TestLogMessageBasic(t *testing.T) {
	if InfoLevel == "" || ErrorLevel == "" {
		t.Fatalf("expected non-empty log levels, got Info=%q Error=%q", InfoLevel, ErrorLevel)
	}

	msg := &LogMessage{
		Message: "test",
	}
	if msg.Message != "test" {
		t.Fatalf("unexpected message: %q", msg.Message)
	}
}
