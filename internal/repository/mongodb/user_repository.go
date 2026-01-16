package mongodb

import (
	"context"
	"time"

	"tp25-api/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	db       *mongo.Database
	users    *mongo.Collection
	secrets  *mongo.Collection
	sessions *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		db:       db,
		users:    db.Collection("user"),
		secrets:  db.Collection("user_secret"),
		sessions: db.Collection("user_sessions"),
	}
}

func (r *UserRepository) GetUser(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := r.users.FindOne(ctx, bson.M{"_id": id, "dtime": bson.M{"$exists": false}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.users.FindOne(ctx, bson.M{"username": username, "dtime": bson.M{"$exists": false}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrUsernameNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	var user domain.User
	err := r.users.FindOne(ctx, bson.M{"phone": phone, "dtime": bson.M{"$exists": false}}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	cursor, err := r.users.Find(ctx, bson.M{"dtime": bson.M{"$exists": false}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) ListUsersWithPagination(ctx context.Context, pagination *domain.Pagination, filter bson.M) ([]domain.User, int64, error) {
	if filter == nil {
		filter = bson.M{}
	}
	filter["dtime"] = bson.M{"$exists": false}

	// Get total count
	total, err := r.users.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find with pagination
	opts := options.Find().
		SetSkip(int64(pagination.GetSkip())).
		SetLimit(int64(pagination.GetLimit())).
		SetSort(bson.D{{Key: "ctime", Value: -1}})

	cursor, err := r.users.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []domain.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	// Check if username already exists
	existing, err := r.GetUserByUsername(ctx, user.Username)
	if err == nil && existing != nil {
		return domain.ErrUsernameExisted
	}

	_, err = r.users.InsertOne(ctx, user)
	return err
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	user.MTime = time.Now().UnixMilli()
	_, err := r.users.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": user},
	)
	return err
}

func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	_, err := r.users.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"dtime": now}},
	)
	return err
}

// Auth-related methods

func (r *UserRepository) SaveUserSecret(ctx context.Context, secret *domain.UserSecret) error {
	opts := options.Update().SetUpsert(true)
	_, err := r.secrets.UpdateOne(
		ctx,
		bson.M{"user_id": secret.UserID, "name": secret.Name},
		bson.M{"$set": secret},
		opts,
	)
	return err
}

func (r *UserRepository) GetUserSecret(ctx context.Context, userID, name string) (*domain.UserSecret, error) {
	var secret domain.UserSecret
	err := r.secrets.FindOne(ctx, bson.M{"user_id": userID, "name": name}).Decode(&secret)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrUserHasNoLogin
		}
		return nil, err
	}
	return &secret, nil
}

func (r *UserRepository) SaveRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	_, err := r.sessions.InsertOne(ctx, token)
	return err
}

func (r *UserRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := r.sessions.FindOne(ctx, bson.M{"_id": token}).Decode(&rt)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrInvalidRefreshToken
		}
		return nil, err
	}
	return &rt, nil
}

func (r *UserRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.sessions.DeleteOne(ctx, bson.M{"_id": token})
	return err
}

func (r *UserRepository) DeleteRefreshTokensByUserID(ctx context.Context, userID string) error {
	_, err := r.sessions.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}
