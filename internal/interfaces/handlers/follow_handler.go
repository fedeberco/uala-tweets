package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"uala-tweets/internal/application"

	"github.com/gin-gonic/gin"
)

// FollowResponse represents a success response for follow operations
type FollowResponse struct {
	Message string `json:"message" example:"successfully followed user"`
}

// FollowErrorResponse represents an error response for follow operations
type FollowErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

type FollowHandler struct {
	followService *application.FollowService
}

func NewFollowHandler(followService *application.FollowService) *FollowHandler {
	return &FollowHandler{
		followService: followService,
	}
}

// FollowRequest represents the request body for follow operations
type FollowRequest struct {
	// ID of the target user to follow/unfollow
	// required: true
	// example: 123
	TargetID int `json:"target_id" binding:"required"`
}

// FollowUser follows another user
// @Summary      Follow a user
// @Description  Follow another user by their ID
// @Tags         follows
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "Follower User ID"
// @Param        target_id  path  int  true  "Target User ID to follow"
// @Success      200  {object}  FollowResponse
// @Failure      400  {object}  FollowErrorResponse
// @Failure      404  {object}  FollowErrorResponse
// @Failure      409  {object}  FollowErrorResponse
// @Failure      500  {object}  FollowErrorResponse
// @Router       /users/{id}/follow/{target_id} [post]
func (h *FollowHandler) FollowUser(c *gin.Context) {
	followerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, FollowErrorResponse{Error: "invalid follower ID"})
		return
	}

	targetID, err := strconv.Atoi(c.Param("target_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, FollowErrorResponse{Error: "invalid target user ID"})
		return
	}

	err = h.followService.Follow(followerID, targetID)
	if err != nil {
		switch {
		case errors.As(err, &application.ErrUserNotFound{}):
			c.JSON(http.StatusNotFound, FollowErrorResponse{Error: err.Error()})
		case errors.As(err, &application.ErrAlreadyFollowing{}):
			c.JSON(http.StatusConflict, FollowErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, FollowErrorResponse{Error: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, FollowResponse{
		Message: "successfully followed user",
	})
}

// UnfollowUser unfollows a user
// @Summary      Unfollow a user
// @Description  Unfollow a user by their ID
// @Tags         follows
// @Accept       json
// @Produce      json
// @Param        id    path      int  true  "Follower User ID"
// @Param        target_id  path  int  true  "Target User ID to unfollow"
// @Success      200  {object}  FollowResponse
// @Failure      400  {object}  FollowErrorResponse
// @Failure      404  {object}  FollowErrorResponse
// @Failure      500  {object}  FollowErrorResponse
// @Router       /users/{id}/unfollow/{target_id} [post]
func (h *FollowHandler) UnfollowUser(c *gin.Context) {
	followerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, FollowErrorResponse{Error: "invalid follower ID"})
		return
	}

	targetID, err := strconv.Atoi(c.Param("target_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, FollowErrorResponse{Error: "invalid target user ID"})
		return
	}

	err = h.followService.Unfollow(followerID, targetID)
	if err != nil {
		switch {
		case errors.As(err, &application.ErrUserNotFound{}):
			c.JSON(http.StatusNotFound, FollowErrorResponse{Error: err.Error()})
		case errors.As(err, &application.ErrNotFollowing{}):
			c.JSON(http.StatusBadRequest, FollowErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, FollowErrorResponse{Error: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, FollowResponse{
		Message: "successfully unfollowed user",
	})
}
