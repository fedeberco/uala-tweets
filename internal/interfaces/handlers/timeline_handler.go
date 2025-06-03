package handlers

import (
	"net/http"
	"strconv"

	"uala-tweets/internal/application"

	"github.com/gin-gonic/gin"
)

// TimelineResponse represents the response for the timeline endpoint
type TimelineResponse struct {
	UserID   int     `json:"user_id" example:"123"`
	TweetIDs []int64 `json:"tweet_ids"`
}

// TimelineErrorResponse represents an error response for timeline operations
type TimelineErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

// TimelineHandler handles timeline-related HTTP requests
type TimelineHandler struct {
	service *application.TimelineService
}

// NewTimelineHandler creates a new TimelineHandler
func NewTimelineHandler(service *application.TimelineService) *TimelineHandler {
	return &TimelineHandler{service: service}
}

// GetTimelineHandler retrieves a user's timeline
// @Summary      Get user timeline
// @Description  Get a paginated list of tweet IDs from users that the specified user follows
// @Tags         timeline
// @Accept       json
// @Produce      json
// @Param        user_id path int true "User ID"
// @Param        limit query int false "Maximum number of tweets to return (default 10)"
// @Success      200  {object}  TimelineResponse
// @Failure      400  {object}  TimelineErrorResponse
// @Failure      500  {object}  TimelineErrorResponse
// @Router       /timeline/{user_id} [get]
func (h *TimelineHandler) GetTimelineHandler(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, TimelineErrorResponse{Error: "invalid user_id"})
		return
	}
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	ids, err := h.service.GetTimeline(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, TimelineErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, TimelineResponse{
		UserID:   userID,
		TweetIDs: ids,
	})
}
