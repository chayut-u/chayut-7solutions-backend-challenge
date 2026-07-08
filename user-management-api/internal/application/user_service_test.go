package application_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"sevensolutions-backend/internal/application"
	"sevensolutions-backend/internal/domain"
	"sevensolutions-backend/pkg/jwt"
)

func newUserServiceWithUser(t *testing.T) (*application.UserService, *domain.User) {
	t.Helper()

	repo := newFakeUserRepository()
	authService := application.NewAuthService(repo, jwt.NewService("test-secret"))

	user, err := authService.Register(context.Background(), domain.RegisterInput{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("setup: register failed: %v", err)
	}

	return application.NewUserService(repo), user
}

func TestGetByID_NotFound(t *testing.T) {
	// Arrange
	service, _ := newUserServiceWithUser(t)
	unknownID := "64b1f2c3d4e5f6a7b8c9d0e1"

	// Act
	_, err := service.GetByID(context.Background(), unknownID)

	// Assert
	if !errors.Is(err, application.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestGetByID_InvalidObjectID(t *testing.T) {
	// Arrange
	service, _ := newUserServiceWithUser(t)

	// Act
	_, err := service.GetByID(context.Background(), "not-a-valid-object-id")

	// Assert
	if !errors.Is(err, application.ErrInvalidUserID) {
		t.Errorf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestGetByID_Success(t *testing.T) {
	// Arrange
	service, user := newUserServiceWithUser(t)

	// Act
	found, err := service.GetByID(context.Background(), user.ID.Hex())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Email != user.Email {
		t.Errorf("expected email %q, got %q", user.Email, found.Email)
	}
}

func TestGetAll_Pagination(t *testing.T) {
	// Arrange
	repo := newFakeUserRepository()
	authService := application.NewAuthService(repo, jwt.NewService("test-secret"))
	for i := range 5 {
		_, err := authService.Register(context.Background(), domain.RegisterInput{
			Name:     fmt.Sprintf("User %d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: "secret123",
		})
		if err != nil {
			t.Fatalf("setup: register failed: %v", err)
		}
	}
	service := application.NewUserService(repo)

	// Act
	result, err := service.GetAll(context.Background(), 1, 2)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("expected 2 items on page 1 with limit 2, got %d", len(result.Items))
	}
	if result.Total != 5 {
		t.Errorf("expected total 5, got %d", result.Total)
	}
}

func TestGetAll_DefaultsOnInvalidParams(t *testing.T) {
	// Arrange
	service, _ := newUserServiceWithUser(t)

	// Act
	result, err := service.GetAll(context.Background(), 0, 0)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Page != 1 {
		t.Errorf("expected page to default to 1, got %d", result.Page)
	}
	if result.Limit != 20 {
		t.Errorf("expected limit to default to 20, got %d", result.Limit)
	}
}

func TestUpdate_NoFieldsProvided(t *testing.T) {
	// Arrange
	service, user := newUserServiceWithUser(t)

	// Act
	_, err := service.Update(context.Background(), user.ID.Hex(), domain.UpdateUserInput{})

	// Assert
	if !errors.Is(err, application.ErrNoFieldsToUpdate) {
		t.Errorf("expected ErrNoFieldsToUpdate, got %v", err)
	}
}

func TestUpdate_Success(t *testing.T) {
	// Arrange
	service, user := newUserServiceWithUser(t)
	newName := "Jane Doe"

	// Act
	updated, err := service.Update(context.Background(), user.ID.Hex(), domain.UpdateUserInput{Name: &newName})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updated.Name)
	}
}

func TestDelete_NotFound(t *testing.T) {
	// Arrange
	service, _ := newUserServiceWithUser(t)
	unknownID := "64b1f2c3d4e5f6a7b8c9d0e1"

	// Act
	err := service.Delete(context.Background(), unknownID)

	// Assert
	if !errors.Is(err, application.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestDelete_Success(t *testing.T) {
	// Arrange
	service, user := newUserServiceWithUser(t)

	// Act
	err := service.Delete(context.Background(), user.ID.Hex())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := service.GetByID(context.Background(), user.ID.Hex()); !errors.Is(err, application.ErrUserNotFound) {
		t.Error("expected user to be gone after delete")
	}
}
