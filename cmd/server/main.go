package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sotremont/taskflow/internal/cache"
	"github.com/sotremont/taskflow/internal/config"
	"github.com/sotremont/taskflow/internal/repository"
	"github.com/sotremont/taskflow/internal/service"
	"github.com/sotremont/taskflow/internal/transport/http"
	_ "github.com/sotremont/taskflow/docs"
)

// @title TaskFlow API
// @version 1.0
// @description This is a task management server.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	cfg := config.LoadConfig()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// DB connection
	dbConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB)
	
	pool, err := pgxpool.New(context.Background(), dbConnStr)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}

	// Repositories
	userRepo := repository.NewUserRepository(pool)
	teamRepo := repository.NewTeamRepository(pool)
	taskRepo := repository.NewTaskRepository(pool)
	eventRepo := repository.NewEventRepository(pool)
	commentRepo := repository.NewCommentRepository(pool)
	redisCache := cache.NewRedisCache(rdb)

	// Services
	authService := service.NewAuthService(userRepo, teamRepo, cfg.JWTSecret)
	teamService := service.NewTeamService(teamRepo, userRepo, redisCache)
	taskService := service.NewTaskService(taskRepo, teamRepo, eventRepo, userRepo)
	commentService := service.NewCommentService(commentRepo, taskRepo, teamRepo)

	// Handlers
	authHandler := http.NewAuthHandler(authService)
	teamHandler := http.NewTeamHandler(teamService)
	taskHandler := http.NewTaskHandler(taskService)
	commentHandler := http.NewCommentHandler(commentService)

	// Server
	server := http.NewServer(cfg, authHandler, teamHandler, taskHandler, commentHandler)

	logger.Info("starting server", "port", cfg.ServerPort)
	if err := server.Run(":" + cfg.ServerPort); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
