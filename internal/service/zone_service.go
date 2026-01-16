package service

import (
	"context"
	"sort"

	"tp25-api/internal/domain"
	"tp25-api/internal/repository/mongodb"

	"go.mongodb.org/mongo-driver/bson"
)

type ZoneService struct {
	repo *mongodb.ZoneRepository
}

func NewZoneService(repo *mongodb.ZoneRepository) *ZoneService {
	return &ZoneService{repo: repo}
}

// Zone operations

func (s *ZoneService) ListZones(ctx context.Context) ([]domain.Zone, error) {
	return s.repo.ListZones(ctx)
}

func (s *ZoneService) ListZonesWithPagination(ctx context.Context, pagination *domain.Pagination, filter bson.M) ([]domain.Zone, int64, error) {
	return s.repo.ListZonesWithPagination(ctx, pagination, filter)
}

func (s *ZoneService) GetZone(ctx context.Context, id string) (*domain.Zone, error) {
	return s.repo.GetZone(ctx, id)
}

func (s *ZoneService) CreateZone(ctx context.Context, params domain.CreateZoneParams) (*domain.Zone, error) {
	zone := domain.NewZone(params)
	if err := s.repo.CreateZone(ctx, zone); err != nil {
		return nil, err
	}
	return zone, nil
}

func (s *ZoneService) UpdateZone(ctx context.Context, id string, params domain.UpdateZoneParams) (*domain.Zone, error) {
	zone, err := s.repo.GetZone(ctx, id)
	if err != nil {
		return nil, err
	}

	if params.Name != nil {
		zone.Name = *params.Name
	}
	if params.Code != nil {
		zone.Code = *params.Code
	}
	if params.Detail != nil {
		zone.Detail = params.Detail
	}
	if params.Center != nil {
		zone.Center = *params.Center
	}

	if err := s.repo.UpdateZone(ctx, zone); err != nil {
		return nil, err
	}

	return zone, nil
}

// BoxGroup operations

func (s *ZoneService) ListGroups(ctx context.Context, zoneID string) ([]domain.ViewBox, error) {
	groups, err := s.repo.ListGroups(ctx, zoneID)
	if err != nil {
		return nil, err
	}

	// Sort groups by sort_order
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].SortOrder < groups[j].SortOrder
	})

	var viewBoxes []domain.ViewBox
	for _, group := range groups {
		// Get boxes for each group
		filter := domain.FilterBoxParams{GroupID: &group.ID}
		boxes, err := s.repo.ListBoxes(ctx, filter)
		if err != nil {
			// Continue on error, return empty boxes
			boxes = []domain.Box{}
		}

		// Sort boxes by sort_order
		sort.Slice(boxes, func(i, j int) bool {
			return boxes[i].SortOrder < boxes[j].SortOrder
		})

		total := len(boxes)
		viewBox := domain.ViewBox{
			BoxGroup: group,
			Boxes:    boxes,
			Total:    &total,
		}
		viewBoxes = append(viewBoxes, viewBox)
	}

	return viewBoxes, nil
}

func (s *ZoneService) GetGroup(ctx context.Context, id string) (*domain.ViewBox, error) {
	group, err := s.repo.GetGroup(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get boxes for this group
	filter := domain.FilterBoxParams{GroupID: &group.ID}
	boxes, err := s.repo.ListBoxes(ctx, filter)
	if err != nil {
		boxes = []domain.Box{}
	}

	// Sort boxes by sort_order
	sort.Slice(boxes, func(i, j int) bool {
		return boxes[i].SortOrder < boxes[j].SortOrder
	})

	total := len(boxes)
	return &domain.ViewBox{
		BoxGroup: *group,
		Boxes:    boxes,
		Total:    &total,
	}, nil
}

func (s *ZoneService) FindGroup(ctx context.Context, id string) (*domain.BoxGroup, error) {
	return s.repo.GetGroup(ctx, id)
}

func (s *ZoneService) CreateGroup(ctx context.Context, params domain.CreateGroupParams) (*domain.BoxGroup, error) {
	// Get max sort_order for auto-increment
	groups, err := s.repo.ListGroups(ctx, params.ZoneID)
	if err != nil {
		return nil, err
	}

	maxSortOrder := 0
	for _, g := range groups {
		if g.SortOrder > maxSortOrder {
			maxSortOrder = g.SortOrder
		}
	}

	group := domain.NewBoxGroup(params)
	group.SortOrder = maxSortOrder + 1

	if err := s.repo.CreateGroup(ctx, group); err != nil {
		return nil, err
	}

	return group, nil
}

func (s *ZoneService) UpdateGroup(ctx context.Context, id string, params domain.UpdateGroupParams) (*domain.ViewBox, error) {
	group, err := s.repo.GetGroup(ctx, id)
	if err != nil {
		return nil, err
	}

	if params.Name != nil {
		group.Name = *params.Name
	}
	if params.SortOrder != nil {
		group.SortOrder = *params.SortOrder
	}
	if params.Center != nil {
		group.Center = params.Center
	}
	if params.Zoom != nil {
		group.Zoom = params.Zoom
	}
	if params.Cameras != nil {
		group.Cameras = params.Cameras
	}
	if params.Subdomain != nil {
		group.Subdomain = params.Subdomain
	}

	if err := s.repo.UpdateGroup(ctx, group); err != nil {
		return nil, err
	}

	return s.GetGroup(ctx, id)
}

func (s *ZoneService) DeleteGroup(ctx context.Context, id string) (*domain.BoxGroup, error) {
	group, err := s.repo.GetGroup(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteGroup(ctx, id); err != nil {
		return nil, err
	}

	return group, nil
}

// Box operations

func (s *ZoneService) ListBoxes(ctx context.Context, filter domain.FilterBoxParams) ([]domain.Box, error) {
	boxes, err := s.repo.ListBoxes(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Sort boxes by sort_order
	sort.Slice(boxes, func(i, j int) bool {
		return boxes[i].SortOrder < boxes[j].SortOrder
	})

	return boxes, nil
}

func (s *ZoneService) ListBoxesWithPagination(ctx context.Context, pagination *domain.Pagination, filter domain.FilterBoxParams) ([]domain.Box, int64, error) {
	return s.repo.ListBoxesWithPagination(ctx, pagination, filter)
}

func (s *ZoneService) GetBox(ctx context.Context, id string) (*domain.Box, error) {
	return s.repo.GetBox(ctx, id)
}

func (s *ZoneService) CreateBox(ctx context.Context, params domain.CreateBoxParams) (*domain.Box, error) {
	// Get max sort_order for auto-increment
	filter := domain.FilterBoxParams{GroupID: &params.GroupID}
	boxes, err := s.repo.ListBoxes(ctx, filter)
	if err != nil {
		return nil, err
	}

	maxSortOrder := 0
	for _, b := range boxes {
		if b.SortOrder > maxSortOrder {
			maxSortOrder = b.SortOrder
		}
	}

	box := domain.NewBox(params)
	box.SortOrder = maxSortOrder + 1

	if err := s.repo.CreateBox(ctx, box); err != nil {
		return nil, err
	}

	return box, nil
}

func (s *ZoneService) UpdateBox(ctx context.Context, id string, params domain.UpdateBoxParams) (*domain.Box, error) {
	box, err := s.repo.GetBox(ctx, id)
	if err != nil {
		return nil, err
	}

	if params.Name != nil {
		box.Name = *params.Name
	}
	if params.Desc != nil {
		box.Desc = *params.Desc
	}
	if params.Type != nil {
		box.Type = params.Type
	}
	if params.GroupID != nil {
		box.GroupID = *params.GroupID
	}
	if params.SortOrder != nil {
		box.SortOrder = *params.SortOrder
	}
	if params.Location != nil {
		box.Location = *params.Location
	}
	if params.DeviceID != nil {
		box.DeviceID = *params.DeviceID
	}
	if params.Metrics != nil {
		box.Metrics = params.Metrics
	}

	if err := s.repo.UpdateBox(ctx, box); err != nil {
		return nil, err
	}

	return box, nil
}

func (s *ZoneService) DeleteBox(ctx context.Context, id string) (*domain.Box, error) {
	box, err := s.repo.GetBox(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.repo.DeleteBox(ctx, id); err != nil {
		return nil, err
	}

	return box, nil
}

// Report operations

func (s *ZoneService) ReportByMetric(ctx context.Context, boxGroupID string, metrics []string) ([]domain.Report, error) {
	// Get all boxes in the group
	filter := domain.FilterBoxParams{GroupID: &boxGroupID}
	boxes, err := s.repo.ListBoxes(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Extract box IDs as data sources
	var sources []string
	for _, box := range boxes {
		sources = append(sources, box.ID)
	}

	return s.repo.ReportByMetric(ctx, sources, metrics)
}
