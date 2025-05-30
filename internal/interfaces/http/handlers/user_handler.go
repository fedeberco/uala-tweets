package handlers

import (
	"net/http"
	"strconv"
	"time"

	"uala-tweets/internal/domain"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo UserRepository
}

type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id int) (*domain.User, error)
}

func NewUserHandler(userRepo UserRepository) *UserHandler {
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
