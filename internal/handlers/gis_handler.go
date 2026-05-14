package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Deprecated: GISHandler is an in-memory store protected by sync.RWMutex and
// does NOT follow the platform's control-plane architecture
// (ResourceStore -> Informer -> WorkQueue -> Controller -> Reconciler with
// etcd as authoritative state). It is retained only because
// APIBuilderHandler currently depends on it for layer/region/marker/dataset
// persistence during the Dashboard<->GIS conversion flow.
//
// MIGRATION TARGET:
//
//	GISHandler (gin)
//	    -> GISService
//	    -> ResourceStore[GISLayer | GISRegion | GISMarker | GISDataset]
//	    -> etcd
//	    -> Reconciler (projects to Postgres / Elasticsearch read-model)
//
// New GIS APIs must be authored via the API Builder, which already writes to
// etcd. Do NOT add new endpoints to this handler. See
// docs/architecture/handler-migration.md for the full migration plan.
//
// GISHandler manages GIS dashboard data (layers, regions, markers, datasets).
type GISHandler struct {
	mu       sync.RWMutex
	layers   []GISLayer
	regions  []GISRegion
	markers  []GISMarker
	datasets []GISDataset
}

// GISLayer represents a map layer
type GISLayer struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Type      string      `json:"type"` // geojson, tile, marker, heatmap
	Visible   bool        `json:"visible"`
	Style     LayerStyle  `json:"style,omitempty"`
	URL       string      `json:"url,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	CreatedAt time.Time   `json:"createdAt"`
}

// LayerStyle defines visual properties for a layer
type LayerStyle struct {
	Color       string  `json:"color,omitempty"`
	Weight      float64 `json:"weight,omitempty"`
	Opacity     float64 `json:"opacity,omitempty"`
	FillColor   string  `json:"fillColor,omitempty"`
	FillOpacity float64 `json:"fillOpacity,omitempty"`
}

// GISRegion represents a geographic region (division/district/upazila)
type GISRegion struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // division, district, upazila
	ParentID   string                 `json:"parentId,omitempty"`
	Center     [2]float64             `json:"center"`           // [lat, lng]
	Bounds     [4]float64             `json:"bounds,omitempty"` // [minLat, minLng, maxLat, maxLng]
	Properties map[string]interface{} `json:"properties,omitempty"`
	GeoJSON    interface{}            `json:"geojson,omitempty"`
}

// GISMarker represents a point marker on the map
type GISMarker struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Lat        float64                `json:"lat"`
	Lng        float64                `json:"lng"`
	Category   string                 `json:"category"`
	Icon       string                 `json:"icon,omitempty"`
	Color      string                 `json:"color,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// GISDataset represents a named data collection for choropleth/analysis
type GISDataset struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Unit        string                   `json:"unit,omitempty"`
	Columns     []DatasetColumn          `json:"columns"`
	Rows        []map[string]interface{} `json:"rows"`
	CreatedAt   time.Time                `json:"createdAt"`
	UpdatedAt   time.Time                `json:"updatedAt"`
}

// DatasetColumn defines a column in a dataset
type DatasetColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"` // string, number, date
}

// GISSummary is returned for the overview endpoint
type GISSummary struct {
	TotalLayers   int            `json:"totalLayers"`
	TotalRegions  int            `json:"totalRegions"`
	TotalMarkers  int            `json:"totalMarkers"`
	TotalDatasets int            `json:"totalDatasets"`
	RegionsByType map[string]int `json:"regionsByType"`
	MapCenter     [2]float64     `json:"mapCenter"`
	DefaultZoom   int            `json:"defaultZoom"`
}

func NewGISHandler() *GISHandler {
	h := &GISHandler{
		layers:   make([]GISLayer, 0),
		regions:  make([]GISRegion, 0),
		markers:  make([]GISMarker, 0),
		datasets: make([]GISDataset, 0),
	}
	h.seedDefaults()
	return h
}

// seedDefaults populates sample data so the dashboard works out of the box
func (h *GISHandler) seedDefaults() {
	// Default layers
	h.layers = []GISLayer{
		{ID: "divisions", Name: "Divisions", Type: "geojson", Visible: true, Style: LayerStyle{Color: "#3388ff", Weight: 2, Opacity: 0.8, FillOpacity: 0.15}, CreatedAt: time.Now()},
		{ID: "districts", Name: "Districts", Type: "geojson", Visible: false, Style: LayerStyle{Color: "#666", Weight: 1, Opacity: 0.6, FillOpacity: 0.1}, CreatedAt: time.Now()},
		{ID: "markers", Name: "Points of Interest", Type: "marker", Visible: true, CreatedAt: time.Now()},
		{ID: "heatmap", Name: "Population Density", Type: "heatmap", Visible: false, CreatedAt: time.Now()},
	}

	// 8 divisions of Bangladesh as example regions
	h.regions = []GISRegion{
		{ID: "dhaka", Name: "Dhaka", Type: "division", Center: [2]float64{23.8103, 90.4125}, Properties: map[string]interface{}{"population": 44218000, "area_km2": 20593, "districts": 13, "color": "#e8f5e9"}},
		{ID: "chattogram", Name: "Chattogram", Type: "division", Center: [2]float64{22.3569, 91.7832}, Properties: map[string]interface{}{"population": 33202000, "area_km2": 33771, "districts": 11, "color": "#e3f2fd"}},
		{ID: "rajshahi", Name: "Rajshahi", Type: "division", Center: [2]float64{24.3745, 88.6042}, Properties: map[string]interface{}{"population": 20353000, "area_km2": 18197, "districts": 8, "color": "#fff3e0"}},
		{ID: "khulna", Name: "Khulna", Type: "division", Center: [2]float64{22.8456, 89.5403}, Properties: map[string]interface{}{"population": 17416000, "area_km2": 22285, "districts": 10, "color": "#f3e5f5"}},
		{ID: "barishal", Name: "Barishal", Type: "division", Center: [2]float64{22.7010, 90.3535}, Properties: map[string]interface{}{"population": 9328000, "area_km2": 13297, "districts": 6, "color": "#e8eaf6"}},
		{ID: "sylhet", Name: "Sylhet", Type: "division", Center: [2]float64{24.8949, 91.8687}, Properties: map[string]interface{}{"population": 11310000, "area_km2": 12596, "districts": 4, "color": "#e0f2f1"}},
		{ID: "rangpur", Name: "Rangpur", Type: "division", Center: [2]float64{25.7439, 89.2752}, Properties: map[string]interface{}{"population": 17610000, "area_km2": 16185, "districts": 8, "color": "#fce4ec"}},
		{ID: "mymensingh", Name: "Mymensingh", Type: "division", Center: [2]float64{24.7471, 90.4203}, Properties: map[string]interface{}{"population": 12334000, "area_km2": 10584, "districts": 4, "color": "#fff8e1"}},
	}

	// Sample district-level data under Dhaka
	dhakaDistricts := []GISRegion{
		{ID: "dhaka-city", Name: "Dhaka", Type: "district", ParentID: "dhaka", Center: [2]float64{23.8103, 90.4125}, Properties: map[string]interface{}{"population": 18900000, "area_km2": 1463}},
		{ID: "faridpur", Name: "Faridpur", Type: "district", ParentID: "dhaka", Center: [2]float64{23.6070, 89.8420}, Properties: map[string]interface{}{"population": 1910000, "area_km2": 2073}},
		{ID: "gazipur", Name: "Gazipur", Type: "district", ParentID: "dhaka", Center: [2]float64{24.0023, 90.4253}, Properties: map[string]interface{}{"population": 4270000, "area_km2": 1741}},
		{ID: "gopalganj", Name: "Gopalganj", Type: "district", ParentID: "dhaka", Center: [2]float64{23.0488, 89.8879}, Properties: map[string]interface{}{"population": 1270000, "area_km2": 1490}},
		{ID: "kishoreganj", Name: "Kishoreganj", Type: "district", ParentID: "dhaka", Center: [2]float64{24.4260, 90.7847}, Properties: map[string]interface{}{"population": 2910000, "area_km2": 2689}},
		{ID: "madaripur", Name: "Madaripur", Type: "district", ParentID: "dhaka", Center: [2]float64{23.1641, 90.1978}, Properties: map[string]interface{}{"population": 1160000, "area_km2": 1125}},
		{ID: "manikganj", Name: "Manikganj", Type: "district", ParentID: "dhaka", Center: [2]float64{23.8617, 89.9836}, Properties: map[string]interface{}{"population": 1460000, "area_km2": 1379}},
		{ID: "munshiganj", Name: "Munshiganj", Type: "district", ParentID: "dhaka", Center: [2]float64{23.5422, 90.5305}, Properties: map[string]interface{}{"population": 1590000, "area_km2": 954}},
		{ID: "narayanganj", Name: "Narayanganj", Type: "district", ParentID: "dhaka", Center: [2]float64{23.6238, 90.5000}, Properties: map[string]interface{}{"population": 3140000, "area_km2": 684}},
		{ID: "narsingdi", Name: "Narsingdi", Type: "district", ParentID: "dhaka", Center: [2]float64{23.9322, 90.7151}, Properties: map[string]interface{}{"population": 2220000, "area_km2": 1141}},
		{ID: "rajbari", Name: "Rajbari", Type: "district", ParentID: "dhaka", Center: [2]float64{23.7574, 89.6445}, Properties: map[string]interface{}{"population": 1070000, "area_km2": 1092}},
		{ID: "shariatpur", Name: "Shariatpur", Type: "district", ParentID: "dhaka", Center: [2]float64{23.2423, 90.4348}, Properties: map[string]interface{}{"population": 1170000, "area_km2": 1174}},
		{ID: "tangail", Name: "Tangail", Type: "district", ParentID: "dhaka", Center: [2]float64{24.2513, 89.9164}, Properties: map[string]interface{}{"population": 3740000, "area_km2": 3414}},
	}
	h.regions = append(h.regions, dhakaDistricts...)

	// Sample markers
	h.markers = []GISMarker{
		{ID: "m1", Name: "Dhaka Central", Lat: 23.8103, Lng: 90.4125, Category: "capital", Icon: "star", Color: "#e74c3c"},
		{ID: "m2", Name: "Chattogram Port", Lat: 22.3300, Lng: 91.8000, Category: "port", Icon: "anchor", Color: "#3498db"},
		{ID: "m3", Name: "Shah Jalal Airport", Lat: 23.8513, Lng: 90.4023, Category: "airport", Icon: "plane", Color: "#2ecc71"},
		{ID: "m4", Name: "Padma Bridge", Lat: 23.4460, Lng: 90.2594, Category: "infrastructure", Icon: "road", Color: "#9b59b6"},
		{ID: "m5", Name: "Cox's Bazar", Lat: 21.4272, Lng: 92.0058, Category: "tourism", Icon: "umbrella-beach", Color: "#f39c12"},
		{ID: "m6", Name: "Sundarbans", Lat: 21.9497, Lng: 89.1833, Category: "nature", Icon: "tree", Color: "#27ae60"},
	}

	// One sample dataset: Division Population
	h.datasets = []GISDataset{
		{
			ID:          "division-population",
			Name:        "Division Population",
			Description: "Population by division (Census 2022 estimate)",
			Unit:        "people",
			Columns: []DatasetColumn{
				{Key: "name", Label: "Division", Type: "string"},
				{Key: "population", Label: "Population", Type: "number"},
				{Key: "area", Label: "Area (km²)", Type: "number"},
				{Key: "density", Label: "Density (/km²)", Type: "number"},
				{Key: "districts", Label: "Districts", Type: "number"},
			},
			Rows: []map[string]interface{}{
				{"name": "Dhaka", "population": 44218000, "area": 20593, "density": 2147, "districts": 13},
				{"name": "Chattogram", "population": 33202000, "area": 33771, "density": 983, "districts": 11},
				{"name": "Rajshahi", "population": 20353000, "area": 18197, "density": 1118, "districts": 8},
				{"name": "Khulna", "population": 17416000, "area": 22285, "density": 781, "districts": 10},
				{"name": "Rangpur", "population": 17610000, "area": 16185, "density": 1088, "districts": 8},
				{"name": "Mymensingh", "population": 12334000, "area": 10584, "density": 1165, "districts": 4},
				{"name": "Sylhet", "population": 11310000, "area": 12596, "density": 898, "districts": 4},
				{"name": "Barishal", "population": 9328000, "area": 13297, "density": 701, "districts": 6},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "district-population",
			Name:        "District Population",
			Description: "Population by district (Dhaka division)",
			Unit:        "people",
			Columns: []DatasetColumn{
				{Key: "name", Label: "Title", Type: "string"},
				{Key: "population", Label: "Value", Type: "number"},
			},
			Rows: []map[string]interface{}{
				{"name": "Dhaka", "population": 18900000},
				{"name": "Faridpur", "population": 1910000},
				{"name": "Gazipur", "population": 4270000},
				{"name": "Gopalganj", "population": 1270000},
				{"name": "Kishoreganj", "population": 2910000},
				{"name": "Madaripur", "population": 1160000},
				{"name": "Manikganj", "population": 1460000},
				{"name": "Munshiganj", "population": 1590000},
				{"name": "Narayanganj", "population": 3140000},
				{"name": "Narsingdi", "population": 2220000},
				{"name": "Rajbari", "population": 1070000},
				{"name": "Shariatpur", "population": 1170000},
				{"name": "Tangail", "population": 3740000},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// GetSummary returns overview stats for the GIS dashboard
func (h *GISHandler) GetSummary(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	regionsByType := make(map[string]int)
	for _, r := range h.regions {
		regionsByType[r.Type]++
	}

	c.JSON(http.StatusOK, GISSummary{
		TotalLayers:   len(h.layers),
		TotalRegions:  len(h.regions),
		TotalMarkers:  len(h.markers),
		TotalDatasets: len(h.datasets),
		RegionsByType: regionsByType,
		MapCenter:     [2]float64{23.6850, 90.3563}, // Bangladesh center
		DefaultZoom:   7,
	})
}

// --- Layers ---

func (h *GISHandler) ListLayers(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	c.JSON(http.StatusOK, h.layers)
}

func (h *GISHandler) CreateLayer(c *gin.Context) {
	var layer GISLayer
	if err := c.ShouldBindJSON(&layer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	layer.CreatedAt = time.Now()
	h.mu.Lock()
	h.layers = append(h.layers, layer)
	h.mu.Unlock()
	c.JSON(http.StatusCreated, layer)
}

func (h *GISHandler) UpdateLayer(c *gin.Context) {
	id := c.Param("id")
	var update GISLayer
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, l := range h.layers {
		if l.ID == id {
			update.ID = id
			update.CreatedAt = l.CreatedAt
			h.layers[i] = update
			c.JSON(http.StatusOK, update)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "layer not found"})
}

func (h *GISHandler) DeleteLayer(c *gin.Context) {
	id := c.Param("id")
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, l := range h.layers {
		if l.ID == id {
			h.layers = append(h.layers[:i], h.layers[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "layer deleted"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "layer not found"})
}

// --- Regions ---

func (h *GISHandler) ListRegions(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	regionType := c.Query("type") // division, district, upazila
	parentID := c.Query("parent")

	result := make([]GISRegion, 0)
	for _, r := range h.regions {
		if regionType != "" && r.Type != regionType {
			continue
		}
		if parentID != "" && r.ParentID != parentID {
			continue
		}
		result = append(result, r)
	}
	c.JSON(http.StatusOK, result)
}

func (h *GISHandler) GetRegion(c *gin.Context) {
	id := c.Param("id")
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, r := range h.regions {
		if r.ID == id {
			c.JSON(http.StatusOK, r)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
}

func (h *GISHandler) CreateRegion(c *gin.Context) {
	var region GISRegion
	if err := c.ShouldBindJSON(&region); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.mu.Lock()
	h.regions = append(h.regions, region)
	h.mu.Unlock()
	c.JSON(http.StatusCreated, region)
}

func (h *GISHandler) UpdateRegion(c *gin.Context) {
	id := c.Param("id")
	var update GISRegion
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, r := range h.regions {
		if r.ID == id {
			update.ID = id
			h.regions[i] = update
			c.JSON(http.StatusOK, update)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
}

func (h *GISHandler) DeleteRegion(c *gin.Context) {
	id := c.Param("id")
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, r := range h.regions {
		if r.ID == id {
			h.regions = append(h.regions[:i], h.regions[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "region deleted"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
}

// --- Markers ---

func (h *GISHandler) ListMarkers(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	category := c.Query("category")
	result := make([]GISMarker, 0)
	for _, m := range h.markers {
		if category != "" && m.Category != category {
			continue
		}
		result = append(result, m)
	}
	c.JSON(http.StatusOK, result)
}

func (h *GISHandler) CreateMarker(c *gin.Context) {
	var marker GISMarker
	if err := c.ShouldBindJSON(&marker); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.mu.Lock()
	h.markers = append(h.markers, marker)
	h.mu.Unlock()
	c.JSON(http.StatusCreated, marker)
}

func (h *GISHandler) DeleteMarker(c *gin.Context) {
	id := c.Param("id")
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, m := range h.markers {
		if m.ID == id {
			h.markers = append(h.markers[:i], h.markers[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "marker deleted"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "marker not found"})
}

// --- Datasets ---

func (h *GISHandler) ListDatasets(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	// Return metadata only (no rows) for the list view
	meta := make([]map[string]interface{}, len(h.datasets))
	for i, d := range h.datasets {
		meta[i] = map[string]interface{}{
			"id":          d.ID,
			"name":        d.Name,
			"description": d.Description,
			"unit":        d.Unit,
			"columns":     d.Columns,
			"rowCount":    len(d.Rows),
			"createdAt":   d.CreatedAt,
			"updatedAt":   d.UpdatedAt,
		}
	}
	c.JSON(http.StatusOK, meta)
}

func (h *GISHandler) GetDataset(c *gin.Context) {
	id := c.Param("id")
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, d := range h.datasets {
		if d.ID == id {
			c.JSON(http.StatusOK, d)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "dataset not found"})
}

func (h *GISHandler) CreateDataset(c *gin.Context) {
	var ds GISDataset
	if err := c.ShouldBindJSON(&ds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ds.CreatedAt = time.Now()
	ds.UpdatedAt = time.Now()
	h.mu.Lock()
	h.datasets = append(h.datasets, ds)
	h.mu.Unlock()
	c.JSON(http.StatusCreated, ds)
}

func (h *GISHandler) UpdateDataset(c *gin.Context) {
	id := c.Param("id")
	var update GISDataset
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, d := range h.datasets {
		if d.ID == id {
			update.ID = id
			update.CreatedAt = d.CreatedAt
			update.UpdatedAt = time.Now()
			h.datasets[i] = update
			c.JSON(http.StatusOK, update)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "dataset not found"})
}

func (h *GISHandler) DeleteDataset(c *gin.Context) {
	id := c.Param("id")
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, d := range h.datasets {
		if d.ID == id {
			h.datasets = append(h.datasets[:i], h.datasets[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "dataset deleted"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "dataset not found"})
}
