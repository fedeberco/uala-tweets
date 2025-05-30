package application

import (
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
	user := &domain.User{
		Username:  input.Username,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUser(id int) (*domain.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *UserService) UserExists(id int) (bool, error) {
	return s.userRepo.Exists(id)
}
