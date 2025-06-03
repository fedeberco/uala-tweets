package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"uala-tweets/internal/application"

	"github.com/gin-gonic/gin"
)

// User represents a user in the system
type UserResponse struct {
	ID       int64  `json:"id" example:"123"`
	Username string `json:"username" example:"johndoe"`
}

// UserErrorResponse represents an error response for user operations
type UserErrorResponse struct {
	Error string `json:"error" example:"error message"`
}

type UserHandler struct {
	userService *application.UserService
}

func NewUserHandler(userService *application.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUserRequest represents the request body for creating a user
// swagger:parameters createUser
type CreateUserRequest struct {
	// Username must be unique and between 3-50 characters
	// required: true
	// example: johndoe
	Username string `json:"username" binding:"required,min=3,max=50"`
}

// CreateUser creates a new user
// @Summary      Create a new user
// @Description  Create a new user with the specified username
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      CreateUserRequest  true  "User to create"
// @Success      201  {object}  UserResponse
// @Failure      400  {object}  UserErrorResponse
// @Failure      409  {object}  UserErrorResponse
// @Failure      500  {object}  UserErrorResponse
// @Router       /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, UserErrorResponse{Error: err.Error()})
		return
	}

	user, err := h.userService.CreateUser(application.CreateUserInput{
		Username: req.Username,
	})

	if err != nil {
		if errors.As(err, &application.ErrUserAlreadyExists{}) {
			c.JSON(http.StatusConflict, UserErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, UserErrorResponse{Error: "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, UserResponse{
		ID:       int64(user.ID),
		Username: user.Username,
	})
}

// GetUser retrieves a user by ID
// @Summary      Get a user
// @Description  Get a user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  UserResponse
// @Failure      400  {object}  UserErrorResponse
// @Failure      404  {object}  UserErrorResponse
// @Failure      500  {object}  UserErrorResponse
// @Router       /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, UserErrorResponse{Error: "invalid user ID"})
		return
	}

	user, err := h.userService.GetUser(id)
	if err != nil {
		if errors.As(err, &application.ErrUserNotFound{}) {
			c.JSON(http.StatusNotFound, UserErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, UserErrorResponse{Error: "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user)
}
