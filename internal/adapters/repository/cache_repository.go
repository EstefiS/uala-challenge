package repository

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
	"github.com/EstefiS/uala-challenge/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

type CachingRepository struct {
	redisClient      *redis.Client
	nextUserRepo     ports.UserRepository     // Repositorio subyacente (PostgreSQL)
	nextTweetRepo    ports.TweetRepository    // Repositorio subyacente (PostgreSQL)
	nextTimelineRepo ports.TimelineRepository // Repositorio subyacente (PostgreSQL)
	ttl              time.Duration
}

// NewCachingRepository crea una nueva instancia del decorator de caché.
func NewCachingRepository(
	client *redis.Client,
	userRepo ports.UserRepository,
	tweetRepo ports.TweetRepository,
	timelineRepo ports.TimelineRepository,
) *CachingRepository {
	return &CachingRepository{
		redisClient:      client,
		nextUserRepo:     userRepo,
		nextTweetRepo:    tweetRepo,
		nextTimelineRepo: timelineRepo,
		ttl:              2 * time.Minute,
	}
}

// --- Implementación de TimelineRepository (con caché) ---

func (r *CachingRepository) Get(ctx context.Context, userID string, limit int) ([]domain.Tweet, error) {
	cacheKey := "timeline:" + userID

	val, err := r.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		log.Printf("Cache HIT para el timeline del usuario: %s", userID)
		var timeline []domain.Tweet
		if json.Unmarshal([]byte(val), &timeline) == nil {
			return timeline, nil
		}
	}

	if err != redis.Nil {
		log.Printf("Error de Redis (no es 'no encontrado'), continuando a la DB: %v", err)
	}

	log.Printf("Cache MISS para el timeline del usuario: %s. Consultando la DB.", userID)
	timeline, err := r.nextTimelineRepo.Get(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	if len(timeline) > 0 {
		data, marshalErr := json.Marshal(timeline)
		if marshalErr == nil {
			r.redisClient.Set(ctx, cacheKey, data, r.ttl)
		}
	}

	return timeline, nil
}

// --- Implementación de TweetRepository (con invalidación de caché) ---

func (r *CachingRepository) PublishTx(ctx context.Context, tweet *domain.Tweet) error {
	// 1. Ejecutar la operación en la base de datos primero.
	err := r.nextTweetRepo.PublishTx(ctx, tweet)
	if err != nil {
		return err
	}

	// 2. Si la escritura fue exitosa, invalidar las cachés de los seguidores.
	followers, err := r.nextUserRepo.GetFollowers(ctx, tweet.UserID)
	if err != nil {
		log.Printf("Error al obtener seguidores para invalidar caché: %v", err)
		return nil // No devolvemos error, la operación principal fue exitosa.
	}

	// --- INICIO DE LA CORRECCIÓN ---
	// Si no hay seguidores, no hay nada que invalidar.
	if len(followers) == 0 {
		return nil
	}

	// Invalidamos la caché SOLAMENTE para los seguidores.
	pipe := r.redisClient.Pipeline()
	for _, followerID := range followers {
		cacheKey := "timeline:" + followerID
		pipe.Del(ctx, cacheKey)
	}
	// --- FIN DE LA CORRECCIÓN ---

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		log.Printf("Error al ejecutar pipeline de invalidación de caché en Redis: %v", err)
	}

	log.Printf("Caché invalidada para %d timelines de seguidores.", len(followers))

	return nil
}

// --- Implementación de UserRepository ---

func (r *CachingRepository) FollowTx(ctx context.Context, userID, userToFollowID string) error {
	err := r.nextUserRepo.FollowTx(ctx, userID, userToFollowID)
	if err == nil {
		log.Printf("Invalidando caché del timeline para el nuevo seguidor: %s", userID)
		r.redisClient.Del(ctx, "timeline:"+userID)
	}
	return err
}

// --- INICIO DE LA CORRECCIÓN ---
// GetFollowers simplemente pasa la llamada al repositorio subyacente.
// No tiene lógica de caché, pero es necesario para cumplir con la interfaz UserRepository.
func (r *CachingRepository) GetFollowers(ctx context.Context, userID string) ([]string, error) {
	return r.nextUserRepo.GetFollowers(ctx, userID)
}
