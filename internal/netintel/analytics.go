package netintel

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// =====================================================
// Network Intelligence — Predictive Analytics Engine
// Movement prediction, anomaly detection, traffic forecasting
// =====================================================

// --- Models ---

type PredictionType string

const (
	PredMovement       PredictionType = "movement"
	PredAnomaly        PredictionType = "anomaly"
	PredTraffic        PredictionType = "traffic_forecast"
	PredDeviceBehavior PredictionType = "device_behavior"
	PredSignalQuality  PredictionType = "signal_quality"
	PredThreat         PredictionType = "threat_detection"
)

type Prediction struct {
	ID         string                 `json:"id"`
	Type       PredictionType         `json:"type"`
	Timestamp  time.Time              `json:"timestamp"`
	DeviceMAC  string                 `json:"device_mac,omitempty"`
	Confidence float64                `json:"confidence"` // 0-1
	Value      interface{}            `json:"value"`
	Details    map[string]interface{} `json:"details,omitempty"`
	ExpiresAt  time.Time              `json:"expires_at"`
}

type MovementPrediction struct {
	DeviceMAC     string       `json:"device_mac"`
	CurrentZone   string       `json:"current_zone"`
	PredictedZone string       `json:"predicted_zone"`
	Confidence    float64      `json:"confidence"`
	PredictedPath []TrackPoint `json:"predicted_path"`
	ETA           string       `json:"eta"`
	Direction     string       `json:"direction"` // N, NE, E, SE, S, SW, W, NW
	SpeedMps      float64      `json:"speed_mps"`
}

type Anomaly struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`     // traffic_spike, rogue_device, port_scan, unusual_pattern, signal_drop, auth_failure
	Severity    string                 `json:"severity"` // low, medium, high, critical
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Score       float64                `json:"score"` // 0-1 anomaly score
	Metrics     map[string]interface{} `json:"metrics,omitempty"`
	Status      string                 `json:"status"` // active, acknowledged, resolved
	DeviceMAC   string                 `json:"device_mac,omitempty"`
	SrcIP       string                 `json:"src_ip,omitempty"`
	RelatedIDs  []string               `json:"related_ids,omitempty"`
}

type TrafficForecast struct {
	Metric     string          `json:"metric"` // bytes, packets, connections, errors
	Period     string          `json:"period"` // 1h, 6h, 24h
	Points     []ForecastPoint `json:"points"`
	Trend      string          `json:"trend"` // increasing, decreasing, stable, cyclic
	Confidence float64         `json:"confidence"`
	Generated  time.Time       `json:"generated_at"`
}

type ForecastPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Predicted float64   `json:"predicted"`
	Lower     float64   `json:"lower_bound"`
	Upper     float64   `json:"upper_bound"`
}

type Alert struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // anomaly, threshold, prediction, system
	Severity  string    `json:"severity"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // active, acknowledged, resolved, muted
	AnomalyID string    `json:"anomaly_id,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
}

type AnalyticsSummary struct {
	TotalDevices       int                    `json:"total_devices"`
	ActiveDevices      int                    `json:"active_devices"`
	TotalEntries       int                    `json:"total_entries"`
	ActiveAlerts       int                    `json:"active_alerts"`
	ActiveAnomalies    int                    `json:"active_anomalies"`
	PredictionAccuracy float64                `json:"prediction_accuracy"`
	TrafficTrend       string                 `json:"traffic_trend"`
	TopDevices         []DeviceSummary        `json:"top_devices"`
	SeverityBreakdown  map[string]int         `json:"severity_breakdown"`
	ZoneActivity       map[string]int         `json:"zone_activity"`
	Metrics            map[string]interface{} `json:"metrics"`
}

type DeviceSummary struct {
	MAC       string  `json:"mac"`
	Name      string  `json:"name,omitempty"`
	Entries   int     `json:"entries"`
	LastSeen  string  `json:"last_seen"`
	AvgSignal int     `json:"avg_signal_dbm"`
	Zone      string  `json:"current_zone,omitempty"`
	RiskScore float64 `json:"risk_score"`
}

// --- Analytics Engine ---

type AnalyticsEngine struct {
	mu          sync.RWMutex
	parser      *ParserEngine
	anomalies   []Anomaly
	alerts      []Alert
	predictions []Prediction
	forecasts   map[string]*TrafficForecast
	seq         int64
}

func NewAnalyticsEngine(parser *ParserEngine) *AnalyticsEngine {
	ae := &AnalyticsEngine{
		parser:      parser,
		anomalies:   make([]Anomaly, 0),
		alerts:      make([]Alert, 0),
		predictions: make([]Prediction, 0),
		forecasts:   make(map[string]*TrafficForecast),
	}
	ae.seedAnomalies()
	ae.seedAlerts()
	ae.generateForecasts()
	return ae
}

// --- Movement Prediction ---

func (ae *AnalyticsEngine) PredictMovement() []MovementPrediction {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	tracks := ae.parser.ListTracks()
	predictions := make([]MovementPrediction, 0, len(tracks))
	directions := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}

	for _, track := range tracks {
		if len(track.Points) < 3 {
			continue
		}
		last := track.Points[len(track.Points)-1]
		prev := track.Points[len(track.Points)-2]

		dLat := last.Lat - prev.Lat
		dLng := last.Lng - prev.Lng
		angle := math.Atan2(dLng, dLat) * 180 / math.Pi
		if angle < 0 {
			angle += 360
		}
		dirIdx := int((angle+22.5)/45) % 8

		predicted := make([]TrackPoint, 0)
		for i := 1; i <= 3; i++ {
			predicted = append(predicted, TrackPoint{
				Lat:       last.Lat + dLat*float64(i),
				Lng:       last.Lng + dLng*float64(i),
				Timestamp: time.Now().Add(time.Duration(i*5) * time.Minute),
				Zone:      "predicted",
			})
		}

		currentZone := last.Zone
		if currentZone == "" && len(track.Zones) > 0 {
			currentZone = track.Zones[len(track.Zones)-1]
		}
		predictedZone := currentZone
		if len(track.Zones) > 1 {
			predictedZone = track.Zones[len(track.Zones)-1]
		}

		confidence := math.Min(0.95, 0.5+float64(len(track.Points))*0.02)

		predictions = append(predictions, MovementPrediction{
			DeviceMAC:     track.DeviceMAC,
			CurrentZone:   currentZone,
			PredictedZone: predictedZone,
			Confidence:    confidence,
			PredictedPath: predicted,
			ETA:           fmt.Sprintf("%d min", 5+len(track.Points)%10),
			Direction:     directions[dirIdx],
			SpeedMps:      track.AvgSpeed,
		})
	}
	return predictions
}

// --- Anomaly Detection ---

func (ae *AnalyticsEngine) ListAnomalies(status string, severity string) []Anomaly {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	result := make([]Anomaly, 0)
	for _, a := range ae.anomalies {
		if status != "" && a.Status != status {
			continue
		}
		if severity != "" && a.Severity != severity {
			continue
		}
		result = append(result, a)
	}
	return result
}

func (ae *AnalyticsEngine) AcknowledgeAnomaly(id string) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	for i, a := range ae.anomalies {
		if a.ID == id {
			ae.anomalies[i].Status = "acknowledged"
			return nil
		}
	}
	return fmt.Errorf("anomaly not found: %s", id)
}

func (ae *AnalyticsEngine) ResolveAnomaly(id string) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	for i, a := range ae.anomalies {
		if a.ID == id {
			ae.anomalies[i].Status = "resolved"
			return nil
		}
	}
	return fmt.Errorf("anomaly not found: %s", id)
}

// --- Alerts ---

func (ae *AnalyticsEngine) ListAlerts(status string) []Alert {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	result := make([]Alert, 0)
	for _, a := range ae.alerts {
		if status != "" && a.Status != status {
			continue
		}
		result = append(result, a)
	}
	return result
}

func (ae *AnalyticsEngine) AcknowledgeAlert(id string) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	for i, a := range ae.alerts {
		if a.ID == id {
			ae.alerts[i].Status = "acknowledged"
			return nil
		}
	}
	return fmt.Errorf("alert not found: %s", id)
}

func (ae *AnalyticsEngine) ResolveAlert(id string) error {
	ae.mu.Lock()
	defer ae.mu.Unlock()
	for i, a := range ae.alerts {
		if a.ID == id {
			ae.alerts[i].Status = "resolved"
			return nil
		}
	}
	return fmt.Errorf("alert not found: %s", id)
}

// --- Forecasts ---

func (ae *AnalyticsEngine) GetForecast(metric string) (*TrafficForecast, bool) {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	f, ok := ae.forecasts[metric]
	return f, ok
}

func (ae *AnalyticsEngine) ListForecasts() map[string]*TrafficForecast {
	ae.mu.RLock()
	defer ae.mu.RUnlock()
	return ae.forecasts
}

func (ae *AnalyticsEngine) generateForecasts() {
	now := time.Now()
	metrics := []struct {
		name  string
		base  float64
		trend string
	}{
		{"traffic_bytes", 50e6, "increasing"},
		{"connections", 1200, "stable"},
		{"errors", 15, "decreasing"},
		{"packets", 85000, "cyclic"},
		{"latency_ms", 12.5, "stable"},
	}

	for _, m := range metrics {
		points := make([]ForecastPoint, 24)
		for i := range points {
			t := now.Add(time.Duration(i) * time.Hour)
			hourFactor := 1.0 + 0.3*math.Sin(float64(i)*math.Pi/12)
			predicted := m.base * hourFactor
			if m.trend == "increasing" {
				predicted *= 1 + float64(i)*0.02
			} else if m.trend == "decreasing" {
				predicted *= 1 - float64(i)*0.01
			}
			points[i] = ForecastPoint{
				Timestamp: t,
				Predicted: predicted,
				Lower:     predicted * 0.85,
				Upper:     predicted * 1.15,
			}
		}
		ae.forecasts[m.name] = &TrafficForecast{
			Metric:     m.name,
			Period:     "24h",
			Points:     points,
			Trend:      m.trend,
			Confidence: 0.82 + float64(len(m.name)%5)*0.03,
			Generated:  now,
		}
	}
}

// --- Dashboard Summary ---

func (ae *AnalyticsEngine) GetSummary() *AnalyticsSummary {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	stats := ae.parser.GetEntryStats()
	tracks := ae.parser.ListTracks()

	activeAlerts := 0
	for _, a := range ae.alerts {
		if a.Status == "active" {
			activeAlerts++
		}
	}
	activeAnomalies := 0
	for _, a := range ae.anomalies {
		if a.Status == "active" {
			activeAnomalies++
		}
	}

	// Severity breakdown from anomalies
	sevBreakdown := map[string]int{"low": 0, "medium": 0, "high": 0, "critical": 0}
	for _, a := range ae.anomalies {
		sevBreakdown[a.Severity]++
	}

	// Zone activity from tracks
	zoneActivity := make(map[string]int)
	for _, t := range tracks {
		for _, z := range t.Zones {
			zoneActivity[z]++
		}
	}

	// Top devices
	type devInfo struct {
		mac     string
		name    string
		entries int
		signal  int
		count   int
		zone    string
		lastTS  time.Time
	}
	devMap := make(map[string]*devInfo)
	entries := ae.parser.ListEntries("", 5000, "")
	for _, e := range entries {
		if e.DeviceMAC == "" {
			continue
		}
		d, ok := devMap[e.DeviceMAC]
		if !ok {
			d = &devInfo{mac: e.DeviceMAC}
			devMap[e.DeviceMAC] = d
		}
		d.entries++
		if e.Signal != 0 {
			d.signal += e.Signal
			d.count++
		}
		if e.Timestamp.After(d.lastTS) {
			d.lastTS = e.Timestamp
		}
	}
	for _, t := range tracks {
		if d, ok := devMap[t.DeviceMAC]; ok {
			d.name = t.DeviceName
			if len(t.Zones) > 0 {
				d.zone = t.Zones[len(t.Zones)-1]
			}
		}
	}

	topDevices := make([]DeviceSummary, 0)
	for _, d := range devMap {
		avgSig := 0
		if d.count > 0 {
			avgSig = d.signal / d.count
		}
		topDevices = append(topDevices, DeviceSummary{
			MAC: d.mac, Name: d.name, Entries: d.entries,
			LastSeen: d.lastTS.Format(time.RFC3339), AvgSignal: avgSig,
			Zone: d.zone, RiskScore: math.Min(1.0, float64(d.entries)/500.0*0.3),
		})
	}
	sort.Slice(topDevices, func(i, j int) bool { return topDevices[i].Entries > topDevices[j].Entries })
	if len(topDevices) > 10 {
		topDevices = topDevices[:10]
	}

	totalEntries := 0
	if v, ok := stats["total_entries"].(int); ok {
		totalEntries = v
	}

	return &AnalyticsSummary{
		TotalDevices:       len(devMap),
		ActiveDevices:      len(tracks),
		TotalEntries:       totalEntries,
		ActiveAlerts:       activeAlerts,
		ActiveAnomalies:    activeAnomalies,
		PredictionAccuracy: 0.87,
		TrafficTrend:       "stable",
		TopDevices:         topDevices,
		SeverityBreakdown:  sevBreakdown,
		ZoneActivity:       zoneActivity,
		Metrics: map[string]interface{}{
			"avg_signal_dbm":      -52,
			"avg_latency_ms":      12.4,
			"bandwidth_mbps":      847.3,
			"packet_loss_pct":     0.12,
			"active_parsers":      stats["active_parsers"],
			"unique_protocols":    len(stats["by_protocol"].(map[string]int)),
			"forecasts_available": len(ae.forecasts),
		},
	}
}

// --- Seed Data ---

func (ae *AnalyticsEngine) seedAnomalies() {
	now := time.Now()
	ae.anomalies = []Anomaly{
		{ID: "anom-001", Type: "traffic_spike", Severity: "high", Timestamp: now.Add(-25 * time.Minute), Source: "router-core-1", Description: "Traffic volume 340% above baseline on port Gi0/1 — possible DDoS or bandwidth abuse", Score: 0.89, Status: "active", Metrics: map[string]interface{}{"baseline_mbps": 120, "current_mbps": 528, "duration_min": 18}},
		{ID: "anom-002", Type: "rogue_device", Severity: "critical", Timestamp: now.Add(-42 * time.Minute), Source: "AP-Floor-2", Description: "Unknown device with MAC de:ad:be:ef:ca:fe broadcasting rogue SSID 'AxiomNizam-Corp'", Score: 0.95, Status: "active", DeviceMAC: "de:ad:be:ef:ca:fe", Metrics: map[string]interface{}{"rogue_ssid": "AxiomNizam-Corp", "signal_dbm": -35, "channel": 6}},
		{ID: "anom-003", Type: "port_scan", Severity: "high", Timestamp: now.Add(-1 * time.Hour), Source: "fw-edge-1", Description: "Sequential port scan detected from 10.42.8.99 targeting 10.0.1.0/24 subnet", Score: 0.92, Status: "active", SrcIP: "10.42.8.99", Metrics: map[string]interface{}{"ports_scanned": 1024, "target_subnet": "10.0.1.0/24", "scan_rate_pps": 500}},
		{ID: "anom-004", Type: "unusual_pattern", Severity: "medium", Timestamp: now.Add(-2 * time.Hour), Source: "esp32-node-3", Description: "CSI amplitude variance exceeds 3σ — possible new movement pattern or environmental change", Score: 0.72, Status: "acknowledged", Metrics: map[string]interface{}{"amplitude_variance": 0.48, "baseline_variance": 0.12, "sigma_distance": 3.2}},
		{ID: "anom-005", Type: "signal_drop", Severity: "medium", Timestamp: now.Add(-3 * time.Hour), Source: "AP-Floor-1", Description: "Signal strength dropped 25dB in 5 min for 8 devices — possible interference or AP failure", Score: 0.78, Status: "acknowledged", Metrics: map[string]interface{}{"affected_devices": 8, "avg_drop_db": 25, "ap_name": "AP-Floor-1"}},
		{ID: "anom-006", Type: "auth_failure", Severity: "high", Timestamp: now.Add(-90 * time.Minute), Source: "radius-srv-1", Description: "47 failed RADIUS authentications from MAC 3a:b1:cc:... in 10 minutes — brute force suspected", Score: 0.88, Status: "active", DeviceMAC: "3a:b1:cc:dd:ee:ff", Metrics: map[string]interface{}{"failure_count": 47, "window_min": 10, "username_attempts": 12}},
		{ID: "anom-007", Type: "traffic_spike", Severity: "low", Timestamp: now.Add(-5 * time.Hour), Source: "dns-resolver-1", Description: "DNS query rate 180% above normal — likely legitimate surge from deployment", Score: 0.45, Status: "resolved", Metrics: map[string]interface{}{"baseline_qps": 450, "current_qps": 810}},
		{ID: "anom-008", Type: "unusual_pattern", Severity: "critical", Timestamp: now.Add(-15 * time.Minute), Source: "fw-edge-1", Description: "Data exfiltration pattern: 2.3GB uploaded to external IP in 8 minutes via encrypted tunnel", Score: 0.97, Status: "active", SrcIP: "10.0.1.55", Metrics: map[string]interface{}{"bytes_sent": 2.3e9, "destination": "external", "protocol": "TLS", "duration_min": 8}},
	}
}

func (ae *AnalyticsEngine) seedAlerts() {
	now := time.Now()
	ae.alerts = []Alert{
		{ID: "alert-001", Type: "anomaly", Severity: "critical", Title: "Rogue AP Detected", Message: "Unauthorized access point broadcasting corporate SSID on Floor 2", Source: "anomaly-engine", Timestamp: now.Add(-42 * time.Minute), Status: "active", AnomalyID: "anom-002", Tags: []string{"wireless", "security", "rogue-ap"}},
		{ID: "alert-002", Type: "anomaly", Severity: "critical", Title: "Data Exfiltration Detected", Message: "2.3GB outbound transfer to unknown external IP via encrypted tunnel from 10.0.1.55", Source: "anomaly-engine", Timestamp: now.Add(-15 * time.Minute), Status: "active", AnomalyID: "anom-008", Tags: []string{"security", "data-loss", "critical"}},
		{ID: "alert-003", Type: "anomaly", Severity: "high", Title: "Active Port Scan", Message: "Sequential port scan from 10.42.8.99 targeting internal subnet 10.0.1.0/24", Source: "anomaly-engine", Timestamp: now.Add(-1 * time.Hour), Status: "active", AnomalyID: "anom-003", Tags: []string{"security", "scan"}},
		{ID: "alert-004", Type: "threshold", Severity: "high", Title: "Traffic Spike on Core Router", Message: "Interface Gi0/1 traffic 340% above baseline for 18 minutes", Source: "threshold-monitor", Timestamp: now.Add(-25 * time.Minute), Status: "active", Tags: []string{"network", "performance"}},
		{ID: "alert-005", Type: "anomaly", Severity: "high", Title: "RADIUS Brute Force", Message: "47 failed auth attempts from single device in 10-minute window", Source: "anomaly-engine", Timestamp: now.Add(-90 * time.Minute), Status: "active", AnomalyID: "anom-006", Tags: []string{"security", "authentication"}},
		{ID: "alert-006", Type: "prediction", Severity: "medium", Title: "Bandwidth Exhaustion Warning", Message: "Traffic forecast predicts 95% link utilization within 4 hours if current trend continues", Source: "forecast-engine", Timestamp: now.Add(-30 * time.Minute), Status: "active", Tags: []string{"network", "capacity"}},
		{ID: "alert-007", Type: "threshold", Severity: "medium", Title: "AP Signal Degradation", Message: "AP-Floor-1 average signal dropped 25dB affecting 8 devices", Source: "threshold-monitor", Timestamp: now.Add(-3 * time.Hour), Status: "acknowledged", Tags: []string{"wireless", "performance"}},
		{ID: "alert-008", Type: "system", Severity: "low", Title: "Parser Error Rate", Message: "Syslog parser error rate increased to 0.15% (threshold: 0.1%)", Source: "parser-monitor", Timestamp: now.Add(-6 * time.Hour), Status: "resolved", Tags: []string{"system", "parser"}},
		{ID: "alert-009", Type: "prediction", Severity: "medium", Title: "Movement Pattern Anomaly", Message: "Device pattern deviates 4.2σ from learned behavior — possible tailgating", Source: "movement-engine", Timestamp: now.Add(-1 * time.Hour), Status: "active", Tags: []string{"security", "physical", "movement"}},
		{ID: "alert-010", Type: "system", Severity: "low", Title: "CSI Node Offline", Message: "ESP32 node esp32-node-4 has not reported CSI data for 12 minutes", Source: "health-monitor", Timestamp: now.Add(-12 * time.Minute), Status: "active", Tags: []string{"system", "csi", "health"}},
	}
}
