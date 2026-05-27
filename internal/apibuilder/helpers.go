package apibuilder

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ===================================================================
// Helper functions for CSV/GIS analysis and widget generation
// ===================================================================

func analyzeColumnTypes(headers []string, rows [][]string) []string {
	types := make([]string, len(headers))
	for i, hdr := range headers {
		lower := strings.ToLower(hdr)

		if strings.Contains(lower, "lat") || strings.Contains(lower, "latitude") {
			types[i] = "geo_lat"
			continue
		}
		if strings.Contains(lower, "lng") || strings.Contains(lower, "lon") || strings.Contains(lower, "longitude") {
			types[i] = "geo_lng"
			continue
		}
		if lower == "region" || lower == "district" || lower == "city" || lower == "country" || lower == "state" || lower == "province" || lower == "area" || lower == "zone" || lower == "location" {
			types[i] = "geo_name"
			continue
		}

		numCount, dateCount := 0, 0
		sampleSize := len(rows)
		if sampleSize > 50 {
			sampleSize = 50
		}
		for j := 0; j < sampleSize; j++ {
			if i >= len(rows[j]) {
				continue
			}
			val := strings.TrimSpace(rows[j][i])
			if val == "" {
				continue
			}
			if _, err := strconv.ParseFloat(val, 64); err == nil {
				numCount++
			}
			if len(val) >= 8 && (strings.Contains(val, "-") || strings.Contains(val, "/")) {
				dateCount++
			}
		}

		threshold := sampleSize / 2
		if threshold < 1 {
			threshold = 1
		}
		if numCount >= threshold {
			types[i] = "number"
		} else if dateCount >= threshold {
			types[i] = "date"
		} else {
			types[i] = "string"
		}
	}
	return types
}

func generateWidgetsFromCSV(headers []string, colTypes []string, rows [][]string) []AnalyticsWidget {
	widgets := []AnalyticsWidget{}
	order := 1

	numericCols := []int{}
	stringCols := []int{}
	dateCols := []int{}
	for i, ct := range colTypes {
		switch ct {
		case "number", "geo_lat", "geo_lng":
			numericCols = append(numericCols, i)
		case "string", "geo_name":
			stringCols = append(stringCols, i)
		case "date":
			dateCols = append(dateCols, i)
		}
	}

	maxKPI := 4
	if len(numericCols) < maxKPI {
		maxKPI = len(numericCols)
	}
	for k := 0; k < maxKPI; k++ {
		col := numericCols[k]
		sum := 0.0
		count := 0
		for _, row := range rows {
			if col < len(row) {
				if v, err := strconv.ParseFloat(strings.TrimSpace(row[col]), 64); err == nil {
					sum += v
					count++
				}
			}
		}
		avg := 0.0
		if count > 0 {
			avg = math.Round(sum/float64(count)*100) / 100
		}

		widgets = append(widgets, AnalyticsWidget{
			ID:     fmt.Sprintf("w-kpi-%d", order),
			Type:   "kpi",
			Title:  fmt.Sprintf("Avg %s", headers[col]),
			Width:  3,
			Height: 1,
			Order:  order,
			Config: WidgetConfig{ShowLegend: false},
			Data:   WidgetData{Value: avg},
		})
		order++
	}

	if len(stringCols) > 0 && len(numericCols) > 0 {
		sc := stringCols[0]
		nc := numericCols[0]

		agg := map[string]float64{}
		for _, row := range rows {
			if sc < len(row) && nc < len(row) {
				key := row[sc]
				if v, err := strconv.ParseFloat(strings.TrimSpace(row[nc]), 64); err == nil {
					agg[key] += v
				}
			}
		}

		labels := make([]string, 0, len(agg))
		values := make([]float64, 0, len(agg))
		for k, v := range agg {
			labels = append(labels, k)
			values = append(values, math.Round(v*100)/100)
		}
		if len(labels) > 15 {
			labels = labels[:15]
			values = values[:15]
		}

		colors := generateColors(len(labels))
		widgets = append(widgets, AnalyticsWidget{
			ID:    fmt.Sprintf("w-bar-%d", order),
			Type:  "bar",
			Title: fmt.Sprintf("%s by %s", headers[nc], headers[sc]),
			Width: 6, Height: 2, Order: order,
			Config: WidgetConfig{XAxis: headers[sc], YAxis: headers[nc], Colors: colors, ShowLegend: true, ShowGrid: true, Animation: true},
			Data: WidgetData{
				Labels:   labels,
				Datasets: []ChartDataset{{Label: headers[nc], Data: values, BackgroundColor: colors, BorderColor: colors[0], BorderWidth: 1}},
			},
		})
		order++
	}

	if len(stringCols) > 0 {
		sc := stringCols[0]
		freq := map[string]int{}
		for _, row := range rows {
			if sc < len(row) {
				freq[row[sc]]++
			}
		}
		labels := make([]string, 0, len(freq))
		values := make([]float64, 0, len(freq))
		for k, v := range freq {
			labels = append(labels, k)
			values = append(values, float64(v))
		}
		if len(labels) > 10 {
			labels = labels[:10]
			values = values[:10]
		}
		colors := generateColors(len(labels))
		widgets = append(widgets, AnalyticsWidget{
			ID:    fmt.Sprintf("w-pie-%d", order),
			Type:  "doughnut",
			Title: fmt.Sprintf("%s Distribution", headers[sc]),
			Width: 6, Height: 2, Order: order,
			Config: WidgetConfig{Colors: colors, ShowLegend: true, Animation: true},
			Data: WidgetData{
				Labels:   labels,
				Datasets: []ChartDataset{{Label: headers[sc], Data: values, BackgroundColor: colors}},
			},
		})
		order++
	}

	if len(dateCols) > 0 && len(numericCols) > 0 {
		dc := dateCols[0]
		nc := numericCols[0]
		labels := make([]string, 0)
		values := make([]float64, 0)
		maxPts := 30
		step := 1
		if len(rows) > maxPts {
			step = len(rows) / maxPts
		}
		for i := 0; i < len(rows); i += step {
			if dc < len(rows[i]) && nc < len(rows[i]) {
				labels = append(labels, rows[i][dc])
				if v, err := strconv.ParseFloat(strings.TrimSpace(rows[i][nc]), 64); err == nil {
					values = append(values, v)
				} else {
					values = append(values, 0)
				}
			}
		}
		widgets = append(widgets, AnalyticsWidget{
			ID:    fmt.Sprintf("w-line-%d", order),
			Type:  "line",
			Title: fmt.Sprintf("%s Over Time", headers[nc]),
			Width: 12, Height: 2, Order: order,
			Config: WidgetConfig{XAxis: headers[dc], YAxis: headers[nc], Colors: []string{"#3b82f6"}, ShowLegend: true, ShowGrid: true, Animation: true},
			Data: WidgetData{
				Labels:   labels,
				Datasets: []ChartDataset{{Label: headers[nc], Data: values, BorderColor: "#3b82f6", Fill: false, Tension: 0.3}},
			},
		})
		order++
	}

	tableCols := make([]TableColumn, len(headers))
	for i, h := range headers {
		colType := "string"
		if i < len(colTypes) {
			switch colTypes[i] {
			case "number", "geo_lat", "geo_lng":
				colType = "number"
			case "date":
				colType = "date"
			}
		}
		tableCols[i] = TableColumn{Key: h, Label: h, Type: colType, Sortable: true}
	}

	tableRows := make([]map[string]interface{}, 0)
	maxRows := 100
	if len(rows) < maxRows {
		maxRows = len(rows)
	}
	for i := 0; i < maxRows; i++ {
		rm := map[string]interface{}{}
		for j, hdr := range headers {
			if j < len(rows[i]) {
				rm[hdr] = rows[i][j]
			}
		}
		tableRows = append(tableRows, rm)
	}

	widgets = append(widgets, AnalyticsWidget{
		ID:    fmt.Sprintf("w-table-%d", order),
		Type:  "table",
		Title: "Data Table",
		Width: 12, Height: 3, Order: order,
		Config: WidgetConfig{ShowGrid: true},
		Data: WidgetData{
			Columns: tableCols,
			Rows:    tableRows,
		},
	})

	return widgets
}

func generateFiltersFromCSV(headers []string, colTypes []string) []DashboardFilter {
	filters := []DashboardFilter{}
	for i, ct := range colTypes {
		if ct == "string" || ct == "geo_name" {
			filters = append(filters, DashboardFilter{
				ID:    fmt.Sprintf("f-%d", i),
				Label: headers[i],
				Type:  "select",
				Key:   headers[i],
			})
		} else if ct == "date" {
			filters = append(filters, DashboardFilter{
				ID:    fmt.Sprintf("f-%d", i),
				Label: headers[i],
				Type:  "date-range",
				Key:   headers[i],
			})
		}
		if len(filters) >= 4 {
			break
		}
	}
	return filters
}

func generateWidgetsFromGISDataset(ds *GISDataset) []AnalyticsWidget {
	widgets := []AnalyticsWidget{}
	order := 1

	widgets = append(widgets, AnalyticsWidget{
		ID: "w-kpi-1", Type: "kpi", Title: "Total Records",
		Width: 3, Height: 1, Order: order,
		Data: WidgetData{Value: len(ds.Rows)},
	})
	order++

	widgets = append(widgets, AnalyticsWidget{
		ID: "w-kpi-2", Type: "kpi", Title: "Data Fields",
		Width: 3, Height: 1, Order: order,
		Data: WidgetData{Value: len(ds.Columns)},
	})
	order++

	numCols := []DatasetColumn{}
	strCols := []DatasetColumn{}
	for _, col := range ds.Columns {
		if col.Type == "number" {
			numCols = append(numCols, col)
		} else {
			strCols = append(strCols, col)
		}
	}

	if len(strCols) > 0 && len(numCols) > 0 {
		sc := strCols[0]
		nc := numCols[0]
		labels := []string{}
		values := []float64{}
		for _, row := range ds.Rows {
			if len(labels) >= 15 {
				break
			}
			if l, ok := row[sc.Key]; ok {
				labels = append(labels, fmt.Sprintf("%v", l))
			}
			if v, ok := row[nc.Key]; ok {
				if fv, err := toFloat64(v); err == nil {
					values = append(values, fv)
				}
			}
		}
		colors := generateColors(len(labels))
		widgets = append(widgets, AnalyticsWidget{
			ID: fmt.Sprintf("w-bar-%d", order), Type: "bar",
			Title: fmt.Sprintf("%s by %s", nc.Label, sc.Label),
			Width: 6, Height: 2, Order: order,
			Config: WidgetConfig{Colors: colors, ShowLegend: true, ShowGrid: true, Animation: true},
			Data: WidgetData{
				Labels:   labels,
				Datasets: []ChartDataset{{Label: nc.Label, Data: values, BackgroundColor: colors}},
			},
		})
		order++
	}

	tableCols := make([]TableColumn, len(ds.Columns))
	for i, c := range ds.Columns {
		tableCols[i] = TableColumn{Key: c.Key, Label: c.Label, Type: c.Type, Sortable: true}
	}
	maxRows := 100
	if len(ds.Rows) < maxRows {
		maxRows = len(ds.Rows)
	}
	widgets = append(widgets, AnalyticsWidget{
		ID: fmt.Sprintf("w-table-%d", order), Type: "table",
		Title: "GIS Data Table", Width: 12, Height: 3, Order: order,
		Config: WidgetConfig{ShowGrid: true},
		Data:   WidgetData{Columns: tableCols, Rows: ds.Rows[:maxRows]},
	})

	return widgets
}

func extractMarkersFromRows(rows []map[string]interface{}, mappings []FieldMapping) []GISMarker {
	latField, lngField, nameField := "", "", ""
	for _, m := range mappings {
		switch m.MappingType {
		case "geo_lat":
			latField = m.SourceField
		case "geo_lng":
			lngField = m.SourceField
		case "geo_region":
			nameField = m.SourceField
		}
	}
	if latField == "" || lngField == "" {
		return nil
	}

	markers := make([]GISMarker, 0)
	for i, row := range rows {
		latV, okLat := row[latField]
		lngV, okLng := row[lngField]
		if !okLat || !okLng {
			continue
		}
		lat, errLat := toFloat64(latV)
		lng, errLng := toFloat64(lngV)
		if errLat != nil || errLng != nil {
			continue
		}
		mName := fmt.Sprintf("Point %d", i+1)
		if nameField != "" {
			if v, ok := row[nameField]; ok {
				mName = fmt.Sprintf("%v", v)
			}
		}
		markers = append(markers, GISMarker{
			ID:         fmt.Sprintf("conv-mkr-%d", i+1),
			Name:       mName,
			Lat:        lat,
			Lng:        lng,
			Category:   "conversion",
			Icon:       "pin",
			Color:      "#ef4444",
			Properties: row,
		})
	}
	return markers
}

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(val), 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func isLikelyGeoLabel(lower string) bool {
	geoTerms := []string{"dhaka", "chittagong", "sylhet", "rajshahi", "khulna", "barisal", "rangpur", "mymensingh",
		"north", "south", "east", "west", "zone", "area", "district", "division", "region", "city", "country"}
	for _, term := range geoTerms {
		if strings.Contains(lower, term) {
			return true
		}
	}
	return false
}

func generateColors(n int) []string {
	palette := []string{
		"#3b82f6", "#ef4444", "#10b981", "#f59e0b", "#8b5cf6",
		"#ec4899", "#06b6d4", "#84cc16", "#f97316", "#6366f1",
		"#14b8a6", "#e11d48", "#a855f7", "#0ea5e9", "#22c55e",
	}
	colors := make([]string, n)
	for i := 0; i < n; i++ {
		colors[i] = palette[i%len(palette)]
	}
	return colors
}

// ===================================================================
// Seed Data
// ===================================================================

func (h *APIBuilderHandler) seedData() {
	now := time.Now()

	sampleAPIs := []*CustomAPI{
		{
			ID: "api-products", Name: "List Products", Method: "GET", Path: "/api/custom/products",
			Description: "Returns paginated product list with filters", Category: "e-commerce",
			AuthRequired: true, RateLimit: 100, Status: "active", CreatedBy: "admin",
			QueryParams: []ParamDef{
				{Name: "page", Type: "number", Required: false, Default: "1"},
				{Name: "limit", Type: "number", Required: false, Default: "20"},
				{Name: "category", Type: "string", Required: false},
			},
			MockResponse: map[string]interface{}{
				"products": []map[string]interface{}{
					{"id": 1, "name": "Widget A", "price": 29.99, "category": "electronics"},
					{"id": 2, "name": "Gadget B", "price": 49.99, "category": "electronics"},
					{"id": 3, "name": "Tool C", "price": 19.99, "category": "tools"},
				},
				"total": 3, "page": 1,
			},
			CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour), HitCount: 142,
		},
		{
			ID: "api-create-order", Name: "Create Order", Method: "POST", Path: "/api/custom/orders",
			Description: "Create a new customer order", Category: "e-commerce",
			AuthRequired: true, RateLimit: 30, Status: "active", CreatedBy: "admin",
			RequestSchema: &SchemaDefinition{
				Type: "object",
				Properties: map[string]SchemaField{
					"customer_id": {Type: "string", Description: "Customer ID"},
					"items":       {Type: "array", Description: "Order items"},
					"total":       {Type: "number", Description: "Order total"},
				},
				Required: []string{"customer_id", "items"},
			},
			MockResponse: map[string]interface{}{"order_id": "ORD-001", "status": "created"},
			CreatedAt:    now.Add(-36 * time.Hour), UpdatedAt: now.Add(-12 * time.Hour), HitCount: 67,
		},
		{
			ID: "api-weather", Name: "Get Weather", Method: "GET", Path: "/api/custom/weather",
			Description: "Weather data for a given city", Category: "external",
			AuthRequired: false, RateLimit: 60, Status: "active", CreatedBy: "admin",
			QueryParams: []ParamDef{
				{Name: "city", Type: "string", Required: true},
			},
			MockResponse: map[string]interface{}{"city": "Dhaka", "temp": 32, "humidity": 78, "condition": "Partly Cloudy"},
			CreatedAt:    now.Add(-24 * time.Hour), UpdatedAt: now, HitCount: 203,
		},
		{
			ID: "api-inventory", Name: "Update Inventory", Method: "PUT", Path: "/api/custom/inventory/:id",
			Description: "Update stock quantity for a product", Category: "warehouse",
			AuthRequired: true, RateLimit: 50, Status: "active", CreatedBy: "admin",
			RequestSchema: &SchemaDefinition{
				Type: "object",
				Properties: map[string]SchemaField{
					"quantity": {Type: "number", Description: "New stock quantity"},
					"location": {Type: "string", Description: "Warehouse location"},
				},
				Required: []string{"quantity"},
			},
			MockResponse: map[string]interface{}{"id": "INV-001", "quantity": 150, "updated": true},
			CreatedAt:    now.Add(-20 * time.Hour), UpdatedAt: now, HitCount: 89,
		},
		{
			ID: "api-user-analytics", Name: "User Analytics", Method: "GET", Path: "/api/custom/analytics/users",
			Description: "User behavior analytics and aggregated metrics", Category: "analytics",
			AuthRequired: true, RateLimit: 20, Status: "draft", CreatedBy: "admin",
			MockResponse: map[string]interface{}{
				"active_users": 1247, "avg_session_min": 12.5, "bounce_rate": 0.32,
				"top_pages": []string{"/dashboard", "/analytics", "/gis"},
			},
			CreatedAt: now.Add(-10 * time.Hour), UpdatedAt: now, HitCount: 0,
		},
	}

	for _, api := range sampleAPIs {
		if strings.TrimSpace(api.APIType) == "" {
			api.APIType = "rest"
		}
		h.customAPIs[api.ID] = api
		h.apiData[api.ID] = []map[string]interface{}{}
	}

	h.csvUploads["csv-demo-001"] = &CSVUpload{
		ID:          "csv-demo-001",
		Filename:    "sales_data_2025.csv",
		FileType:    "csv",
		Rows:        250,
		Columns:     6,
		ColumnNames: []string{"Region", "Product", "Sales", "Revenue", "Latitude", "Longitude"},
		ColumnTypes: []string{"geo_name", "string", "number", "number", "geo_lat", "geo_lng"},
		SampleData: []map[string]interface{}{
			{"Region": "Dhaka", "Product": "Widget A", "Sales": "150", "Revenue": "4500.00", "Latitude": "23.8103", "Longitude": "90.4125"},
			{"Region": "Chittagong", "Product": "Gadget B", "Sales": "98", "Revenue": "4802.00", "Latitude": "22.3569", "Longitude": "91.7832"},
			{"Region": "Sylhet", "Product": "Tool C", "Sales": "65", "Revenue": "1300.00", "Latitude": "24.8949", "Longitude": "91.8687"},
		},
		HasGeoData: true,
		Status:     "analyzed",
		CreatedAt:  now.Add(-5 * time.Hour),
	}
}
