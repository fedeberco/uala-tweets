package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"uala-tweets/internal/application"
	"uala-tweets/internal/domain"
	"uala-tweets/internal/ports/repositories"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo repositories.UserRepository
}

func NewUserHandler(userRepo repositories.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
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

	user := &domain.User{
		Username:  req.Username,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := h.userRepo.Create(user); err != nil {
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

	user, err := h.userRepo.GetByID(id)
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
