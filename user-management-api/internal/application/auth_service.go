package application

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"sevensolutions-backend/internal/domain"
	"sevensolutions-backend/internal/ports"
	"sevensolutions-backend/pkg/jwt"
	"sevensolutions-backend/pkg/password"
)

type AuthService struct {
	repo       ports.UserRepository
	jwtService *jwt.Service
}

func NewAuthService(repo ports.UserRepository, jwtService *jwt.Service) *AuthService {
	return &AuthService{repo: repo, jwtService: jwtService}
}

// Register = ทั้ง registration และ create user ตาม spec, endpoint เดียวกัน
func (s *AuthService) Register(ctx context.Context, input domain.RegisterInput) (*domain.User, error) {
	existing, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailExists
	}

	hash, err := password.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:      input.Name,
		Email:     input.Email,
		Password:  hash,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		// unique index กันซ้ำจริง ๆ ตรงนี้แค่แปลง error ให้เป็น ErrEmailExists แทนที่จะหลุดเป็น 500
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	return user, nil
}

// error เดียวกันทั้งสองเคส กัน user enumeration
func (s *AuthService) Login(ctx context.Context, input domain.LoginInput) (string, error) {
	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if !password.Compare(user.Password, input.Password) {
		return "", ErrInvalidCredentials
	}

	return s.jwtService.Sign(user.ID.Hex())
}
