package application_test

import (
	"context"
	"errors"
	"testing"

	"sevensolutions-backend/internal/application"
	"sevensolutions-backend/internal/domain"
	"sevensolutions-backend/pkg/jwt"
)

func newAuthService() *application.AuthService {
	repo := newFakeUserRepository()
	jwtService := jwt.NewService("test-secret")
	return application.NewAuthService(repo, jwtService)
}

func TestRegister_Success(t *testing.T) {
	// Arrange
	service := newAuthService()
	input := domain.RegisterInput{Name: "John Doe", Email: "john@example.com", Password: "secret123"}

	// Act
	user, err := service.Register(context.Background(), input)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != input.Email {
		t.Errorf("expected email %q, got %q", input.Email, user.Email)
	}
	if user.Password == input.Password {
		t.Error("expected password to be hashed, got plaintext")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	// Arrange
	service := newAuthService()
	input := domain.RegisterInput{Name: "John Doe", Email: "john@example.com", Password: "secret123"}
	if _, err := service.Register(context.Background(), input); err != nil {
		t.Fatalf("setup: first register failed: %v", err)
	}

	// Act
	_, err := service.Register(context.Background(), input)

	// Assert
	if !errors.Is(err, application.ErrEmailExists) {
		t.Errorf("expected ErrEmailExists, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	// Arrange
	service := newAuthService()
	registerInput := domain.RegisterInput{Name: "John Doe", Email: "john@example.com", Password: "secret123"}
	if _, err := service.Register(context.Background(), registerInput); err != nil {
		t.Fatalf("setup: register failed: %v", err)
	}

	// Act
	token, err := service.Login(context.Background(), domain.LoginInput{Email: "john@example.com", Password: "secret123"})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Error("expected a non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	// Arrange
	service := newAuthService()
	registerInput := domain.RegisterInput{Name: "John Doe", Email: "john@example.com", Password: "secret123"}
	if _, err := service.Register(context.Background(), registerInput); err != nil {
		t.Fatalf("setup: register failed: %v", err)
	}

	// Act
	_, err := service.Login(context.Background(), domain.LoginInput{Email: "john@example.com", Password: "wrong-password"})

	// Assert
	if !errors.Is(err, application.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	// Arrange
	service := newAuthService()

	// Act
	_, err := service.Login(context.Background(), domain.LoginInput{Email: "nobody@example.com", Password: "secret123"})

	// Assert - error เดียวกับ password ผิด กัน user enumeration
	if !errors.Is(err, application.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
