package main

import (
	"context"
	"log"

	"github.com/EstefiS/uala-challenge/configs"
	"github.com/EstefiS/uala-challenge/internal/adapters/http"
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
func main() {
	ctx := context.Background()
	cfg := configs.LoadConfig()
	// Starting application in mode: %s
	log.Printf("Starting application in mode: %s", cfg.AppEnv)

	var userRepo ports.UserRepository
	var tweetRepo ports.TweetRepository
	var timelineRepo ports.TimelineRepository

	if cfg.AppEnv == "prod" {
		// Using production configuration: PostgreSQL + Redis Cache
		log.Println("Using production configuration: PostgreSQL + Redis Cache")
		dbpool, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			// Could not connect to PostgreSQL: %v
			log.Fatalf("Could not connect to PostgreSQL: %v", err)
		}
		defer dbpool.Close()

		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			// Could not parse the Redis URL: %v
			log.Fatalf("Could not parse the Redis URL: %v", err)
		}
		redisClient := redis.NewClient(opt)
		if _, err := redisClient.Ping(ctx).Result(); err != nil {
			// Could not connect to Redis: %v
			log.Fatalf("Could not connect to Redis: %v", err)
		}

		postgresRepo := repository.NewPostgresRepository(dbpool)
		cachingRepo := repository.NewCachingRepository(redisClient, postgresRepo, postgresRepo, postgresRepo)

		userRepo = cachingRepo
		tweetRepo = cachingRepo
		timelineRepo = cachingRepo

	} else {
		// Using development configuration: In-memory Mock Repository
		log.Println("Using development configuration: In-memory Mock Repository")
		mockRepo := repository.NewMockRepository()
		userRepo = mockRepo
		tweetRepo = mockRepo
		timelineRepo = mockRepo
	}

	tweetSvc := services.NewTweetService(tweetRepo)
	followSvc := services.NewFollowService(userRepo)
	timelineSvc := services.NewTimelineService(timelineRepo)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	httpHandler := http.NewGinHandler(tweetSvc, followSvc, timelineSvc)
	httpHandler.SetupRoutes(router)

	serverAddr := ":" + cfg.Port
	// Server listening on %s
	log.Printf("Server listening on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		// Could not start the server: %v
		log.Fatalf("Could not start the server: %v", err)
	}
}
