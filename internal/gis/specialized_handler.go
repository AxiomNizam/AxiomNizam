package gis

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// GISSpecializedHandler serves category-specific GIS dashboard data.
// Categories: agriculture, industries, medical (domestic Bangladesh),
// satellite, airplane, ship (international).
//
// Deprecated: This is an in-memory map guarded by sync.RWMutex and does NOT
// follow the platform's control-plane architecture. All specialized dashboards
// are seeded at startup and lost on restart. New dashboard types should be
// authored through the API Builder instead.
type GISSpecializedHandler struct {
	mu         sync.RWMutex
	dashboards map[string]*GISDashboardData
}

// GISDashboardData bundles all data for one specialized dashboard.
type GISDashboardData struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	MapCenter   [2]float64             `json:"mapCenter"`
	DefaultZoom int                    `json:"defaultZoom"`
	MaxBounds   [2][2]float64          `json:"maxBounds"`
	Layers      []GISLayer             `json:"layers"`
	Regions     []GISRegion            `json:"regions"`
	Markers     []GISMarker            `json:"markers"`
	Datasets    []GISDataset           `json:"datasets"`
	Config      map[string]interface{} `json:"config"`
}

// NewGISSpecializedHandler creates and seeds the specialized GIS handler.
func NewGISSpecializedHandler() *GISSpecializedHandler {
	h := &GISSpecializedHandler{
		dashboards: make(map[string]*GISDashboardData),
	}
	h.seedAgriculture()
	h.seedIndustries()
	h.seedMedical()
	h.seedSatellite()
	h.seedAirplane()
	h.seedShip()
	h.seedTrainPlaceholder()
	return h
}

// ListDashboardTypes returns available dashboard categories.
func (h *GISSpecializedHandler) ListDashboardTypes(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	types := make([]map[string]interface{}, 0, len(h.dashboards))
	for key, d := range h.dashboards {
		types = append(types, map[string]interface{}{
			"id":          key,
			"title":       d.Title,
			"description": d.Description,
			"scope":       categorizeScope(key),
			"layers":      len(d.Layers),
			"regions":     len(d.Regions),
			"markers":     len(d.Markers),
			"datasets":    len(d.Datasets),
		})
	}
	c.JSON(http.StatusOK, types)
}

// GetDashboard returns all data for one dashboard type.
func (h *GISSpecializedHandler) GetDashboard(c *gin.Context) {
	dashType := c.Param("type")
	h.mu.RLock()
	defer h.mu.RUnlock()

	dash, ok := h.dashboards[dashType]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard type not found", "available": dashboardKeys(h.dashboards)})
		return
	}
	c.JSON(http.StatusOK, dash)
}

// GetDashboardSummary returns summary stats for one dashboard type.
func (h *GISSpecializedHandler) GetDashboardSummary(c *gin.Context) {
	dashType := c.Param("type")
	h.mu.RLock()
	defer h.mu.RUnlock()

	dash, ok := h.dashboards[dashType]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard type not found"})
		return
	}

	regionsByType := make(map[string]int)
	for _, r := range dash.Regions {
		regionsByType[r.Type]++
	}

	c.JSON(http.StatusOK, GISSummary{
		TotalLayers:   len(dash.Layers),
		TotalRegions:  len(dash.Regions),
		TotalMarkers:  len(dash.Markers),
		TotalDatasets: len(dash.Datasets),
		RegionsByType: regionsByType,
		MapCenter:     dash.MapCenter,
		DefaultZoom:   dash.DefaultZoom,
	})
}

func categorizeScope(key string) string {
	switch key {
	case "agriculture", "industries", "medical":
		return "domestic"
	case "train", "bd-train":
		return "railway"
	default:
		return "international"
	}
}

func dashboardKeys(m map[string]*GISDashboardData) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// --- Seed methods ---

func (h *GISSpecializedHandler) seedAgriculture() {
	now := time.Now()
	d := &GISDashboardData{
		Type:        "agriculture",
		Title:       "Agriculture Dashboard",
		Description: "Crop production, irrigation, soil types and agricultural infrastructure across Bangladesh",
		MapCenter:   [2]float64{23.6850, 90.3563},
		DefaultZoom: 7,
		MaxBounds:   [2][2]float64{{18, 85}, {28, 96}},
		Config: map[string]interface{}{
			"legendTitle": "Rice Production (MT)",
			"colorField":  "rice_production",
			"iconTheme":   "agriculture",
		},
		Layers: []GISLayer{
			{ID: "agri-zones", Name: "Agricultural Zones", Type: "geojson", Visible: true, Style: LayerStyle{Color: "#2ecc71", Weight: 2, Opacity: 0.8, FillColor: "#27ae60", FillOpacity: 0.2}, CreatedAt: now},
			{ID: "irrigation", Name: "Irrigation Coverage", Type: "geojson", Visible: false, Style: LayerStyle{Color: "#3498db", Weight: 1, Opacity: 0.6, FillColor: "#2980b9", FillOpacity: 0.15}, CreatedAt: now},
			{ID: "agri-markers", Name: "Agricultural Facilities", Type: "marker", Visible: true, CreatedAt: now},
			{ID: "soil-types", Name: "Soil Type Overlay", Type: "heatmap", Visible: false, CreatedAt: now},
		},
		Regions: []GISRegion{
			{ID: "ag-dhaka", Name: "Dhaka", Type: "division", Center: [2]float64{23.8103, 90.4125}, Properties: map[string]interface{}{"rice_production": 5800000, "wheat_production": 320000, "jute_production": 890000, "tea_production": 0, "arable_land_pct": 72, "irrigation_pct": 65, "crop_intensity": 192, "color": "#a8e6cf"}},
			{ID: "ag-chattogram", Name: "Chattogram", Type: "division", Center: [2]float64{22.3569, 91.7832}, Properties: map[string]interface{}{"rice_production": 4200000, "wheat_production": 85000, "jute_production": 340000, "tea_production": 82000, "arable_land_pct": 58, "irrigation_pct": 42, "crop_intensity": 168, "color": "#dcedc1"}},
			{ID: "ag-rajshahi", Name: "Rajshahi", Type: "division", Center: [2]float64{24.3745, 88.6042}, Properties: map[string]interface{}{"rice_production": 7100000, "wheat_production": 980000, "jute_production": 210000, "tea_production": 0, "arable_land_pct": 82, "irrigation_pct": 78, "crop_intensity": 218, "color": "#ffd3b6"}},
			{ID: "ag-khulna", Name: "Khulna", Type: "division", Center: [2]float64{22.8456, 89.5403}, Properties: map[string]interface{}{"rice_production": 3900000, "wheat_production": 190000, "jute_production": 520000, "tea_production": 0, "arable_land_pct": 64, "irrigation_pct": 55, "crop_intensity": 175, "color": "#ffaaa5"}},
			{ID: "ag-barishal", Name: "Barishal", Type: "division", Center: [2]float64{22.7010, 90.3535}, Properties: map[string]interface{}{"rice_production": 2800000, "wheat_production": 45000, "jute_production": 180000, "tea_production": 0, "arable_land_pct": 68, "irrigation_pct": 38, "crop_intensity": 155, "color": "#b5ead7"}},
			{ID: "ag-sylhet", Name: "Sylhet", Type: "division", Center: [2]float64{24.8949, 91.8687}, Properties: map[string]interface{}{"rice_production": 3400000, "wheat_production": 15000, "jute_production": 95000, "tea_production": 66000, "arable_land_pct": 52, "irrigation_pct": 35, "crop_intensity": 148, "color": "#c7ceea"}},
			{ID: "ag-rangpur", Name: "Rangpur", Type: "division", Center: [2]float64{25.7439, 89.2752}, Properties: map[string]interface{}{"rice_production": 6200000, "wheat_production": 720000, "jute_production": 380000, "tea_production": 0, "arable_land_pct": 78, "irrigation_pct": 72, "crop_intensity": 205, "color": "#e2f0cb"}},
			{ID: "ag-mymensingh", Name: "Mymensingh", Type: "division", Center: [2]float64{24.7471, 90.4203}, Properties: map[string]interface{}{"rice_production": 4500000, "wheat_production": 180000, "jute_production": 620000, "tea_production": 0, "arable_land_pct": 74, "irrigation_pct": 58, "crop_intensity": 185, "color": "#ffeead"}},
		},
		Markers: []GISMarker{
			{ID: "ag-m1", Name: "BRRI Headquarters", Lat: 23.9923, Lng: 90.4075, Category: "research", Color: "#2ecc71", Properties: map[string]interface{}{"type": "Bangladesh Rice Research Institute", "established": 1970}},
			{ID: "ag-m2", Name: "BARI Research Center", Lat: 24.3850, Lng: 88.5922, Category: "research", Color: "#27ae60", Properties: map[string]interface{}{"type": "Agricultural Research Institute", "focus": "Wheat & Vegetables"}},
			{ID: "ag-m3", Name: "Rangpur Cold Storage Hub", Lat: 25.7544, Lng: 89.2445, Category: "cold_storage", Color: "#3498db", Properties: map[string]interface{}{"capacity_mt": 50000, "type": "Potato & Grain Storage"}},
			{ID: "ag-m4", Name: "Khulna Shrimp Processing", Lat: 22.8200, Lng: 89.5500, Category: "processing", Color: "#e67e22", Properties: map[string]interface{}{"type": "Aquaculture Processing Zone", "export_value_usd": 580000000}},
			{ID: "ag-m5", Name: "Sylhet Tea Gardens", Lat: 24.3636, Lng: 91.7329, Category: "tea_estate", Color: "#16a085", Properties: map[string]interface{}{"area_hectares": 23500, "annual_production_kg": 82000000}},
			{ID: "ag-m6", Name: "Mymensingh Dairy Hub", Lat: 24.7540, Lng: 90.4073, Category: "dairy", Color: "#f39c12", Properties: map[string]interface{}{"type": "Milk Collection & Processing", "daily_capacity_liters": 120000}},
			{ID: "ag-m7", Name: "Comilla Seed Multiplication Farm", Lat: 23.4607, Lng: 91.1809, Category: "seed_bank", Color: "#8e44ad", Properties: map[string]interface{}{"type": "Government Seed Farm", "varieties": 45}},
			{ID: "ag-m8", Name: "Rajshahi Mango Processing", Lat: 24.3636, Lng: 88.6241, Category: "processing", Color: "#d35400", Properties: map[string]interface{}{"type": "Fruit Processing & Export", "seasonal_output_mt": 35000}},
			{ID: "ag-m9", Name: "Barishal Floating Market", Lat: 22.7100, Lng: 90.3700, Category: "market", Color: "#1abc9c", Properties: map[string]interface{}{"type": "Traditional Floating Market", "daily_traders": 200}},
			{ID: "ag-m10", Name: "Dinajpur Wheat Belt Center", Lat: 25.6279, Lng: 88.6332, Category: "research", Color: "#2ecc71", Properties: map[string]interface{}{"type": "Wheat Research Station", "focus": "High-yield varieties"}},
		},
		Datasets: []GISDataset{
			{
				ID: "crop-production", Name: "Crop Production by Division", Description: "Annual crop output in metric tons", Unit: "MT",
				Columns: []DatasetColumn{
					{Key: "name", Label: "Division", Type: "string"},
					{Key: "rice_production", Label: "Rice (MT)", Type: "number"},
					{Key: "wheat_production", Label: "Wheat (MT)", Type: "number"},
					{Key: "jute_production", Label: "Jute (MT)", Type: "number"},
					{Key: "tea_production", Label: "Tea (MT)", Type: "number"},
				},
				Rows: []map[string]interface{}{
					{"name": "Dhaka", "rice_production": 5800000, "wheat_production": 320000, "jute_production": 890000, "tea_production": 0},
					{"name": "Chattogram", "rice_production": 4200000, "wheat_production": 85000, "jute_production": 340000, "tea_production": 82000},
					{"name": "Rajshahi", "rice_production": 7100000, "wheat_production": 980000, "jute_production": 210000, "tea_production": 0},
					{"name": "Khulna", "rice_production": 3900000, "wheat_production": 190000, "jute_production": 520000, "tea_production": 0},
					{"name": "Barishal", "rice_production": 2800000, "wheat_production": 45000, "jute_production": 180000, "tea_production": 0},
					{"name": "Sylhet", "rice_production": 3400000, "wheat_production": 15000, "jute_production": 95000, "tea_production": 66000},
					{"name": "Rangpur", "rice_production": 6200000, "wheat_production": 720000, "jute_production": 380000, "tea_production": 0},
					{"name": "Mymensingh", "rice_production": 4500000, "wheat_production": 180000, "jute_production": 620000, "tea_production": 0},
				},
				CreatedAt: now, UpdatedAt: now,
			},
			{
				ID: "land-irrigation", Name: "Land & Irrigation Coverage", Description: "Arable land and irrigation statistics", Unit: "%",
				Columns: []DatasetColumn{
					{Key: "name", Label: "Division", Type: "string"},
					{Key: "arable_land_pct", Label: "Arable Land %", Type: "number"},
					{Key: "irrigation_pct", Label: "Irrigated %", Type: "number"},
					{Key: "crop_intensity", Label: "Crop Intensity", Type: "number"},
				},
				Rows: []map[string]interface{}{
					{"name": "Dhaka", "arable_land_pct": 72, "irrigation_pct": 65, "crop_intensity": 192},
					{"name": "Chattogram", "arable_land_pct": 58, "irrigation_pct": 42, "crop_intensity": 168},
					{"name": "Rajshahi", "arable_land_pct": 82, "irrigation_pct": 78, "crop_intensity": 218},
					{"name": "Khulna", "arable_land_pct": 64, "irrigation_pct": 55, "crop_intensity": 175},
					{"name": "Barishal", "arable_land_pct": 68, "irrigation_pct": 38, "crop_intensity": 155},
					{"name": "Sylhet", "arable_land_pct": 52, "irrigation_pct": 35, "crop_intensity": 148},
					{"name": "Rangpur", "arable_land_pct": 78, "irrigation_pct": 72, "crop_intensity": 205},
					{"name": "Mymensingh", "arable_land_pct": 74, "irrigation_pct": 58, "crop_intensity": 185},
				},
				CreatedAt: now, UpdatedAt: now,
			},
		},
	}
	h.dashboards["agriculture"] = d
}

func (h *GISSpecializedHandler) seedIndustries() {
	now := time.Now()
	d := &GISDashboardData{
		Type:        "industries",
		Title:       "Industries Dashboard",
		Description: "Industrial zones, factories, economic corridors and manufacturing data across Bangladesh",
		MapCenter:   [2]float64{23.6850, 90.3563},
		DefaultZoom: 7,
		MaxBounds:   [2][2]float64{{18, 85}, {28, 96}},
		Config: map[string]interface{}{
			"legendTitle": "Industrial Output (Cr BDT)",
			"colorField":  "industrial_output",
			"iconTheme":   "industries",
		},
		Layers: []GISLayer{
			{ID: "ind-zones", Name: "Industrial Zones", Type: "geojson", Visible: true, Style: LayerStyle{Color: "#e74c3c", Weight: 2, Opacity: 0.8, FillColor: "#c0392b", FillOpacity: 0.2}, CreatedAt: now},
			{ID: "epz", Name: "Export Processing Zones", Type: "geojson", Visible: true, Style: LayerStyle{Color: "#f39c12", Weight: 2, Opacity: 0.8, FillColor: "#e67e22", FillOpacity: 0.15}, CreatedAt: now},
			{ID: "ind-markers", Name: "Industrial Facilities", Type: "marker", Visible: true, CreatedAt: now},
			{ID: "transport", Name: "Transport Network", Type: "geojson", Visible: false, Style: LayerStyle{Color: "#95a5a6", Weight: 1, Opacity: 0.5}, CreatedAt: now},
		},
		Regions: []GISRegion{
			{ID: "in-dhaka", Name: "Dhaka", Type: "division", Center: [2]float64{23.8103, 90.4125}, Properties: map[string]interface{}{"industrial_output": 285000, "factories": 12400, "employment": 4200000, "garment_units": 3200, "gdp_share_pct": 36.2, "sez_count": 3, "color": "#fadbd8"}},
			{ID: "in-chattogram", Name: "Chattogram", Type: "division", Center: [2]float64{22.3569, 91.7832}, Properties: map[string]interface{}{"industrial_output": 198000, "factories": 7800, "employment": 2800000, "garment_units": 1100, "gdp_share_pct": 22.5, "sez_count": 4, "color": "#d5f5e3"}},
			{ID: "in-rajshahi", Name: "Rajshahi", Type: "division", Center: [2]float64{24.3745, 88.6042}, Properties: map[string]interface{}{"industrial_output": 62000, "factories": 2900, "employment": 890000, "garment_units": 180, "gdp_share_pct": 8.1, "sez_count": 1, "color": "#fef9e7"}},
			{ID: "in-khulna", Name: "Khulna", Type: "division", Center: [2]float64{22.8456, 89.5403}, Properties: map[string]interface{}{"industrial_output": 78000, "factories": 3400, "employment": 1100000, "garment_units": 320, "gdp_share_pct": 9.8, "sez_count": 2, "color": "#eaf2f8"}},
			{ID: "in-barishal", Name: "Barishal", Type: "division", Center: [2]float64{22.7010, 90.3535}, Properties: map[string]interface{}{"industrial_output": 21000, "factories": 980, "employment": 340000, "garment_units": 45, "gdp_share_pct": 3.2, "sez_count": 0, "color": "#f5eef8"}},
			{ID: "in-sylhet", Name: "Sylhet", Type: "division", Center: [2]float64{24.8949, 91.8687}, Properties: map[string]interface{}{"industrial_output": 45000, "factories": 1800, "employment": 520000, "garment_units": 85, "gdp_share_pct": 5.4, "sez_count": 1, "color": "#eafaf1"}},
			{ID: "in-rangpur", Name: "Rangpur", Type: "division", Center: [2]float64{25.7439, 89.2752}, Properties: map[string]interface{}{"industrial_output": 38000, "factories": 1600, "employment": 480000, "garment_units": 120, "gdp_share_pct": 4.6, "sez_count": 1, "color": "#fef5e7"}},
			{ID: "in-mymensingh", Name: "Mymensingh", Type: "division", Center: [2]float64{24.7471, 90.4203}, Properties: map[string]interface{}{"industrial_output": 32000, "factories": 1400, "employment": 410000, "garment_units": 95, "gdp_share_pct": 3.9, "sez_count": 0, "color": "#fdedec"}},
		},
		Markers: []GISMarker{
			{ID: "in-m1", Name: "Dhaka EPZ", Lat: 23.7461, Lng: 90.3742, Category: "epz", Color: "#e74c3c", Properties: map[string]interface{}{"type": "Export Processing Zone", "enterprises": 94, "employment": 78000, "established": 1993}},
			{ID: "in-m2", Name: "Chattogram EPZ", Lat: 22.3467, Lng: 91.8123, Category: "epz", Color: "#e74c3c", Properties: map[string]interface{}{"type": "Export Processing Zone", "enterprises": 158, "employment": 210000, "established": 1983}},
			{ID: "in-m3", Name: "Karnaphuli EPZ", Lat: 22.2900, Lng: 91.8300, Category: "epz", Color: "#c0392b", Properties: map[string]interface{}{"type": "Export Processing Zone", "enterprises": 52, "employment": 42000, "established": 2006}},
			{ID: "in-m4", Name: "Gazipur Industrial Belt", Lat: 24.0023, Lng: 90.4253, Category: "industrial_zone", Color: "#f39c12", Properties: map[string]interface{}{"type": "Garment Manufacturing Hub", "factories": 1200, "employment": 850000}},
			{ID: "in-m5", Name: "Narayanganj Textile Hub", Lat: 23.6238, Lng: 90.5000, Category: "textile", Color: "#9b59b6", Properties: map[string]interface{}{"type": "Textile & Dyeing Hub", "factories": 450, "export_usd": 2800000000}},
			{ID: "in-m6", Name: "Mongla Port Industrial Area", Lat: 22.4700, Lng: 89.6000, Category: "port_industry", Color: "#3498db", Properties: map[string]interface{}{"type": "Port-Based Industrial Zone", "cargo_mt": 4500000}},
			{ID: "in-m7", Name: "Bangabandhu Hi-Tech City", Lat: 24.0100, Lng: 90.4200, Category: "tech_park", Color: "#1abc9c", Properties: map[string]interface{}{"type": "IT & Hi-Tech Park", "companies": 120, "it_professionals": 15000}},
			{ID: "in-m8", Name: "Mirsarai Economic Zone", Lat: 22.7700, Lng: 91.5700, Category: "sez", Color: "#e67e22", Properties: map[string]interface{}{"type": "Special Economic Zone", "area_acres": 30000, "investment_usd": 15000000000}},
			{ID: "in-m9", Name: "Ishwardi EPZ", Lat: 24.1300, Lng: 89.0700, Category: "epz", Color: "#e74c3c", Properties: map[string]interface{}{"type": "Export Processing Zone", "enterprises": 38, "employment": 18000}},
			{ID: "in-m10", Name: "Payra Power Plant", Lat: 22.3500, Lng: 90.3200, Category: "power", Color: "#f1c40f", Properties: map[string]interface{}{"type": "Coal Power Plant", "capacity_mw": 1320, "status": "Operational"}},
		},
		Datasets: []GISDataset{
			{
				ID: "industrial-output", Name: "Industrial Output by Division", Description: "Manufacturing & industrial output data", Unit: "Cr BDT",
				Columns: []DatasetColumn{
					{Key: "name", Label: "Division", Type: "string"},
					{Key: "industrial_output", Label: "Output (Cr BDT)", Type: "number"},
					{Key: "factories", Label: "Factories", Type: "number"},
					{Key: "employment", Label: "Employment", Type: "number"},
					{Key: "garment_units", Label: "Garment Units", Type: "number"},
				},
				Rows: []map[string]interface{}{
					{"name": "Dhaka", "industrial_output": 285000, "factories": 12400, "employment": 4200000, "garment_units": 3200},
					{"name": "Chattogram", "industrial_output": 198000, "factories": 7800, "employment": 2800000, "garment_units": 1100},
					{"name": "Rajshahi", "industrial_output": 62000, "factories": 2900, "employment": 890000, "garment_units": 180},
					{"name": "Khulna", "industrial_output": 78000, "factories": 3400, "employment": 1100000, "garment_units": 320},
					{"name": "Barishal", "industrial_output": 21000, "factories": 980, "employment": 340000, "garment_units": 45},
					{"name": "Sylhet", "industrial_output": 45000, "factories": 1800, "employment": 520000, "garment_units": 85},
					{"name": "Rangpur", "industrial_output": 38000, "factories": 1600, "employment": 480000, "garment_units": 120},
					{"name": "Mymensingh", "industrial_output": 32000, "factories": 1400, "employment": 410000, "garment_units": 95},
				},
				CreatedAt: now, UpdatedAt: now,
			},
			{
				ID: "gdp-contribution", Name: "GDP Contribution & SEZ", Description: "Economic contribution and special zones", Unit: "%",
				Columns: []DatasetColumn{
					{Key: "name", Label: "Division", Type: "string"},
					{Key: "gdp_share_pct", Label: "GDP Share %", Type: "number"},
					{Key: "sez_count", Label: "SEZ Count", Type: "number"},
				},
				Rows: []map[string]interface{}{
					{"name": "Dhaka", "gdp_share_pct": 36.2, "sez_count": 3},
					{"name": "Chattogram", "gdp_share_pct": 22.5, "sez_count": 4},
					{"name": "Rajshahi", "gdp_share_pct": 8.1, "sez_count": 1},
					{"name": "Khulna", "gdp_share_pct": 9.8, "sez_count": 2},
					{"name": "Barishal", "gdp_share_pct": 3.2, "sez_count": 0},
					{"name": "Sylhet", "gdp_share_pct": 5.4, "sez_count": 1},
					{"name": "Rangpur", "gdp_share_pct": 4.6, "sez_count": 1},
					{"name": "Mymensingh", "gdp_share_pct": 3.9, "sez_count": 0},
				},
				CreatedAt: now, UpdatedAt: now,
			},
		},
	}
	h.dashboards["industries"] = d
}

func (h *GISSpecializedHandler) seedMedical() {
	now := time.Now()
	d := &GISDashboardData{
		Type:        "medical",
		Title:       "Medical & Health Dashboard",
		Description: "Healthcare infrastructure, disease surveillance, EPI coverage and hospital data across Bangladesh",
		MapCenter:   [2]float64{23.6850, 90.3563},
		DefaultZoom: 7,
		MaxBounds:   [2][2]float64{{18, 85}, {28, 96}},
		Config: map[string]interface{}{
			"legendTitle": "EPI Coverage %",
			"colorField":  "epi_coverage",
			"iconTheme":   "medical",
		},
		Layers: []GISLayer{
			{ID: "health-zones", Name: "Health Districts", Type: "geojson", Visible: true, Style: LayerStyle{Color: "#e74c3c", Weight: 2, Opacity: 0.8, FillColor: "#e74c3c", FillOpacity: 0.15}, CreatedAt: now},
			{ID: "epi-coverage", Name: "EPI Vaccination Coverage", Type: "heatmap", Visible: true, Style: LayerStyle{Color: "#2ecc71", FillOpacity: 0.3}, CreatedAt: now},
			{ID: "med-markers", Name: "Health Facilities", Type: "marker", Visible: true, CreatedAt: now},
			{ID: "disease-hotspots", Name: "Disease Hotspots", Type: "heatmap", Visible: false, Style: LayerStyle{Color: "#e74c3c", FillOpacity: 0.4}, CreatedAt: now},
		},
		Regions: []GISRegion{
			{ID: "md-dhaka", Name: "Dhaka", Type: "division", Center: [2]float64{23.8103, 90.4125}, Properties: map[string]interface{}{"hospitals": 245, "beds": 42000, "doctors": 18500, "epi_coverage": 92, "community_clinics": 3200, "maternal_mortality": 156, "child_mortality": 22, "color": "#fadbd8"}},
			{ID: "md-chattogram", Name: "Chattogram", Type: "division", Center: [2]float64{22.3569, 91.7832}, Properties: map[string]interface{}{"hospitals": 178, "beds": 28000, "doctors": 11200, "epi_coverage": 88, "community_clinics": 2800, "maternal_mortality": 172, "child_mortality": 26, "color": "#d5f5e3"}},
			{ID: "md-rajshahi", Name: "Rajshahi", Type: "division", Center: [2]float64{24.3745, 88.6042}, Properties: map[string]interface{}{"hospitals": 98, "beds": 14500, "doctors": 5800, "epi_coverage": 85, "community_clinics": 2100, "maternal_mortality": 189, "child_mortality": 30, "color": "#fef9e7"}},
			{ID: "md-khulna", Name: "Khulna", Type: "division", Center: [2]float64{22.8456, 89.5403}, Properties: map[string]interface{}{"hospitals": 112, "beds": 16800, "doctors": 6200, "epi_coverage": 83, "community_clinics": 2400, "maternal_mortality": 195, "child_mortality": 28, "color": "#eaf2f8"}},
			{ID: "md-barishal", Name: "Barishal", Type: "division", Center: [2]float64{22.7010, 90.3535}, Properties: map[string]interface{}{"hospitals": 52, "beds": 7200, "doctors": 2600, "epi_coverage": 78, "community_clinics": 1400, "maternal_mortality": 215, "child_mortality": 34, "color": "#f5eef8"}},
			{ID: "md-sylhet", Name: "Sylhet", Type: "division", Center: [2]float64{24.8949, 91.8687}, Properties: map[string]interface{}{"hospitals": 65, "beds": 9800, "doctors": 3400, "epi_coverage": 80, "community_clinics": 1200, "maternal_mortality": 205, "child_mortality": 32, "color": "#eafaf1"}},
			{ID: "md-rangpur", Name: "Rangpur", Type: "division", Center: [2]float64{25.7439, 89.2752}, Properties: map[string]interface{}{"hospitals": 78, "beds": 11200, "doctors": 4200, "epi_coverage": 86, "community_clinics": 1800, "maternal_mortality": 185, "child_mortality": 29, "color": "#fef5e7"}},
			{ID: "md-mymensingh", Name: "Mymensingh", Type: "division", Center: [2]float64{24.7471, 90.4203}, Properties: map[string]interface{}{"hospitals": 58, "beds": 8500, "doctors": 3100, "epi_coverage": 82, "community_clinics": 1500, "maternal_mortality": 198, "child_mortality": 31, "color": "#fdedec"}},
		},
		Markers: []GISMarker{
			{ID: "md-m1", Name: "Dhaka Medical College Hospital", Lat: 23.7259, Lng: 90.3961, Category: "hospital", Color: "#e74c3c", Properties: map[string]interface{}{"type": "Teaching Hospital", "beds": 2600, "established": 1946, "specialties": 32}},
			{ID: "md-m2", Name: "BSMMU", Lat: 23.7395, Lng: 90.3933, Category: "hospital", Color: "#c0392b", Properties: map[string]interface{}{"type": "Specialized University Hospital", "beds": 1850, "departments": 45}},
			{ID: "md-m3", Name: "Chattogram Medical College", Lat: 22.3559, Lng: 91.8300, Category: "hospital", Color: "#e74c3c", Properties: map[string]interface{}{"type": "Teaching Hospital", "beds": 1900, "established": 1957}},
			{ID: "md-m4", Name: "ICDDR,B", Lat: 23.7466, Lng: 90.3728, Category: "research", Color: "#3498db", Properties: map[string]interface{}{"type": "International Research Center", "focus": "Diarrheal Disease & Nutrition", "established": 1960}},
			{ID: "md-m5", Name: "National Institute of Diseases", Lat: 23.7450, Lng: 90.3800, Category: "research", Color: "#2980b9", Properties: map[string]interface{}{"type": "Disease Surveillance Center", "monitoring": "Dengue, TB, Malaria"}},
			{ID: "md-m6", Name: "Rajshahi Medical College", Lat: 24.3700, Lng: 88.6000, Category: "hospital", Color: "#e74c3c", Properties: map[string]interface{}{"type": "Teaching Hospital", "beds": 1100, "established": 1958}},
			{ID: "md-m7", Name: "Cox's Bazar Health Complex", Lat: 21.4350, Lng: 92.0095, Category: "clinic", Color: "#2ecc71", Properties: map[string]interface{}{"type": "Refugee Health Hub", "patients_daily": 3500, "ngos": 12}},
			{ID: "md-m8", Name: "Sylhet MAG Osmani Medical", Lat: 24.8980, Lng: 91.8720, Category: "hospital", Color: "#e74c3c", Properties: map[string]interface{}{"type": "Teaching Hospital", "beds": 800, "established": 1962}},
			{ID: "md-m9", Name: "EPI Central Warehouse", Lat: 23.7500, Lng: 90.3700, Category: "vaccine", Color: "#f1c40f", Properties: map[string]interface{}{"type": "National Vaccine Storage", "vaccines_stored": 14, "cold_chain_capacity": 500000}},
			{ID: "md-m10", Name: "Khulna Blood Transfusion Center", Lat: 22.8200, Lng: 89.5600, Category: "blood_bank", Color: "#e74c3c", Properties: map[string]interface{}{"type": "Regional Blood Bank", "daily_collections": 120, "blood_types": 8}},
		},
		Datasets: []GISDataset{
			{
				ID: "health-infrastructure", Name: "Health Infrastructure by Division", Description: "Hospitals, beds, doctors per division", Unit: "count",
				Columns: []DatasetColumn{
					{Key: "name", Label: "Division", Type: "string"},
					{Key: "hospitals", Label: "Hospitals", Type: "number"},
					{Key: "beds", Label: "Hospital Beds", Type: "number"},
					{Key: "doctors", Label: "Doctors", Type: "number"},
					{Key: "community_clinics", Label: "Community Clinics", Type: "number"},
				},
				Rows: []map[string]interface{}{
					{"name": "Dhaka", "hospitals": 245, "beds": 42000, "doctors": 18500, "community_clinics": 3200},
					{"name": "Chattogram", "hospitals": 178, "beds": 28000, "doctors": 11200, "community_clinics": 2800},
					{"name": "Rajshahi", "hospitals": 98, "beds": 14500, "doctors": 5800, "community_clinics": 2100},
					{"name": "Khulna", "hospitals": 112, "beds": 16800, "doctors": 6200, "community_clinics": 2400},
					{"name": "Barishal", "hospitals": 52, "beds": 7200, "doctors": 2600, "community_clinics": 1400},
					{"name": "Sylhet", "hospitals": 65, "beds": 9800, "doctors": 3400, "community_clinics": 1200},
					{"name": "Rangpur", "hospitals": 78, "beds": 11200, "doctors": 4200, "community_clinics": 1800},
					{"name": "Mymensingh", "hospitals": 58, "beds": 8500, "doctors": 3100, "community_clinics": 1500},
				},
				CreatedAt: now, UpdatedAt: now,
			},
			{
				ID: "epi-vaccination", Name: "EPI & Mortality Statistics", Description: "Vaccination coverage and mortality rates", Unit: "mixed",
				Columns: []DatasetColumn{
					{Key: "name", Label: "Division", Type: "string"},
					{Key: "epi_coverage", Label: "EPI Coverage %", Type: "number"},
					{Key: "maternal_mortality", Label: "Maternal Mortality", Type: "number"},
					{Key: "child_mortality", Label: "Child Mortality (per 1000)", Type: "number"},
				},
				Rows: []map[string]interface{}{
					{"name": "Dhaka", "epi_coverage": 92, "maternal_mortality": 156, "child_mortality": 22},
					{"name": "Chattogram", "epi_coverage": 88, "maternal_mortality": 172, "child_mortality": 26},
					{"name": "Rajshahi", "epi_coverage": 85, "maternal_mortality": 189, "child_mortality": 30},
					{"name": "Khulna", "epi_coverage": 83, "maternal_mortality": 195, "child_mortality": 28},
					{"name": "Barishal", "epi_coverage": 78, "maternal_mortality": 215, "child_mortality": 34},
					{"name": "Sylhet", "epi_coverage": 80, "maternal_mortality": 205, "child_mortality": 32},
					{"name": "Rangpur", "epi_coverage": 86, "maternal_mortality": 185, "child_mortality": 29},
					{"name": "Mymensingh", "epi_coverage": 82, "maternal_mortality": 198, "child_mortality": 31},
				},
				CreatedAt: now, UpdatedAt: now,
			},
		},
	}
	h.dashboards["medical"] = d
}

func (h *GISSpecializedHandler) seedSatellite() {
	now := time.Now()
	d := &GISDashboardData{
		Type:        "satellite",
		Title:       "Satellite Tracking Dashboard",
		Description: "Real-time satellite positions, orbital tracks, ground stations and coverage zones worldwide",
		MapCenter:   [2]float64{20, 0},
		DefaultZoom: 2,
		MaxBounds:   [2][2]float64{{-85, -180}, {85, 180}},
		Config: map[string]interface{}{
			"legendTitle": "Orbit Type",
			"colorField":  "orbit_type",
			"iconTheme":   "satellite",
			"scope":       "international",
		},
		Layers: []GISLayer{
			{ID: "sat-orbits", Name: "Orbital Tracks", Type: "geojson", Visible: true, Style: LayerStyle{Color: "#00bcd4", Weight: 2, Opacity: 0.7}, CreatedAt: now},
			{ID: "sat-coverage", Name: "Coverage Footprints", Type: "geojson", Visible: false, Style: LayerStyle{Color: "#ff9800", Weight: 1, Opacity: 0.5, FillColor: "#ff9800", FillOpacity: 0.1}, CreatedAt: now},
			{ID: "sat-ground", Name: "Ground Stations", Type: "marker", Visible: true, CreatedAt: now},
			{ID: "sat-positions", Name: "Live Satellite Positions", Type: "marker", Visible: true, CreatedAt: now},
		},
		Regions: []GISRegion{
			{ID: "leo-belt", Name: "LEO Belt (160-2000km)", Type: "orbit_zone", Center: [2]float64{0, 0}, Properties: map[string]interface{}{"altitude_min": 160, "altitude_max": 2000, "satellites": 7800, "orbit_type": "LEO", "period_min": 90, "color": "#e3f2fd"}},
			{ID: "meo-belt", Name: "MEO Belt (2000-35786km)", Type: "orbit_zone", Center: [2]float64{0, 60}, Properties: map[string]interface{}{"altitude_min": 2000, "altitude_max": 35786, "satellites": 145, "orbit_type": "MEO", "period_min": 720, "color": "#fff3e0"}},
			{ID: "geo-belt", Name: "GEO Belt (35786km)", Type: "orbit_zone", Center: [2]float64{0, -60}, Properties: map[string]interface{}{"altitude_min": 35786, "altitude_max": 35786, "satellites": 565, "orbit_type": "GEO", "period_min": 1440, "color": "#fce4ec"}},
			{ID: "polar-orbit", Name: "Polar/SSO", Type: "orbit_zone", Center: [2]float64{75, 0}, Properties: map[string]interface{}{"inclination": 97.8, "satellites": 1200, "orbit_type": "SSO", "usage": "Earth Observation", "color": "#e8eaf6"}},
		},
		Markers: []GISMarker{
			{ID: "sat-m1", Name: "ISS (International Space Station)", Lat: 28.5, Lng: -80.6, Category: "satellite", Color: "#00bcd4", Properties: map[string]interface{}{"orbit": "LEO", "altitude_km": 408, "speed_kph": 27600, "crew": 7, "inclination": 51.6}},
			{ID: "sat-m2", Name: "Hubble Space Telescope", Lat: 35.0, Lng: -118.0, Category: "satellite", Color: "#9c27b0", Properties: map[string]interface{}{"orbit": "LEO", "altitude_km": 547, "launched": 1990, "mission": "Astronomy"}},
			{ID: "sat-m3", Name: "Starlink Constellation Node", Lat: 47.6, Lng: -122.3, Category: "constellation", Color: "#2196f3", Properties: map[string]interface{}{"orbit": "LEO", "altitude_km": 550, "total_sats": 5500, "operator": "SpaceX"}},
			{ID: "sat-m4", Name: "GPS III SV06", Lat: 38.9, Lng: -77.0, Category: "navigation", Color: "#4caf50", Properties: map[string]interface{}{"orbit": "MEO", "altitude_km": 20200, "constellation": "GPS", "launched": 2023}},
			{ID: "sat-m5", Name: "INSAT-3DR", Lat: 23.5, Lng: 90.0, Category: "weather", Color: "#ff9800", Properties: map[string]interface{}{"orbit": "GEO", "altitude_km": 35786, "operator": "ISRO", "purpose": "Meteorology"}},
			{ID: "sat-m6", Name: "NASA Goddard Ground Station", Lat: 38.99, Lng: -76.85, Category: "ground_station", Color: "#f44336", Properties: map[string]interface{}{"type": "Primary Tracking", "antennas": 12, "operator": "NASA"}},
			{ID: "sat-m7", Name: "ESA Kourou Launch Site", Lat: 5.236, Lng: -52.769, Category: "launch_site", Color: "#e91e63", Properties: map[string]interface{}{"type": "Launch Complex", "launches_per_year": 12, "operator": "Arianespace"}},
			{ID: "sat-m8", Name: "Baikonur Cosmodrome", Lat: 45.965, Lng: 63.305, Category: "launch_site", Color: "#e91e63", Properties: map[string]interface{}{"type": "Launch Complex", "established": 1955, "operator": "Roscosmos"}},
			{ID: "sat-m9", Name: "JAXA Tanegashima Center", Lat: 30.4, Lng: 131.0, Category: "launch_site", Color: "#e91e63", Properties: map[string]interface{}{"type": "Launch Complex", "operator": "JAXA", "country": "Japan"}},
			{ID: "sat-m10", Name: "Svalbard Ground Station", Lat: 78.23, Lng: 15.39, Category: "ground_station", Color: "#f44336", Properties: map[string]interface{}{"type": "Polar Tracking", "operator": "KSAT", "antennas": 31}},
			{ID: "sat-m11", Name: "Sentinel-2A", Lat: -10.0, Lng: 30.0, Category: "earth_observation", Color: "#8bc34a", Properties: map[string]interface{}{"orbit": "SSO", "altitude_km": 786, "operator": "ESA", "revisit_days": 5}},
			{ID: "sat-m12", Name: "ISRO Sriharikota Launch", Lat: 13.72, Lng: 80.23, Category: "launch_site", Color: "#e91e63", Properties: map[string]interface{}{"type": "Launch Complex", "operator": "ISRO", "country": "India"}},
		},
		Datasets: []GISDataset{
			{
				ID: "orbit-catalog", Name: "Satellite Orbit Catalog", Description: "Active satellites by orbit type", Unit: "count",
				Columns: []DatasetColumn{
					{Key: "name", Label: "Orbit Zone", Type: "string"},
					{Key: "satellites", Label: "Active Satellites", Type: "number"},
					{Key: "altitude_min", Label: "Min Altitude (km)", Type: "number"},
					{Key: "altitude_max", Label: "Max Altitude (km)", Type: "number"},
				},
				Rows: []map[string]interface{}{
					{"name": "LEO (Low Earth)", "satellites": 7800, "altitude_min": 160, "altitude_max": 2000},
					{"name": "MEO (Medium Earth)", "satellites": 145, "altitude_min": 2000, "altitude_max": 35786},
					{"name": "GEO (Geostationary)", "satellites": 565, "altitude_min": 35786, "altitude_max": 35786},
					{"name": "SSO (Sun-Synchronous)", "satellites": 1200, "altitude_min": 400, "altitude_max": 900},
					{"name": "HEO (Highly Elliptical)", "satellites": 42, "altitude_min": 500, "altitude_max": 40000},
				},
				CreatedAt: now, UpdatedAt: now,
			},
			{
				ID: "ground-stations", Name: "Major Ground Stations", Description: "Satellite tracking and communication stations", Unit: "facility",
				Columns: []DatasetColumn{
					{Key: "name", Label: "Station", Type: "string"},
					{Key: "operator", Label: "Operator", Type: "string"},
					{Key: "antennas", Label: "Antennas", Type: "number"},
				},
				Rows: []map[string]interface{}{
					{"name": "NASA Goddard", "operator": "NASA", "antennas": 12},
					{"name": "Svalbard SvalSat", "operator": "KSAT", "antennas": 31},
					{"name": "ESA Darmstadt", "operator": "ESA", "antennas": 8},
					{"name": "ISRO Bangalore", "operator": "ISRO", "antennas": 6},
					{"name": "JAXA Tsukuba", "operator": "JAXA", "antennas": 5},
				},
				CreatedAt: now, UpdatedAt: now,
			},
		},
	}
	h.dashboards["satellite"] = d
}

func (h *GISSpecializedHandler) seedAirplane() {
	now := time.Now()
	d := &GISDashboardData{
		Type:        "airplane",
		Title:       "Aviation Tracking Dashboard",
		Description: "Flight routes, international airports, air traffic data and aviation corridors worldwide",
		MapCenter:   [2]float64{30, 50},
		DefaultZoom: 3,
		MaxBounds:   [2][2]float64{{-85, -180}, {85, 180}},
		Config: map[string]interface{}{
			"legendTitle": "Traffic Density",
			"colorField":  "annual_passengers",
			"iconTheme":   "airplane",
			"scope":       "international",
		},
		Layers: []GISLayer{
			{ID: "air-routes", Name: "Major Air Routes", Type: "geojson", Visible: true, Style: LayerStyle{Color: "#ff5722", Weight: 1, Opacity: 0.4}, CreatedAt: now},
			{ID: "air-corridors", Name: "IATA Corridors", Type: "geojson", Visible: false, Style: LayerStyle{Color: "#ffc107", Weight: 2, Opacity: 0.5, FillColor: "#ffc107", FillOpacity: 0.1}, CreatedAt: now},
			{ID: "airports", Name: "International Airports", Type: "marker", Visible: true, CreatedAt: now},
			{ID: "aircraft", Name: "Active Aircraft", Type: "marker", Visible: true, CreatedAt: now},
		},
		Regions: []GISRegion{
			{ID: "air-asia", Name: "Asia-Pacific", Type: "air_region", Center: [2]float64{25, 105}, Properties: map[string]interface{}{"annual_passengers": 3800000000, "airports": 2450, "airlines": 320, "market_share_pct": 37, "color": "#ffecb3"}},
			{ID: "air-europe", Name: "Europe", Type: "air_region", Center: [2]float64{50, 10}, Properties: map[string]interface{}{"annual_passengers": 2400000000, "airports": 1800, "airlines": 240, "market_share_pct": 26, "color": "#bbdefb"}},
			{ID: "air-namerica", Name: "North America", Type: "air_region", Center: [2]float64{42, -95}, Properties: map[string]interface{}{"annual_passengers": 1900000000, "airports": 1200, "airlines": 95, "market_share_pct": 22, "color": "#c8e6c9"}},
			{ID: "air-mideast", Name: "Middle East", Type: "air_region", Center: [2]float64{25, 48}, Properties: map[string]interface{}{"annual_passengers": 420000000, "airports": 180, "airlines": 48, "market_share_pct": 5, "color": "#f8bbd0"}},
			{ID: "air-africa", Name: "Africa", Type: "air_region", Center: [2]float64{0, 25}, Properties: map[string]interface{}{"annual_passengers": 230000000, "airports": 420, "airlines": 68, "market_share_pct": 3, "color": "#d1c4e9"}},
			{ID: "air-latam", Name: "Latin America", Type: "air_region", Center: [2]float64{-15, -60}, Properties: map[string]interface{}{"annual_passengers": 380000000, "airports": 650, "airlines": 85, "market_share_pct": 7, "color": "#ffe0b2"}},
		},
		Markers: []GISMarker{
			{ID: "ap-m1", Name: "Hartsfield-Jackson Atlanta (ATL)", Lat: 33.6407, Lng: -84.4277, Category: "airport", Color: "#f44336", Properties: map[string]interface{}{"iata": "ATL", "passengers": 93700000, "rank": 1, "country": "USA"}},
			{ID: "ap-m2", Name: "Dubai International (DXB)", Lat: 25.2532, Lng: 55.3657, Category: "airport", Color: "#ff9800", Properties: map[string]interface{}{"iata": "DXB", "passengers": 87000000, "rank": 2, "country": "UAE"}},
			{ID: "ap-m3", Name: "London Heathrow (LHR)", Lat: 51.4700, Lng: -0.4543, Category: "airport", Color: "#2196f3", Properties: map[string]interface{}{"iata": "LHR", "passengers": 79200000, "rank": 4, "country": "UK"}},
			{ID: "ap-m4", Name: "Tokyo Haneda (HND)", Lat: 35.5494, Lng: 139.7798, Category: "airport", Color: "#4caf50", Properties: map[string]interface{}{"iata": "HND", "passengers": 75500000, "rank": 5, "country": "Japan"}},
			{ID: "ap-m5", Name: "Istanbul Airport (IST)", Lat: 41.2752, Lng: 28.7519, Category: "airport", Color: "#9c27b0", Properties: map[string]interface{}{"iata": "IST", "passengers": 64300000, "rank": 7, "country": "Turkey"}},
			{ID: "ap-m6", Name: "Singapore Changi (SIN)", Lat: 1.3644, Lng: 103.9915, Category: "airport", Color: "#00bcd4", Properties: map[string]interface{}{"iata": "SIN", "passengers": 62900000, "rank": 8, "country": "Singapore"}},
			{ID: "ap-m7", Name: "Hazrat Shahjalal Intl (DAC)", Lat: 23.8513, Lng: 90.4023, Category: "airport", Color: "#e91e63", Properties: map[string]interface{}{"iata": "DAC", "passengers": 8200000, "rank": 142, "country": "Bangladesh"}},
			{ID: "ap-m8", Name: "Los Angeles (LAX)", Lat: 33.9425, Lng: -118.4081, Category: "airport", Color: "#ff5722", Properties: map[string]interface{}{"iata": "LAX", "passengers": 88100000, "rank": 3, "country": "USA"}},
			{ID: "ap-m9", Name: "Beijing Capital (PEK)", Lat: 40.0799, Lng: 116.6031, Category: "airport", Color: "#795548", Properties: map[string]interface{}{"iata": "PEK", "passengers": 59700000, "rank": 10, "country": "China"}},
			{ID: "ap-m10", Name: "Sydney Kingsford Smith (SYD)", Lat: -33.9461, Lng: 151.1772, Category: "airport", Color: "#607d8b", Properties: map[string]interface{}{"iata": "SYD", "passengers": 42600000, "rank": 25, "country": "Australia"}},
			{ID: "ap-m11", Name: "Indira Gandhi Intl (DEL)", Lat: 28.5562, Lng: 77.1000, Category: "airport", Color: "#ff9800", Properties: map[string]interface{}{"iata": "DEL", "passengers": 72200000, "rank": 6, "country": "India"}},
			{ID: "ap-m12", Name: "Sao Paulo Guarulhos (GRU)", Lat: -23.4356, Lng: -46.4731, Category: "airport", Color: "#8bc34a", Properties: map[string]interface{}{"iata": "GRU", "passengers": 35700000, "rank": 35, "country": "Brazil"}},
			{ID: "ac-1", Name: "BA-217 London→Delhi", Lat: 38.5, Lng: 42.0, Category: "aircraft", Color: "#2196f3", Properties: map[string]interface{}{"flight": "BA-217", "airline": "British Airways", "aircraft": "Boeing 787-9", "altitude_ft": 38000, "speed_kts": 480, "route": "LHR→DEL"}},
			{ID: "ac-2", Name: "EK-584 Dubai→Dhaka", Lat: 22.8, Lng: 78.5, Category: "aircraft", Color: "#ff9800", Properties: map[string]interface{}{"flight": "EK-584", "airline": "Emirates", "aircraft": "Boeing 777-300ER", "altitude_ft": 36000, "speed_kts": 510, "route": "DXB→DAC"}},
			{ID: "ac-3", Name: "SQ-322 Singapore→London", Lat: 25.0, Lng: 68.0, Category: "aircraft", Color: "#00bcd4", Properties: map[string]interface{}{"flight": "SQ-322", "airline": "Singapore Airlines", "aircraft": "A350-900ULR", "altitude_ft": 41000, "speed_kts": 490, "route": "SIN→LHR"}},
		},
		Datasets: []GISDataset{
			{
				ID: "airport-traffic", Name: "Top Airport Traffic", Description: "Busiest airports by passenger count", Unit: "passengers",
				Columns: []DatasetColumn{
					{Key: "name", Label: "Airport", Type: "string"},
					{Key: "iata", Label: "IATA", Type: "string"},
					{Key: "passengers", Label: "Annual Passengers", Type: "number"},
					{Key: "country", Label: "Country", Type: "string"},
					{Key: "rank", Label: "Rank", Type: "number"},
				},
				Rows: []map[string]interface{}{
					{"name": "Hartsfield-Jackson Atlanta", "iata": "ATL", "passengers": 93700000, "country": "USA", "rank": 1},
					{"name": "Dubai International", "iata": "DXB", "passengers": 87000000, "country": "UAE", "rank": 2},
					{"name": "Los Angeles", "iata": "LAX", "passengers": 88100000, "country": "USA", "rank": 3},
					{"name": "London Heathrow", "iata": "LHR", "passengers": 79200000, "country": "UK", "rank": 4},
					{"name": "Tokyo Haneda", "iata": "HND", "passengers": 75500000, "country": "Japan", "rank": 5},
					{"name": "Indira Gandhi Delhi", "iata": "DEL", "passengers": 72200000, "country": "India", "rank": 6},
					{"name": "Istanbul Airport", "iata": "IST", "passengers": 64300000, "country": "Turkey", "rank": 7},
					{"name": "Singapore Changi", "iata": "SIN", "passengers": 62900000, "country": "Singapore", "rank": 8},
					{"name": "Beijing Capital", "iata": "PEK", "passengers": 59700000, "country": "China", "rank": 10},
					{"name": "Hazrat Shahjalal Dhaka", "iata": "DAC", "passengers": 8200000, "country": "Bangladesh", "rank": 142},
				},
				CreatedAt: now, UpdatedAt: now,
			},
			{
				ID: "air-regions", Name: "Aviation Market by Region", Description: "Regional aviation statistics", Unit: "mixed",
				Columns: []DatasetColumn{
					{Key: "name", Label: "Region", Type: "string"},
					{Key: "annual_passengers", Label: "Passengers/Year", Type: "number"},
					{Key: "airports", Label: "Airports", Type: "number"},
					{Key: "airlines", Label: "Airlines", Type: "number"},
					{Key: "market_share_pct", Label: "Market Share %", Type: "number"},
				},
				Rows: []map[string]interface{}{
					{"name": "Asia-Pacific", "annual_passengers": 3800000000, "airports": 2450, "airlines": 320, "market_share_pct": 37},
					{"name": "Europe", "annual_passengers": 2400000000, "airports": 1800, "airlines": 240, "market_share_pct": 26},
					{"name": "North America", "annual_passengers": 1900000000, "airports": 1200, "airlines": 95, "market_share_pct": 22},
					{"name": "Middle East", "annual_passengers": 420000000, "airports": 180, "airlines": 48, "market_share_pct": 5},
					{"name": "Latin America", "annual_passengers": 380000000, "airports": 650, "airlines": 85, "market_share_pct": 7},
					{"name": "Africa", "annual_passengers": 230000000, "airports": 420, "airlines": 68, "market_share_pct": 3},
				},
				CreatedAt: now, UpdatedAt: now,
			},
		},
	}
	h.dashboards["airplane"] = d
}

func (h *GISSpecializedHandler) seedShip() {
	now := time.Now()
	d := &GISDashboardData{
		Type:        "ship",
		Title:       "Maritime Shipping Dashboard",
		Description: "Global shipping lanes, major ports, vessel tracking and maritime trade data",
		MapCenter:   [2]float64{15, 65},
		DefaultZoom: 3,
		MaxBounds:   [2][2]float64{{-85, -180}, {85, 180}},
		Config: map[string]interface{}{
			"legendTitle": "Port Throughput (TEU)",
			"colorField":  "throughput_teu",
			"iconTheme":   "ship",
			"scope":       "international",
		},
		Layers: []GISLayer{
			{ID: "shipping-lanes", Name: "Major Shipping Lanes", Type: "geojson", Visible: true, Style: LayerStyle{Color: "#0277bd", Weight: 2, Opacity: 0.6}, CreatedAt: now},
			{ID: "trade-zones", Name: "Maritime Trade Zones", Type: "geojson", Visible: false, Style: LayerStyle{Color: "#00897b", Weight: 1, Opacity: 0.4, FillColor: "#00897b", FillOpacity: 0.1}, CreatedAt: now},
			{ID: "ports", Name: "Major Ports", Type: "marker", Visible: true, CreatedAt: now},
			{ID: "vessels", Name: "Active Vessels", Type: "marker", Visible: true, CreatedAt: now},
		},
		Regions: []GISRegion{
			{ID: "sea-indian", Name: "Indian Ocean", Type: "ocean_zone", Center: [2]float64{-5, 70}, Properties: map[string]interface{}{"trade_volume_mt": 12500000000, "shipping_routes": 42, "piracy_incidents": 18, "avg_transit_days": 15, "color": "#b3e5fc"}},
			{ID: "sea-pacific", Name: "Pacific Ocean", Type: "ocean_zone", Center: [2]float64{5, 170}, Properties: map[string]interface{}{"trade_volume_mt": 28000000000, "shipping_routes": 68, "piracy_incidents": 5, "avg_transit_days": 22, "color": "#c8e6c9"}},
			{ID: "sea-atlantic", Name: "Atlantic Ocean", Type: "ocean_zone", Center: [2]float64{25, -35}, Properties: map[string]interface{}{"trade_volume_mt": 18500000000, "shipping_routes": 55, "piracy_incidents": 12, "avg_transit_days": 18, "color": "#dcedc8"}},
			{ID: "sea-mediterranean", Name: "Mediterranean Sea", Type: "ocean_zone", Center: [2]float64{36, 18}, Properties: map[string]interface{}{"trade_volume_mt": 6200000000, "shipping_routes": 38, "piracy_incidents": 2, "avg_transit_days": 8, "color": "#fff9c4"}},
			{ID: "sea-south-china", Name: "South China Sea", Type: "ocean_zone", Center: [2]float64{12, 115}, Properties: map[string]interface{}{"trade_volume_mt": 9800000000, "shipping_routes": 52, "piracy_incidents": 32, "avg_transit_days": 6, "color": "#ffe0b2"}},
			{ID: "sea-bengal", Name: "Bay of Bengal", Type: "ocean_zone", Center: [2]float64{13, 88}, Properties: map[string]interface{}{"trade_volume_mt": 3200000000, "shipping_routes": 18, "piracy_incidents": 7, "avg_transit_days": 5, "color": "#e1bee7"}},
		},
		Markers: []GISMarker{
			{ID: "sh-m1", Name: "Port of Shanghai", Lat: 31.3500, Lng: 121.5000, Category: "port", Color: "#f44336", Properties: map[string]interface{}{"throughput_teu": 47000000, "rank": 1, "country": "China", "type": "Container Terminal"}},
			{ID: "sh-m2", Name: "Port of Singapore", Lat: 1.2646, Lng: 103.8200, Category: "port", Color: "#e91e63", Properties: map[string]interface{}{"throughput_teu": 37200000, "rank": 2, "country": "Singapore", "type": "Transshipment Hub"}},
			{ID: "sh-m3", Name: "Port of Rotterdam", Lat: 51.9225, Lng: 4.4792, Category: "port", Color: "#2196f3", Properties: map[string]interface{}{"throughput_teu": 14400000, "rank": 10, "country": "Netherlands", "type": "Gateway Port"}},
			{ID: "sh-m4", Name: "Port of Chattogram", Lat: 22.3300, Lng: 91.8000, Category: "port", Color: "#ff9800", Properties: map[string]interface{}{"throughput_teu": 3200000, "rank": 58, "country": "Bangladesh", "type": "Main National Port"}},
			{ID: "sh-m5", Name: "Jebel Ali Dubai", Lat: 25.0174, Lng: 55.0644, Category: "port", Color: "#9c27b0", Properties: map[string]interface{}{"throughput_teu": 13500000, "rank": 11, "country": "UAE", "type": "Transshipment Hub"}},
			{ID: "sh-m6", Name: "Port of Busan", Lat: 35.0957, Lng: 129.0360, Category: "port", Color: "#00bcd4", Properties: map[string]interface{}{"throughput_teu": 22000000, "rank": 5, "country": "South Korea", "type": "Container Terminal"}},
			{ID: "sh-m7", Name: "Suez Canal", Lat: 30.5852, Lng: 32.2654, Category: "canal", Color: "#ff5722", Properties: map[string]interface{}{"type": "Strategic Chokepoint", "vessels_daily": 52, "annual_tonnage_mt": 1200000000}},
			{ID: "sh-m8", Name: "Strait of Malacca", Lat: 2.5, Lng: 101.5, Category: "strait", Color: "#ff5722", Properties: map[string]interface{}{"type": "Strategic Chokepoint", "vessels_daily": 94, "world_trade_pct": 25}},
			{ID: "sh-m9", Name: "Port of Mongla", Lat: 22.4700, Lng: 89.6000, Category: "port", Color: "#4caf50", Properties: map[string]interface{}{"throughput_teu": 450000, "country": "Bangladesh", "type": "Secondary Port"}},
			{ID: "sh-m10", Name: "Panama Canal", Lat: 9.08, Lng: -79.68, Category: "canal", Color: "#ff5722", Properties: map[string]interface{}{"type": "Strategic Chokepoint", "vessels_daily": 36, "annual_tonnage_mt": 500000000}},
			{ID: "vs-1", Name: "MV Ever Given (Container)", Lat: 8.5, Lng: 78.0, Category: "vessel", Color: "#0277bd", Properties: map[string]interface{}{"imo": "9811000", "type": "Container Ship", "teu_capacity": 20124, "speed_kts": 14.5, "route": "Shanghai→Rotterdam"}},
			{ID: "vs-2", Name: "MV Pacific Star (Tanker)", Lat: 22.0, Lng: 66.0, Category: "vessel", Color: "#558b2f", Properties: map[string]interface{}{"type": "Oil Tanker", "dwt": 320000, "speed_kts": 12.0, "route": "Jeddah→Mumbai", "cargo": "Crude Oil"}},
			{ID: "vs-3", Name: "MV Banglar Shourabh", Lat: 21.5, Lng: 90.5, Category: "vessel", Color: "#ff9800", Properties: map[string]interface{}{"type": "Bulk Carrier", "dwt": 45000, "speed_kts": 11.0, "flag": "Bangladesh", "route": "Chattogram→Singapore"}},
		},
		Datasets: []GISDataset{
			{
				ID: "port-throughput", Name: "Top Port Throughput", Description: "Container throughput of major world ports", Unit: "TEU",
				Columns: []DatasetColumn{
					{Key: "name", Label: "Port", Type: "string"},
					{Key: "throughput_teu", Label: "Throughput (TEU)", Type: "number"},
					{Key: "country", Label: "Country", Type: "string"},
					{Key: "rank", Label: "Rank", Type: "number"},
				},
				Rows: []map[string]interface{}{
					{"name": "Shanghai", "throughput_teu": 47000000, "country": "China", "rank": 1},
					{"name": "Singapore", "throughput_teu": 37200000, "country": "Singapore", "rank": 2},
					{"name": "Ningbo-Zhoushan", "throughput_teu": 33500000, "country": "China", "rank": 3},
					{"name": "Shenzhen", "throughput_teu": 28800000, "country": "China", "rank": 4},
					{"name": "Busan", "throughput_teu": 22000000, "country": "South Korea", "rank": 5},
					{"name": "Rotterdam", "throughput_teu": 14400000, "country": "Netherlands", "rank": 10},
					{"name": "Jebel Ali Dubai", "throughput_teu": 13500000, "country": "UAE", "rank": 11},
					{"name": "Chattogram", "throughput_teu": 3200000, "country": "Bangladesh", "rank": 58},
					{"name": "Mongla", "throughput_teu": 450000, "country": "Bangladesh", "rank": 120},
				},
				CreatedAt: now, UpdatedAt: now,
			},
			{
				ID: "ocean-trade", Name: "Ocean Trade Volume", Description: "Trade volume and routes by ocean", Unit: "MT",
				Columns: []DatasetColumn{
					{Key: "name", Label: "Ocean/Sea", Type: "string"},
					{Key: "trade_volume_mt", Label: "Trade Volume (MT)", Type: "number"},
					{Key: "shipping_routes", Label: "Routes", Type: "number"},
					{Key: "piracy_incidents", Label: "Piracy Incidents", Type: "number"},
				},
				Rows: []map[string]interface{}{
					{"name": "Pacific Ocean", "trade_volume_mt": 28000000000, "shipping_routes": 68, "piracy_incidents": 5},
					{"name": "Atlantic Ocean", "trade_volume_mt": 18500000000, "shipping_routes": 55, "piracy_incidents": 12},
					{"name": "Indian Ocean", "trade_volume_mt": 12500000000, "shipping_routes": 42, "piracy_incidents": 18},
					{"name": "South China Sea", "trade_volume_mt": 9800000000, "shipping_routes": 52, "piracy_incidents": 32},
					{"name": "Mediterranean", "trade_volume_mt": 6200000000, "shipping_routes": 38, "piracy_incidents": 2},
					{"name": "Bay of Bengal", "trade_volume_mt": 3200000000, "shipping_routes": 18, "piracy_incidents": 7},
				},
				CreatedAt: now, UpdatedAt: now,
			},
		},
	}
	h.dashboards["ship"] = d
}

func (h *GISSpecializedHandler) seedTrainPlaceholder() {
	h.dashboards["train"] = &GISDashboardData{
		Type:        "train",
		Title:       "Indian Railway Network Dashboard",
		Description: "Train routes, stations, and schedules — live data from PostgreSQL train database. Full API at /api/v1/gis/trains",
		MapCenter:   [2]float64{22.5, 82.0},
		DefaultZoom: 5,
		MaxBounds:   [2][2]float64{{6, 65}, {38, 98}},
		Config: map[string]interface{}{
			"data_source": "postgresql",
			"database":    "train",
			"country":     "IN",
			"api_base":    "/api/v1/gis/trains",
		},
	}
	h.dashboards["bd-train"] = &GISDashboardData{
		Type:        "bd-train",
		Title:       "Bangladesh Railway Network Dashboard",
		Description: "Train routes, stations, trips and fares — live data from PostgreSQL bd-train database. Full API at /api/v1/gis/bd-trains",
		MapCenter:   [2]float64{23.6850, 90.3563},
		DefaultZoom: 7,
		MaxBounds:   [2][2]float64{{20.5, 87.5}, {26.7, 92.8}},
		Config: map[string]interface{}{
			"data_source": "postgresql",
			"database":    "bd-train",
			"country":     "BD",
			"api_base":    "/api/v1/gis/bd-trains",
		},
	}
}
