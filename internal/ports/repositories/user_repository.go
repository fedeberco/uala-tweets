package repositories

import "uala-tweets/internal/domain"

type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id int) (*domain.User, error)
	Exists(id int) (bool, error)
	// TODO Update(user *domain.User) error
	// TODO Delete(id int) error
}
