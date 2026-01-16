package domain

import (
	"errors"
	"time"
	"tp25-api/lib"
)

type Location struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lng float64 `json:"lng" bson:"lng"`
}

type Zone struct {
	ID     string      `json:"id" bson:"_id"`
	Code   string      `json:"code" bson:"code"`
	Name   string      `json:"name" bson:"name"`
	Detail interface{} `json:"detail" bson:"detail"`
	Center Location    `json:"center" bson:"center"`
	CTime  int64       `json:"ctime" bson:"ctime"`
	MTime  int64       `json:"mtime" bson:"mtime"`
}

type CreateZoneParams struct {
	Name   string      `json:"name" binding:"required"`
	Code   string      `json:"code" binding:"required"`
	Detail interface{} `json:"detail"`
	Center *Location   `json:"center"`
}

type UpdateZoneParams struct {
	Name   *string     `json:"name"`
	Code   *string     `json:"code"`
	Detail interface{} `json:"detail"`
	Center *Location   `json:"center"`
}

type MissionGroup struct {
	Name      string `json:"name" bson:"name"`
	Unit      string `json:"unit" bson:"unit"`           // Đơn vị
	Parameter string `json:"parameter" bson:"parameter"` // Thông số
}

type NoteGroup struct {
	// Nhiệm vụ công trình
	Mission []MissionGroup `json:"mission,omitempty" bson:"mission,omitempty"`

	// Cấp công trình
	Region string `json:"region,omitempty" bson:"region,omitempty"` // Khu vực
	City   string `json:"city,omitempty" bson:"city,omitempty"`     // Thành phố

	Level             string `json:"level,omitempty" bson:"level,omitempty"`                                   // Cấp công trình
	Watering          string `json:"watering,omitempty" bson:"watering,omitempty"`                             // Tần suất tưới thiết kế %
	FrequencyFlood    string `json:"frequency_flood,omitempty" bson:"frequency_flood,omitempty"`               // Tần suất lũ thiết kế(%)
	FrequencyFloodOri string `json:"frequency_flood_origin,omitempty" bson:"frequency_flood_origin,omitempty"` // Tần suất lũ kiểm tra(%)

	// Hồ chứa
	Area               string `json:"area,omitempty" bson:"area,omitempty"`                                 // Diện tích lưu vực (km2)
	WaterRiseDie       string `json:"water_rise_die,omitempty" bson:"water_rise_die,omitempty"`             // Mực nước chết MNC
	WaterRiseNormal    string `json:"water_rise_normal,omitempty" bson:"water_rise_normal,omitempty"`       // Mực nước dâng bình thường MNDBT (m)
	WaterRiseHight     string `json:"water_rise_hight,omitempty" bson:"water_rise_hight,omitempty"`         // Mực nước dâng gia cường MNDGC (P = 1,50%)
	CapacityRiseDie    string `json:"capacity_rise_die,omitempty" bson:"capacity_rise_die,omitempty"`       // Dung tích hồ ứng với MNC(10^6m3)
	CapacityRiseNormal string `json:"capacity_rise_normal,omitempty" bson:"capacity_rise_normal,omitempty"` // Dung tích hồ ứng với MNDBT(10^6m3)
	CapacityRiseHight  string `json:"capacity_rise_hight,omitempty" bson:"capacity_rise_hight,omitempty"`   // Dung tích hồ ứng với MNDGC(10^6m3)
	AreaRiseDie        string `json:"area_rise_die,omitempty" bson:"area_rise_die,omitempty"`               // Diện tích hồ ứng với MNC(Ha)
	AreaRiseNormal     string `json:"area_rise_normal,omitempty" bson:"area_rise_normal,omitempty"`         // Diện tích hồ ứng với MNDBT(Ha)
	AreaRiseHight      string `json:"area_rise_hight,omitempty" bson:"area_rise_hight,omitempty"`           // Diện tích hồ ứng với MNDGC(Ha)

	// Đập chính
	Structure         string `json:"structure,omitempty" bson:"structure,omitempty"`                         // Kết cấu đập
	Elevation         string `json:"elevation,omitempty" bson:"elevation,omitempty"`                         // Cao trình đập
	Height            string `json:"height,omitempty" bson:"height,omitempty"`                               // Chiều cao đập lớn nhất(m)
	Longs             string `json:"longs,omitempty" bson:"longs,omitempty"`                                 // Chiều dài đập(m)
	Width             string `json:"width,omitempty" bson:"width,omitempty"`                                 // Bề rộng mặt đập(m)
	RoofCoefficientUp string `json:"roof_coefficient_up,omitempty" bson:"roof_coefficient_up,omitempty"`     // Hệ số mái thượng lưu
	RoofCoefficientDn string `json:"roof_coefficient_down,omitempty" bson:"roof_coefficient_down,omitempty"` // Hệ số mái hạ lưu

	// Tràn xả lũ
	StructuralDischargeDam  string `json:"structural_discharge_dam,omitempty" bson:"structural_discharge_dam,omitempty"`     // Đặc điểm cấu kết cấu tràn xả lũ
	FormatDischargeDam      string `json:"format_discharge_dam,omitempty" bson:"format_discharge_dam,omitempty"`             // Hình thức tràn
	ElevationDischargeDam   string `json:"elevation_discharge_dam,omitempty" bson:"elevation_discharge_dam,omitempty"`       // Cao trình đập(m)
	WidthDischargeDam       string `json:"width_discharge_dam,omitempty" bson:"width_discharge_dam,omitempty"`               // Chiều rộng tràn nước(m)
	WaterColumnDischargeDam string `json:"water_column_discharge_dam,omitempty" bson:"water_column_discharge_dam,omitempty"` // Cột nước tràn thiết kế(m)

	// Cống lấy nước
	StructuralDrain       string `json:"structural_drain,omitempty" bson:"structural_drain,omitempty"`               // Đặc điểm kết cấu cống
	ElevationDrain        string `json:"elevation_drain,omitempty" bson:"elevation_drain,omitempty"`                 // Cao trình cấu cống(m)
	OpeningDrain          string `json:"opening_drain,omitempty" bson:"opening_drain,omitempty"`                     // Khẩu diện cống(m)
	MaximumDischargeDrain string `json:"maximum_discharge_drain,omitempty" bson:"maximum_discharge_drain,omitempty"` // Lưu lượng xả max(m3/s);

	// Tràn sự cố
	StructuralIncident      string `json:"structural_incident,omitempty" bson:"structural_incident,omitempty"`             // Đặc điểm kết cấu tràn sự cố
	ElevationIncident       string `json:"elevation_incident,omitempty" bson:"elevation_incident,omitempty"`               // Cao trình ngưỡng tràn(m)
	BottomWidthIncident     string `json:"bottom_width_incident,omitempty" bson:"bottom_width_incident,omitempty"`         // Bề rộng đáy kênh tràn(m)
	RoofCoefficientIncident string `json:"roof_coefficient_incident,omitempty" bson:"roof_coefficient_incident,omitempty"` // Hệ số mái kênh tràn
	WaterColumnIncident     string `json:"water_column_incident,omitempty" bson:"water_column_incident,omitempty"`         // Cột nước tràn thiết kế(m)

	// Nhà quản lý
	StructuralManage string `json:"structural_manage,omitempty" bson:"structural_manage,omitempty"` // Kết cấu nhà quản lý
	AreaManage       string `json:"area_manage,omitempty" bson:"area_manage,omitempty"`             // Diện tích phòng quản lý (m2)
	QuantityManage   string `json:"quantity_manage,omitempty" bson:"quantity_manage,omitempty"`     // Số phòng
}

type BoxGroup struct {
	ID        string     `json:"id" bson:"_id"`
	Name      string     `json:"name" bson:"name"`
	SortOrder int        `json:"sort_order" bson:"sort_order"`
	ZoneID    string     `json:"zone_id" bson:"zone_id"`
	Center    *Location  `json:"center,omitempty" bson:"center,omitempty"`
	Zoom      *int       `json:"zoom,omitempty" bson:"zoom,omitempty"` // 10-16
	Cameras   []string   `json:"cameras,omitempty" bson:"cameras,omitempty"`
	Note      *NoteGroup `json:"note,omitempty" bson:"note,omitempty"`
	CTime     int64      `json:"ctime" bson:"ctime"`
	MTime     int64      `json:"mtime" bson:"mtime"`
	DTime     *int64     `json:"dtime,omitempty" bson:"dtime,omitempty"`
	Subdomain *string    `json:"subdomain,omitempty" bson:"subdomain,omitempty"`
}

type CreateGroupParams struct {
	Name      string    `json:"name" binding:"required"`
	ZoneID    string    `json:"zone_id" binding:"required"`
	Center    *Location `json:"center"`
	Zoom      *int      `json:"zoom"`
	Cameras   []string  `json:"cameras"`
	Subdomain *string   `json:"subdomain"`
}

type UpdateGroupParams struct {
	Name      *string   `json:"name"`
	SortOrder *int      `json:"sort_order"`
	Center    *Location `json:"center"`
	Zoom      *int      `json:"zoom"`
	Cameras   []string  `json:"cameras"`
	Subdomain *string   `json:"subdomain"`
}

type BoxMetric struct {
	Code     string   `json:"code" bson:"code"`
	Name     *string  `json:"name,omitempty" bson:"name,omitempty"`
	Metric   *string  `json:"metric,omitempty" bson:"metric,omitempty"`
	Lat      *float64 `json:"lat,omitempty" bson:"lat,omitempty"`
	Lng      *float64 `json:"lng,omitempty" bson:"lng,omitempty"`
	Warning1 *string  `json:"warning1,omitempty" bson:"warning1,omitempty"`
	Warning2 *string  `json:"warning2,omitempty" bson:"warning2,omitempty"`
	Warning3 *string  `json:"warning3,omitempty" bson:"warning3,omitempty"`
}

type Box struct {
	ID        string      `json:"id" bson:"_id"`
	Name      string      `json:"name" bson:"name"`
	Desc      string      `json:"desc" bson:"desc"`
	GroupID   string      `json:"group_id" bson:"group_id"`
	SortOrder int         `json:"sort_order" bson:"sort_order"`
	ZoneID    string      `json:"zone_id" bson:"zone_id"`
	Location  Location    `json:"location" bson:"location"`
	DeviceID  string      `json:"device_id" bson:"device_id"`
	Metrics   []BoxMetric `json:"metrics" bson:"metrics"`
	Type      *string     `json:"type,omitempty" bson:"type,omitempty"`
	CTime     int64       `json:"ctime" bson:"ctime"`
	MTime     int64       `json:"mtime" bson:"mtime"`
	DTime     *int64      `json:"dtime,omitempty" bson:"dtime,omitempty"`
}

type CreateBoxParams struct {
	Name     string      `json:"name" binding:"required"`
	GroupID  string      `json:"group_id" binding:"required"`
	ZoneID   string      `json:"zone_id" binding:"required"`
	Location Location    `json:"location" binding:"required"`
	DeviceID string      `json:"device_id" binding:"required"`
	Metrics  []BoxMetric `json:"metrics" binding:"required"`
	Desc     string      `json:"desc"`
	Type     *string     `json:"type"`
}

type UpdateBoxParams struct {
	Name      *string     `json:"name"`
	Desc      *string     `json:"desc"`
	Type      *string     `json:"type"`
	GroupID   *string     `json:"group_id"`
	SortOrder *int        `json:"sort_order"`
	Location  *Location   `json:"location"`
	DeviceID  *string     `json:"device_id"`
	Metrics   []BoxMetric `json:"metrics"`
}

type FilterBoxParams struct {
	GroupID *string `json:"group_id" form:"group_id"`
}

type ViewBox struct {
	BoxGroup
	Boxes []Box `json:"boxs" bson:"boxs"`
	Total *int  `json:"total,omitempty" bson:"total,omitempty"`
}

type Report struct {
	Info struct {
		Month  int    `json:"month" bson:"month"`
		Year   int    `json:"year" bson:"year"`
		Metric string `json:"metric" bson:"metric"`
	} `json:"info" bson:"info"`
	Count int     `json:"count" bson:"count"`
	Total float64 `json:"total" bson:"total"`
}

var (
	ErrZoneNotFound     = errors.New("zone not found")
	ErrZoneCodeExisted  = errors.New("zone code existed")
	ErrBoxNotFound      = errors.New("box not found")
	ErrBoxDeviceExisted = errors.New("box device existed")
	ErrBoxGroupNotFound = errors.New("box group not found")
	ErrBoxGroupExisted  = errors.New("box group existed")
)

// NewZone creates a new zone with timestamps
func NewZone(params CreateZoneParams) *Zone {
	now := time.Now().UnixMilli()
	center := Location{Lat: 0, Lng: 0}
	if params.Center != nil {
		center = *params.Center
	}
	return &Zone{
		ID:     lib.Rand.Char(12),
		Code:   params.Code,
		Name:   params.Name,
		Detail: params.Detail,
		Center: center,
		CTime:  now,
		MTime:  now,
	}
}

// NewBoxGroup creates a new box group with timestamps
func NewBoxGroup(params CreateGroupParams) *BoxGroup {
	now := time.Now().UnixMilli()
	return &BoxGroup{
		ID:        lib.Rand.Char(12),
		Name:      params.Name,
		ZoneID:    params.ZoneID,
		Center:    params.Center,
		Zoom:      params.Zoom,
		Cameras:   params.Cameras,
		Subdomain: params.Subdomain,
		SortOrder: 0,
		CTime:     now,
		MTime:     now,
	}
}

// NewBox creates a new box with timestamps
func NewBox(params CreateBoxParams) *Box {
	now := time.Now().UnixMilli()
	return &Box{
		ID:        lib.Rand.Char(12),
		Name:      params.Name,
		Desc:      params.Desc,
		GroupID:   params.GroupID,
		ZoneID:    params.ZoneID,
		Location:  params.Location,
		DeviceID:  params.DeviceID,
		Metrics:   params.Metrics,
		Type:      params.Type,
		SortOrder: 0,
		CTime:     now,
		MTime:     now,
	}
}

// RoundValue rounds a float to 2 decimal places
func RoundValue(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
