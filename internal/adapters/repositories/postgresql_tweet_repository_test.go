package repositories

import (
	"testing"
	"time"

	"uala-tweets/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgreSQLTweetRepository_CreateAndGet(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Create a user first since tweets require a user
	userRepo := NewPostgreSQLUserRepository(db)
	user := &domain.User{
		Username:  "testuser",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	repo := NewPostgreSQLTweetRepository(db)

	tests := []struct {
		name    string
		tweet   *domain.Tweet
		wantErr bool
	}{
		{
			name: "create valid tweet",
			tweet: &domain.Tweet{
				UserID:  int64(user.ID),
				Content: "Hello, world!",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Create
			err := repo.Create(tt.tweet)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotZero(t, tt.tweet.ID)
			assert.False(t, tt.tweet.CreatedAt.IsZero())
			assert.False(t, tt.tweet.UpdatedAt.IsZero())

			// Test GetByID
			found, err := repo.GetByID(tt.tweet.ID)
			require.NoError(t, err)
			require.NotNil(t, found)
			assert.Equal(t, tt.tweet.UserID, found.UserID)
			assert.Equal(t, tt.tweet.Content, found.Content)
		})
	}
}

func TestPostgreSQLTweetRepository_GetByUserID(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Create a user
	userRepo := NewPostgreSQLUserRepository(db)
	user := &domain.User{
		Username:  "testuser",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	repo := NewPostgreSQLTweetRepository(db)

	// Create test tweets
	tweets := []*domain.Tweet{
		{UserID: int64(user.ID), Content: "First tweet"},
		{UserID: int64(user.ID), Content: "Second tweet"},
	}

	for _, tweet := range tweets {
		err := repo.Create(tweet)
		require.NoError(t, err)
	}

	// Test GetByUserID
	found, err := repo.GetByUserID(int64(user.ID))
	require.NoError(t, err)
	assert.Len(t, found, 2)

	// Verify tweet contents
	contents := make(map[string]bool)
	for _, tweet := range found {
		contents[tweet.Content] = true
	}
	assert.True(t, contents["First tweet"])
	assert.True(t, contents["Second tweet"])
}

func TestPostgreSQLTweetRepository_GetTweetIDsByUser(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	// Create a user
	userRepo := NewPostgreSQLUserRepository(db)
	user := &domain.User{
		Username:  "testuser",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	err := userRepo.Create(user)
	require.NoError(t, err)

	repo := NewPostgreSQLTweetRepository(db)

	// Create test tweets
	tweets := []*domain.Tweet{
		{UserID: int64(user.ID), Content: "First tweet"},
		{UserID: int64(user.ID), Content: "Second tweet"},
	}

	var tweetIDs []int64
	for _, tweet := range tweets {
		err := repo.Create(tweet)
		require.NoError(t, err)
		tweetIDs = append(tweetIDs, tweet.ID)
	}

	// Test GetTweetIDsByUser
	foundIDs, err := repo.GetTweetIDsByUser(user.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, tweetIDs, foundIDs)
}

func TestPostgreSQLTweetRepository_NonExistentTweet(t *testing.T) {
	db, teardown := setupTestDB(t)
	defer teardown()

	repo := NewPostgreSQLTweetRepository(db)

	// Test GetByID with non-existent tweet
	tweet, err := repo.GetByID(999)
	require.Error(t, err)
	assert.Nil(t, tweet)

	// Test GetByUserID with no tweets
	tweets, err := repo.GetByUserID(999)
	require.NoError(t, err)
	assert.Empty(t, tweets)

	// Test GetTweetIDsByUser with no tweets
	tweetIDs, err := repo.GetTweetIDsByUser(999)
	require.NoError(t, err)
	assert.Empty(t, tweetIDs)
}
