package repositories

import (
	"testing"
	"time"

	"uala-tweets/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLUserRepository_CreateAndGet(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewPostgreSQLUserRepository(db)

	tests := []struct {
		name    string
		user    *domain.User
		wantErr bool
	}{
		{
			name: "create valid user",
			user: &domain.User{
				Username:  "testuser",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Create
			err := repo.Create(tt.user)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotZero(t, tt.user.ID)

			// Test GetByID
			found, err := repo.GetByID(tt.user.ID)
			require.NoError(t, err)
			require.NotNil(t, found)
			assert.Equal(t, tt.user.Username, found.Username)

			// Test Exists
			exists, err := repo.Exists(tt.user.ID)
			require.NoError(t, err)
			assert.True(t, exists)
		})
	}
}

func TestPostgreSQLUserRepository_NonExistentUser(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewPostgreSQLUserRepository(db)
	nonExistentUserID := 999

	tests := []struct {
		name     string
		testFunc func(*testing.T, *PostgreSQLUserRepository, int)
	}{
		{
			name: "get non-existent user",
			testFunc: func(t *testing.T, r *PostgreSQLUserRepository, id int) {
				user, err := r.GetByID(id)
				require.Error(t, err)
				assert.Nil(t, user)
			},
		},
		{
			name: "check non-existent user exists",
			testFunc: func(t *testing.T, r *PostgreSQLUserRepository, id int) {
				exists, err := r.Exists(id)
				require.NoError(t, err)
				assert.False(t, exists)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t, repo, nonExistentUserID)
		})
	}
}
