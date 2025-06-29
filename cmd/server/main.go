package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/EstefiS/uala-challenge/configs"
	httpAdapter "github.com/EstefiS/uala-challenge/internal/adapters/http"
	"github.com/EstefiS/uala-challenge/internal/adapters/repository"
	"github.com/EstefiS/uala-challenge/internal/core/ports"
	"github.com/EstefiS/uala-challenge/internal/core/services"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	_ "github.com/EstefiS/uala-challenge/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Uala Challenge - Microblogging API
// @version         1.0
// @description     This is an API for a microblogging platform, similar to Twitter, built with Go and Hexagonal Architecture..
// @termsOfService  http://swagger.io/terms/

// @contact.name   Estefania Sack
// @contact.url    https://github.com/EstefiS/uala-challenge
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1
func setupDependencies(ctx context.Context, cfg *configs.Config, logger *slog.Logger) (ports.UserRepository, ports.TweetRepository, ports.TimelineRepository) {
	if cfg.AppEnv == "prod" {
		logger.Info("Using production configuration: PostgreSQL + Redis Cache")

		dbpool, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			logger.Error("Could not connect to PostgreSQL", "error", err)
			os.Exit(1)
		}

		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			logger.Error("Could not parse the Redis URL", "error", err)
			os.Exit(1)
		}
		redisClient := redis.NewClient(opt)
		if _, err := redisClient.Ping(ctx).Result(); err != nil {
			logger.Error("Could not connect to Redis", "error", err)
			os.Exit(1)
		}

		postgresRepo := repository.NewPostgresRepository(dbpool, logger)
		cachingRepo := repository.NewCachingRepository(redisClient, postgresRepo, postgresRepo, postgresRepo, logger)

		return cachingRepo, cachingRepo, cachingRepo
	}

	logger.Info("Using development configuration: In-memory Mock Repository")
	mockRepo := repository.NewMockRepository()
	return mockRepo, mockRepo, mockRepo
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx := context.Background()
	cfg := configs.LoadConfig()
	logger.Info("Starting application", "environment", cfg.AppEnv)

	userRepo, tweetRepo, timelineRepo := setupDependencies(ctx, cfg, logger)

	tweetSvc := services.NewTweetService(tweetRepo)
	followSvc := services.NewFollowService(userRepo)
	timelineSvc := services.NewTimelineService(timelineRepo)

	apiDeps := httpAdapter.HandlerDependencies{
		TweetSvc:    tweetSvc,
		FollowSvc:   followSvc,
		TimelineSvc: timelineSvc,
		Logger:      logger,
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	httpHandler := httpAdapter.NewGinHandler(apiDeps)
	httpHandler.SetupRoutes(router)

	serverAddr := ":" + cfg.Port
	logger.Info("Server listening", "address", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		logger.Error("Could not start the server", "error", err)
		os.Exit(1)
	}
}
