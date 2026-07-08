package ports

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"sevensolutions-backend/internal/domain"
)

// port - application layer รู้จักแค่ interface นี้ ไม่รู้จัก storage จริงเลย
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindAll(ctx context.Context, skip, limit int64) ([]*domain.User, error)
	Update(ctx context.Context, id primitive.ObjectID, input domain.UpdateUserInput) (*domain.User, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}
