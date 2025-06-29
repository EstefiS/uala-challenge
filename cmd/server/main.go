package server

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

// @title           Uala Challenge API
// @version         1.0
// @description     Esta es una API para una plataforma de microblogging similar a Twitter.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Estefania Sack
// @contact.url    https://github.com/EstefiS

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1
func main() {
	ctx := context.Background()
	cfg := configs.LoadConfig()
	log.Printf("Iniciando aplicación en modo: %s", cfg.AppEnv)

	var userRepo ports.UserRepository
	var tweetRepo ports.TweetRepository
	var timelineRepo ports.TimelineRepository

	if cfg.AppEnv == "prod" {
		log.Println("Usando configuración de producción: PostgreSQL + Redis Cache")
		dbpool, err := pgxpool.New(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("No se pudo conectar a PostgreSQL: %v", err)
		}
		defer dbpool.Close()

		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			log.Fatalf("No se pudo parsear la URL de Redis: %v", err)
		}
		redisClient := redis.NewClient(opt)
		if _, err := redisClient.Ping(ctx).Result(); err != nil {
			log.Fatalf("No se pudo conectar a Redis: %v", err)
		}

		postgresRepo := repository.NewPostgresRepository(dbpool)
		cachingRepo := repository.NewCachingRepository(redisClient, postgresRepo, postgresRepo, postgresRepo)

		userRepo = cachingRepo
		tweetRepo = cachingRepo
		timelineRepo = cachingRepo

	} else {
		log.Println("Usando configuración de desarrollo: Mock Repository en memoria")
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
	log.Printf("Servidor escuchando en %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("No se pudo iniciar el servidor: %v", err)
	}
}
