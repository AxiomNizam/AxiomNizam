package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GISTrainHandler serves train/railway data from PostgreSQL for the GIS dashboard.
type GISTrainHandler struct {
	db *gorm.DB
}

// Train represents a train record.
type Train struct {
	ID              int    `json:"id" gorm:"primaryKey"`
	TrainNumber     string `json:"train_number" gorm:"column:train_number"`
	TrainName       string `json:"train_name" gorm:"column:train_name"`
	FromStationCode string `json:"from_station_code" gorm:"column:from_station_code"`
	ToStationCode   string `json:"to_station_code" gorm:"column:to_station_code"`
	TrainOwner      string `json:"train_owner" gorm:"column:train_owner"`
	RunsMon         bool   `json:"runs_mon" gorm:"column:runs_mon"`
	RunsTue         bool   `json:"runs_tue" gorm:"column:runs_tue"`
	RunsWed         bool   `json:"runs_wed" gorm:"column:runs_wed"`
	RunsThu         bool   `json:"runs_thu" gorm:"column:runs_thu"`
	RunsFri         bool   `json:"runs_fri" gorm:"column:runs_fri"`
	RunsSat         bool   `json:"runs_sat" gorm:"column:runs_sat"`
	RunsSun         bool   `json:"runs_sun" gorm:"column:runs_sun"`
}

func (Train) TableName() string { return "trains" }

// Station represents a railway station.
type Station struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Code string `json:"code" gorm:"column:code"`
	Name string `json:"name" gorm:"column:name"`
}

func (Station) TableName() string { return "stations" }

// TrainSchedule represents a stop in a train route.
type TrainSchedule struct {
	ID               int    `json:"id" gorm:"primaryKey"`
	TrainID          int    `json:"train_id" gorm:"column:train_id"`
	StationID        int    `json:"station_id" gorm:"column:station_id"`
	StopOrder        int    `json:"stop_order" gorm:"column:stop_order"`
	ArrivalTime      string `json:"arrival_time" gorm:"column:arrival_time"`
	DepartureTime    string `json:"departure_time" gorm:"column:departure_time"`
	HaltTime         string `json:"halt_time" gorm:"column:halt_time"`
	DistanceKM       int    `json:"distance_km" gorm:"column:distance_km"`
	DayCount         int    `json:"day_count" gorm:"column:day_count"`
	RouteNumber      int    `json:"route_number" gorm:"column:route_number"`
	BoardingDisabled bool   `json:"boarding_disabled" gorm:"column:boarding_disabled"`
}

func (TrainSchedule) TableName() string { return "train_schedules" }

// SeatClass represents a seat class code.
type SeatClass struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Code string `json:"code" gorm:"column:code"`
}

func (SeatClass) TableName() string { return "seat_classes" }

// StationTrainCount represents station-level train counts.
type StationTrainCount struct {
	ID            int `json:"id" gorm:"primaryKey"`
	StationID     int `json:"station_id" gorm:"column:station_id"`
	TotalTrains   int `json:"total_trains" gorm:"column:total_trains"`
	DepartTrains  int `json:"depart_trains" gorm:"column:depart_trains"`
	ArriveTrains  int `json:"arrive_trains" gorm:"column:arrive_trains"`
	TransitTrains int `json:"transit_trains" gorm:"column:transit_trains"`
}

func (StationTrainCount) TableName() string { return "station_train_counts" }

// NewGISTrainHandler creates a new GIS train handler with a database connection.
func NewGISTrainHandler(db *gorm.DB) *GISTrainHandler {
	return &GISTrainHandler{db: db}
}

// -- Trains CRUD --

// ListTrains GET /api/v1/gis/trains
func (h *GISTrainHandler) ListTrains(c *gin.Context) {
	var trains []Train
	q := h.db

	if search := strings.TrimSpace(c.Query("search")); search != "" {
		pattern := "%" + search + "%"
		q = q.Where("train_number ILIKE ? OR train_name ILIKE ?", pattern, pattern)
	}
	if from := strings.TrimSpace(c.Query("from")); from != "" {
		q = q.Where("from_station_code = ?", strings.ToUpper(from))
	}
	if to := strings.TrimSpace(c.Query("to")); to != "" {
		q = q.Where("to_station_code = ?", strings.ToUpper(to))
	}

	limit := 100
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 500 {
		limit = l
	}
	offset := 0
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}

	var total int64
	q.Model(&Train{}).Count(&total)

	if err := q.Order("train_number").Limit(limit).Offset(offset).Find(&trains).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trains": trains,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetTrain GET /api/v1/gis/trains/:id
func (h *GISTrainHandler) GetTrain(c *gin.Context) {
	id := c.Param("id")
	var train Train

	// Try by ID first, then by train number
	if err := h.db.Where("id = ? OR train_number = ?", id, id).First(&train).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "train not found"})
		return
	}

	// Get route (schedule stops)
	var schedule []struct {
		TrainSchedule
		StationCode string `json:"station_code" gorm:"column:code"`
		StationName string `json:"station_name" gorm:"column:name"`
	}
	h.db.Table("train_schedules ts").
		Select("ts.*, s.code, s.name").
		Joins("JOIN stations s ON s.id = ts.station_id").
		Where("ts.train_id = ?", train.ID).
		Order("ts.stop_order").
		Find(&schedule)

	type stopInfo struct {
		StopOrder        int    `json:"stop_order"`
		StationCode      string `json:"station_code"`
		StationName      string `json:"station_name"`
		ArrivalTime      string `json:"arrival_time,omitempty"`
		DepartureTime    string `json:"departure_time,omitempty"`
		HaltTime         string `json:"halt_time,omitempty"`
		DistanceKM       int    `json:"distance_km"`
		DayCount         int    `json:"day_count"`
		BoardingDisabled bool   `json:"boarding_disabled"`
	}

	stops := make([]stopInfo, 0, len(schedule))
	for _, s := range schedule {
		stops = append(stops, stopInfo{
			StopOrder:        s.StopOrder,
			StationCode:      s.StationCode,
			StationName:      s.StationName,
			ArrivalTime:      s.ArrivalTime,
			DepartureTime:    s.DepartureTime,
			HaltTime:         s.HaltTime,
			DistanceKM:       s.DistanceKM,
			DayCount:         s.DayCount,
			BoardingDisabled: s.BoardingDisabled,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"train": train,
		"route": stops,
		"stops": len(stops),
	})
}

// -- Stations --

// ListStations GET /api/v1/gis/trains/stations
func (h *GISTrainHandler) ListStations(c *gin.Context) {
	var stations []Station
	q := h.db

	if search := strings.TrimSpace(c.Query("search")); search != "" {
		pattern := "%" + search + "%"
		q = q.Where("code ILIKE ? OR name ILIKE ?", pattern, pattern)
	}

	limit := 100
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}
	offset := 0
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}

	var total int64
	q.Model(&Station{}).Count(&total)

	if err := q.Order("name").Limit(limit).Offset(offset).Find(&stations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stations": stations,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetStation GET /api/v1/gis/trains/stations/:code
func (h *GISTrainHandler) GetStation(c *gin.Context) {
	code := strings.ToUpper(c.Param("code"))
	var station Station
	if err := h.db.Where("code = ?", code).First(&station).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "station not found"})
		return
	}

	// Get trains that pass through this station
	type trainBrief struct {
		TrainNumber string `json:"train_number"`
		TrainName   string `json:"train_name"`
		StopOrder   int    `json:"stop_order"`
		Arrival     string `json:"arrival_time"`
		Departure   string `json:"departure_time"`
		DistanceKM  int    `json:"distance_km"`
	}
	var passingTrains []trainBrief
	h.db.Table("train_schedules ts").
		Select("t.train_number, t.train_name, ts.stop_order, ts.arrival_time, ts.departure_time, ts.distance_km").
		Joins("JOIN trains t ON t.id = ts.train_id").
		Where("ts.station_id = ?", station.ID).
		Order("ts.departure_time").
		Find(&passingTrains)

	c.JSON(http.StatusOK, gin.H{
		"station": station,
		"trains":  passingTrains,
		"total":   len(passingTrains),
	})
}

// -- Schedules/Routes --

// GetTrainRoute GET /api/v1/gis/trains/:id/route
func (h *GISTrainHandler) GetTrainRoute(c *gin.Context) {
	id := c.Param("id")
	var train Train
	if err := h.db.Where("id = ? OR train_number = ?", id, id).First(&train).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "train not found"})
		return
	}

	type routeStop struct {
		StopOrder     int    `json:"stop_order"`
		StationCode   string `json:"station_code" gorm:"column:code"`
		StationName   string `json:"station_name" gorm:"column:name"`
		ArrivalTime   string `json:"arrival_time"`
		DepartureTime string `json:"departure_time"`
		HaltTime      string `json:"halt_time"`
		DistanceKM    int    `json:"distance_km"`
		DayCount      int    `json:"day_count"`
	}
	var route []routeStop
	h.db.Table("train_schedules ts").
		Select("ts.stop_order, s.code, s.name, ts.arrival_time, ts.departure_time, ts.halt_time, ts.distance_km, ts.day_count").
		Joins("JOIN stations s ON s.id = ts.station_id").
		Where("ts.train_id = ?", train.ID).
		Order("ts.stop_order").
		Find(&route)

	c.JSON(http.StatusOK, gin.H{
		"train": train,
		"route": route,
		"stops": len(route),
	})
}

// -- Search --

// SearchTrains GET /api/v1/gis/trains/search?from=XX&to=YY
func (h *GISTrainHandler) SearchTrains(c *gin.Context) {
	from := strings.ToUpper(strings.TrimSpace(c.Query("from")))
	to := strings.ToUpper(strings.TrimSpace(c.Query("to")))

	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "both 'from' and 'to' station codes are required"})
		return
	}

	type searchResult struct {
		TrainID     int    `json:"train_id"`
		TrainNumber string `json:"train_number"`
		TrainName   string `json:"train_name"`
		FromStop    int    `json:"from_stop_order"`
		ToStop      int    `json:"to_stop_order"`
		FromDepart  string `json:"from_departure"`
		ToArrive    string `json:"to_arrival"`
		DistanceKM  int    `json:"distance_km"`
	}

	var results []searchResult
	h.db.Raw(`
		SELECT t.id AS train_id, t.train_number, t.train_name,
		       ts1.stop_order AS from_stop_order, ts2.stop_order AS to_stop_order,
		       ts1.departure_time AS from_departure, ts2.arrival_time AS to_arrival,
		       COALESCE(ts2.distance_km, 0) - COALESCE(ts1.distance_km, 0) AS distance_km
		FROM train_schedules ts1
		JOIN train_schedules ts2 ON ts1.train_id = ts2.train_id AND ts2.stop_order > ts1.stop_order
		JOIN stations s1 ON s1.id = ts1.station_id AND s1.code = ?
		JOIN stations s2 ON s2.id = ts2.station_id AND s2.code = ?
		JOIN trains t ON t.id = ts1.train_id
		ORDER BY ts1.departure_time
	`, from, to).Find(&results)

	c.JSON(http.StatusOK, gin.H{
		"from":   from,
		"to":     to,
		"trains": results,
		"total":  len(results),
	})
}

// -- Stats/Summary --

// GetTrainStats GET /api/v1/gis/trains/stats
func (h *GISTrainHandler) GetTrainStats(c *gin.Context) {
	var trainCount, stationCount, scheduleCount, seatClassCount int64
	h.db.Model(&Train{}).Count(&trainCount)
	h.db.Model(&Station{}).Count(&stationCount)
	h.db.Model(&TrainSchedule{}).Count(&scheduleCount)
	h.db.Model(&SeatClass{}).Count(&seatClassCount)

	// Top stations by number of trains
	type stationStat struct {
		StationCode string `json:"station_code" gorm:"column:code"`
		StationName string `json:"station_name" gorm:"column:name"`
		TrainCount  int    `json:"train_count" gorm:"column:train_count"`
	}
	var topStations []stationStat
	h.db.Raw(`
		SELECT s.code, s.name, COUNT(DISTINCT ts.train_id) AS train_count
		FROM train_schedules ts
		JOIN stations s ON s.id = ts.station_id
		GROUP BY s.code, s.name
		ORDER BY train_count DESC
		LIMIT 20
	`).Find(&topStations)

	// Seat classes
	var seatClasses []SeatClass
	h.db.Find(&seatClasses)

	c.JSON(http.StatusOK, gin.H{
		"total_trains":       trainCount,
		"total_stations":     stationCount,
		"total_schedules":    scheduleCount,
		"total_seat_classes": seatClassCount,
		"top_stations":       topStations,
		"seat_classes":       seatClasses,
	})
}

// -- GIS Dashboard Integration --

// GetTrainDashboard GET /api/v1/gis/dashboards/train
// Returns train data in GIS dashboard format
func (h *GISTrainHandler) GetTrainDashboard(c *gin.Context) {
	var trains []Train
	h.db.Find(&trains)

	var stations []Station
	h.db.Find(&stations)

	stationMap := make(map[string]Station)
	for _, s := range stations {
		stationMap[s.Code] = s
	}

	// Train routes as markers (origin → destination)
	markers := make([]GISMarker, 0, len(trains))
	for i, t := range trains {
		markers = append(markers, GISMarker{
			ID:       fmt.Sprintf("train-%d", t.ID),
			Name:     fmt.Sprintf("%s (%s)", t.TrainName, t.TrainNumber),
			Lat:      0, // Placeholder — no lat/lng in station data
			Lng:      0,
			Category: "train",
			Icon:     "🚂",
			Color:    "#3b82f6",
			Properties: map[string]interface{}{
				"train_number": t.TrainNumber,
				"from":         t.FromStationCode,
				"to":           t.ToStationCode,
				"index":        i,
			},
		})
	}

	// Top stations as another marker set
	type stationStat struct {
		Code       string `gorm:"column:code"`
		Name       string `gorm:"column:name"`
		TrainCount int    `gorm:"column:train_count"`
	}
	var topStations []stationStat
	h.db.Raw(`
		SELECT s.code, s.name, COUNT(DISTINCT ts.train_id) AS train_count
		FROM train_schedules ts
		JOIN stations s ON s.id = ts.station_id
		GROUP BY s.code, s.name
		ORDER BY train_count DESC
		LIMIT 50
	`).Find(&topStations)

	stationMarkers := make([]GISMarker, 0, len(topStations))
	for _, s := range topStations {
		stationMarkers = append(stationMarkers, GISMarker{
			ID:       fmt.Sprintf("station-%s", s.Code),
			Name:     s.Name,
			Category: "station",
			Icon:     "🚉",
			Color:    "#10b981",
			Properties: map[string]interface{}{
				"code":        s.Code,
				"train_count": s.TrainCount,
			},
		})
	}

	// Build datasets
	var trainCount, stationCount, scheduleCount int64
	h.db.Model(&Train{}).Count(&trainCount)
	h.db.Model(&Station{}).Count(&stationCount)
	h.db.Model(&TrainSchedule{}).Count(&scheduleCount)

	now := time.Now()
	datasets := []GISDataset{
		{
			ID:          "train-stats",
			Name:        "Railway Statistics",
			Description: "Overall railway network statistics",
			Unit:        "count",
			Columns: []DatasetColumn{
				{Key: "metric", Label: "Metric"},
				{Key: "value", Label: "Value"},
			},
			Rows: []map[string]interface{}{
				{"metric": "Total Trains", "value": trainCount},
				{"metric": "Total Stations", "value": stationCount},
				{"metric": "Total Schedule Entries", "value": scheduleCount},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "top-stations",
			Name:        "Top Stations by Trains",
			Description: "Stations with the most train routes passing through",
			Unit:        "trains",
			Columns: []DatasetColumn{
				{Key: "station", Label: "Station"},
				{Key: "code", Label: "Code"},
				{Key: "train_count", Label: "Train Count"},
			},
			Rows: func() []map[string]interface{} {
				rows := make([]map[string]interface{}, 0, len(topStations))
				for _, s := range topStations {
					rows = append(rows, map[string]interface{}{
						"station":     s.Name,
						"code":        s.Code,
						"train_count": s.TrainCount,
					})
				}
				return rows
			}(),
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	dash := &GISDashboardData{
		Type:        "train",
		Title:       "Railway Network Dashboard",
		Description: "Train routes, stations, and schedules from the railway database",
		MapCenter:   [2]float64{22.5, 82.0}, // Center of India
		DefaultZoom: 5,
		MaxBounds:   [2][2]float64{{6, 65}, {38, 98}},
		Layers: []GISLayer{
			{ID: "train-routes", Name: "Train Routes", Type: "marker", Visible: true, Style: LayerStyle{Color: "#3b82f6", Weight: 2, Opacity: 0.8}, CreatedAt: time.Now()},
			{ID: "stations", Name: "Railway Stations", Type: "marker", Visible: true, Style: LayerStyle{Color: "#10b981", Weight: 2, Opacity: 0.8}, CreatedAt: time.Now()},
		},
		Regions:  []GISRegion{},
		Markers:  append(stationMarkers, markers...),
		Datasets: datasets,
		Config: map[string]interface{}{
			"data_source": "postgresql",
			"database":    "train",
		},
	}

	c.JSON(http.StatusOK, dash)
}

// GetTrainDashboardSummary GET /api/v1/gis/dashboards/train/summary
func (h *GISTrainHandler) GetTrainDashboardSummary(c *gin.Context) {
	var trainCount, stationCount, scheduleCount, seatClassCount int64
	h.db.Model(&Train{}).Count(&trainCount)
	h.db.Model(&Station{}).Count(&stationCount)
	h.db.Model(&TrainSchedule{}).Count(&scheduleCount)
	h.db.Model(&SeatClass{}).Count(&seatClassCount)

	c.JSON(http.StatusOK, gin.H{
		"total_trains":    trainCount,
		"total_stations":  stationCount,
		"total_schedules": scheduleCount,
		"seat_classes":    seatClassCount,
		"total_layers":    2,
		"total_markers":   trainCount + stationCount,
		"total_datasets":  2,
	})
}
