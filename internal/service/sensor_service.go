package service

import (
	"context"

	"tp25-api/internal/domain"
	"tp25-api/internal/repository/mongodb"
	"tp25-api/lib/interpolation"

	"go.mongodb.org/mongo-driver/bson"
)

type SensorService struct {
	repo       *mongodb.SensorRepository
	zoneRepo   *mongodb.ZoneRepository
	calculator *interpolation.HydraulicCalculator
}

func NewSensorService(repo *mongodb.SensorRepository, zoneRepo *mongodb.ZoneRepository) *SensorService {
	return &SensorService{
		repo:       repo,
		zoneRepo:   zoneRepo,
		calculator: interpolation.NewHydraulicCalculator(),
	}
}

// Metric operations

func (s *SensorService) ListMetrics(ctx context.Context) ([]domain.Metric, error) {
	return s.repo.ListMetrics(ctx)
}

func (s *SensorService) ListMetricsWithPagination(ctx context.Context, pagination *domain.Pagination, filter bson.M) ([]domain.Metric, int64, error) {
	return s.repo.ListMetricsWithPagination(ctx, pagination, filter)
}

func (s *SensorService) GetMetric(ctx context.Context, filter bson.M) (*domain.Metric, error) {
	return s.repo.GetMetric(ctx, filter)
}

func (s *SensorService) CreateMetric(ctx context.Context, params domain.CreateMetricParams) (*domain.Metric, error) {
	if params.Code == "" {
		return nil, domain.ErrMetricMustHaveCode
	}

	metric := domain.NewMetric(params)
	if err := s.repo.CreateMetric(ctx, metric); err != nil {
		return nil, err
	}
	return metric, nil
}

func (s *SensorService) UpdateMetric(ctx context.Context, id string, params domain.UpdateMetricParams) (*domain.Metric, error) {
	metric, err := s.repo.GetMetric(ctx, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}

	if params.Unit != nil {
		metric.Unit = *params.Unit
	}
	if params.Code != nil {
		metric.Code = *params.Code
	}
	if params.Name != nil {
		metric.Name = *params.Name
	}
	if params.Range != nil {
		metric.Range = params.Range
	}

	if err := s.repo.UpdateMetric(ctx, metric); err != nil {
		return nil, err
	}

	return metric, nil
}

func (s *SensorService) DeleteMetric(ctx context.Context, id string) (*domain.Metric, error) {
	metric, err := s.repo.GetMetric(ctx, bson.M{"_id": id})
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteMetric(ctx, id); err != nil {
		return nil, err
	}

	return metric, nil
}

// Record operations

func (s *SensorService) ListRecords(ctx context.Context, boxID string, query *domain.QueryRecord) (*domain.RecordsResult, error) {
	return s.repo.ListRecords(ctx, boxID, query)
}

func (s *SensorService) CountRecords(ctx context.Context, boxID string, query *domain.QueryRecord) (int64, error) {
	return s.repo.CountRecords(ctx, boxID, query)
}

func (s *SensorService) AddRecord(ctx context.Context, boxID string, record domain.Record) error {
	// Apply interpolation calculations if needed
	record = s.applyInterpolation(record)
	return s.repo.AddRecord(ctx, boxID, record)
}

func (s *SensorService) ImportRecord(ctx context.Context, boxID string, record domain.Record) error {
	// Apply interpolation calculations if needed
	record = s.applyInterpolation(record)
	return s.repo.ImportRecord(ctx, boxID, record)
}

func (s *SensorService) ReportRecords(ctx context.Context, boxID string, query *domain.QueryRecord) ([]domain.DailyReport, error) {
	return s.repo.ReportRecords(ctx, boxID, query)
}

func (s *SensorService) ListRecordsByGroup(ctx context.Context, groupID string, query *domain.QueryRecord) (*domain.RecordsResult, error) {
	filter := domain.FilterBoxParams{GroupID: &groupID}
	boxes, err := s.zoneRepo.ListBoxes(ctx, filter)
	if err != nil {
		return nil, err
	}

	var boxIDs []string
	for _, box := range boxes {
		boxIDs = append(boxIDs, box.ID)
	}

	return s.repo.ListRecordsByGroup(ctx, boxIDs, query)
}

func (s *SensorService) ListRecordsLatestByGroup(ctx context.Context, groupID string) (*domain.RecordsResult, error) {
	filter := domain.FilterBoxParams{GroupID: &groupID}
	boxes, err := s.zoneRepo.ListBoxes(ctx, filter)
	if err != nil {
		return nil, err
	}

	var boxIDs []string
	for _, box := range boxes {
		boxIDs = append(boxIDs, box.ID)
	}

	return s.repo.ListRecordsLatestByGroup(ctx, boxIDs)
}

// applyInterpolation applies hydraulic calculations to sensor records
// Calculates V (volume), Q (flow), Q_of (overflow) from WAU and DR
func (s *SensorService) applyInterpolation(record domain.Record) domain.Record {
	// Get WAU (water level) if exists
	wau := record.GetFloat("WAU")
	if wau == 0 {
		return record
	}

	// Calculate V (volume) from WAU
	v := s.calculator.CalculateWaterIndex(wau)
	record["V"] = domain.RoundValue(v)

	// Calculate Q (flow) from WAU and DR if DR exists
	dr := record.GetFloat("DR")
	if dr > 0 {
		q := s.calculator.CalculateWaterFlow(wau, dr)
		record["Q"] = domain.RoundValue(q)
	}

	// Calculate Q_of (overflow) from WAU
	qOf := s.calculator.CalculateWaterOverFlow(wau)
	record["Q_of"] = domain.RoundValue(qOf)

	return record
}

// SetVolumeCurve allows configuring custom volume curve for a specific box/zone
func (s *SensorService) SetVolumeCurve(points []interpolation.Point) {
	s.calculator.SetVolumeCurve(points)
}

// SetFlowCurve allows configuring custom flow curve
func (s *SensorService) SetFlowCurve(points []interpolation.Point) {
	s.calculator.SetFlowCurve(points)
}

// SetOverflowParams allows configuring overflow parameters
func (s *SensorService) SetOverflowParams(m, b float64) {
	s.calculator.SetOverflowParams(m, b)
}
