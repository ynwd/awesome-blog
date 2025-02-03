package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ynwd/awesome-blog/internal/users/domain"
)

// MockUserRepository is a mock implementation of UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByUsernameAndPassword(ctx context.Context, username string, password string) (domain.User, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserRepository) IsUsernameExists(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func TestNewUserService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.(*userService).repo)
}

func TestCreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		user := domain.User{
			Username: "testuser",
			Password: "testpass",
		}

		mockRepo.On("IsUsernameExists", ctx, user.Username).Return(false, nil)
		mockRepo.On("Create", ctx, user).Return(nil)

		err := service.CreateUser(ctx, user)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty Username", func(t *testing.T) {
		user := domain.User{
			Password: "testpass",
		}

		err := service.CreateUser(ctx, user)

		assert.Equal(t, ErrInvalidInput, err)
		mockRepo.AssertNotCalled(t, "IsUsernameExists")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Empty Password", func(t *testing.T) {
		user := domain.User{
			Username: "testuser",
		}

		err := service.CreateUser(ctx, user)

		assert.Equal(t, ErrInvalidInput, err)
		mockRepo.AssertNotCalled(t, "IsUsernameExists")
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Username Already Exists", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		service := NewUserService(mockRepo)
		ctx := context.Background()

		user := domain.User{
			Username: "testuser",
			Password: "testpass",
		}

		// Setup mock expectation
		mockRepo.On("IsUsernameExists", ctx, user.Username).
			Return(true, nil).
			Once()

		err := service.CreateUser(ctx, user)

		assert.Equal(t, ErrUsernameExists, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthenticateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewUserService(mockRepo)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		expectedUser := domain.User{
			Id:       "123",
			Username: "testuser",
			Password: "testpass",
		}

		mockRepo.On("GetByUsernameAndPassword", ctx, "testuser", "testpass").Return(expectedUser, nil)

		user, err := service.AuthenticateUser(ctx, "testuser", "testpass")

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty Username", func(t *testing.T) {
		user, err := service.AuthenticateUser(ctx, "", "testpass")

		assert.Equal(t, ErrInvalidInput, err)
		assert.Empty(t, user)
		mockRepo.AssertNotCalled(t, "GetByUsernameAndPassword")
	})

	t.Run("Empty Password", func(t *testing.T) {
		user, err := service.AuthenticateUser(ctx, "testuser", "")

		assert.Equal(t, ErrInvalidInput, err)
		assert.Empty(t, user)
		mockRepo.AssertNotCalled(t, "GetByUsernameAndPassword")
	})
}

func TestValidateUser(t *testing.T) {
	t.Run("Valid User", func(t *testing.T) {
		user := domain.User{
			Username: "testuser",
			Password: "testpass",
		}

		err := validateUser(user)
		assert.NoError(t, err)
	})

	t.Run("Empty Username", func(t *testing.T) {
		user := domain.User{
			Password: "testpass",
		}

		err := validateUser(user)
		assert.Equal(t, ErrInvalidInput, err)
	})

	t.Run("Empty Password", func(t *testing.T) {
		user := domain.User{
			Username: "testuser",
		}

		err := validateUser(user)
		assert.Equal(t, ErrInvalidInput, err)
	})
}
