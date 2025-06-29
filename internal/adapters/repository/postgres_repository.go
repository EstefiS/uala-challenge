package repository

import (
	"context"
	"errors"
	"fmt"
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

	if err := ensureUserExistsTx(ctx, tx, userID); err != nil {
		return err
	}
	if err := ensureUserExistsTx(ctx, tx, userToFollowID); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO followers (user_id, follower_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userToFollowID, userID)
	if err != nil {
		return fmt.Errorf("error saving follow relationship: %w", err)
	}

	return tx.Commit(ctx)
}

// --- TweetRepository ---
func (r *PostgresRepository) PublishTx(ctx context.Context, tweet *domain.Tweet) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := ensureUserExistsTx(ctx, tx, tweet.UserID); err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO tweets (id, user_id, text, created_at) VALUES ($1, $2, $3, $4)",
		tweet.ID, tweet.UserID, tweet.Text, tweet.CreatedAt)
	if err != nil {
		return fmt.Errorf("error saving tweet: %w", err)
	}

	rows, err := tx.Query(ctx, "SELECT follower_id FROM followers WHERE user_id=$1", tweet.UserID)
	if err != nil {
		return fmt.Errorf("error getting followers: %w", err)
	}

	followerIDs, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return fmt.Errorf("error al recolectar seguidores: %w", err)
	}

	followerIDs = append(followerIDs, tweet.UserID)

	if len(followerIDs) > 0 {
		var b pgx.Batch
		query := "INSERT INTO timelines (user_id, tweet_id, tweet_created_at) VALUES ($1, $2, $3)"
		for _, followerID := range followerIDs {
			b.Queue(query, followerID, tweet.ID, tweet.CreatedAt)
		}
		br := tx.SendBatch(ctx, &b)
		if err := br.Close(); err != nil {
			return fmt.Errorf("error in bulk timeline insertion: %w", err)
		}
	}

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
