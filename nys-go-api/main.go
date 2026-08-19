package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nys-go-api/internal/cache"
	"nys-go-api/internal/config"
	"nys-go-api/internal/controller"
	"nys-go-api/internal/database"
	"nys-go-api/internal/repository"
	"nys-go-api/internal/security"
	"nys-go-api/internal/service"
	"nys-go-api/internal/session"
)

func main() {
	configPath := os.Getenv("NYS_CONFIG")
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal(err)
	}
	appLocation, err := database.ResolveLocation(cfg.Database.Location)
	if err != nil {
		log.Fatal(err)
	}
	// Docker 容器的系统时区通常是 UTC。显式设置 Go 进程时区，保证接口、RSS、SEO
	// 以及所有使用 time.Now/ParseInLocation 的业务都按同一北京时间工作。
	time.Local = appLocation

	db, err := database.ConnectMySQL(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	startupContext, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Database.ConnectTimeoutSeconds)*time.Second)
	defer cancel()
	store, err := cache.NewStore(startupContext, cfg.Redis)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	jwt := security.NewJWT(cfg.Security.JWTSecret, time.Duration(cfg.Security.JWTExpirationSeconds)*time.Second)
	aesCipher, err := security.NewAESCipher(cfg.Security.AESKey, cfg.Security.AESIV)
	if err != nil {
		log.Fatal(err)
	}
	repo := repository.NewSQLRepository(db)
	sessionManager := session.NewManager(store, cfg.Security)
	services := service.New(cfg, repo, store, sessionManager, jwt)
	router := controller.NewRouter(cfg, services, aesCipher)

	server := &http.Server{
		Addr:         cfg.Address(),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
	}

	go func() {
		log.Printf("NYS Go API 已启动: http://%s", cfg.Address())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP 服务异常退出: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeoutSec)*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf("HTTP 服务未能平滑关闭: %v", err)
	}
}
