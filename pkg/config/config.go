package config

import (
	"log"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/lib/pq"
)

// Config описывает общую конфигурацию сервиса GophKeeper, включая адрес сервера,
// параметры JWT, настройки БД и ключ шифрования пользовательских данных.
type Config struct {
	ServiceName   string        `yaml:"service_name"`
	Env           string        `yaml:"env"`
	ServerAddr    string        `yaml:"server_addr"`
	BaseURL       string        `yaml:"base_url"`
	JWTCookieName string        `yaml:"jwt_cookie_name"`
	JWTSecretKey  string        `yaml:"jwt_secret_key"`
	JWTTokenExp   time.Duration `yaml:"jwt_token_exp"`
	DBConfig      *DBConfig     `yaml:"db"`
	DataEncKey    string        `yaml:"data_enc_key"`
}

// DBConfig содержит настройки подключения к базе данных Postgres.
type DBConfig struct {
	DatabaseUri string `yaml:"database_uri" json:"database-uri"`
	PoolSize    int    `yaml:"pool_size" json:"poolSize"`
	DBTimeout   int    `yaml:"db_timeout" json:"db-timeout"`
}

var (
	cfgOnce   sync.Once
	globalCfg *Config
)

const configPath = "config.yml"

// InitConfig читает конфигурацию из YAML-файла и кэширует её для повторного использования.
func InitConfig() *Config {
	cfgOnce.Do(func() {
		var cfg Config

		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("cannot read config file %q: %v", configPath, err)
		}

		globalCfg = &cfg
	})

	return globalCfg
}
