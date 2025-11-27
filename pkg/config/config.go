package config

import (
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"sync"
	"time"
)

type Config struct {
	ServiceName        string
	Env                string
	ServerAddr         string
	BaseURL            string
	JWTCookieName      string
	JWTSecretKey       string
	JWTTokenExp        time.Duration
	DBConfig           *DBConfig
	AccrualBaseURL     string
	AccrualHTTPTimeout time.Duration
	AccrualRPM         int
	DataEncKey         string // ключ для шифрования пользовательских данных
}

type DBConfig struct {
	DatabaseUri string `json:"database-uri"`
	PoolSize    int    `json:"poolSize"`
	DBTimeout   int    `json:"db-timeout"`
}

var (
	baseURLFlag  = flag.String("a", "", "base URL for short links")
	dbConfigFlag = flag.String("d", "", "database URI")

	cfgOnce   sync.Once
	globalCfg *Config
)

// InitConfig собирается один раз
func InitConfig() *Config {
	cfgOnce.Do(func() {
		secret := os.Getenv("JWT_SECRET_KEY")
		if secret == "" {
			log.Fatal("ENV JWT_SECRET_KEY is required")
		}
		dataKey := os.Getenv("DATA_ENC_KEY")
		if dataKey == "" {
			dataKey = "dev-insecure-data-key"
		}

		// дефолты
		baseURL := "http://localhost:8080"
		dbUri := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			"localhost", 5432, "postgres", "postgres", "aviato",
		)

		// флаги (значения уже будут установлены, если где-то был flag.Parse(), например в main)
		if *baseURLFlag != "" {
			baseURL = *baseURLFlag
		}
		if *dbConfigFlag != "" {
			dbUri = *dbConfigFlag
		}

		// окружение
		if env := os.Getenv("RUN_ADDRESS"); env != "" {
			baseURL = env
		}
		if env := os.Getenv("DATABASE_URI"); env != "" {
			dbUri = env
		}

		dbConfig := DBConfig{
			DatabaseUri: dbUri,
			PoolSize:    50,
			DBTimeout:   5000,
		}

		cfg := &Config{
			ServerAddr:         ":8080",
			ServiceName:        "Diploma-1",
			Env:                "prod",
			BaseURL:            baseURL,
			JWTCookieName:      "auth_token",
			JWTSecretKey:       "supersecretkey",
			JWTTokenExp:        time.Hour * 720,
			DBConfig:           &dbConfig,
			AccrualBaseURL:     "http://accrual:8080",
			AccrualHTTPTimeout: 30 * time.Second,
			AccrualRPM:         60,
			DataEncKey:         dataKey,
		}

		// при желании можно переопределять секрет и окружение из env:
		if secret := os.Getenv("JWT_SECRET_KEY"); secret != "" {
			cfg.JWTSecretKey = secret
		}
		if env := os.Getenv("ENV"); env != "" {
			cfg.Env = env
		}

		globalCfg = cfg
	})

	return globalCfg
}
