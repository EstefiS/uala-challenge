package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func ensureUserExistsTx(ctx context.Context, tx pgx.Tx, userID string) error {
	query := "INSERT INTO users (id, created_at) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING"
	_, err := tx.Exec(ctx, query, userID, time.Now())
	return err
}

// --- UserRepository ---
func (r *PostgresRepository) FollowTx(ctx context.Context, userID, userToFollowID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	// 1. Asegurarse de que ambos usuarios existen en la tabla de usuarios.
	batch.Queue("INSERT INTO users (id, created_at) VALUES ($1, NOW()) ON CONFLICT (id) DO NOTHING", userID)
	batch.Queue("INSERT INTO users (id, created_at) VALUES ($1, NOW()) ON CONFLICT (id) DO NOTHING", userToFollowID)

	// 2. Crear la relación de seguimiento.
	// La clave primaria previene duplicados, por lo que no necesitamos ON CONFLICT.
	batch.Queue("INSERT INTO followers (user_id, follower_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userToFollowID, userID)

	// --- INICIO DE LA CORRECCIÓN ---
	// 3. "Rellenar" (Backfill) el timeline del nuevo seguidor (userID) con los tweets más recientes
	//    del usuario al que ahora sigue (userToFollowID).
	//    La consulta ahora inserta en las columnas correctas: user_id, tweet_id, tweet_created_at.
	backfillQuery := `
		INSERT INTO timelines (user_id, tweet_id, tweet_created_at)
		SELECT $1, id, created_at
		FROM tweets
		WHERE user_id = $2
		ORDER BY created_at DESC
		LIMIT 50
		ON CONFLICT (user_id, tweet_id) DO NOTHING
	`
	batch.Queue(backfillQuery, userID, userToFollowID)
	// --- FIN DE LA CORRECCIÓN ---

	br := tx.SendBatch(ctx, batch)
	// Cerramos el batch para liberar recursos. Si hubo un error en el batch, br.Close() lo devolverá.
	if err := br.Close(); err != nil {
		// No es necesario hacer tx.Rollback(ctx) aquí, el defer ya se encarga de eso.
		return fmt.Errorf("error en el batch de seguimiento: %w", err)
	}

	// Si todo fue bien, hacemos commit de la transacción.
	return tx.Commit(ctx)
}
func (r *PostgresRepository) GetFollowers(ctx context.Context, userID string) ([]string, error) {
	query := "SELECT follower_id FROM followers WHERE user_id=$1"
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowTo[string])
}

// --- TweetRepository ---
func (r *PostgresRepository) PublishTx(ctx context.Context, tweet *domain.Tweet) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// --- INICIO DE LA CORRECCIÓN ---
	// 1. Asegurarnos de que el autor del tweet exista en la tabla 'users'.
	//    Si no existe, lo crea. Si ya existe, ON CONFLICT no hace nada.
	userQuery := "INSERT INTO users (id, created_at) VALUES ($1, NOW()) ON CONFLICT (id) DO NOTHING"
	if _, err := tx.Exec(ctx, userQuery, tweet.UserID); err != nil {
		return fmt.Errorf("error al asegurar la existencia del usuario autor: %w", err)
	}
	// --- FIN DE LA CORRECCIÓN ---

	// 2. Ahora sí, insertar el nuevo tweet en la tabla de tweets.
	tweetQuery := "INSERT INTO tweets (id, user_id, text, created_at) VALUES ($1, $2, $3, $4)"
	if _, err := tx.Exec(ctx, tweetQuery, tweet.ID, tweet.UserID, tweet.Text, tweet.CreatedAt); err != nil {
		return fmt.Errorf("error al insertar el tweet: %w", err)
	}

	// 3. Obtener la lista de todos los usuarios que siguen al autor del tweet.
	followersQuery := "SELECT follower_id FROM followers WHERE user_id = $1"
	rows, err := tx.Query(ctx, followersQuery, tweet.UserID)
	if err != nil {
		// Logueamos pero no fallamos la transacción, el tweet ya se guardó.
		log.Printf("Error al obtener seguidores para fan-out: %v", err)
	} else {
		followers, err := pgx.CollectRows(rows, pgx.RowTo[string])
		rows.Close() // Importante cerrar aquí
		if err != nil {
			log.Printf("Error al recolectar seguidores para fan-out: %v", err)
		} else if len(followers) > 0 {
			// 4. Si hay seguidores, distribuir el tweet a SUS timelines.
			batch := &pgx.Batch{}
			fanOutQuery := "INSERT INTO timelines (user_id, tweet_id, tweet_created_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING"
			for _, followerID := range followers {
				batch.Queue(fanOutQuery, followerID, tweet.ID, tweet.CreatedAt)
			}
			br := tx.SendBatch(ctx, batch)
			if err := br.Close(); err != nil {
				log.Printf("Error en el fan-out a los timelines de los seguidores: %v", err)
			}
		}
	}

	// 5. Si todo fue bien, hacemos commit.
	return tx.Commit(ctx)
}

// --- TimelineRepository ---
func (r *PostgresRepository) Get(ctx context.Context, userID string, limit int) ([]domain.Tweet, error) {
	query := `
        SELECT t.id, t.user_id, t.text, t.created_at
        FROM timelines tl JOIN tweets t ON tl.tweet_id = t.id
        WHERE tl.user_id = $1 ORDER BY tl.tweet_created_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Tweet])
}
