package handlers

import (
	"net/http"
	"strconv"
	"time"

	"uala-tweets/internal/application"

	"github.com/gin-gonic/gin"
)

// TweetResponse represents a tweet in the system
type TweetResponse struct {
	ID        int64     `json:"id" example:"123"`
	UserID    int64     `json:"user_id" example:"456"`
	Content   string    `json:"content" example:"Hello, world!"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TweetErrorResponse represents an error response for tweet operations
type TweetErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

type TweetHandler struct {
	tweetService *application.TweetService
}

func NewTweetHandler(tweetService *application.TweetService) *TweetHandler {
	return &TweetHandler{tweetService: tweetService}
}

// CreateTweetRequest represents the request body for creating a tweet
// swagger:parameters createTweet
type CreateTweetRequest struct {
	// ID of the user creating the tweet
	// required: true
	// example: 123
	UserID  int64  `json:"user_id" binding:"required"`
	
	// Content of the tweet (max 280 characters)
	// required: true
	// example: Hello, this is my first tweet!
	// max length: 280
	Content string `json:"content" binding:"required,max=280"`
}

// CreateTweet creates a new tweet
// @Summary      Create a new tweet
// @Description  Create a new tweet with the specified content
// @Tags         tweets
// @Accept       json
// @Produce      json
// @Param        tweet  body      CreateTweetRequest  true  "Tweet to create"
// @Success      201  {object}  TweetResponse
// @Failure      400  {object}  TweetErrorResponse
// @Failure      500  {object}  TweetErrorResponse
// @Router       /tweets [post]
func (h *TweetHandler) CreateTweet(c *gin.Context) {
	var req CreateTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TweetErrorResponse{Error: err.Error()})
		return
	}

	tweet, err := h.tweetService.CreateTweet(c.Request.Context(), application.CreateTweetInput{
		UserID:  req.UserID,
		Content: req.Content,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, TweetErrorResponse{Error: err.Error()})
		return
	}

	response := TweetResponse{
		ID:        tweet.ID,
		UserID:    tweet.UserID,
		Content:   tweet.Content,
		CreatedAt: tweet.CreatedAt,
		UpdatedAt: tweet.UpdatedAt,
	}
	c.JSON(http.StatusCreated, response)
}

// GetTweet retrieves a tweet by ID
// @Summary      Get a tweet
// @Description  Get a tweet by its ID
// @Tags         tweets
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Tweet ID"
// @Success      200  {object}  TweetResponse
// @Failure      400  {object}  TweetErrorResponse
// @Failure      404  {object}  TweetErrorResponse
// @Router       /tweets/{id} [get]
func (h *TweetHandler) GetTweet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, TweetErrorResponse{Error: "invalid tweet id"})
		return
	}

	tweet, err := h.tweetService.GetTweet(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, TweetErrorResponse{Error: "tweet not found"})
		return
	}

	response := TweetResponse{
		ID:        tweet.ID,
		UserID:    tweet.UserID,
		Content:   tweet.Content,
		CreatedAt: tweet.CreatedAt,
		UpdatedAt: tweet.UpdatedAt,
	}
	c.JSON(http.StatusOK, response)
}

// GetUserTweets retrieves all tweets for a specific user
// @Summary      Get user tweets
// @Description  Get all tweets for a specific user
// @Tags         tweets
// @Accept       json
// @Produce      json
// @Param        user_id   path      int  true  "User ID"
// @Success      200  {array}   TweetResponse
// @Failure      400  {object}  TweetErrorResponse
// @Failure      500  {object}  TweetErrorResponse
// @Router       /users/{user_id}/tweets [get]
func (h *TweetHandler) GetUserTweets(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, TweetErrorResponse{Error: "invalid user id"})
		return
	}

	tweets, err := h.tweetService.GetUserTweets(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, TweetErrorResponse{Error: err.Error()})
		return
	}

	response := make([]TweetResponse, len(tweets))
	for i, tweet := range tweets {
		response[i] = TweetResponse{
			ID:        tweet.ID,
			UserID:    tweet.UserID,
			Content:   tweet.Content,
			CreatedAt: tweet.CreatedAt,
			UpdatedAt: tweet.UpdatedAt,
		}
	}
	c.JSON(http.StatusOK, response)
}
