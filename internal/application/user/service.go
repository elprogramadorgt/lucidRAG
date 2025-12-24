package user

import (
	"context"
	"fmt"

	userDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"github.com/sirupsen/logrus"
)

type service struct {
	repo userDomain.Repository
}

func NewService(repo userDomain.Repository) userDomain.Service {
	return &service{repo: repo}
}

func (s *service) Register(ctx context.Context, user *userDomain.User) error {
	emailExists, err := s.repo.FindByEmail(ctx, user.Email)
	if err != nil {
		logrus.Errorf("Error checking email existence: %v", err)
		return err
	}
	if emailExists != nil {
		logrus.Infof("Attempt to register with existing email: %s", user.Email)
		return fmt.Errorf(userDomain.ErrEmailAlreadyInUse)
	}
	return s.repo.Create(ctx, user)
}
func (s *service) Login(ctx context.Context, email, password string) (*userDomain.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	// Here you would add logic to compare the hashed password
	// if user.Password != password {
	// 	return nil, ErrInvalidCredentials
	// }
	return user, nil
}
