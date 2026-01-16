package mongodb

import (
	"context"
	"fmt"
	"time"

	"tp25-api/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SensorRepository struct {
	db      *mongo.Database
	metrics *mongo.Collection
}

func NewSensorRepository(db *mongo.Database) *SensorRepository {
	return &SensorRepository{
		db:      db,
		metrics: db.Collection("metric"),
	}
}

// Metric operations

func (r *SensorRepository) ListMetrics(ctx context.Context) ([]domain.Metric, error) {
	cursor, err := r.metrics.Find(ctx, bson.M{"dtime": bson.M{"$exists": false}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var metrics []domain.Metric
	if err := cursor.All(ctx, &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}

func (r *SensorRepository) ListMetricsWithPagination(ctx context.Context, pagination *domain.Pagination, filter bson.M) ([]domain.Metric, int64, error) {
	if filter == nil {
		filter = bson.M{}
	}
	filter["dtime"] = bson.M{"$exists": false}

	// Get total count
	total, err := r.metrics.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find with pagination
	opts := options.Find().
		SetSkip(int64(pagination.GetSkip())).
		SetLimit(int64(pagination.GetLimit())).
		SetSort(bson.D{{Key: "ctime", Value: -1}})

	cursor, err := r.metrics.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var metrics []domain.Metric
	if err := cursor.All(ctx, &metrics); err != nil {
		return nil, 0, err
	}

	return metrics, total, nil
}

func (r *SensorRepository) GetMetric(ctx context.Context, filter bson.M) (*domain.Metric, error) {
	// Add dtime filter
	filter["dtime"] = bson.M{"$exists": false}

	var metric domain.Metric
	err := r.metrics.FindOne(ctx, filter).Decode(&metric)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrMetricNotFound
		}
		return nil, err
	}
	return &metric, nil
}

func (r *SensorRepository) CreateMetric(ctx context.Context, metric *domain.Metric) error {
	// Check if code already exists
	existing, err := r.GetMetric(ctx, bson.M{"code": metric.Code})
	if err == nil && existing != nil {
		return domain.ErrMetricCodeExisted
	}

	_, err = r.metrics.InsertOne(ctx, metric)
	return err
}

func (r *SensorRepository) UpdateMetric(ctx context.Context, metric *domain.Metric) error {
	metric.MTime = time.Now().UnixMilli()
	_, err := r.metrics.UpdateOne(
		ctx,
		bson.M{"_id": metric.ID},
		bson.M{"$set": metric},
	)
	return err
}

func (r *SensorRepository) DeleteMetric(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	_, err := r.metrics.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"dtime": now}},
	)
	return err
}

// Record operations

// getRecordCollection returns the collection name for a box's sensor data
func (r *SensorRepository) getRecordCollection(boxID string) *mongo.Collection {
	collectionName := "sensor_data_" + boxID
	return r.db.Collection(collectionName)
}

func (r *SensorRepository) ListRecords(ctx context.Context, boxID string, query *domain.QueryRecord) (*domain.RecordsResult, error) {
	collection := r.getRecordCollection(boxID)

	filter := bson.M{}
	if query != nil && len(query.Time) == 2 {
		filter["_id"] = bson.M{
			"$gte": query.Time[0],
			"$lte": query.Time[1],
		}
	}

	skip := int64(0)
	limit := int64(20)
	if query != nil {
		if query.Skip != nil {
			skip = int64(*query.Skip)
		}
		if query.Limit != nil {
			limit = int64(*query.Limit)
		}
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$facet", Value: bson.M{
			"records": []bson.M{
				{"$sort": bson.M{"_id": -1}},
				{"$skip": skip},
				{"$limit": limit},
				{"$addFields": bson.M{"id": "$_id"}},
				{"$unset": "_id"},
			},
			"total": []bson.M{
				{"$count": "count"},
			},
		}}},
	}

	opts := options.Aggregate().SetAllowDiskUse(true)
	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	var records []domain.Record
	var totalCount int64

	if len(result) > 0 {
		if recordsData, ok := result[0]["records"].(bson.A); ok {
			for _, item := range recordsData {
				if record, ok := item.(bson.M); ok {
					records = append(records, domain.Record(record))
				}
			}
		}

		if totalData, ok := result[0]["total"].(bson.A); ok && len(totalData) > 0 {
			if countDoc, ok := totalData[0].(bson.M); ok {
				if count, ok := countDoc["count"].(int32); ok {
					totalCount = int64(count)
				}
			}
		}
	}

	return &domain.RecordsResult{
		Records: records,
		Total:   totalCount,
	}, nil
}

func (r *SensorRepository) CountRecords(ctx context.Context, boxID string, query *domain.QueryRecord) (int64, error) {
	collection := r.getRecordCollection(boxID)

	filter := bson.M{}
	if query != nil && len(query.Time) == 2 {
		filter["_id"] = bson.M{
			"$gte": query.Time[0],
			"$lte": query.Time[1],
		}
	}

	return collection.CountDocuments(ctx, filter)
}

func (r *SensorRepository) AddRecord(ctx context.Context, boxID string, record domain.Record) error {
	collection := r.getRecordCollection(boxID)

	// Set server create timestamp if not exists
	if _, exists := record["c"]; !exists {
		record["c"] = time.Now().UnixMilli()
	}

	_, err := collection.InsertOne(ctx, record)
	return err
}

func (r *SensorRepository) ImportRecord(ctx context.Context, boxID string, record domain.Record) error {
	collection := r.getRecordCollection(boxID)

	// Set server create timestamp if not exists
	if _, exists := record["c"]; !exists {
		record["c"] = time.Now().UnixMilli()
	}

	_, err := collection.InsertOne(ctx, record)
	return err
}

// ReportRecords generates daily reports for a box within a time range
// This implementation FIXES the N+1 query problem from the TypeScript version
func (r *SensorRepository) ReportRecords(ctx context.Context, boxID string, query *domain.QueryRecord) ([]domain.DailyReport, error) {
	collection := r.getRecordCollection(boxID)

	matchStage := bson.M{"_id": bson.M{"$exists": true}}
	if query != nil && len(query.Time) == 2 {
		matchStage["_id"] = bson.M{
			"$gte": query.Time[0],
			"$lte": query.Time[1],
		}
	}

	// Use aggregation pipeline to generate daily reports in a single query
	// This FIXES the N+1 query problem from the original TypeScript implementation
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$addFields", Value: bson.M{
			"date": bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m-%d",
					"date":   bson.M{"$toDate": bson.M{"$multiply": []interface{}{"$_id", 1000}}},
				},
			},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$date",
			"count": bson.M{"$sum": 1},
			"data":  bson.M{"$push": "$$ROOT"},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	var reports []domain.DailyReport
	for _, result := range results {
		date := result["_id"].(string)
		count := int(result["count"].(int32))
		data := result["data"].([]interface{})

		report := domain.DailyReport{
			Date:  date,
			Count: count,
			Avg:   make(map[string]float64),
			Min:   make(map[string]float64),
			Max:   make(map[string]float64),
		}

		// Calculate statistics for all numeric fields
		metricSums := make(map[string]float64)
		metricCounts := make(map[string]int)
		metricMins := make(map[string]float64)
		metricMaxs := make(map[string]float64)

		for _, item := range data {
			if record, ok := item.(bson.M); ok {
				for key, value := range record {
					// Skip non-metric fields
					if key == "_id" || key == "c" || key == "date" {
						continue
					}

					if floatVal, ok := value.(float64); ok {
						metricSums[key] += floatVal
						metricCounts[key]++

						if _, exists := metricMins[key]; !exists {
							metricMins[key] = floatVal
							metricMaxs[key] = floatVal
						} else {
							if floatVal < metricMins[key] {
								metricMins[key] = floatVal
							}
							if floatVal > metricMaxs[key] {
								metricMaxs[key] = floatVal
							}
						}
					}
				}
			}
		}

		// Calculate averages
		for key, sum := range metricSums {
			if count := metricCounts[key]; count > 0 {
				report.Avg[key] = domain.RoundValue(sum / float64(count))
			}
		}

		for key, min := range metricMins {
			report.Min[key] = domain.RoundValue(min)
		}

		for key, max := range metricMaxs {
			report.Max[key] = domain.RoundValue(max)
		}

		reports = append(reports, report)
	}

	return reports, nil
}

func (r *SensorRepository) ListRecordsByGroup(ctx context.Context, boxIDs []string, query *domain.QueryRecord) (*domain.RecordsResult, error) {
	if len(boxIDs) == 0 {
		return &domain.RecordsResult{Records: []domain.Record{}, Total: 0}, nil
	}

	skip := int64(0)
	limit := int64(20)
	if query != nil {
		if query.Skip != nil {
			skip = int64(*query.Skip)
		}
		if query.Limit != nil {
			limit = int64(*query.Limit)
		}
	}

	matchStage := bson.M{}
	if query != nil && len(query.Time) == 2 {
		matchStage["_id"] = bson.M{
			"$gte": query.Time[0],
			"$lte": query.Time[1],
		}
	}

	firstBoxID := boxIDs[0]
	collection := r.getRecordCollection(firstBoxID)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$addFields", Value: bson.M{"box_id": firstBoxID}}},
	}

	for i := 1; i < len(boxIDs); i++ {
		pipeline = append(pipeline, bson.D{{Key: "$unionWith", Value: bson.M{
			"coll": "sensor_data_" + boxIDs[i],
			"pipeline": []bson.M{
				{"$match": matchStage},
				{"$addFields": bson.M{"box_id": boxIDs[i]}},
			},
		}}})
	}

	pipeline = append(pipeline, bson.D{{Key: "$facet", Value: bson.M{
		"records": []bson.M{
			{"$sort": bson.M{"_id": -1}},
			{"$skip": skip},
			{"$limit": limit},
			{"$addFields": bson.M{"id": "$_id"}},
			{"$unset": "_id"},
		},
		"total": []bson.M{
			{"$count": "count"},
		},
	}}})

	opts := options.Aggregate().SetAllowDiskUse(true)
	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, fmt.Errorf("aggregation failed: %w", err)
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("reading aggregation result failed: %w", err)
	}

	var allRecords []domain.Record
	var totalCount int64

	if len(result) > 0 {
		if recordsData, ok := result[0]["records"].(bson.A); ok {
			for _, item := range recordsData {
				if record, ok := item.(bson.M); ok {
					allRecords = append(allRecords, domain.Record(record))
				}
			}
		}

		if totalData, ok := result[0]["total"].(bson.A); ok && len(totalData) > 0 {
			if countDoc, ok := totalData[0].(bson.M); ok {
				if count, ok := countDoc["count"].(int32); ok {
					totalCount = int64(count)
				}
			}
		}
	}

	return &domain.RecordsResult{
		Records: allRecords,
		Total:   totalCount,
	}, nil
}

func (r *SensorRepository) ListRecordsLatestByGroup(ctx context.Context, boxIDs []string) (*domain.RecordsResult, error) {
	if len(boxIDs) == 0 {
		return &domain.RecordsResult{Records: []domain.Record{}, Total: 0}, nil
	}

	firstBoxID := boxIDs[0]
	collection := r.getRecordCollection(firstBoxID)

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$sort", Value: bson.M{"_id": -1}}},
		bson.D{{Key: "$limit", Value: 1}},
		bson.D{{Key: "$addFields", Value: bson.M{"box_id": firstBoxID}}},
	}

	for i := 1; i < len(boxIDs); i++ {
		pipeline = append(pipeline, bson.D{{Key: "$unionWith", Value: bson.M{
			"coll": "sensor_data_" + boxIDs[i],
			"pipeline": mongo.Pipeline{
				bson.D{{Key: "$sort", Value: bson.M{"_id": -1}}},
				bson.D{{Key: "$limit", Value: 1}},
				bson.D{{Key: "$addFields", Value: bson.M{"box_id": boxIDs[i]}}},
			},
		}}})
	}

	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.M{"_id": -1}}})

	pipeline = append(pipeline,
		bson.D{{Key: "$addFields", Value: bson.M{"id": "$_id"}}},
		bson.D{{Key: "$unset", Value: "_id"}},
	)

	opts := options.Aggregate()
	cursor, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return nil, fmt.Errorf("aggregation failed: %w", err)
	}
	defer cursor.Close(ctx)

	var result []domain.Record
	if err := cursor.All(ctx, &result); err != nil {
		return nil, fmt.Errorf("reading aggregation result failed: %w", err)
	}

	return &domain.RecordsResult{
		Records: result,
		Total:   int64(len(result)),
	}, nil
}
