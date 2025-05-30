package handlers

import (
	"net/http"
	"strconv"
	"uala-tweets/internal/application"

	"github.com/gin-gonic/gin"
)

type TweetHandler struct {
	tweetService *application.TweetService
}

func NewTweetHandler(tweetService *application.TweetService) *TweetHandler {
	return &TweetHandler{tweetService: tweetService}
}

type CreateTweetRequest struct {
	UserID  int64  `json:"user_id" binding:"required"`
	Content string `json:"content" binding:"required,max=280"`
}

func (h *TweetHandler) CreateTweet(c *gin.Context) {
	var req CreateTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tweet, err := h.tweetService.CreateTweet(c.Request.Context(), application.CreateTweetInput{
		UserID:  req.UserID,
		Content: req.Content,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tweet)
}

func (h *TweetHandler) GetTweet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tweet id"})
		return
	}

	tweet, err := h.tweetService.GetTweet(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tweet not found"})
		return
	}

	c.JSON(http.StatusOK, tweet)
}

func (h *TweetHandler) GetUserTweets(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	tweets, err := h.tweetService.GetUserTweets(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tweets)
}
