package handlers

import (
	"net/http"
	"strconv"

	"uala-tweets/internal/application"

	"github.com/gin-gonic/gin"
)

type FollowHandler struct {
	followService *application.FollowService
}

func NewFollowHandler(followService *application.FollowService) *FollowHandler {
	return &FollowHandler{
		followService: followService,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *FollowHandler) FollowUser(c *gin.Context) {
	followerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid follower ID"})
		return
	}

	targetID, err := strconv.Atoi(c.Param("target_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid target user ID"})
		return
	}

	err = h.followService.Follow(followerID, targetID)
	if err != nil {
		switch err.Error() {
		case "follower user does not exist", "followed user does not exist":
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case "already following this user":
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully followed user",
	})
}

func (h *FollowHandler) UnfollowUser(c *gin.Context) {
	followerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid follower ID"})
		return
	}

	targetID, err := strconv.Atoi(c.Param("target_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid target user ID"})
		return
	}

	err = h.followService.Unfollow(followerID, targetID)
	if err != nil {
		switch err.Error() {
		case "follower user does not exist", "followed user does not exist":
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case "not currently following this user":
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully unfollowed user",
	})
}
