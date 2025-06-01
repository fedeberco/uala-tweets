package repositories

import (
	"database/sql"
	"time"
	"uala-tweets/internal/domain"
)

type PostgreSQLTweetRepository struct {
	db *sql.DB
}

func NewPostgreSQLTweetRepository(db *sql.DB) *PostgreSQLTweetRepository {
	return &PostgreSQLTweetRepository{db: db}
}

func (r *PostgreSQLTweetRepository) Create(tweet *domain.Tweet) error {
	query := `
		INSERT INTO tweets (user_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	now := time.Now().UTC()
	err := r.db.QueryRow(
		query,
		tweet.UserID,
		tweet.Content,
		now,
		now,
	).Scan(&tweet.ID, &tweet.CreatedAt, &tweet.UpdatedAt)

	return err
}

func (r *PostgreSQLTweetRepository) GetByID(id int64) (*domain.Tweet, error) {
	query := `
		SELECT id, user_id, content, created_at, updated_at
		FROM tweets
		WHERE id = $1
	`

	tweet := &domain.Tweet{}
	err := r.db.QueryRow(query, id).Scan(
		&tweet.ID,
		&tweet.UserID,
		&tweet.Content,
		&tweet.CreatedAt,
		&tweet.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}

	return tweet, err
}

func (r *PostgreSQLTweetRepository) GetByUserID(userID int64) ([]*domain.Tweet, error) {
	query := `
		SELECT id, user_id, content, created_at, updated_at
		FROM tweets
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tweets := make([]*domain.Tweet, 0)
	for rows.Next() {
		tweet := &domain.Tweet{}
		err := rows.Scan(
			&tweet.ID,
			&tweet.UserID,
			&tweet.Content,
			&tweet.CreatedAt,
			&tweet.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tweets = append(tweets, tweet)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tweets, nil
}

func (r *PostgreSQLTweetRepository) GetTweetIDsByUser(userID int) ([]int64, error) {
	query := `
		SELECT id
		FROM tweets
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tweetIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		tweetIDs = append(tweetIDs, id)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tweetIDs, nil
}
