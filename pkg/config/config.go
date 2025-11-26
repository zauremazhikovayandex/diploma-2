package config

import (
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"os"
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
}

type DBConfig struct {
	DatabaseUri string `json:"database-uri"`
	PoolSize    int    `json:"poolSize"`
	DBTimeout   int    `json:"db-timeout"`
}

func InitConfig() *Config {
	// Парсим флаги во временные переменные
	baseURLFlag := flag.String("a", "", "base URL for short links")
	dbConfigFlag := flag.String("d", "", "base URL for short links")
	flag.Parse()

	// Устанавливаем значения по умолчанию
	baseURL := "http://localhost:8080"
	dbUri := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "postgres", "aviato",
	)

	if *baseURLFlag != "" {
		baseURL = *baseURLFlag
	}

	if *dbConfigFlag != "" {
		dbUri = *dbConfigFlag
	}

	// Окружением (имеет самый высокий приоритет)
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

	return &Config{
		ServerAddr:         ":8080",
		ServiceName:        fmt.Sprintf("Diploma-1"),
		Env:                "prod",
		BaseURL:            baseURL,
		JWTCookieName:      "auth_token",
		JWTSecretKey:       "supersecretkey",
		JWTTokenExp:        time.Hour * 720,
		DBConfig:           &dbConfig,
		AccrualBaseURL:     "http://accrual:8080",
		AccrualHTTPTimeout: 30 * time.Second,
		AccrualRPM:         60,
	}
}
