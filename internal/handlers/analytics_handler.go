package handlers

import (
	"encoding/csv"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// =============================================
// Analytics Dashboard Handler
// Provides dynamic charts, graphs, tables with
// editable widget layouts and CSV/Excel export
// =============================================

// --- Models ---

type AnalyticsDashboard struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Widgets     []AnalyticsWidget `json:"widgets"`
	Filters     []DashboardFilter `json:"filters"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type AnalyticsWidget struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"` // bar, line, pie, doughnut, area, heatmap, table, gauge, scatter, radar, funnel, kpi, log
	Title   string                 `json:"title"`
	Width   int                    `json:"width"`  // grid columns 1-12
	Height  int                    `json:"height"` // grid rows 1-4
	Order   int                    `json:"order"`
	Config  WidgetConfig           `json:"config"`
	Data    WidgetData             `json:"data"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type WidgetConfig struct {
	XAxis      string   `json:"xAxis,omitempty"`
	YAxis      string   `json:"yAxis,omitempty"`
	GroupBy    string   `json:"groupBy,omitempty"`
	Colors     []string `json:"colors,omitempty"`
	ShowLegend bool     `json:"showLegend"`
	ShowGrid   bool     `json:"showGrid"`
	Stacked    bool     `json:"stacked,omitempty"`
	Animation  bool     `json:"animation"`
	DataSource string   `json:"dataSource,omitempty"`
}

type WidgetData struct {
	Labels   []string                 `json:"labels,omitempty"`
	Datasets []ChartDataset           `json:"datasets,omitempty"`
	Rows     []map[string]interface{} `json:"rows,omitempty"`
	Columns  []TableColumn            `json:"columns,omitempty"`
	Value    interface{}              `json:"value,omitempty"`
	Min      float64                  `json:"min,omitempty"`
	Max      float64                  `json:"max,omitempty"`
	Entries  []LogEntry               `json:"entries,omitempty"`
}

type ChartDataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BackgroundColor []string  `json:"backgroundColor,omitempty"`
	BorderColor     string    `json:"borderColor,omitempty"`
	BorderWidth     int       `json:"borderWidth,omitempty"`
	Fill            bool      `json:"fill,omitempty"`
	Tension         float64   `json:"tension,omitempty"`
}

type TableColumn struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"` // string, number, date, status, currency
	Sortable bool   `json:"sortable"`
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"` // info, warn, error, debug
	Message   string `json:"message"`
	Source    string `json:"source"`
}

type DashboardFilter struct {
	ID      string         `json:"id"`
	Label   string         `json:"label"`
	Type    string         `json:"type"` // select, date-range, multi-select, search
	Key     string         `json:"key"`
	Options []FilterOption `json:"options,omitempty"`
	Default string         `json:"default,omitempty"`
}

type FilterOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// --- Handler ---

type AnalyticsHandler struct {
	mu         sync.RWMutex
	dashboards map[string]*AnalyticsDashboard
}

func NewAnalyticsHandler() *AnalyticsHandler {
	h := &AnalyticsHandler{
		dashboards: make(map[string]*AnalyticsDashboard),
	}
	h.seedDashboards()
	return h
}

// --- API Endpoints ---

// ListDashboards returns all dashboards (metadata only, no widget data)
func (h *AnalyticsHandler) ListDashboards(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	category := c.Query("category")
	result := make([]gin.H, 0, len(h.dashboards))
	for _, d := range h.dashboards {
		if category != "" && d.Category != category {
			continue
		}
		result = append(result, gin.H{
			"id":          d.ID,
			"name":        d.Name,
			"description": d.Description,
			"category":    d.Category,
			"widgetCount": len(d.Widgets),
			"updated_at":  d.UpdatedAt,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i]["name"].(string) < result[j]["name"].(string)
	})
	c.JSON(http.StatusOK, result)
}

// GetDashboard returns full dashboard with all widget data
func (h *AnalyticsHandler) GetDashboard(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	id := c.Param("id")
	d, ok := h.dashboards[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}
	c.JSON(http.StatusOK, d)
}

// UpdateWidget updates a single widget's configuration (for GUI editing)
func (h *AnalyticsHandler) UpdateWidget(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	dashID := c.Param("id")
	widgetID := c.Param("widgetId")

	d, ok := h.dashboards[dashID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}

	var update struct {
		Type    string                 `json:"type"`
		Title   string                 `json:"title"`
		Width   int                    `json:"width"`
		Height  int                    `json:"height"`
		Order   int                    `json:"order"`
		Config  *WidgetConfig          `json:"config"`
		Options map[string]interface{} `json:"options"`
	}
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for i, w := range d.Widgets {
		if w.ID == widgetID {
			if update.Type != "" {
				d.Widgets[i].Type = update.Type
			}
			if update.Title != "" {
				d.Widgets[i].Title = update.Title
			}
			if update.Width > 0 && update.Width <= 12 {
				d.Widgets[i].Width = update.Width
			}
			if update.Height > 0 && update.Height <= 4 {
				d.Widgets[i].Height = update.Height
			}
			if update.Order >= 0 {
				d.Widgets[i].Order = update.Order
			}
			if update.Config != nil {
				d.Widgets[i].Config = *update.Config
			}
			if update.Options != nil {
				d.Widgets[i].Options = update.Options
			}
			d.UpdatedAt = time.Now()
			c.JSON(http.StatusOK, d.Widgets[i])
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "widget not found"})
}

// ReorderWidgets updates widget ordering
func (h *AnalyticsHandler) ReorderWidgets(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	dashID := c.Param("id")
	d, ok := h.dashboards[dashID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}

	var body struct {
		Order []struct {
			WidgetID string `json:"widgetId"`
			Order    int    `json:"order"`
			Width    int    `json:"width"`
			Height   int    `json:"height"`
		} `json:"order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderMap := make(map[string]struct{ Order, Width, Height int })
	for _, o := range body.Order {
		orderMap[o.WidgetID] = struct{ Order, Width, Height int }{o.Order, o.Width, o.Height}
	}

	for i, w := range d.Widgets {
		if o, found := orderMap[w.ID]; found {
			d.Widgets[i].Order = o.Order
			if o.Width > 0 {
				d.Widgets[i].Width = o.Width
			}
			if o.Height > 0 {
				d.Widgets[i].Height = o.Height
			}
		}
	}
	d.UpdatedAt = time.Now()
	sort.Slice(d.Widgets, func(i, j int) bool {
		return d.Widgets[i].Order < d.Widgets[j].Order
	})
	c.JSON(http.StatusOK, gin.H{"status": "ok", "widgets": len(d.Widgets)})
}

// ExportCSV exports widget data as CSV
func (h *AnalyticsHandler) ExportCSV(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	dashID := c.Param("id")
	widgetID := c.Param("widgetId")

	d, ok := h.dashboards[dashID]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}

	var widget *AnalyticsWidget
	for i := range d.Widgets {
		if d.Widgets[i].ID == widgetID {
			widget = &d.Widgets[i]
			break
		}
	}
	if widget == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "widget not found"})
		return
	}

	filename := strings.ReplaceAll(strings.ToLower(widget.Title), " ", "_") + ".csv"
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	if widget.Type == "table" || widget.Type == "log" {
		// Table/log export
		if len(widget.Data.Columns) > 0 {
			header := make([]string, len(widget.Data.Columns))
			for i, col := range widget.Data.Columns {
				header[i] = col.Label
			}
			writer.Write(header)
			for _, row := range widget.Data.Rows {
				record := make([]string, len(widget.Data.Columns))
				for i, col := range widget.Data.Columns {
					record[i] = fmt.Sprintf("%v", row[col.Key])
				}
				writer.Write(record)
			}
		} else if len(widget.Data.Entries) > 0 {
			writer.Write([]string{"Timestamp", "Level", "Source", "Message"})
			for _, e := range widget.Data.Entries {
				writer.Write([]string{e.Timestamp, e.Level, e.Source, e.Message})
			}
		}
	} else {
		// Chart export: labels + dataset values
		header := []string{widget.Config.XAxis}
		for _, ds := range widget.Data.Datasets {
			header = append(header, ds.Label)
		}
		if len(header) == 1 {
			header[0] = "Label"
		}
		writer.Write(header)
		for i, label := range widget.Data.Labels {
			row := []string{label}
			for _, ds := range widget.Data.Datasets {
				val := ""
				if i < len(ds.Data) {
					val = fmt.Sprintf("%.2f", ds.Data[i])
				}
				row = append(row, val)
			}
			writer.Write(row)
		}
	}
}

// GetWidgetTypes returns available widget types for the editor
func (h *AnalyticsHandler) GetWidgetTypes(c *gin.Context) {
	types := []gin.H{
		{"type": "bar", "label": "Bar Chart", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="12" width="4" height="9" rx="1"/><rect x="10" y="7" width="4" height="14" rx="1"/><rect x="17" y="3" width="4" height="18" rx="1"/></svg>`, "category": "chart"},
		{"type": "line", "label": "Line Chart", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><polyline points="3 17 9 11 13 15 21 7"/><polyline points="17 7 21 7 21 11"/></svg>`, "category": "chart"},
		{"type": "area", "label": "Area Chart", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 20l6-8 4 4 8-10v14H3z" fill="currentColor" opacity="0.15"/><polyline points="3 12 9 4 13 8 21 2"/></svg>`, "category": "chart"},
		{"type": "pie", "label": "Pie Chart", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 2a10 10 0 1 0 10 10h-10V2z"/><path d="M20 12A8 8 0 0 0 12 4v8h8z"/></svg>`, "category": "chart"},
		{"type": "doughnut", "label": "Doughnut Chart", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="9"/><circle cx="12" cy="12" r="4"/><path d="M12 3v4"/><path d="M21 12h-4"/></svg>`, "category": "chart"},
		{"type": "radar", "label": "Radar Chart", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><polygon points="12 2 20 8 18 18 6 18 4 8"/><polygon points="12 6 16 9 15 15 9 15 8 9" opacity="0.4"/><circle cx="12" cy="12" r="1" fill="currentColor"/></svg>`, "category": "chart"},
		{"type": "scatter", "label": "Scatter Plot", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 3v18h18"/><circle cx="8" cy="14" r="1.5" fill="currentColor"/><circle cx="12" cy="9" r="1.5" fill="currentColor"/><circle cx="16" cy="12" r="1.5" fill="currentColor"/><circle cx="18" cy="8" r="1.5" fill="currentColor"/></svg>`, "category": "chart"},
		{"type": "heatmap", "label": "Heatmap", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="5" height="5" rx="1" fill="currentColor" opacity="0.8"/><rect x="10" y="3" width="5" height="5" rx="1" fill="currentColor" opacity="0.4"/><rect x="17" y="3" width="5" height="5" rx="1" fill="currentColor" opacity="0.6"/><rect x="3" y="10" width="5" height="5" rx="1" fill="currentColor" opacity="0.3"/><rect x="10" y="10" width="5" height="5" rx="1" fill="currentColor" opacity="0.9"/><rect x="17" y="10" width="5" height="5" rx="1" fill="currentColor" opacity="0.5"/></svg>`, "category": "chart"},
		{"type": "funnel", "label": "Funnel Chart", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M3 4h18l-6 7v6l-4 3V11L3 4z"/></svg>`, "category": "chart"},
		{"type": "gauge", "label": "Gauge", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 21a9 9 0 1 1 0-18 9 9 0 0 1 0 18z"/><path d="M12 12l4-4"/><circle cx="12" cy="12" r="1.5" fill="currentColor"/></svg>`, "category": "indicator"},
		{"type": "kpi", "label": "KPI Card", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="18" height="18" rx="3"/><path d="M8 12h8"/><path d="M12 8v8"/><circle cx="12" cy="12" r="3" opacity="0.3"/></svg>`, "category": "indicator"},
		{"type": "table", "label": "Data Table", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 9h18"/><path d="M3 15h18"/><path d="M9 3v18"/></svg>`, "category": "data"},
		{"type": "log", "label": "Log Viewer", "icon": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M7 8h10"/><path d="M7 12h7"/><path d="M7 16h10"/></svg>`, "category": "data"},
	}
	c.JSON(http.StatusOK, types)
}

// --- Seed Data ---

func (h *AnalyticsHandler) seedDashboards() {
	h.seedBusinessOverview()
	h.seedBDEconomy()
	h.seedStartupEcosystem()
	h.seedOperationalMetrics()
}

func (h *AnalyticsHandler) seedBusinessOverview() {
	now := time.Now()
	d := &AnalyticsDashboard{
		ID:          "business-overview",
		Name:        "Business Overview",
		Description: "Key business metrics, revenue, customers, and performance KPIs",
		Category:    "business",
		CreatedAt:   now,
		UpdatedAt:   now,
		Filters: []DashboardFilter{
			{ID: "f1", Label: "Year", Type: "select", Key: "year", Default: "2025", Options: []FilterOption{
				{Value: "2023", Label: "2023"}, {Value: "2024", Label: "2024"}, {Value: "2025", Label: "2025"}, {Value: "2026", Label: "2026"},
			}},
			{ID: "f2", Label: "Quarter", Type: "select", Key: "quarter", Default: "", Options: []FilterOption{
				{Value: "", Label: "All"}, {Value: "Q1", Label: "Q1"}, {Value: "Q2", Label: "Q2"}, {Value: "Q3", Label: "Q3"}, {Value: "Q4", Label: "Q4"},
			}},
			{ID: "f3", Label: "Region", Type: "multi-select", Key: "region", Options: []FilterOption{
				{Value: "dhaka", Label: "Dhaka"}, {Value: "chittagong", Label: "Chittagong"}, {Value: "sylhet", Label: "Sylhet"}, {Value: "rajshahi", Label: "Rajshahi"},
				{Value: "khulna", Label: "Khulna"}, {Value: "barisal", Label: "Barisal"}, {Value: "rangpur", Label: "Rangpur"}, {Value: "mymensingh", Label: "Mymensingh"},
			}},
		},
		Widgets: []AnalyticsWidget{
			// KPI cards row
			{ID: "w1", Type: "kpi", Title: "Total Revenue", Width: 3, Height: 1, Order: 0, Config: WidgetConfig{Colors: []string{"#3b82f6"}, Animation: true}, Data: WidgetData{Value: "৳2.45B", Labels: []string{"+12.5% from last quarter"}}},
			{ID: "w2", Type: "kpi", Title: "Active Customers", Width: 3, Height: 1, Order: 1, Config: WidgetConfig{Colors: []string{"#10b981"}, Animation: true}, Data: WidgetData{Value: "11,847", Labels: []string{"+8.3% growth"}}},
			{ID: "w3", Type: "kpi", Title: "Orders Today", Width: 3, Height: 1, Order: 2, Config: WidgetConfig{Colors: []string{"#f59e0b"}, Animation: true}, Data: WidgetData{Value: "1,234", Labels: []string{"+5.7% vs yesterday"}}},
			{ID: "w4", Type: "kpi", Title: "Avg Response Time", Width: 3, Height: 1, Order: 3, Config: WidgetConfig{Colors: []string{"#8b5cf6"}, Animation: true}, Data: WidgetData{Value: "142ms", Labels: []string{"-18% improvement"}}},
			// Revenue bar chart
			{ID: "w5", Type: "bar", Title: "Monthly Revenue (Crore BDT)", Width: 8, Height: 2, Order: 4,
				Config: WidgetConfig{XAxis: "Month", YAxis: "Revenue (Cr)", Colors: []string{"#3b82f6", "#93c5fd"}, ShowLegend: true, ShowGrid: true, Stacked: true, Animation: true},
				Data: WidgetData{
					Labels: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
					Datasets: []ChartDataset{
						{Label: "Product Sales", Data: []float64{185, 210, 195, 220, 245, 260, 275, 290, 310, 295, 320, 340}, BorderColor: "#3b82f6", BackgroundColor: []string{"#3b82f6"}, BorderWidth: 1},
						{Label: "Services", Data: []float64{85, 90, 92, 95, 100, 105, 110, 115, 120, 118, 125, 130}, BorderColor: "#93c5fd", BackgroundColor: []string{"#93c5fd"}, BorderWidth: 1},
					},
				},
			},
			// Pie chart
			{ID: "w6", Type: "doughnut", Title: "Revenue by Division", Width: 4, Height: 2, Order: 5,
				Config: WidgetConfig{ShowLegend: true, Colors: []string{"#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6", "#ec4899", "#06b6d4", "#f97316"}, Animation: true},
				Data: WidgetData{
					Labels: []string{"Dhaka", "Chittagong", "Sylhet", "Rajshahi", "Khulna", "Barisal", "Rangpur", "Mymensingh"},
					Datasets: []ChartDataset{
						{Label: "Revenue Share", Data: []float64{42, 18, 8, 7, 6, 5, 8, 6}, BackgroundColor: []string{"#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6", "#ec4899", "#06b6d4", "#f97316"}},
					},
				},
			},
			// Line chart — growth trend
			{ID: "w7", Type: "line", Title: "Customer Growth Trend", Width: 6, Height: 2, Order: 6,
				Config: WidgetConfig{XAxis: "Month", YAxis: "Customers", Colors: []string{"#10b981", "#f59e0b"}, ShowLegend: true, ShowGrid: true, Animation: true},
				Data: WidgetData{
					Labels: []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
					Datasets: []ChartDataset{
						{Label: "New Customers", Data: []float64{450, 520, 480, 610, 580, 640, 720, 680, 750, 810, 790, 870}, BorderColor: "#10b981", BackgroundColor: []string{"rgba(16,185,129,0.1)"}, Fill: true, Tension: 0.4},
						{Label: "Churned", Data: []float64{120, 100, 130, 90, 110, 95, 80, 105, 85, 70, 90, 75}, BorderColor: "#f59e0b", BackgroundColor: []string{"rgba(245,158,11,0.1)"}, Fill: true, Tension: 0.4},
					},
				},
			},
			// Gauge
			{ID: "w8", Type: "gauge", Title: "Customer Satisfaction", Width: 3, Height: 2, Order: 7,
				Config: WidgetConfig{Colors: []string{"#ef4444", "#f59e0b", "#10b981"}, Animation: true},
				Data:   WidgetData{Value: 79.0, Min: 0, Max: 100, Labels: []string{"Poor", "Average", "Good", "Excellent"}},
			},
			// Radar
			{ID: "w9", Type: "radar", Title: "Performance Metrics", Width: 3, Height: 2, Order: 8,
				Config: WidgetConfig{ShowLegend: true, Colors: []string{"#3b82f6", "#f59e0b"}, Animation: true},
				Data: WidgetData{
					Labels: []string{"Sales", "Marketing", "Support", "Engineering", "HR", "Finance"},
					Datasets: []ChartDataset{
						{Label: "Current", Data: []float64{85, 72, 90, 88, 68, 78}, BorderColor: "#3b82f6", BackgroundColor: []string{"rgba(59,130,246,0.2)"}},
						{Label: "Target", Data: []float64{90, 85, 85, 90, 80, 85}, BorderColor: "#f59e0b", BackgroundColor: []string{"rgba(245,158,11,0.2)"}},
					},
				},
			},
			// Table
			{ID: "w10", Type: "table", Title: "Top Customers by Revenue", Width: 7, Height: 2, Order: 9,
				Config: WidgetConfig{ShowGrid: true, Animation: true},
				Data: WidgetData{
					Columns: []TableColumn{
						{Key: "rank", Label: "#", Type: "number", Sortable: true},
						{Key: "company", Label: "Company", Type: "string", Sortable: true},
						{Key: "sector", Label: "Sector", Type: "string", Sortable: true},
						{Key: "revenue", Label: "Revenue (BDT)", Type: "currency", Sortable: true},
						{Key: "growth", Label: "Growth", Type: "string", Sortable: true},
						{Key: "status", Label: "Status", Type: "status", Sortable: false},
					},
					Rows: []map[string]interface{}{
						{"rank": 1, "company": "Grameenphone", "sector": "Telecom", "revenue": "৳45.2Cr", "growth": "+15%", "status": "active"},
						{"rank": 2, "company": "Square Pharma", "sector": "Pharma", "revenue": "৳38.7Cr", "growth": "+12%", "status": "active"},
						{"rank": 3, "company": "BRAC Bank", "sector": "Finance", "revenue": "৳32.1Cr", "growth": "+8%", "status": "active"},
						{"rank": 4, "company": "Walton Group", "sector": "Electronics", "revenue": "৳28.5Cr", "growth": "+22%", "status": "active"},
						{"rank": 5, "company": "bKash Ltd", "sector": "FinTech", "revenue": "৳24.8Cr", "growth": "+35%", "status": "active"},
						{"rank": 6, "company": "Robi Axiata", "sector": "Telecom", "revenue": "৳21.3Cr", "growth": "+6%", "status": "active"},
						{"rank": 7, "company": "ACI Ltd", "sector": "Conglomerate", "revenue": "৳18.9Cr", "growth": "+10%", "status": "active"},
						{"rank": 8, "company": "Pathao", "sector": "Tech/Ride", "revenue": "৳15.2Cr", "growth": "+45%", "status": "active"},
						{"rank": 9, "company": "Chaldal", "sector": "E-commerce", "revenue": "৳12.8Cr", "growth": "+28%", "status": "active"},
						{"rank": 10, "company": "ShopUp", "sector": "B2B Commerce", "revenue": "৳10.5Cr", "growth": "+52%", "status": "active"},
					},
				},
			},
			// Heatmap
			{ID: "w11", Type: "heatmap", Title: "Orders by Hour & Day", Width: 5, Height: 2, Order: 10,
				Config: WidgetConfig{XAxis: "Hour", YAxis: "Day", Colors: []string{"#eff6ff", "#3b82f6", "#1e3a5f"}, ShowGrid: true, Animation: true},
				Data: WidgetData{
					Labels: []string{"12am", "3am", "6am", "9am", "12pm", "3pm", "6pm", "9pm"},
					Datasets: []ChartDataset{
						{Label: "Sunday", Data: []float64{5, 2, 8, 45, 78, 65, 42, 18}},
						{Label: "Monday", Data: []float64{3, 1, 12, 85, 120, 95, 68, 25}},
						{Label: "Tuesday", Data: []float64{4, 2, 15, 90, 115, 100, 72, 28}},
						{Label: "Wednesday", Data: []float64{3, 1, 14, 88, 125, 98, 70, 22}},
						{Label: "Thursday", Data: []float64{5, 2, 16, 92, 130, 108, 75, 30}},
						{Label: "Friday", Data: []float64{8, 3, 10, 55, 95, 120, 85, 45}},
						{Label: "Saturday", Data: []float64{12, 5, 8, 38, 68, 85, 92, 55}},
					},
				},
			},
			// Log viewer
			{ID: "w12", Type: "log", Title: "Recent Activity Log", Width: 12, Height: 2, Order: 11,
				Config: WidgetConfig{Animation: true},
				Data: WidgetData{
					Entries: generateSampleLogs(),
				},
			},
		},
	}
	h.dashboards[d.ID] = d
}

func (h *AnalyticsHandler) seedBDEconomy() {
	now := time.Now()
	d := &AnalyticsDashboard{
		ID:          "bd-economy",
		Name:        "Bangladesh Economy",
		Description: "National economic indicators — GDP, trade, remittance, population statistics",
		Category:    "economy",
		CreatedAt:   now,
		UpdatedAt:   now,
		Filters: []DashboardFilter{
			{ID: "f1", Label: "Fiscal Year", Type: "select", Key: "fy", Default: "2024-25", Options: []FilterOption{
				{Value: "2022-23", Label: "FY 2022-23"}, {Value: "2023-24", Label: "FY 2023-24"}, {Value: "2024-25", Label: "FY 2024-25"},
			}},
			{ID: "f2", Label: "Sector", Type: "select", Key: "sector", Default: "", Options: []FilterOption{
				{Value: "", Label: "All Sectors"}, {Value: "agriculture", Label: "Agriculture"}, {Value: "industry", Label: "Industry"}, {Value: "service", Label: "Service"},
			}},
		},
		Widgets: []AnalyticsWidget{
			{ID: "e1", Type: "kpi", Title: "GDP (Nominal)", Width: 3, Height: 1, Order: 0, Config: WidgetConfig{Colors: []string{"#059669"}, Animation: true}, Data: WidgetData{Value: "$460.2B", Labels: []string{"+6.5% growth rate"}}},
			{ID: "e2", Type: "kpi", Title: "Per Capita Income", Width: 3, Height: 1, Order: 1, Config: WidgetConfig{Colors: []string{"#2563eb"}, Animation: true}, Data: WidgetData{Value: "$2,784", Labels: []string{"+9.2% YoY"}}},
			{ID: "e3", Type: "kpi", Title: "Remittance", Width: 3, Height: 1, Order: 2, Config: WidgetConfig{Colors: []string{"#d97706"}, Animation: true}, Data: WidgetData{Value: "$23.1B", Labels: []string{"+14.7% increase"}}},
			{ID: "e4", Type: "kpi", Title: "FDI Inflow", Width: 3, Height: 1, Order: 3, Config: WidgetConfig{Colors: []string{"#dc2626"}, Animation: true}, Data: WidgetData{Value: "$3.89B", Labels: []string{"+18.3% growth"}}},
			// GDP trend
			{ID: "e5", Type: "area", Title: "GDP Growth Rate (%)", Width: 8, Height: 2, Order: 4,
				Config: WidgetConfig{XAxis: "Year", YAxis: "Growth %", Colors: []string{"#059669"}, ShowLegend: true, ShowGrid: true, Animation: true},
				Data: WidgetData{
					Labels: []string{"2015", "2016", "2017", "2018", "2019", "2020", "2021", "2022", "2023", "2024", "2025"},
					Datasets: []ChartDataset{
						{Label: "GDP Growth %", Data: []float64{6.6, 7.1, 7.3, 7.9, 8.2, 3.5, 6.9, 7.1, 6.0, 6.5, 6.8}, BorderColor: "#059669", BackgroundColor: []string{"rgba(5,150,105,0.15)"}, Fill: true, Tension: 0.4},
					},
				},
			},
			// GDP composition pie
			{ID: "e6", Type: "pie", Title: "GDP by Sector", Width: 4, Height: 2, Order: 5,
				Config: WidgetConfig{ShowLegend: true, Colors: []string{"#16a34a", "#2563eb", "#d97706"}, Animation: true},
				Data: WidgetData{
					Labels:   []string{"Services (52.5%)", "Industry (35.4%)", "Agriculture (12.1%)"},
					Datasets: []ChartDataset{{Label: "GDP Share", Data: []float64{52.5, 35.4, 12.1}, BackgroundColor: []string{"#2563eb", "#d97706", "#16a34a"}}},
				},
			},
			// Export bar
			{ID: "e7", Type: "bar", Title: "Export Earnings by Sector (Billion USD)", Width: 6, Height: 2, Order: 6,
				Config: WidgetConfig{XAxis: "Sector", YAxis: "Billion USD", Colors: []string{"#2563eb"}, ShowLegend: false, ShowGrid: true, Animation: true},
				Data: WidgetData{
					Labels: []string{"RMG", "Pharma", "Leather", "Frozen Food", "IT Services", "Jute", "Others"},
					Datasets: []ChartDataset{
						{Label: "Export", Data: []float64{47.4, 0.23, 1.2, 0.58, 1.8, 1.1, 3.5}, BackgroundColor: []string{"#3b82f6", "#60a5fa", "#93c5fd", "#bfdbfe", "#2563eb", "#1d4ed8", "#dbeafe"}},
					},
				},
			},
			// Remittance line
			{ID: "e8", Type: "line", Title: "Remittance Trend (Billion USD)", Width: 6, Height: 2, Order: 7,
				Config: WidgetConfig{XAxis: "Year", YAxis: "Billion USD", Colors: []string{"#d97706", "#16a34a"}, ShowLegend: true, ShowGrid: true, Animation: true},
				Data: WidgetData{
					Labels: []string{"2018", "2019", "2020", "2021", "2022", "2023", "2024", "2025"},
					Datasets: []ChartDataset{
						{Label: "Remittance", Data: []float64{15.5, 18.4, 21.7, 22.1, 21.0, 21.6, 23.1, 25.2}, BorderColor: "#d97706", Fill: false, Tension: 0.3},
						{Label: "FDI", Data: []float64{2.58, 2.87, 2.56, 2.89, 3.24, 3.48, 3.89, 4.15}, BorderColor: "#16a34a", Fill: false, Tension: 0.3},
					},
				},
			},
			// Population table
			{ID: "e9", Type: "table", Title: "Division-wise Economic Indicators", Width: 12, Height: 2, Order: 8,
				Config: WidgetConfig{ShowGrid: true, Animation: true},
				Data: WidgetData{
					Columns: []TableColumn{
						{Key: "division", Label: "Division", Type: "string", Sortable: true},
						{Key: "population", Label: "Population", Type: "string", Sortable: true},
						{Key: "gdp_share", Label: "GDP Share %", Type: "number", Sortable: true},
						{Key: "literacy", Label: "Literacy %", Type: "number", Sortable: true},
						{Key: "poverty", Label: "Poverty %", Type: "number", Sortable: true},
						{Key: "industries", Label: "Industrial Units", Type: "number", Sortable: true},
					},
					Rows: []map[string]interface{}{
						{"division": "Dhaka", "population": "44.2M", "gdp_share": 38.5, "literacy": 78.2, "poverty": 18.5, "industries": 45200},
						{"division": "Chittagong", "population": "34.6M", "gdp_share": 22.1, "literacy": 68.5, "poverty": 22.3, "industries": 28900},
						{"division": "Rajshahi", "population": "21.3M", "gdp_share": 10.8, "literacy": 62.1, "poverty": 28.7, "industries": 12400},
						{"division": "Khulna", "population": "17.4M", "gdp_share": 8.5, "literacy": 65.8, "poverty": 25.1, "industries": 9800},
						{"division": "Sylhet", "population": "12.6M", "gdp_share": 6.2, "literacy": 58.4, "poverty": 20.8, "industries": 6200},
						{"division": "Rangpur", "population": "18.5M", "gdp_share": 6.5, "literacy": 55.2, "poverty": 35.4, "industries": 7100},
						{"division": "Barisal", "population": "9.4M", "gdp_share": 4.2, "literacy": 60.1, "poverty": 30.2, "industries": 4300},
						{"division": "Mymensingh", "population": "13.1M", "gdp_share": 3.2, "literacy": 52.8, "poverty": 32.1, "industries": 5100},
					},
				},
			},
		},
	}
	h.dashboards[d.ID] = d
}

func (h *AnalyticsHandler) seedStartupEcosystem() {
	now := time.Now()
	d := &AnalyticsDashboard{
		ID:          "startup-ecosystem",
		Name:        "Startup Ecosystem",
		Description: "Bangladesh startup funding landscape, deals, investor activity",
		Category:    "startup",
		CreatedAt:   now,
		UpdatedAt:   now,
		Filters: []DashboardFilter{
			{ID: "f1", Label: "Year", Type: "select", Key: "year", Default: "2025", Options: []FilterOption{
				{Value: "2020", Label: "2020"}, {Value: "2021", Label: "2021"}, {Value: "2022", Label: "2022"}, {Value: "2023", Label: "2023"}, {Value: "2024", Label: "2024"}, {Value: "2025", Label: "2025"},
			}},
			{ID: "f2", Label: "Sector", Type: "multi-select", Key: "sector", Options: []FilterOption{
				{Value: "fintech", Label: "FinTech"}, {Value: "ecommerce", Label: "E-commerce"}, {Value: "logistics", Label: "Logistics"}, {Value: "healthtech", Label: "HealthTech"},
				{Value: "edtech", Label: "EdTech"}, {Value: "agritech", Label: "AgriTech"}, {Value: "saas", Label: "SaaS"},
			}},
			{ID: "f3", Label: "Round", Type: "select", Key: "round", Default: "", Options: []FilterOption{
				{Value: "", Label: "All Rounds"}, {Value: "seed", Label: "Seed"}, {Value: "pre-seed", Label: "Pre-Seed"}, {Value: "series-a", Label: "Series A"}, {Value: "grant", Label: "Grant"},
			}},
		},
		Widgets: []AnalyticsWidget{
			{ID: "s1", Type: "kpi", Title: "Total Funding (USD)", Width: 3, Height: 1, Order: 0, Config: WidgetConfig{Colors: []string{"#1e3a5f"}, Animation: true}, Data: WidgetData{Value: "$1.02B", Labels: []string{"Cumulative since 2010"}}},
			{ID: "s2", Type: "kpi", Title: "Total Deals", Width: 3, Height: 1, Order: 1, Config: WidgetConfig{Colors: []string{"#0e7490"}, Animation: true}, Data: WidgetData{Value: "456", Labels: []string{"Publicly disclosed"}}},
			{ID: "s3", Type: "kpi", Title: "Unique Startups", Width: 3, Height: 1, Order: 2, Config: WidgetConfig{Colors: []string{"#059669"}, Animation: true}, Data: WidgetData{Value: "167", Labels: []string{"Funded startups"}}},
			{ID: "s4", Type: "kpi", Title: "Avg Deal Size", Width: 3, Height: 1, Order: 3, Config: WidgetConfig{Colors: []string{"#d97706"}, Animation: true}, Data: WidgetData{Value: "$2.24M", Labels: []string{"Median: $500K"}}},
			// Funding bar chart — yearly
			{ID: "s5", Type: "bar", Title: "Startup Funding Raised (USD)", Width: 8, Height: 2, Order: 4,
				Config: WidgetConfig{XAxis: "Year", YAxis: "USD", Colors: []string{"#1e3a5f", "#2563eb"}, ShowLegend: true, ShowGrid: true, Stacked: true, Animation: true},
				Data: WidgetData{
					Labels: []string{"2010", "2011", "2013", "2014", "2015", "2016", "2017", "2018", "2019", "2020", "2021", "2022", "2023", "2024", "2025"},
					Datasets: []ChartDataset{
						{Label: "Global", Data: []float64{4, 0.1, 8, 11, 13, 13, 16, 100, 70, 40, 220, 160, 100, 55, 95}, BackgroundColor: []string{"#1e3a5f"}},
						{Label: "Local", Data: []float64{1, 0.02, 2, 2, 2, 2, 3, 19, 12, 11, 30, 25, 25, 16, 29}, BackgroundColor: []string{"#60a5fa"}},
					},
				},
			},
			// Sector pie
			{ID: "s6", Type: "doughnut", Title: "Funding by Sector", Width: 4, Height: 2, Order: 5,
				Config: WidgetConfig{ShowLegend: true, Colors: []string{"#1e3a5f", "#2563eb", "#0ea5e9", "#06b6d4", "#14b8a6", "#d97706", "#f59e0b", "#ef4444"}, Animation: true},
				Data: WidgetData{
					Labels:   []string{"FinTech", "E-commerce", "Logistics", "HealthTech", "EdTech", "AgriTech", "SaaS", "Others"},
					Datasets: []ChartDataset{{Label: "Share", Data: []float64{35, 18, 12, 8, 7, 6, 5, 9}, BackgroundColor: []string{"#1e3a5f", "#2563eb", "#0ea5e9", "#06b6d4", "#14b8a6", "#d97706", "#f59e0b", "#ef4444"}}},
				},
			},
			// Deals table
			{ID: "s7", Type: "table", Title: "Recent Publicly Announced Deals", Width: 12, Height: 2, Order: 6,
				Config: WidgetConfig{ShowGrid: true, Animation: true},
				Data: WidgetData{
					Columns: []TableColumn{
						{Key: "year", Label: "Year", Type: "number", Sortable: true},
						{Key: "company", Label: "Company", Type: "string", Sortable: true},
						{Key: "round", Label: "Round", Type: "string", Sortable: true},
						{Key: "investor", Label: "Lead Investor", Type: "string", Sortable: true},
						{Key: "source", Label: "Source", Type: "string", Sortable: true},
						{Key: "amount", Label: "Amount (USD)", Type: "currency", Sortable: true},
					},
					Rows: []map[string]interface{}{
						{"year": 2025, "company": "Markopolo", "round": "Seed", "investor": "Joa Capital", "source": "Global", "amount": "$2,000,000"},
						{"year": 2025, "company": "Palki Motors", "round": "Grant", "investor": "Zayed Sustainability", "source": "Global", "amount": "$1,000,000"},
						{"year": 2024, "company": "iFarmer", "round": "Seed", "investor": "Razor Capital", "source": "Global", "amount": "$3,000,000"},
						{"year": 2024, "company": "Chhaya", "round": "Pre-Seed", "investor": "Accelerating Asia", "source": "Global", "amount": "$1,000,000"},
						{"year": 2024, "company": "Relaxy", "round": "Seed", "investor": "Accelerating Asia", "source": "Global", "amount": "$149,142"},
						{"year": 2024, "company": "WeCre8", "round": "Seed", "investor": "IJX", "source": "Local", "amount": "$200,000"},
						{"year": 2023, "company": "ShopUp", "round": "Series B", "investor": "Peter Thiel", "source": "Global", "amount": "$65,000,000"},
						{"year": 2023, "company": "Pathao", "round": "Pre-IPO", "investor": "Startup BD", "source": "Local", "amount": "$10,000,000"},
						{"year": 2022, "company": "bKash", "round": "Series D", "investor": "SoftBank", "source": "Global", "amount": "$250,000,000"},
						{"year": 2022, "company": "Chaldal", "round": "Series C", "investor": "IFC", "source": "Global", "amount": "$10,000,000"},
					},
				},
			},
			// Funnel
			{ID: "s8", Type: "funnel", Title: "Deal Pipeline Funnel", Width: 6, Height: 2, Order: 7,
				Config: WidgetConfig{Colors: []string{"#1e3a5f", "#2563eb", "#3b82f6", "#60a5fa", "#93c5fd"}, ShowLegend: false, Animation: true},
				Data: WidgetData{
					Labels:   []string{"Applicants", "Screening", "Due Diligence", "Term Sheet", "Funded"},
					Datasets: []ChartDataset{{Label: "Count", Data: []float64{1200, 450, 180, 65, 28}}},
				},
			},
			// Investor type radar
			{ID: "s9", Type: "radar", Title: "Investor Activity by Type", Width: 6, Height: 2, Order: 8,
				Config: WidgetConfig{ShowLegend: true, Colors: []string{"#1e3a5f", "#d97706"}, Animation: true},
				Data: WidgetData{
					Labels: []string{"VC", "Angel", "Accelerator", "Corporate", "Impact", "Government"},
					Datasets: []ChartDataset{
						{Label: "Deal Volume", Data: []float64{172, 85, 95, 42, 38, 24}, BorderColor: "#1e3a5f", BackgroundColor: []string{"rgba(30,58,95,0.2)"}},
						{Label: "Deal Count", Data: []float64{45, 120, 65, 28, 55, 18}, BorderColor: "#d97706", BackgroundColor: []string{"rgba(217,119,6,0.2)"}},
					},
				},
			},
		},
	}
	h.dashboards[d.ID] = d
}

func (h *AnalyticsHandler) seedOperationalMetrics() {
	now := time.Now()
	d := &AnalyticsDashboard{
		ID:          "operational-metrics",
		Name:        "Operational Metrics",
		Description: "System performance, API health, infrastructure monitoring",
		Category:    "operations",
		CreatedAt:   now,
		UpdatedAt:   now,
		Filters: []DashboardFilter{
			{ID: "f1", Label: "Time Range", Type: "select", Key: "range", Default: "24h", Options: []FilterOption{
				{Value: "1h", Label: "Last 1 Hour"}, {Value: "6h", Label: "Last 6 Hours"}, {Value: "24h", Label: "Last 24 Hours"}, {Value: "7d", Label: "Last 7 Days"}, {Value: "30d", Label: "Last 30 Days"},
			}},
		},
		Widgets: []AnalyticsWidget{
			{ID: "o1", Type: "kpi", Title: "Uptime", Width: 2, Height: 1, Order: 0, Config: WidgetConfig{Colors: []string{"#059669"}, Animation: true}, Data: WidgetData{Value: "99.97%", Labels: []string{"Last 30 days"}}},
			{ID: "o2", Type: "kpi", Title: "Avg Latency", Width: 2, Height: 1, Order: 1, Config: WidgetConfig{Colors: []string{"#2563eb"}, Animation: true}, Data: WidgetData{Value: "48ms", Labels: []string{"P95: 142ms"}}},
			{ID: "o3", Type: "kpi", Title: "Requests/min", Width: 2, Height: 1, Order: 2, Config: WidgetConfig{Colors: []string{"#d97706"}, Animation: true}, Data: WidgetData{Value: "12,450", Labels: []string{"Peak: 28K"}}},
			{ID: "o4", Type: "kpi", Title: "Error Rate", Width: 2, Height: 1, Order: 3, Config: WidgetConfig{Colors: []string{"#dc2626"}, Animation: true}, Data: WidgetData{Value: "0.03%", Labels: []string{"3 errors/10K"}}},
			{ID: "o5", Type: "kpi", Title: "CPU Usage", Width: 2, Height: 1, Order: 4, Config: WidgetConfig{Colors: []string{"#7c3aed"}, Animation: true}, Data: WidgetData{Value: "34%", Labels: []string{"4 cores"}}},
			{ID: "o6", Type: "kpi", Title: "Memory", Width: 2, Height: 1, Order: 5, Config: WidgetConfig{Colors: []string{"#0891b2"}, Animation: true}, Data: WidgetData{Value: "2.4GB", Labels: []string{"of 8GB (30%)"}}},
			// Request rate line
			{ID: "o7", Type: "line", Title: "Request Rate (req/min)", Width: 8, Height: 2, Order: 6,
				Config: WidgetConfig{XAxis: "Time", YAxis: "Requests/min", Colors: []string{"#2563eb", "#dc2626"}, ShowLegend: true, ShowGrid: true, Animation: true},
				Data: WidgetData{
					Labels: generateTimeLabels(24),
					Datasets: []ChartDataset{
						{Label: "Success", Data: generateSineData(24, 12000, 3000), BorderColor: "#2563eb", Fill: false, Tension: 0.3},
						{Label: "Errors", Data: generateSineData(24, 15, 10), BorderColor: "#dc2626", Fill: false, Tension: 0.3},
					},
				},
			},
			// Response time gauge
			{ID: "o8", Type: "gauge", Title: "API Response Time (ms)", Width: 4, Height: 2, Order: 7,
				Config: WidgetConfig{Colors: []string{"#059669", "#d97706", "#dc2626"}, Animation: true},
				Data:   WidgetData{Value: 48.0, Min: 0, Max: 500, Labels: []string{"Fast", "Normal", "Slow", "Critical"}},
			},
			// Endpoint bar
			{ID: "o9", Type: "bar", Title: "API Endpoints by Request Volume", Width: 6, Height: 2, Order: 8,
				Config: WidgetConfig{XAxis: "Endpoint", YAxis: "Requests", Colors: []string{"#3b82f6"}, ShowLegend: false, ShowGrid: true, Animation: true},
				Data: WidgetData{
					Labels:   []string{"/api/health", "/api/v1/gis/*", "/api/status", "/api/users", "/api/mysql/*", "/api/v1/analytics/*"},
					Datasets: []ChartDataset{{Label: "Requests", Data: []float64{45200, 32100, 28400, 19800, 15600, 12300}, BackgroundColor: []string{"#3b82f6", "#60a5fa", "#93c5fd", "#bfdbfe", "#dbeafe", "#eff6ff"}}},
				},
			},
			// Status code doughnut
			{ID: "o10", Type: "doughnut", Title: "HTTP Status Codes", Width: 6, Height: 2, Order: 9,
				Config: WidgetConfig{ShowLegend: true, Colors: []string{"#059669", "#3b82f6", "#d97706", "#dc2626", "#6b7280"}, Animation: true},
				Data: WidgetData{
					Labels:   []string{"200 OK", "201 Created", "304 Not Modified", "400 Bad Request", "500 Internal"},
					Datasets: []ChartDataset{{Label: "Count", Data: []float64{85420, 12340, 5620, 890, 32}, BackgroundColor: []string{"#059669", "#3b82f6", "#d97706", "#dc2626", "#6b7280"}}},
				},
			},
			// Log viewer
			{ID: "o11", Type: "log", Title: "System Logs", Width: 12, Height: 2, Order: 10,
				Config: WidgetConfig{Animation: true},
				Data:   WidgetData{Entries: generateSystemLogs()},
			},
		},
	}
	h.dashboards[d.ID] = d
}

// --- Helpers ---

func generateSampleLogs() []LogEntry {
	entries := []LogEntry{
		{Timestamp: "2026-03-11 14:32:15", Level: "info", Source: "OrderService", Message: "New order #ORD-28471 created — ৳45,200 — Dhaka"},
		{Timestamp: "2026-03-11 14:31:48", Level: "info", Source: "PaymentGateway", Message: "Payment confirmed for order #ORD-28470 via bKash"},
		{Timestamp: "2026-03-11 14:30:22", Level: "warn", Source: "InventoryService", Message: "Low stock alert: SKU-1847 (Wireless Mouse) — 12 units remaining"},
		{Timestamp: "2026-03-11 14:29:55", Level: "info", Source: "UserService", Message: "New user registered: user_id=U-9847, region=Chittagong"},
		{Timestamp: "2026-03-11 14:28:10", Level: "error", Source: "NotificationService", Message: "SMS delivery failed to +880171XXXX — provider timeout"},
		{Timestamp: "2026-03-11 14:27:33", Level: "info", Source: "AnalyticsEngine", Message: "Daily report generated: 1,234 orders, ৳2.1Cr revenue"},
		{Timestamp: "2026-03-11 14:26:45", Level: "debug", Source: "CacheService", Message: "Cache hit ratio: 94.2% — last 5 minutes"},
		{Timestamp: "2026-03-11 14:25:18", Level: "warn", Source: "APIGateway", Message: "Rate limit approaching for client IP 103.15.XX.XX — 890/1000 req/min"},
		{Timestamp: "2026-03-11 14:24:02", Level: "info", Source: "DeliveryService", Message: "Order #ORD-28465 delivered to Mirpur, Dhaka — delivery time 2h 15m"},
		{Timestamp: "2026-03-11 14:23:30", Level: "info", Source: "AuthService", Message: "Admin login successful: admin@axiom.bd from 103.108.XX.XX"},
		{Timestamp: "2026-03-11 14:22:11", Level: "error", Source: "DatabasePool", Message: "Connection pool exhausted briefly — recovered in 200ms"},
		{Timestamp: "2026-03-11 14:21:45", Level: "info", Source: "SearchService", Message: "Index rebuilt: 45,200 products indexed in 3.2s"},
	}
	return entries
}

func generateSystemLogs() []LogEntry {
	return []LogEntry{
		{Timestamp: "2026-03-11 14:35:02", Level: "info", Source: "APIServer", Message: "Request processed: GET /api/v1/gis/dashboards — 12ms"},
		{Timestamp: "2026-03-11 14:34:58", Level: "info", Source: "APIServer", Message: "Request processed: GET /api/health — 2ms"},
		{Timestamp: "2026-03-11 14:34:45", Level: "warn", Source: "RateLimiter", Message: "Client 103.15.XX.XX approaching rate limit (850/1000 req/min)"},
		{Timestamp: "2026-03-11 14:34:30", Level: "info", Source: "GinRouter", Message: "POST /api/mysql/users — 201 Created — 45ms"},
		{Timestamp: "2026-03-11 14:34:12", Level: "debug", Source: "ConnectionPool", Message: "Active DB connections: MySQL=5, Postgres=3, MongoDB=2"},
		{Timestamp: "2026-03-11 14:33:55", Level: "info", Source: "MetricsCollector", Message: "Metrics flushed to Valkey — 1,240 data points"},
		{Timestamp: "2026-03-11 14:33:40", Level: "error", Source: "WebSocketHub", Message: "Client disconnected unexpectedly: ws_id=WS-4821"},
		{Timestamp: "2026-03-11 14:33:22", Level: "info", Source: "CronScheduler", Message: "Job 'cleanup_expired_tokens' completed — removed 45 tokens"},
		{Timestamp: "2026-03-11 14:33:05", Level: "info", Source: "APIServer", Message: "Request processed: GET /api/v1/analytics/dashboards — 8ms"},
		{Timestamp: "2026-03-11 14:32:48", Level: "warn", Source: "DiskMonitor", Message: "Disk usage at 72% — /var/lib/postgresql"},
	}
}

func generateTimeLabels(count int) []string {
	labels := make([]string, count)
	now := time.Now()
	for i := count - 1; i >= 0; i-- {
		t := now.Add(-time.Duration(i) * time.Hour)
		labels[count-1-i] = t.Format("15:04")
	}
	return labels
}

func generateSineData(count int, base, amplitude float64) []float64 {
	data := make([]float64, count)
	for i := range data {
		data[i] = math.Max(0, base+amplitude*math.Sin(float64(i)*0.5)+amplitude*0.3*math.Cos(float64(i)*1.2))
	}
	return data
}

// ensure uuid import is used
var _ = uuid.New
