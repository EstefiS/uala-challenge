package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type PostgresRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgresRepository(db *pgxpool.Pool, logger *slog.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger.With("component", "PostgresRepository"),
	}
}

func (r *PostgresRepository) FollowTx(ctx context.Context, userID, userToFollowID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	userInsertQuery := "INSERT INTO users (id, created_at) VALUES ($1, NOW()) ON CONFLICT (id) DO NOTHING"
	batch.Queue(userInsertQuery, userID)
	batch.Queue(userInsertQuery, userToFollowID)

	followerInsertQuery := "INSERT INTO followers (user_id, follower_id) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	batch.Queue(followerInsertQuery, userToFollowID, userID)

	backfillQuery := `
		INSERT INTO timelines (user_id, tweet_id, tweet_created_at)
		SELECT $1, id, created_at FROM tweets WHERE user_id = $2
		ORDER BY created_at DESC LIMIT 50
		ON CONFLICT (user_id, tweet_id) DO NOTHING`
	batch.Queue(backfillQuery, userID, userToFollowID)

	br := tx.SendBatch(ctx, batch)
	if err := br.Close(); err != nil {
		return fmt.Errorf("error in follow batch transaction: %w", err)
	}

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

func (r *PostgresRepository) PublishTx(ctx context.Context, tweet *domain.Tweet) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	userQuery := "INSERT INTO users (id, created_at) VALUES ($1, NOW()) ON CONFLICT (id) DO NOTHING"
	if _, err := tx.Exec(ctx, userQuery, tweet.UserID); err != nil {
		return fmt.Errorf("error ensuring author user existence: %w", err)
	}

	tweetQuery := "INSERT INTO tweets (id, user_id, text, created_at) VALUES ($1, $2, $3, $4)"
	if _, err := tx.Exec(ctx, tweetQuery, tweet.ID, tweet.UserID, tweet.Text, tweet.CreatedAt); err != nil {
		return fmt.Errorf("error inserting tweet: %w", err)
	}

	followersQuery := "SELECT follower_id FROM followers WHERE user_id = $1"
	rows, err := tx.Query(ctx, followersQuery, tweet.UserID)
	if err != nil {
		return fmt.Errorf("error getting followers for fan-out: %w", err)
	}

	followers, err := pgx.CollectRows(rows, pgx.RowTo[string])
	rows.Close()
	if err != nil {
		return fmt.Errorf("error collecting followers for fan-out: %w", err)
	}

	if len(followers) > 0 {
		batch := &pgx.Batch{}
		fanOutQuery := "INSERT INTO timelines (user_id, tweet_id, tweet_created_at) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING"
		for _, followerID := range followers {
			batch.Queue(fanOutQuery, followerID, tweet.ID, tweet.CreatedAt)
		}
		br := tx.SendBatch(ctx, batch)
		if err := br.Close(); err != nil {
			return fmt.Errorf("error in fan-out to follower timelines: %w", err)
		}
	}

	return tx.Commit(ctx)
}

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
