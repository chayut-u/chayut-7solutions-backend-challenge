package application_test

import (
	"context"
	"sort"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"sevensolutions-backend/internal/domain"
)

// fake ของ ports.UserRepository ทำให้ test ไม่ต้องพึ่ง MongoDB จริง
type fakeUserRepository struct {
	usersByID map[primitive.ObjectID]*domain.User
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{usersByID: make(map[primitive.ObjectID]*domain.User)}
}

func (r *fakeUserRepository) Create(_ context.Context, user *domain.User) error {
	user.ID = primitive.NewObjectID()
	stored := *user
	r.usersByID[user.ID] = &stored
	return nil
}

func (r *fakeUserRepository) FindByID(_ context.Context, id primitive.ObjectID) (*domain.User, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	found := *user
	return &found, nil
}

func (r *fakeUserRepository) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, user := range r.usersByID {
		if user.Email == email {
			found := *user
			return &found, nil
		}
	}
	return nil, mongo.ErrNoDocuments
}

func (r *fakeUserRepository) FindAll(_ context.Context, skip, limit int64) ([]*domain.User, error) {
	users := make([]*domain.User, 0, len(r.usersByID))
	for _, user := range r.usersByID {
		found := *user
		users = append(users, &found)
	}
	// map iteration order สุ่ม ต้อง sort เองให้ pagination ทดสอบได้แน่นอน
	sort.Slice(users, func(i, j int) bool { return users[i].ID.Hex() < users[j].ID.Hex() })

	start := min(int(skip), len(users))
	end := min(start+int(limit), len(users))
	return users[start:end], nil
}

func (r *fakeUserRepository) Update(_ context.Context, id primitive.ObjectID, input domain.UpdateUserInput) (*domain.User, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return nil, mongo.ErrNoDocuments
	}
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	found := *user
	return &found, nil
}

func (r *fakeUserRepository) Delete(_ context.Context, id primitive.ObjectID) error {
	if _, ok := r.usersByID[id]; !ok {
		return mongo.ErrNoDocuments
	}
	delete(r.usersByID, id)
	return nil
}

func (r *fakeUserRepository) Count(_ context.Context) (int64, error) {
	return int64(len(r.usersByID)), nil
}
