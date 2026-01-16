package service

import (
	"context"
	"time"

	"tp25-api/internal/domain"
	"tp25-api/internal/repository/mongodb"
	"tp25-api/lib"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      *mongodb.UserRepository
	jwtSecret string
}

func NewUserService(repo *mongodb.UserRepository, jwtSecret string) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.repo.GetUserByUsername(ctx, username)
}

func (s *UserService) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return s.repo.GetUserByPhone(ctx, phone)
}

func (s *UserService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.ListUsers(ctx)
}

func (s *UserService) ListUsersWithPagination(ctx context.Context, pagination *domain.Pagination, filter bson.M) ([]domain.User, int64, error) {
	return s.repo.ListUsersWithPagination(ctx, pagination, filter)
}

func (s *UserService) CreateUser(ctx context.Context, params domain.CreateUserParams) (*domain.User, error) {
	user := domain.NewUser(params)
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, params domain.UpdateUserParams) (*domain.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	if params.FullName != nil {
		user.FullName = *params.FullName
	}
	if params.Phone != nil {
		user.Phone = *params.Phone
	}
	if params.Groups != nil {
		user.Groups = params.Groups
	}
	if params.ZaloID != nil {
		user.ZaloID = params.ZaloID
	}

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteUser(ctx, id); err != nil {
		return nil, err
	}

	return user, nil
}

// Authentication methods

func (s *UserService) SetPassword(ctx context.Context, userID, password string) error {
	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	secret := &domain.UserSecret{
		UserID: userID,
		Name:   "password",
		Value:  string(hashedPassword),
		Encode: "bcrypt",
	}

	return s.repo.SaveUserSecret(ctx, secret)
}

func (s *UserService) Login(ctx context.Context, username, password string) (*domain.User, string, error) {
	// Get user by username
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, "", err
	}

	// Get user password secret
	secret, err := s.repo.GetUserSecret(ctx, user.ID, "password")
	if err != nil {
		return nil, "", err
	}

	// Compare passwords with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(secret.Value), []byte(password)); err != nil {
		return nil, "", domain.ErrWrongPassword
	}

	// Create refresh token record (7 days)
	now := time.Now().UnixMilli()
	refreshTokenID := lib.Rand.Char(12)
	refreshTokenRecord := &domain.RefreshToken{
		ID:        refreshTokenID,
		UserID:    user.ID,
		ExpiresAt: now + (7 * 24 * 60 * 60 * 1000),
		CTime:     now,
	}

	if err := s.repo.SaveRefreshToken(ctx, refreshTokenRecord); err != nil {
		return nil, "", err
	}

	// Generate JWT refresh token
	refreshTokenClaims := jwt.RegisteredClaims{
		Subject:   user.ID,
		ID:        refreshTokenID,
		ExpiresAt: jwt.NewNumericDate(time.Unix(refreshTokenRecord.ExpiresAt/1000, 0)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, "", err
	}

	return user, refreshTokenString, nil
}

func (s *UserService) RefreshToken(ctx context.Context, tokenString string) (*domain.User, string, error) {
	// Parse and validate JWT refresh token
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, "", domain.ErrInvalidRefreshToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, "", domain.ErrInvalidRefreshToken
	}

	// Get refresh token from DB
	rt, err := s.repo.GetRefreshToken(ctx, claims.ID)
	if err != nil {
		return nil, "", err
	}

	// Check if expired
	if time.Now().UnixMilli() > rt.ExpiresAt {
		_ = s.repo.DeleteRefreshToken(ctx, rt.ID)
		return nil, "", domain.ErrInvalidRefreshToken
	}

	// Verify user ID matches
	if claims.Subject != rt.UserID {
		return nil, "", domain.ErrInvalidRefreshToken
	}

	// Get user
	user, err := s.repo.GetUser(ctx, rt.UserID)
	if err != nil {
		return nil, "", err
	}

	// Delete old refresh token
	_ = s.repo.DeleteRefreshToken(ctx, rt.ID)

	// Create new refresh token record (7 days)
	now := time.Now().UnixMilli()
	newRefreshTokenID := lib.Rand.Char(12)
	newRefreshTokenRecord := &domain.RefreshToken{
		ID:        newRefreshTokenID,
		UserID:    user.ID,
		ExpiresAt: now + (7 * 24 * 60 * 60 * 1000),
		CTime:     now,
	}

	if err := s.repo.SaveRefreshToken(ctx, newRefreshTokenRecord); err != nil {
		return nil, "", err
	}

	// Generate new JWT refresh token
	newRefreshTokenClaims := jwt.RegisteredClaims{
		Subject:   user.ID,
		ID:        newRefreshTokenID,
		ExpiresAt: jwt.NewNumericDate(time.Unix(newRefreshTokenRecord.ExpiresAt/1000, 0)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	newRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newRefreshTokenClaims)
	newRefreshTokenString, err := newRefreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, "", err
	}

	return user, newRefreshTokenString, nil
}

func (s *UserService) Logout(ctx context.Context, userID string) error {
	return s.repo.DeleteRefreshTokensByUserID(ctx, userID)
}
