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

// GISBDTrainHandler serves Bangladesh Railway data from PostgreSQL bd-train database.
type GISBDTrainHandler struct {
	db *gorm.DB
}

// BDTrain represents a Bangladesh train.
type BDTrain struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"column:name"`
	TrainNumber int    `json:"train_number" gorm:"column:train_number"`
}

func (BDTrain) TableName() string { return "trains" }

// BDStation represents a Bangladesh railway station.
type BDStation struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"column:name"`
}

func (BDStation) TableName() string { return "stations" }

// BDStationGeo represents geographic coordinates for a station.
type BDStationGeo struct {
	StationID   int     `json:"station_id" gorm:"primaryKey;column:station_id"`
	Latitude    float64 `json:"latitude" gorm:"column:latitude"`
	Longitude   float64 `json:"longitude" gorm:"column:longitude"`
	ElevationM  float64 `json:"elevation_m" gorm:"column:elevation_m"`
	District    string  `json:"district" gorm:"column:district"`
	Division    string  `json:"division" gorm:"column:division"`
	CountryCode string  `json:"country_code" gorm:"column:country_code"`
}

func (BDStationGeo) TableName() string { return "station_geo" }

// BDTrainRoute represents a stop in a train route.
type BDTrainRoute struct {
	ID             int    `json:"id" gorm:"primaryKey"`
	TrainID        int    `json:"train_id" gorm:"column:train_id"`
	StationID      int    `json:"station_id" gorm:"column:station_id"`
	StopOrder      int    `json:"stop_order" gorm:"column:stop_order"`
	ArrivalTime    string `json:"arrival_time" gorm:"column:arrival_time"`
	DepartureTime  string `json:"departure_time" gorm:"column:departure_time"`
	HaltMinutes    int    `json:"halt_minutes" gorm:"column:halt_minutes"`
	TravelDuration string `json:"travel_duration" gorm:"column:travel_duration"`
}

func (BDTrainRoute) TableName() string { return "train_routes" }

// BDTrip represents a scheduled trip.
type BDTrip struct {
	ID                   int    `json:"id" gorm:"primaryKey"`
	TrainID              int    `json:"train_id" gorm:"column:train_id"`
	TripNumber           string `json:"trip_number" gorm:"column:trip_number"`
	OriginStationID      int    `json:"origin_station_id" gorm:"column:origin_station_id"`
	DestinationStationID int    `json:"destination_station_id" gorm:"column:destination_station_id"`
	DepartureDatetime    string `json:"departure_datetime" gorm:"column:departure_datetime"`
	ArrivalDatetime      string `json:"arrival_datetime" gorm:"column:arrival_datetime"`
	TravelTime           string `json:"travel_time" gorm:"column:travel_time"`
	IsEidTrip            bool   `json:"is_eid_trip" gorm:"column:is_eid_trip"`
	IsOpenForAll         bool   `json:"is_open_for_all" gorm:"column:is_open_for_all"`
	IsInternational      bool   `json:"is_international" gorm:"column:is_international"`
}

func (BDTrip) TableName() string { return "trips" }

// BDSeatClass represents a seat class.
type BDSeatClass struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Code string `json:"code" gorm:"column:code"`
}

func (BDSeatClass) TableName() string { return "seat_classes" }

// BDRailLine represents a rail line.
type BDRailLine struct {
	ID           int    `json:"id" gorm:"primaryKey"`
	Code         string `json:"code" gorm:"column:code"`
	Name         string `json:"name" gorm:"column:name"`
	ColorHex     string `json:"color_hex" gorm:"column:color_hex"`
	Gauge        string `json:"gauge" gorm:"column:gauge"`
	OperatorName string `json:"operator_name" gorm:"column:operator_name"`
	IsActive     bool   `json:"is_active" gorm:"column:is_active"`
}

func (BDRailLine) TableName() string { return "rail_lines" }

func NewGISBDTrainHandler(db *gorm.DB) *GISBDTrainHandler {
	return &GISBDTrainHandler{db: db}
}

// -- Trains --

// ListTrains GET /api/v1/gis/bd-trains
func (h *GISBDTrainHandler) ListTrains(c *gin.Context) {
	var trains []BDTrain
	q := h.db

	if search := strings.TrimSpace(c.Query("search")); search != "" {
		pattern := "%" + search + "%"
		q = q.Where("name ILIKE ? OR CAST(train_number AS TEXT) ILIKE ?", pattern, pattern)
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
	q.Model(&BDTrain{}).Count(&total)

	if err := q.Order("train_number").Limit(limit).Offset(offset).Find(&trains).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trains":  trains,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
		"country": "BD",
	})
}

// GetTrain GET /api/v1/gis/bd-trains/:id
func (h *GISBDTrainHandler) GetTrain(c *gin.Context) {
	id := c.Param("id")
	var train BDTrain
	if err := h.db.Where("id = ? OR CAST(train_number AS TEXT) = ?", id, id).First(&train).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "train not found"})
		return
	}

	// Get route
	type routeStop struct {
		StopOrder      int    `json:"stop_order"`
		StationName    string `json:"station_name" gorm:"column:name"`
		ArrivalTime    string `json:"arrival_time,omitempty"`
		DepartureTime  string `json:"departure_time,omitempty"`
		HaltMinutes    int    `json:"halt_minutes"`
		TravelDuration string `json:"travel_duration,omitempty"`
	}
	var route []routeStop
	h.db.Table("train_routes tr").
		Select("tr.stop_order, s.name, tr.arrival_time, tr.departure_time, tr.halt_minutes, tr.travel_duration").
		Joins("JOIN stations s ON s.id = tr.station_id").
		Where("tr.train_id = ?", train.ID).
		Order("tr.stop_order").
		Find(&route)

	// Get upcoming trips
	type tripInfo struct {
		TripNumber        string `json:"trip_number"`
		Origin            string `json:"origin" gorm:"column:origin_name"`
		Destination       string `json:"destination" gorm:"column:dest_name"`
		DepartureDatetime string `json:"departure_datetime"`
		ArrivalDatetime   string `json:"arrival_datetime"`
		IsEidTrip         bool   `json:"is_eid_trip"`
	}
	var trips []tripInfo
	h.db.Table("trips t").
		Select("t.trip_number, s1.name AS origin_name, s2.name AS dest_name, t.departure_datetime, t.arrival_datetime, t.is_eid_trip").
		Joins("JOIN stations s1 ON s1.id = t.origin_station_id").
		Joins("JOIN stations s2 ON s2.id = t.destination_station_id").
		Where("t.train_id = ?", train.ID).
		Order("t.departure_datetime").
		Limit(20).
		Find(&trips)

	c.JSON(http.StatusOK, gin.H{
		"train":   train,
		"route":   route,
		"stops":   len(route),
		"trips":   trips,
		"country": "BD",
	})
}

// -- Stations --

// ListStations GET /api/v1/gis/bd-trains/stations
func (h *GISBDTrainHandler) ListStations(c *gin.Context) {
	var stations []BDStation
	q := h.db

	if search := strings.TrimSpace(c.Query("search")); search != "" {
		pattern := "%" + search + "%"
		q = q.Where("name ILIKE ?", pattern)
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
	q.Model(&BDStation{}).Count(&total)

	if err := q.Order("name").Limit(limit).Offset(offset).Find(&stations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Attach geo data where available
	type stationWithGeo struct {
		BDStation
		Latitude  *float64 `json:"latitude,omitempty"`
		Longitude *float64 `json:"longitude,omitempty"`
		District  string   `json:"district,omitempty"`
		Division  string   `json:"division,omitempty"`
	}
	stationIDs := make([]int, len(stations))
	for i, s := range stations {
		stationIDs[i] = s.ID
	}
	var geos []BDStationGeo
	h.db.Where("station_id IN ?", stationIDs).Find(&geos)
	geoMap := make(map[int]BDStationGeo)
	for _, g := range geos {
		geoMap[g.StationID] = g
	}

	result := make([]stationWithGeo, len(stations))
	for i, s := range stations {
		result[i] = stationWithGeo{BDStation: s}
		if g, ok := geoMap[s.ID]; ok {
			result[i].Latitude = &g.Latitude
			result[i].Longitude = &g.Longitude
			result[i].District = g.District
			result[i].Division = g.Division
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"stations": result,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"country":  "BD",
	})
}

// GetStation GET /api/v1/gis/bd-trains/stations/:name
func (h *GISBDTrainHandler) GetStation(c *gin.Context) {
	name := c.Param("name")
	var station BDStation
	if err := h.db.Where("LOWER(name) = LOWER(?)", name).First(&station).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "station not found"})
		return
	}

	// Geo
	var geo BDStationGeo
	hasGeo := h.db.Where("station_id = ?", station.ID).First(&geo).Error == nil

	// Trains passing through
	type trainBrief struct {
		TrainName     string `json:"train_name" gorm:"column:name"`
		TrainNumber   int    `json:"train_number" gorm:"column:train_number"`
		StopOrder     int    `json:"stop_order"`
		ArrivalTime   string `json:"arrival_time"`
		DepartureTime string `json:"departure_time"`
		HaltMinutes   int    `json:"halt_minutes"`
	}
	var trains []trainBrief
	h.db.Table("train_routes tr").
		Select("t.name, t.train_number, tr.stop_order, tr.arrival_time, tr.departure_time, tr.halt_minutes").
		Joins("JOIN trains t ON t.id = tr.train_id").
		Where("tr.station_id = ?", station.ID).
		Order("tr.departure_time").
		Find(&trains)

	resp := gin.H{
		"station": station,
		"trains":  trains,
		"total":   len(trains),
		"country": "BD",
	}
	if hasGeo {
		resp["geo"] = geo
	}
	c.JSON(http.StatusOK, resp)
}

// -- Routes --

// GetTrainRoute GET /api/v1/gis/bd-trains/:id/route
func (h *GISBDTrainHandler) GetTrainRoute(c *gin.Context) {
	id := c.Param("id")
	var train BDTrain
	if err := h.db.Where("id = ? OR CAST(train_number AS TEXT) = ?", id, id).First(&train).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "train not found"})
		return
	}

	type routeStop struct {
		StopOrder      int     `json:"stop_order"`
		StationName    string  `json:"station_name" gorm:"column:sname"`
		ArrivalTime    string  `json:"arrival_time,omitempty"`
		DepartureTime  string  `json:"departure_time,omitempty"`
		HaltMinutes    int     `json:"halt_minutes"`
		Latitude       float64 `json:"latitude,omitempty" gorm:"column:latitude"`
		Longitude      float64 `json:"longitude,omitempty" gorm:"column:longitude"`
		TravelDuration string  `json:"travel_duration,omitempty"`
	}
	var route []routeStop
	h.db.Raw(`
		SELECT tr.stop_order, s.name AS sname, tr.arrival_time, tr.departure_time,
		       tr.halt_minutes, tr.travel_duration,
		       COALESCE(g.latitude, 0) AS latitude, COALESCE(g.longitude, 0) AS longitude
		FROM train_routes tr
		JOIN stations s ON s.id = tr.station_id
		LEFT JOIN station_geo g ON g.station_id = tr.station_id
		WHERE tr.train_id = ?
		ORDER BY tr.stop_order
	`, train.ID).Find(&route)

	c.JSON(http.StatusOK, gin.H{
		"train":   train,
		"route":   route,
		"stops":   len(route),
		"country": "BD",
	})
}

// -- Search --

// SearchTrains GET /api/v1/gis/bd-trains/search?from=XX&to=YY
func (h *GISBDTrainHandler) SearchTrains(c *gin.Context) {
	from := strings.TrimSpace(c.Query("from"))
	to := strings.TrimSpace(c.Query("to"))

	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "both 'from' and 'to' station names are required"})
		return
	}

	type searchResult struct {
		TrainID     int    `json:"train_id"`
		TrainName   string `json:"train_name" gorm:"column:name"`
		TrainNumber int    `json:"train_number" gorm:"column:train_number"`
		FromStop    int    `json:"from_stop_order" gorm:"column:from_stop"`
		ToStop      int    `json:"to_stop_order" gorm:"column:to_stop"`
		FromDepart  string `json:"from_departure" gorm:"column:from_dep"`
		ToArrive    string `json:"to_arrival" gorm:"column:to_arr"`
	}

	var results []searchResult
	h.db.Raw(`
		SELECT t.id AS train_id, t.name, t.train_number,
		       tr1.stop_order AS from_stop, tr2.stop_order AS to_stop,
		       tr1.departure_time AS from_dep, tr2.arrival_time AS to_arr
		FROM train_routes tr1
		JOIN train_routes tr2 ON tr1.train_id = tr2.train_id AND tr2.stop_order > tr1.stop_order
		JOIN stations s1 ON s1.id = tr1.station_id AND LOWER(s1.name) = LOWER(?)
		JOIN stations s2 ON s2.id = tr2.station_id AND LOWER(s2.name) = LOWER(?)
		JOIN trains t ON t.id = tr1.train_id
		ORDER BY tr1.departure_time
	`, from, to).Find(&results)

	c.JSON(http.StatusOK, gin.H{
		"from":    from,
		"to":      to,
		"trains":  results,
		"total":   len(results),
		"country": "BD",
	})
}

// -- Trips --

// ListTrips GET /api/v1/gis/bd-trains/trips
func (h *GISBDTrainHandler) ListTrips(c *gin.Context) {
	limit := 50
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 200 {
		limit = l
	}
	offset := 0
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}

	type tripResult struct {
		ID                int     `json:"id"`
		TripNumber        string  `json:"trip_number"`
		TrainName         string  `json:"train_name" gorm:"column:tname"`
		TrainNumber       int     `json:"train_number" gorm:"column:tnumber"`
		Origin            string  `json:"origin" gorm:"column:origin_name"`
		Destination       string  `json:"destination" gorm:"column:dest_name"`
		DepartureDatetime string  `json:"departure_datetime"`
		ArrivalDatetime   string  `json:"arrival_datetime"`
		TravelTime        string  `json:"travel_time,omitempty"`
		IsEidTrip         bool    `json:"is_eid_trip"`
		IsInternational   bool    `json:"is_international"`
		Fare              float64 `json:"fare,omitempty" gorm:"column:min_fare"`
	}

	var total int64
	h.db.Model(&BDTrip{}).Count(&total)

	var trips []tripResult
	h.db.Raw(`
		SELECT t.id, t.trip_number, tr.name AS tname, tr.train_number AS tnumber,
		       s1.name AS origin_name, s2.name AS dest_name,
		       t.departure_datetime, t.arrival_datetime, t.travel_time,
		       t.is_eid_trip, t.is_international,
		       (SELECT MIN(ts.fare) FROM trip_seats ts WHERE ts.trip_id = t.id) AS min_fare
		FROM trips t
		JOIN trains tr ON tr.id = t.train_id
		JOIN stations s1 ON s1.id = t.origin_station_id
		JOIN stations s2 ON s2.id = t.destination_station_id
		ORDER BY t.departure_datetime
		LIMIT ? OFFSET ?
	`, limit, offset).Find(&trips)

	c.JSON(http.StatusOK, gin.H{
		"trips":   trips,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
		"country": "BD",
	})
}

// GetTripSeats GET /api/v1/gis/bd-trains/trips/:id/seats
func (h *GISBDTrainHandler) GetTripSeats(c *gin.Context) {
	id := c.Param("id")

	type seatInfo struct {
		SeatClass    string  `json:"seat_class" gorm:"column:code"`
		Fare         float64 `json:"fare"`
		VATPercent   float64 `json:"vat_percent"`
		VATAmount    float64 `json:"vat_amount"`
		OnlineSeats  int     `json:"online_seats"`
		OfflineSeats int     `json:"offline_seats"`
	}
	var seats []seatInfo
	h.db.Table("trip_seats ts").
		Select("sc.code, ts.fare, ts.vat_percent, ts.vat_amount, ts.online_seats, ts.offline_seats").
		Joins("JOIN seat_classes sc ON sc.id = ts.seat_class_id").
		Where("ts.trip_id = ?", id).
		Order("ts.fare").
		Find(&seats)

	if len(seats) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "trip not found or has no seat data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trip_id": id,
		"seats":   seats,
		"total":   len(seats),
	})
}

// -- Stats --

// GetStats GET /api/v1/gis/bd-trains/stats
func (h *GISBDTrainHandler) GetStats(c *gin.Context) {
	var trainCount, stationCount, routeCount, tripCount, seatClassCount, tripSeatCount int64
	h.db.Model(&BDTrain{}).Count(&trainCount)
	h.db.Model(&BDStation{}).Count(&stationCount)
	h.db.Model(&BDTrainRoute{}).Count(&routeCount)
	h.db.Model(&BDTrip{}).Count(&tripCount)
	h.db.Model(&BDSeatClass{}).Count(&seatClassCount)
	h.db.Table("trip_seats").Count(&tripSeatCount)

	// Top stations by trains
	type stationStat struct {
		Name       string `json:"name" gorm:"column:name"`
		TrainCount int    `json:"train_count" gorm:"column:train_count"`
	}
	var topStations []stationStat
	h.db.Raw(`
		SELECT s.name, COUNT(DISTINCT tr.train_id) AS train_count
		FROM train_routes tr
		JOIN stations s ON s.id = tr.station_id
		GROUP BY s.name
		ORDER BY train_count DESC
		LIMIT 20
	`).Find(&topStations)

	// Seat classes
	var seatClasses []BDSeatClass
	h.db.Find(&seatClasses)

	// Rail lines
	var railLines []BDRailLine
	h.db.Find(&railLines)

	c.JSON(http.StatusOK, gin.H{
		"total_trains":       trainCount,
		"total_stations":     stationCount,
		"total_routes":       routeCount,
		"total_trips":        tripCount,
		"total_seat_classes": seatClassCount,
		"total_trip_seats":   tripSeatCount,
		"top_stations":       topStations,
		"seat_classes":       seatClasses,
		"rail_lines":         railLines,
		"country":            "BD",
	})
}

// -- GIS Dashboard --

// GetDashboard GET /api/v1/gis/bd-trains/dashboard
func (h *GISBDTrainHandler) GetDashboard(c *gin.Context) {
	now := time.Now()

	var trains []BDTrain
	h.db.Find(&trains)

	// Stations with geo data
	type stationWithGeo struct {
		ID        int     `gorm:"column:id"`
		Name      string  `gorm:"column:name"`
		Latitude  float64 `gorm:"column:latitude"`
		Longitude float64 `gorm:"column:longitude"`
		District  string  `gorm:"column:district"`
		Division  string  `gorm:"column:division"`
	}
	var geoStations []stationWithGeo
	h.db.Raw(`
		SELECT s.id, s.name, g.latitude, g.longitude, g.district, g.division
		FROM stations s
		JOIN station_geo g ON g.station_id = s.id
		WHERE g.latitude IS NOT NULL
	`).Find(&geoStations)

	// All stations for markers (even without geo)
	var allStations []BDStation
	h.db.Find(&allStations)

	// Top stations
	type stationStat struct {
		Name       string `gorm:"column:name"`
		TrainCount int    `gorm:"column:train_count"`
	}
	var topStations []stationStat
	h.db.Raw(`
		SELECT s.name, COUNT(DISTINCT tr.train_id) AS train_count
		FROM train_routes tr
		JOIN stations s ON s.id = tr.station_id
		GROUP BY s.name
		ORDER BY train_count DESC
		LIMIT 30
	`).Find(&topStations)

	// Build station markers (geo stations with coordinates)
	stationMarkers := make([]GISMarker, 0, len(geoStations))
	for _, s := range geoStations {
		stationMarkers = append(stationMarkers, GISMarker{
			ID:       fmt.Sprintf("bd-station-%d", s.ID),
			Name:     s.Name,
			Lat:      s.Latitude,
			Lng:      s.Longitude,
			Category: "station",
			Icon:     "🚉",
			Color:    "#10b981",
			Properties: map[string]interface{}{
				"district": s.District,
				"division": s.Division,
			},
		})
	}

	// Counts
	var trainCount, stationCount, routeCount, tripCount int64
	h.db.Model(&BDTrain{}).Count(&trainCount)
	h.db.Model(&BDStation{}).Count(&stationCount)
	h.db.Model(&BDTrainRoute{}).Count(&routeCount)
	h.db.Model(&BDTrip{}).Count(&tripCount)

	datasets := []GISDataset{
		{
			ID:          "bd-train-stats",
			Name:        "Bangladesh Railway Statistics",
			Description: "Overall railway network statistics",
			Unit:        "count",
			Columns: []DatasetColumn{
				{Key: "metric", Label: "Metric"},
				{Key: "value", Label: "Value"},
			},
			Rows: []map[string]interface{}{
				{"metric": "Total Trains", "value": trainCount},
				{"metric": "Total Stations", "value": stationCount},
				{"metric": "Total Route Stops", "value": routeCount},
				{"metric": "Total Trips", "value": tripCount},
			},
			CreatedAt: now, UpdatedAt: now,
		},
		{
			ID:          "bd-top-stations",
			Name:        "Top Stations by Trains",
			Description: "Bangladesh stations with the most train routes",
			Unit:        "trains",
			Columns: []DatasetColumn{
				{Key: "station", Label: "Station"},
				{Key: "train_count", Label: "Train Count"},
			},
			Rows: func() []map[string]interface{} {
				rows := make([]map[string]interface{}, 0, len(topStations))
				for _, s := range topStations {
					rows = append(rows, map[string]interface{}{
						"station":     s.Name,
						"train_count": s.TrainCount,
					})
				}
				return rows
			}(),
			CreatedAt: now, UpdatedAt: now,
		},
	}

	dash := &GISDashboardData{
		Type:        "bd-train",
		Title:       "Bangladesh Railway Network Dashboard",
		Description: "Train routes, stations, trips and fares from Bangladesh Railway database",
		MapCenter:   [2]float64{23.6850, 90.3563}, // Dhaka
		DefaultZoom: 7,
		MaxBounds:   [2][2]float64{{20.5, 87.5}, {26.7, 92.8}},
		Layers: []GISLayer{
			{ID: "bd-train-routes", Name: "Train Routes", Type: "marker", Visible: true, Style: LayerStyle{Color: "#3b82f6", Weight: 2, Opacity: 0.8}, CreatedAt: now},
			{ID: "bd-stations", Name: "Railway Stations", Type: "marker", Visible: true, Style: LayerStyle{Color: "#10b981", Weight: 2, Opacity: 0.8}, CreatedAt: now},
		},
		Regions:  []GISRegion{},
		Markers:  stationMarkers,
		Datasets: datasets,
		Config: map[string]interface{}{
			"data_source": "postgresql",
			"database":    "bd-train",
			"country":     "BD",
		},
	}

	c.JSON(http.StatusOK, dash)
}

// GetDashboardSummary GET /api/v1/gis/bd-trains/dashboard/summary
func (h *GISBDTrainHandler) GetDashboardSummary(c *gin.Context) {
	var trainCount, stationCount, routeCount, tripCount, seatClassCount int64
	h.db.Model(&BDTrain{}).Count(&trainCount)
	h.db.Model(&BDStation{}).Count(&stationCount)
	h.db.Model(&BDTrainRoute{}).Count(&routeCount)
	h.db.Model(&BDTrip{}).Count(&tripCount)
	h.db.Model(&BDSeatClass{}).Count(&seatClassCount)

	var geoCount int64
	h.db.Table("station_geo").Where("latitude IS NOT NULL").Count(&geoCount)

	c.JSON(http.StatusOK, gin.H{
		"total_trains":       trainCount,
		"total_stations":     stationCount,
		"total_routes":       routeCount,
		"total_trips":        tripCount,
		"total_seat_classes": seatClassCount,
		"geo_stations":       geoCount,
		"total_layers":       2,
		"total_markers":      geoCount,
		"total_datasets":     2,
		"country":            "BD",
	})
}
