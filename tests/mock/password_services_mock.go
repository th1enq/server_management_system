package mock

import "github.com/stretchr/testify/mock"

type PasswordServicesMock struct {
	mock.Mock
}

func (m *PasswordServicesMock) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *PasswordServicesMock) Verify(hashedPassword, password string) (bool, error) {
	args := m.Called(hashedPassword, password)
	return args.Bool(0), args.Error(1)
}
