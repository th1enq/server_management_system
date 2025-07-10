package services

type PasswordService interface {
	Hash(password string) (string, error)
	Verify(hashedPassword, password string) (bool, error)
}
