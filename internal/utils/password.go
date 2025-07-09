package utils

import "golang.org/x/crypto/bcrypt"

func CheckPassword(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
