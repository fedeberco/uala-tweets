package application

import (
	"errors"
	"fmt"
	"time"

	"uala-tweets/internal/domain"
	"uala-tweets/internal/ports/repositories"
)

type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

type CreateUserInput struct {
	Username string
}

func (s *UserService) CreateUser(input CreateUserInput) (*domain.User, error) {
	if input.Username == "" {
		return nil, &ErrInvalidInput{Message: "username cannot be empty"}
	}

	user := &domain.User{
		Username:  input.Username,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if user.ID == 0 {
		return nil, errors.New("user created but ID not set")
	}

	return user, nil
}

func (s *UserService) GetUser(id int) (*domain.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &ErrUserNotFound{UserID: id}
	}
	return user, nil
}

func (s *UserService) UserExists(id int) (bool, error) {
	return s.userRepo.Exists(id)
}
