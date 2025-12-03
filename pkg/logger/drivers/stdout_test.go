package drivers

import (
	"diploma-2/pkg/logger/message"
	"testing"
)

// TestMakeStdoutLogger проверяет, что драйвер создаётся и умеет писать логи
// разных уровней без паники.
func TestMakeStdoutLogger(t *testing.T) {
	logger := MakeStdoutLogger(message.InfoLevel)
	if logger == nil {
		t.Fatal("MakeStdoutLogger returned nil")
	}

	msg := &message.LogMessage{Message: "test"}

	// Важно: мы не проверяем stdout, просто убеждаемся, что вызовы не паникуют.
	logger.Debug(msg)
	logger.Info(msg)
	logger.Warn(msg)
	logger.Error(msg)
	logger.Fatal(msg)
	logger.Panic(msg)
}
