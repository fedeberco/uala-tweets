package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"uala-tweets/internal/application"
	"uala-tweets/internal/domain"
)

type PostgreSQLUserRepository struct {
	db *sql.DB
}

func NewPostgreSQLUserRepository(db *sql.DB) *PostgreSQLUserRepository {
	return &PostgreSQLUserRepository{db: db}
}

func (r *PostgreSQLUserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (username, created_at, updated_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		// Check for unique violation (duplicate username)
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_username_key\"" {
			return application.NewErrUserAlreadyExists(user.Username)
		}
		return err
	}

	return nil
}

func (r *PostgreSQLUserRepository) GetByID(id int) (*domain.User, error) {
	query := `
		SELECT id, username, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, application.NewErrUserNotFound(id)
		}
		return nil, err
	}

	return &user, nil
}

func (r *PostgreSQLUserRepository) Exists(id int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM users WHERE id = $1
		)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
