package mongodb

import (
	"context"
	"time"

	"tp25-api/internal/domain"
	"tp25-api/lib"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SettingRepository struct {
	collection *mongo.Collection
}

func NewSettingRepository(db *mongo.Database) *SettingRepository {
	return &SettingRepository{
		collection: db.Collection("settings"),
	}
}

func (r *SettingRepository) List(ctx context.Context) ([]domain.Setting, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var settings []domain.Setting
	if err := cursor.All(ctx, &settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *SettingRepository) ListWithPagination(ctx context.Context, pagination *domain.Pagination, filter bson.M) ([]domain.Setting, int64, error) {
	if filter == nil {
		filter = bson.M{}
	}

	// Get total count
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find with pagination
	opts := options.Find().
		SetSkip(int64(pagination.GetSkip())).
		SetLimit(int64(pagination.GetLimit())).
		SetSort(bson.D{{Key: "ctime", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var settings []domain.Setting
	if err := cursor.All(ctx, &settings); err != nil {
		return nil, 0, err
	}

	return settings, total, nil
}

func (r *SettingRepository) GetByID(ctx context.Context, id string) (*domain.Setting, error) {
	var setting domain.Setting
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&setting)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrSettingNotFound
		}
		return nil, err
	}
	return &setting, nil
}

func (r *SettingRepository) GetByKey(ctx context.Context, key string) (*domain.Setting, error) {
	var setting domain.Setting
	err := r.collection.FindOne(ctx, bson.M{"key": key}).Decode(&setting)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrSettingNotFound
		}
		return nil, err
	}
	return &setting, nil
}

func (r *SettingRepository) Create(ctx context.Context, params domain.CreateSettingParams) (*domain.Setting, error) {
	// Check if key already exists
	existing, err := r.GetByKey(ctx, params.Key)
	if err == nil && existing != nil {
		return nil, domain.ErrSettingKeyExists
	}

	now := time.Now().Unix()
	setting := domain.Setting{
		ID:    lib.Rand.Char(16),
		Key:   params.Key,
		Value: params.Value,
		CTime: now,
		MTime: now,
	}

	_, err = r.collection.InsertOne(ctx, setting)
	if err != nil {
		return nil, err
	}

	return &setting, nil
}

func (r *SettingRepository) Update(ctx context.Context, id string, params domain.UpdateSettingParams) (*domain.Setting, error) {
	update := bson.M{
		"$set": bson.M{
			"value": params.Value,
			"mtime": time.Now().Unix(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var setting domain.Setting
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, opts).Decode(&setting)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrSettingNotFound
		}
		return nil, err
	}

	return &setting, nil
}

func (r *SettingRepository) UpdateByKey(ctx context.Context, key string, params domain.UpdateSettingParams) (*domain.Setting, error) {
	update := bson.M{
		"$set": bson.M{
			"value": params.Value,
			"mtime": time.Now().Unix(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var setting domain.Setting
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"key": key}, update, opts).Decode(&setting)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrSettingNotFound
		}
		return nil, err
	}

	return &setting, nil
}

func (r *SettingRepository) Delete(ctx context.Context, id string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrSettingNotFound
	}
	return nil
}
