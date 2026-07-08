package application

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"sevensolutions-backend/internal/domain"
	"sevensolutions-backend/internal/ports"
)

type UserService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetByID(ctx context.Context, idHex string) (*domain.User, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

type PagedUsers struct {
	Items []*domain.User
	Total int64
	Page  int
	Limit int
}

// page/limit ที่ผิด (0, ลบ, เกิน max) ให้ fallback เป็นค่า default แทน error
// เพราะ pagination เป็น bonus ที่ spec ไม่ได้ขอ ไม่อยาก over-engineer ด้วย 400 response
func (s *UserService) GetAll(ctx context.Context, page, limit int) (*PagedUsers, error) {
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 || limit > maxLimit {
		limit = defaultLimit
	}

	skip := int64(page-1) * int64(limit)

	items, err := s.repo.FindAll(ctx, skip, int64(limit))
	if err != nil {
		return nil, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, err
	}

	return &PagedUsers{Items: items, Total: total, Page: page, Limit: limit}, nil
}

// เช็คว่ามี field ส่งมาก่อนแตะ repository เลย
func (s *UserService) Update(ctx context.Context, idHex string, input domain.UpdateUserInput) (*domain.User, error) {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	if input.Name == nil && input.Email == nil {
		return nil, ErrNoFieldsToUpdate
	}

	if input.Email != nil {
		existing, err := s.repo.FindByEmail(ctx, *input.Email)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrEmailExists
		}
	}

	user, err := s.repo.Update(ctx, id, input)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		// unique index กันซ้ำจริง ๆ ตรงนี้แค่แปลง error ให้ handler map เป็น 409 ได้
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	return user, nil
}

func (s *UserService) Delete(ctx context.Context, idHex string) error {
	id, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return ErrInvalidUserID
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (s *UserService) Count(ctx context.Context) (int64, error) {
	return s.repo.Count(ctx)
}
