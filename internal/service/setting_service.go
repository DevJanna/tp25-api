package service

import (
	"context"

	"tp25-api/internal/domain"
	"tp25-api/internal/repository/mongodb"

	"go.mongodb.org/mongo-driver/bson"
)

type SettingService struct {
	repo *mongodb.SettingRepository
}

func NewSettingService(repo *mongodb.SettingRepository) *SettingService {
	return &SettingService{repo: repo}
}

func (s *SettingService) List(ctx context.Context) ([]domain.Setting, error) {
	return s.repo.List(ctx)
}

func (s *SettingService) ListWithPagination(ctx context.Context, pagination *domain.Pagination, filter bson.M) ([]domain.Setting, int64, error) {
	return s.repo.ListWithPagination(ctx, pagination, filter)
}

func (s *SettingService) GetByID(ctx context.Context, id string) (*domain.Setting, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SettingService) GetByKey(ctx context.Context, key string) (*domain.Setting, error) {
	return s.repo.GetByKey(ctx, key)
}

func (s *SettingService) Create(ctx context.Context, params domain.CreateSettingParams) (*domain.Setting, error) {
	// Validate key
	if params.Key == "" {
		return nil, domain.ErrInvalidSettingKey
	}

	return s.repo.Create(ctx, params)
}

func (s *SettingService) Update(ctx context.Context, id string, params domain.UpdateSettingParams) (*domain.Setting, error) {
	return s.repo.Update(ctx, id, params)
}

func (s *SettingService) UpdateByKey(ctx context.Context, key string, params domain.UpdateSettingParams) (*domain.Setting, error) {
	return s.repo.UpdateByKey(ctx, key, params)
}

func (s *SettingService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
