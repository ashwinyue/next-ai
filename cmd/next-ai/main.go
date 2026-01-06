package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/handler"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/ashwinyue/next-ai/internal/router"
	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 加载配置
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./configs/config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	db, err := repository.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	log.Printf("Database connected: %s", cfg.Database.DBName)

	// 初始化 Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// 初始化各层
	repos := repository.NewRepositories(db.DB)
	services, err := service.NewServices(repos, cfg, redisClient)
	if err != nil {
		log.Fatalf("Failed to init services: %v", err)
	}
	handlers := handler.NewHandlers(services)

	// 初始化路由
	r := router.SetupRouter(handlers, services)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         cfg.Server.GetAddr(),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// 启动服务器
	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
