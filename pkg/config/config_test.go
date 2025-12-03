package config

import (
	"testing"
)

// TestInitConfig проверяет, что конфигурация читается из config.yml
// и кешируется без паники.
func TestInitConfig(t *testing.T) {
	// Первый вызов должен прочитать config.yml
	cfg1 := InitConfig()
	if cfg1 == nil {
		t.Fatal("InitConfig returned nil")
	}
	if cfg1.ServiceName == "" {
		t.Errorf("ServiceName is empty")
	}
	if cfg1.DBConfig == nil {
		t.Errorf("DBConfig is nil")
	}

	// Второй вызов должен вернуть тот же указатель (singleton по cfgOnce)
	cfg2 := InitConfig()
	if cfg1 != cfg2 {
		t.Errorf("InitConfig should return same instance on repeated calls")
	}
}
