package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"sevensolutions-backend/internal/domain"
)

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(collection *mongo.Collection) *MongoUserRepository {
	return &MongoUserRepository{collection: collection}
}

// unique index กัน email ซ้ำ เป็นด่านสุดท้ายถ้า race หลุดผ่าน check ชั้น application
func EnsureIndexes(ctx context.Context, collection *mongo.Collection) error {
	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := collection.Indexes().CreateMany(ctx, indexModels)
	return err
}

func (r *MongoUserRepository) Create(ctx context.Context, user *domain.User) error {
	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *MongoUserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MongoUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MongoUserRepository) FindAll(ctx context.Context, skip, limit int64) ([]*domain.User, error) {
	// sort ตาม _id คงที่เสมอ ไม่งั้น Mongo ไม่การันตีลำดับข้าม page
	opts := options.Find().SetSkip(skip).SetLimit(limit).SetSort(bson.D{{Key: "_id", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// เริ่มจาก slice ว่าง ไม่ใช่ nil เพื่อให้ JSON ออกเป็น [] ไม่ใช่ null
	users := []*domain.User{}
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *MongoUserRepository) Update(ctx context.Context, id primitive.ObjectID, input domain.UpdateUserInput) (*domain.User, error) {
	set := bson.M{}
	if input.Name != nil {
		set["name"] = *input.Name
	}
	if input.Email != nil {
		set["email"] = *input.Email
	}

	after := options.After
	opts := options.FindOneAndUpdate().SetReturnDocument(after)

	var user domain.User
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": set}, opts).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *MongoUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *MongoUserRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
