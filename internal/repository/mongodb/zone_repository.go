package mongodb

import (
	"context"
	"time"

	"tp25-api/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ZoneRepository struct {
	db     *mongo.Database
	zones  *mongo.Collection
	groups *mongo.Collection
	boxes  *mongo.Collection
}

func NewZoneRepository(db *mongo.Database) *ZoneRepository {
	return &ZoneRepository{
		db:     db,
		zones:  db.Collection("zone"),
		groups: db.Collection("groups"),
		boxes:  db.Collection("box"),
	}
}

// Zone operations

func (r *ZoneRepository) ListZones(ctx context.Context) ([]domain.Zone, error) {
	cursor, err := r.zones.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var zones []domain.Zone
	if err := cursor.All(ctx, &zones); err != nil {
		return nil, err
	}
	return zones, nil
}

func (r *ZoneRepository) ListZonesWithPagination(ctx context.Context, pagination *domain.Pagination, filter bson.M) ([]domain.Zone, int64, error) {
	if filter == nil {
		filter = bson.M{}
	}

	// Get total count
	total, err := r.zones.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Find with pagination
	opts := options.Find().
		SetSkip(int64(pagination.GetSkip())).
		SetLimit(int64(pagination.GetLimit())).
		SetSort(bson.D{{Key: "ctime", Value: -1}})

	cursor, err := r.zones.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var zones []domain.Zone
	if err := cursor.All(ctx, &zones); err != nil {
		return nil, 0, err
	}

	return zones, total, nil
}

func (r *ZoneRepository) GetZone(ctx context.Context, id string) (*domain.Zone, error) {
	var zone domain.Zone
	err := r.zones.FindOne(ctx, bson.M{"_id": id}).Decode(&zone)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrZoneNotFound
		}
		return nil, err
	}
	return &zone, nil
}

func (r *ZoneRepository) CreateZone(ctx context.Context, zone *domain.Zone) error {
	// Check if code already exists
	var existing domain.Zone
	err := r.zones.FindOne(ctx, bson.M{"code": zone.Code}).Decode(&existing)
	if err == nil {
		return domain.ErrZoneCodeExisted
	}

	_, err = r.zones.InsertOne(ctx, zone)
	return err
}

func (r *ZoneRepository) UpdateZone(ctx context.Context, zone *domain.Zone) error {
	zone.MTime = time.Now().UnixMilli()
	_, err := r.zones.UpdateOne(
		ctx,
		bson.M{"_id": zone.ID},
		bson.M{"$set": zone},
	)
	return err
}

// BoxGroup operations

func (r *ZoneRepository) ListGroups(ctx context.Context, zoneID string) ([]domain.BoxGroup, error) {
	filter := bson.M{"dtime": bson.M{"$exists": false}}
	if zoneID != "" {
		filter["zone_id"] = zoneID
	}

	cursor, err := r.groups.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []domain.BoxGroup
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *ZoneRepository) GetGroup(ctx context.Context, id string) (*domain.BoxGroup, error) {
	var group domain.BoxGroup
	err := r.groups.FindOne(ctx, bson.M{"_id": id, "dtime": bson.M{"$exists": false}}).Decode(&group)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrBoxGroupNotFound
		}
		return nil, err
	}
	return &group, nil
}

func (r *ZoneRepository) CreateGroup(ctx context.Context, group *domain.BoxGroup) error {
	_, err := r.groups.InsertOne(ctx, group)
	return err
}

func (r *ZoneRepository) UpdateGroup(ctx context.Context, group *domain.BoxGroup) error {
	group.MTime = time.Now().UnixMilli()
	_, err := r.groups.UpdateOne(
		ctx,
		bson.M{"_id": group.ID},
		bson.M{"$set": group},
	)
	return err
}

func (r *ZoneRepository) DeleteGroup(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	_, err := r.groups.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"dtime": now}},
	)
	return err
}

// Box operations

func (r *ZoneRepository) ListBoxes(ctx context.Context, filter domain.FilterBoxParams) ([]domain.Box, error) {
	query := bson.M{"dtime": bson.M{"$exists": false}}
	if filter.GroupID != nil && *filter.GroupID != "" {
		query["group_id"] = *filter.GroupID
	}

	cursor, err := r.boxes.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var boxes []domain.Box
	if err := cursor.All(ctx, &boxes); err != nil {
		return nil, err
	}
	return boxes, nil
}

func (r *ZoneRepository) ListBoxesWithPagination(ctx context.Context, pagination *domain.Pagination, filter domain.FilterBoxParams) ([]domain.Box, int64, error) {
	query := bson.M{"dtime": bson.M{"$exists": false}}
	if filter.GroupID != nil && *filter.GroupID != "" {
		query["group_id"] = *filter.GroupID
	}

	// Get total count
	total, err := r.boxes.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// Find with pagination
	opts := options.Find().
		SetSkip(int64(pagination.GetSkip())).
		SetLimit(int64(pagination.GetLimit())).
		SetSort(bson.D{{Key: "ctime", Value: -1}})

	cursor, err := r.boxes.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var boxes []domain.Box
	if err := cursor.All(ctx, &boxes); err != nil {
		return nil, 0, err
	}

	return boxes, total, nil
}

func (r *ZoneRepository) GetBox(ctx context.Context, id string) (*domain.Box, error) {
	var box domain.Box
	err := r.boxes.FindOne(ctx, bson.M{"_id": id, "dtime": bson.M{"$exists": false}}).Decode(&box)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrBoxNotFound
		}
		return nil, err
	}
	return &box, nil
}

func (r *ZoneRepository) CreateBox(ctx context.Context, box *domain.Box) error {
	// Check if device_id already exists
	var existing domain.Box
	err := r.boxes.FindOne(ctx, bson.M{"device_id": box.DeviceID, "dtime": bson.M{"$exists": false}}).Decode(&existing)
	if err == nil {
		return domain.ErrBoxDeviceExisted
	}

	_, err = r.boxes.InsertOne(ctx, box)
	return err
}

func (r *ZoneRepository) UpdateBox(ctx context.Context, box *domain.Box) error {
	box.MTime = time.Now().UnixMilli()
	_, err := r.boxes.UpdateOne(
		ctx,
		bson.M{"_id": box.ID},
		bson.M{"$set": box},
	)
	return err
}

func (r *ZoneRepository) DeleteBox(ctx context.Context, id string) error {
	now := time.Now().UnixMilli()
	_, err := r.boxes.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"dtime": now}},
	)
	return err
}

// ReportByMetric generates monthly reports for given metrics across multiple data sources
func (r *ZoneRepository) ReportByMetric(ctx context.Context, sources []string, metrics []string) ([]domain.Report, error) {
	var allReports []domain.Report

	for _, source := range sources {
		collectionName := "sensor_data_" + source
		collection := r.db.Collection(collectionName)

		pipeline := mongo.Pipeline{
			{{Key: "$match", Value: bson.M{
				"t": bson.M{"$exists": true},
			}}},
			{{Key: "$project", Value: bson.M{
				"year":    bson.M{"$year": bson.M{"$toDate": bson.M{"$multiply": []interface{}{"$t", 1000}}}},
				"month":   bson.M{"$month": bson.M{"$toDate": bson.M{"$multiply": []interface{}{"$t", 1000}}}},
				"metrics": "$$ROOT",
			}}},
			{{Key: "$group", Value: bson.M{
				"_id": bson.M{
					"year":  "$year",
					"month": "$month",
				},
				"count": bson.M{"$sum": 1},
				"data":  bson.M{"$push": "$metrics"},
			}}},
		}

		cursor, err := collection.Aggregate(ctx, pipeline)
		if err != nil {
			continue
		}

		var results []bson.M
		if err := cursor.All(ctx, &results); err != nil {
			cursor.Close(ctx)
			continue
		}
		cursor.Close(ctx)

		// Process results for each metric
		for _, result := range results {
			if id, ok := result["_id"].(bson.M); ok {
				year := int(id["year"].(int32))
				month := int(id["month"].(int32))
				count := int(result["count"].(int32))
				data := result["data"].([]interface{})

				for _, metric := range metrics {
					var total float64
					validCount := 0

					for _, item := range data {
						if record, ok := item.(bson.M); ok {
							if val, exists := record[metric]; exists {
								if floatVal, ok := val.(float64); ok {
									total += floatVal
									validCount++
								}
							}
						}
					}

					if validCount > 0 {
						report := domain.Report{
							Count: count,
							Total: domain.RoundValue(total),
						}
						report.Info.Year = year
						report.Info.Month = month
						report.Info.Metric = metric
						allReports = append(allReports, report)
					}
				}
			}
		}
	}

	return allReports, nil
}
