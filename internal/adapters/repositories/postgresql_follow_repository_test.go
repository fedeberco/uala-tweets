package repositories

import (
	"testing"
	"time"

	"uala-tweets/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestUsers(t *testing.T, repo *PostgreSQLUserRepository) (int, int) {
	// Create two test users
	user1 := &domain.User{
		Username:  "follower",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := repo.Create(user1)
	require.NoError(t, err)

	user2 := &domain.User{
		Username:  "followed",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err = repo.Create(user2)
	require.NoError(t, err)

	return user1.ID, user2.ID
}

func TestPostgreSQLFollowRepository_Follow(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	userRepo := NewPostgreSQLUserRepository(db)
	followRepo := NewPostgreSQLFollowRepository(db)

	followerID, followedID := setupTestUsers(t, userRepo)

	tests := []struct {
		name       string
		setup      func() error
		followerID int
		followedID int
		expectErr  bool
		expectFunc func(t *testing.T, err error)
	}{
		{
			name:       "successful follow",
			setup:      func() error { return nil },
			followerID: followerID,
			followedID: followedID,
			expectErr:  false,
			expectFunc: func(t *testing.T, err error) {
				isFollowing, err := followRepo.IsFollowing(followerID, followedID)
				require.NoError(t, err)
				assert.True(t, isFollowing)
			},
		},
		// TODO FIX TEST {
		// 	name: "duplicate follow",
		// 	setup: func() error {
		// 		// First follow should succeed
		// 		return followRepo.Follow(followerID, followedID)
		// 	},
		// 	followerID: followerID,
		// 	followedID: followedID,
		// 	expectErr:  true,
		// 	expectFunc: func(t *testing.T, err error) {
		// 		require.Error(t, err)
		// 		// Check that the error contains the expected message
		// 		assert.Contains(t, err.Error(), "already following")
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				err := tt.setup()
				require.NoError(t, err)
			}

			err := followRepo.Follow(tt.followerID, tt.followedID)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectFunc != nil {
				tt.expectFunc(t, err)
			}
		})
	}
}

func TestPostgreSQLFollowRepository_Unfollow(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	userRepo := NewPostgreSQLUserRepository(db)
	followRepo := NewPostgreSQLFollowRepository(db)

	followerID, followedID := setupTestUsers(t, userRepo)

	tests := []struct {
		name       string
		setup      func() error
		followerID int
		followedID int
		expectErr  bool
		expectFunc func(t *testing.T, err error)
	}{
		{
			name:       "successful unfollow",
			setup:      func() error { return followRepo.Follow(followerID, followedID) },
			followerID: followerID,
			followedID: followedID,
			expectErr:  false,
			expectFunc: func(t *testing.T, err error) {
				isFollowing, err := followRepo.IsFollowing(followerID, followedID)
				require.NoError(t, err)
				assert.False(t, isFollowing)
			},
		},
		{
			name:       "unfollow not following",
			setup:      func() error { return nil },
			followerID: followerID,
			followedID: followedID,
			expectErr:  true, // Expect error when not following
			expectFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				err := tt.setup()
				require.NoError(t, err)
			}

			err := followRepo.Unfollow(tt.followerID, tt.followedID)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectFunc != nil {
				tt.expectFunc(t, err)
			}
		})
	}
}

func TestPostgreSQLFollowRepository_IsFollowing(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	userRepo := NewPostgreSQLUserRepository(db)
	followRepo := NewPostgreSQLFollowRepository(db)

	followerID, followedID := setupTestUsers(t, userRepo)

	tests := []struct {
		name       string
		setup      func() error
		followerID int
		followedID int
		expected   bool
		expectErr  bool
	}{
		{
			name:       "not following",
			setup:      func() error { return nil },
			followerID: followerID,
			followedID: followedID,
			expected:   false,
			expectErr:  false,
		},
		{
			name:       "is following",
			setup:      func() error { return followRepo.Follow(followerID, followedID) },
			followerID: followerID,
			followedID: followedID,
			expected:   true,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				err := tt.setup()
				require.NoError(t, err)
			}

			result, err := followRepo.IsFollowing(tt.followerID, tt.followedID)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
