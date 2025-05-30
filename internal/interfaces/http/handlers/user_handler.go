package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"uala-tweets/internal/application"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *application.UserService
}

func NewUserHandler(userService *application.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.userService.CreateUser(application.CreateUserInput{
		Username: req.Username,
	})

	if err != nil {
		if errors.As(err, &application.ErrUserAlreadyExists{}) {
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user ID"})
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		if errors.As(err, &application.ErrUserNotFound{}) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user)
}
