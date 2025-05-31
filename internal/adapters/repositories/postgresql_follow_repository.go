package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type PostgreSQLFollowRepository struct {
	db *sql.DB
}

func NewPostgreSQLFollowRepository(db *sql.DB) *PostgreSQLFollowRepository {
	return &PostgreSQLFollowRepository{db: db}
}

func (r *PostgreSQLFollowRepository) Follow(followerID, followedID int) error {
	if followerID == followedID {
		return errors.New("cannot follow yourself")
	}

	query := `
		INSERT INTO follows (follower_id, followed_id, created_at)
		VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(
		ctx,
		query,
		followerID,
		followedID,
		time.Now().UTC(),
	)

	if err != nil {
		// Check for duplicate key violation (already following)
		if err.Error() == "pq: duplicate key value violates unique constraint \"follows_pkey\"" {
			return errors.New("already following this user")
		}
		return err
	}

	return nil
}

func (r *PostgreSQLFollowRepository) Unfollow(followerID, followedID int) error {
	query := `
		DELETE FROM follows
		WHERE follower_id = $1 AND followed_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, query, followerID, followedID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("not following this user")
	}

	return nil
}

func (r *PostgreSQLFollowRepository) GetFollowers(userID int) ([]int, error) {
	query := `SELECT follower_id FROM follows WHERE followed_id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var followers []int
	for rows.Next() {
		var followerID int
		if err := rows.Scan(&followerID); err != nil {
			return nil, err
		}
		followers = append(followers, followerID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return followers, nil
}

func (r *PostgreSQLFollowRepository) IsFollowing(followerID, followedID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM follows 
			WHERE follower_id = $1 AND followed_id = $2
		)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	err := r.db.QueryRowContext(ctx, query, followerID, followedID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
