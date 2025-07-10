package services

import (
	"github.com/th1enq/server_management_system/internal/domain/services"
	"golang.org/x/crypto/bcrypt"
)

type bcryptService struct{}

func NewBcryptService() services.PasswordService {
	return &bcryptService{}
}

func (b *bcryptService) Hash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func (b *bcryptService) Verify(hashedPassword, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, err
	}
	return true, nil
}
