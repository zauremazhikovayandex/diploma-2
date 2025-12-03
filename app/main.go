package main

import (
	"context"
	"diploma-2/internal/auth"
	"diploma-2/internal/handlers"
	"diploma-2/pkg/config"
	"diploma-2/pkg/db"
	"diploma-2/pkg/logger"
	"diploma-2/pkg/logger/message"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func waitForShutdown(srv *http.Server, closeDB func()) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Gracefully shutting down...")

	// Останавливаем HTTP-сервер с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		// Если graceful не успел — принудительно закрываем
		fmt.Printf("HTTP server shutdown error: %v\n", err)
		if cerr := srv.Close(); cerr != nil {
			fmt.Printf("HTTP server close error: %v\n", cerr)
		}
	}

	// Закрываем БД
	if closeDB != nil {
		closeDB()
	}
}

func addrFromBaseURL(raw, fallback string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return fallback // например ":8080"
	}
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return net.JoinHostPort(host, port)
}

func main() {

	// Init Config
	flag.Parse()
	cfg := config.InitConfig()

	// Init logger
	_ = logger.New("info")

	// Init Postgres
	instance, err := db.SqlInstance(cfg.DBConfig)
	if err != nil {
		log.Fatal("DB prepare issues", err)
	}
	db.PrepareDB(instance)

	api := handlers.New(handlers.Deps{
		Cfg: cfg,
		DB:  instance,
	})

	// Start app
	gin.SetMode(gin.ReleaseMode)
	rout := gin.New()
	rout.Use(gin.Recovery())
	rout.Use(auth.MiddlewareAuth(cfg))
	rout.GET("health", api.GetHealthCheck)
	rout.POST("/api/user/register", api.PostRegister)
	rout.POST("/api/user/login", api.PostLogin)

	// passwords
	rout.POST("/api/v1/passwords", api.PostPassword)
	rout.GET("/api/v1/passwords/:id", api.GetPasswordHandler)
	rout.GET("/api/v1/passwords", api.GetPasswordsListHandler)

	addr := addrFromBaseURL(cfg.BaseURL, cfg.ServerAddr)
	Srv := &http.Server{
		Addr:    addr,
		Handler: rout,
	}

	// Gracefully shutting down
	go waitForShutdown(Srv, func() {
		if instance == nil {
			logger.Log.Error(&message.LogMessage{Message: "DB instance is nil on shutdown"})
			return
		}
		instance.CloseSqlInstance()
	})

	logger.Log.Info(&message.LogMessage{Message: fmt.Sprintf("Starting server on port %s", cfg.ServerAddr)})
	if err := Srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
