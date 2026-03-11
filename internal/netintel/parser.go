package netintel

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// =====================================================
// Network Intelligence — Log Parser Engine
// Parses Wi-Fi probes, IP flows, syslog, pcap, CSI data
// =====================================================

// --- Log Types ---

type LogType string

const (
	LogWiFiProbe LogType = "wifi_probe"
	LogIPFlow    LogType = "ip_flow"
	LogSyslog    LogType = "syslog"
	LogPcap      LogType = "pcap"
	LogCSI       LogType = "csi" // Channel State Information
	LogNetFlow   LogType = "netflow"
	LogDHCP      LogType = "dhcp"
	LogDNS       LogType = "dns"
	LogRadius    LogType = "radius"
	LogSNMP      LogType = "snmp"
	LogFirewall  LogType = "firewall"
	LogBluetooth LogType = "bluetooth"
)

// --- Parser Configuration ---

type ParserConfig struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	LogType     LogType                `json:"log_type"`
	Description string                 `json:"description"`
	Source      ParserSource           `json:"source"`
	Filters     ParserFilters          `json:"filters,omitempty"`
	Status      string                 `json:"status"` // active, paused, stopped
	Config      map[string]interface{} `json:"config,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ParsedCount int64                  `json:"parsed_count"`
	ErrorCount  int64                  `json:"error_count"`
	Tags        []string               `json:"tags,omitempty"`
}

type ParserSource struct {
	Type     string                 `json:"type"` // file, stream, api, agent, syslog_receiver
	Endpoint string                 `json:"endpoint,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

type ParserFilters struct {
	MinSignal     *int     `json:"min_signal,omitempty"`
	MacPrefixes   []string `json:"mac_prefixes,omitempty"`
	Protocols     []string `json:"protocols,omitempty"`
	SeverityLevel string   `json:"severity_level,omitempty"` // debug, info, warn, error, critical
	IPRanges      []string `json:"ip_ranges,omitempty"`
}

// --- Parsed Log Entry (normalized) ---

type ParsedEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	LogType   LogType                `json:"log_type"`
	Source    string                 `json:"source"`
	Severity  string                 `json:"severity"` // debug, info, warn, error, critical
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	DeviceMAC string                 `json:"device_mac,omitempty"`
	SSID      string                 `json:"ssid,omitempty"`
	Signal    int                    `json:"signal,omitempty"` // dBm
	SrcIP     string                 `json:"src_ip,omitempty"`
	DstIP     string                 `json:"dst_ip,omitempty"`
	Protocol  string                 `json:"protocol,omitempty"`
	Port      int                    `json:"port,omitempty"`
	Bytes     int64                  `json:"bytes,omitempty"`
	Location  *GeoPoint              `json:"location,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
}

type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// --- Log Type Definitions ---

type LogTypeInfo struct {
	ID          LogType  `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"` // wireless, network, security, infrastructure
	Fields      []string `json:"fields"`
	Icon        string   `json:"icon"`
}

// --- Signal Heatmap ---

type HeatmapPoint struct {
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Intensity float64 `json:"intensity"` // 0-1
	Label     string  `json:"label,omitempty"`
	DeviceMAC string  `json:"device_mac,omitempty"`
	SignalDBm int     `json:"signal_dbm,omitempty"`
}

type HeatmapData struct {
	Points    []HeatmapPoint `json:"points"`
	GridSizeM float64        `json:"grid_size_meters"`
	MaxSignal int            `json:"max_signal_dbm"`
	MinSignal int            `json:"min_signal_dbm"`
	Generated time.Time      `json:"generated_at"`
	Category  string         `json:"category"` // wifi_signal, movement_density, traffic_volume
}

// --- Movement Track ---

type MovementTrack struct {
	DeviceMAC  string       `json:"device_mac"`
	DeviceName string       `json:"device_name,omitempty"`
	Points     []TrackPoint `json:"points"`
	FirstSeen  time.Time    `json:"first_seen"`
	LastSeen   time.Time    `json:"last_seen"`
	Distance   float64      `json:"total_distance_m"`
	AvgSpeed   float64      `json:"avg_speed_mps"`
	Zones      []string     `json:"zones_visited"`
	Predicted  []TrackPoint `json:"predicted_path,omitempty"`
}

type TrackPoint struct {
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Timestamp time.Time `json:"timestamp"`
	Signal    int       `json:"signal_dbm,omitempty"`
	Zone      string    `json:"zone,omitempty"`
}

// --- Parser Engine ---

type ParserEngine struct {
	mu         sync.RWMutex
	parsers    map[string]*ParserConfig
	entries    []ParsedEntry
	logTypes   []LogTypeInfo
	tracks     map[string]*MovementTrack
	sequence   int64
	maxEntries int
}

func NewParserEngine() *ParserEngine {
	pe := &ParserEngine{
		parsers:    make(map[string]*ParserConfig),
		entries:    make([]ParsedEntry, 0),
		tracks:     make(map[string]*MovementTrack),
		maxEntries: 50000,
	}
	pe.registerLogTypes()
	pe.seedParsers()
	pe.seedEntries()
	pe.seedTracks()
	return pe
}

func (pe *ParserEngine) registerLogTypes() {
	pe.logTypes = []LogTypeInfo{
		{ID: LogWiFiProbe, Name: "Wi-Fi Probe Requests", Description: "802.11 probe requests from devices scanning for networks", Category: "wireless", Fields: []string{"device_mac", "ssid", "signal_dbm", "channel", "vendor"}, Icon: "📡"},
		{ID: LogIPFlow, Name: "IP Flow Records", Description: "Aggregated network flow data (NetFlow/sFlow/IPFIX format)", Category: "network", Fields: []string{"src_ip", "dst_ip", "src_port", "dst_port", "protocol", "bytes", "packets"}, Icon: "🔀"},
		{ID: LogSyslog, Name: "Syslog Messages", Description: "System log messages from network devices and servers", Category: "infrastructure", Fields: []string{"facility", "severity", "hostname", "process", "message"}, Icon: "📋"},
		{ID: LogPcap, Name: "Packet Captures", Description: "Raw packet capture data in pcap/pcapng format", Category: "network", Fields: []string{"src_mac", "dst_mac", "ethertype", "protocol", "payload_size"}, Icon: "📦"},
		{ID: LogCSI, Name: "Channel State Information", Description: "WiFi CSI data for signal analysis and movement detection", Category: "wireless", Fields: []string{"subcarrier_count", "amplitude", "phase", "rssi", "antenna_count"}, Icon: "🌊"},
		{ID: LogNetFlow, Name: "NetFlow/IPFIX", Description: "Cisco NetFlow or IETF IPFIX flow export records", Category: "network", Fields: []string{"src_ip", "dst_ip", "bytes", "packets", "duration", "tos"}, Icon: "🔄"},
		{ID: LogDHCP, Name: "DHCP Leases", Description: "DHCP lease events and address assignments", Category: "infrastructure", Fields: []string{"mac", "ip", "hostname", "lease_time", "action"}, Icon: "🏠"},
		{ID: LogDNS, Name: "DNS Queries", Description: "DNS query and response logs", Category: "network", Fields: []string{"query", "type", "response", "client_ip", "latency_ms"}, Icon: "🔍"},
		{ID: LogRadius, Name: "RADIUS Auth", Description: "802.1X/RADIUS authentication logs", Category: "security", Fields: []string{"username", "mac", "nas_ip", "result", "reason"}, Icon: "🔐"},
		{ID: LogSNMP, Name: "SNMP Traps", Description: "SNMP trap and inform notifications from network devices", Category: "infrastructure", Fields: []string{"oid", "community", "agent_ip", "trap_type", "value"}, Icon: "⚡"},
		{ID: LogFirewall, Name: "Firewall Logs", Description: "Firewall accept/deny/drop action logs", Category: "security", Fields: []string{"action", "src_ip", "dst_ip", "src_port", "dst_port", "rule_id", "protocol"}, Icon: "🛡️"},
		{ID: LogBluetooth, Name: "Bluetooth LE", Description: "Bluetooth Low Energy advertisement and scan data", Category: "wireless", Fields: []string{"device_mac", "rssi", "uuid", "major", "minor", "tx_power"}, Icon: "🔵"},
	}
}

func (pe *ParserEngine) GetLogTypes() []LogTypeInfo { return pe.logTypes }

// --- Parser CRUD ---

func (pe *ParserEngine) CreateParser(p *ParserConfig) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	if p.ID == "" {
		pe.sequence++
		p.ID = fmt.Sprintf("parser-%d", pe.sequence)
	}
	if p.Status == "" {
		p.Status = "active"
	}
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	pe.parsers[p.ID] = p
	return nil
}

func (pe *ParserEngine) GetParser(id string) (*ParserConfig, bool) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	p, ok := pe.parsers[id]
	return p, ok
}

func (pe *ParserEngine) ListParsers() []*ParserConfig {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	result := make([]*ParserConfig, 0, len(pe.parsers))
	for _, p := range pe.parsers {
		result = append(result, p)
	}
	return result
}

func (pe *ParserEngine) UpdateParser(id string, updates map[string]interface{}) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	p, ok := pe.parsers[id]
	if !ok {
		return fmt.Errorf("parser not found: %s", id)
	}
	if name, ok := updates["name"].(string); ok && name != "" {
		p.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		p.Description = desc
	}
	if status, ok := updates["status"].(string); ok {
		p.Status = status
	}
	p.UpdatedAt = time.Now()
	return nil
}

func (pe *ParserEngine) DeleteParser(id string) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	if _, ok := pe.parsers[id]; !ok {
		return fmt.Errorf("parser not found: %s", id)
	}
	delete(pe.parsers, id)
	return nil
}

// --- Ingest (dynamic API endpoint) ---

func (pe *ParserEngine) IngestLog(entry ParsedEntry) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	if entry.ID == "" {
		b := make([]byte, 8)
		rand.Read(b)
		entry.ID = hex.EncodeToString(b)
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if len(pe.entries) >= pe.maxEntries {
		pe.entries = pe.entries[1:]
	}
	pe.entries = append(pe.entries, entry)

	// Update parser counters
	for _, p := range pe.parsers {
		if p.LogType == entry.LogType && p.Status == "active" {
			p.ParsedCount++
		}
	}

	// Track movement for Wi-Fi probes with location
	if entry.DeviceMAC != "" && entry.Location != nil {
		pe.updateTrack(entry)
	}
	return nil
}

func (pe *ParserEngine) updateTrack(e ParsedEntry) {
	track, ok := pe.tracks[e.DeviceMAC]
	if !ok {
		track = &MovementTrack{
			DeviceMAC: e.DeviceMAC,
			FirstSeen: e.Timestamp,
			Points:    make([]TrackPoint, 0),
			Zones:     make([]string, 0),
		}
		pe.tracks[e.DeviceMAC] = track
	}
	tp := TrackPoint{
		Lat:       e.Location.Lat,
		Lng:       e.Location.Lng,
		Timestamp: e.Timestamp,
		Signal:    e.Signal,
	}
	track.Points = append(track.Points, tp)
	track.LastSeen = e.Timestamp

	if len(track.Points) > 1 {
		prev := track.Points[len(track.Points)-2]
		d := haversine(prev.Lat, prev.Lng, tp.Lat, tp.Lng)
		track.Distance += d
		elapsed := track.LastSeen.Sub(track.FirstSeen).Seconds()
		if elapsed > 0 {
			track.AvgSpeed = track.Distance / elapsed
		}
	}
}

// --- Query Entries ---

func (pe *ParserEngine) ListEntries(logType string, limit int, severity string) []ParsedEntry {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	result := make([]ParsedEntry, 0)
	for i := len(pe.entries) - 1; i >= 0 && len(result) < limit; i-- {
		e := pe.entries[i]
		if logType != "" && string(e.LogType) != logType {
			continue
		}
		if severity != "" && e.Severity != severity {
			continue
		}
		result = append(result, e)
	}
	return result
}

func (pe *ParserEngine) GetEntryStats() map[string]interface{} {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	byType := make(map[string]int)
	bySeverity := make(map[string]int)
	byProtocol := make(map[string]int)
	var totalBytes int64
	for _, e := range pe.entries {
		byType[string(e.LogType)]++
		if e.Severity != "" {
			bySeverity[e.Severity]++
		}
		if e.Protocol != "" {
			byProtocol[e.Protocol]++
		}
		totalBytes += e.Bytes
	}
	return map[string]interface{}{
		"total_entries":  len(pe.entries),
		"by_type":        byType,
		"by_severity":    bySeverity,
		"by_protocol":    byProtocol,
		"total_bytes":    totalBytes,
		"unique_devices": len(pe.tracks),
		"active_parsers": pe.countActiveParsers(),
	}
}

func (pe *ParserEngine) countActiveParsers() int {
	count := 0
	for _, p := range pe.parsers {
		if p.Status == "active" {
			count++
		}
	}
	return count
}

// --- Heatmap Generation ---

func (pe *ParserEngine) GenerateHeatmap(category string) *HeatmapData {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	points := make([]HeatmapPoint, 0)
	minSig, maxSig := 0, -100

	switch category {
	case "wifi_signal":
		for _, e := range pe.entries {
			if e.Location != nil && e.Signal != 0 {
				intensity := math.Max(0, math.Min(1, float64(e.Signal+100)/70.0))
				points = append(points, HeatmapPoint{
					Lat: e.Location.Lat, Lng: e.Location.Lng,
					Intensity: intensity, DeviceMAC: e.DeviceMAC, SignalDBm: e.Signal,
				})
				if e.Signal < minSig {
					minSig = e.Signal
				}
				if e.Signal > maxSig {
					maxSig = e.Signal
				}
			}
		}
	case "movement_density":
		density := make(map[string]int)
		for _, t := range pe.tracks {
			for _, p := range t.Points {
				key := fmt.Sprintf("%.4f,%.4f", p.Lat, p.Lng)
				density[key]++
			}
		}
		maxDensity := 1
		for _, c := range density {
			if c > maxDensity {
				maxDensity = c
			}
		}
		for _, t := range pe.tracks {
			for _, p := range t.Points {
				key := fmt.Sprintf("%.4f,%.4f", p.Lat, p.Lng)
				count := density[key]
				points = append(points, HeatmapPoint{
					Lat: p.Lat, Lng: p.Lng,
					Intensity: float64(count) / float64(maxDensity),
					DeviceMAC: t.DeviceMAC,
				})
			}
		}
	case "traffic_volume":
		for _, e := range pe.entries {
			if e.Location != nil && e.Bytes > 0 {
				intensity := math.Min(1, float64(e.Bytes)/1e6)
				points = append(points, HeatmapPoint{
					Lat: e.Location.Lat, Lng: e.Location.Lng,
					Intensity: intensity, Label: fmt.Sprintf("%dB", e.Bytes),
				})
			}
		}
	}

	return &HeatmapData{
		Points:    points,
		GridSizeM: 5.0,
		MaxSignal: maxSig,
		MinSignal: minSig,
		Generated: time.Now(),
		Category:  category,
	}
}

// --- Movement Tracks ---

func (pe *ParserEngine) ListTracks() []*MovementTrack {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	result := make([]*MovementTrack, 0, len(pe.tracks))
	for _, t := range pe.tracks {
		result = append(result, t)
	}
	return result
}

func (pe *ParserEngine) GetTrack(mac string) (*MovementTrack, bool) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	t, ok := pe.tracks[mac]
	return t, ok
}

// --- Trend Data ---

type TrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label,omitempty"`
}

func (pe *ParserEngine) GetTrend(metric string, hours int) []TrendPoint {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	cutoff := time.Now().Add(-time.Duration(hours) * time.Hour)
	buckets := make(map[int64]float64)

	for _, e := range pe.entries {
		if e.Timestamp.Before(cutoff) {
			continue
		}
		key := e.Timestamp.Truncate(time.Hour).Unix()
		switch metric {
		case "traffic":
			buckets[key] += float64(e.Bytes)
		case "errors":
			if e.Severity == "error" || e.Severity == "critical" {
				buckets[key]++
			}
		case "devices":
			buckets[key]++
		case "connectivity":
			if e.LogType == LogWiFiProbe || e.LogType == LogCSI {
				buckets[key]++
			}
		}
	}

	keys := make([]int64, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	points := make([]TrendPoint, 0, len(keys))
	for _, k := range keys {
		points = append(points, TrendPoint{
			Timestamp: time.Unix(k, 0),
			Value:     buckets[k],
		})
	}
	return points
}

// --- Utility ---

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371e3
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func randMAC() string {
	b := make([]byte, 6)
	rand.Read(b)
	b[0] = (b[0] | 0x02) & 0xfe
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", b[0], b[1], b[2], b[3], b[4], b[5])
}

func randIP() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("10.%d.%d.%d", b[1]%255, b[2]%255, b[3]%254+1)
}

// --- Seed Data ---

func (pe *ParserEngine) seedParsers() {
	now := time.Now()
	configs := []*ParserConfig{
		{ID: "parser-wifi-probe", Name: "Wi-Fi Probe Scanner", LogType: LogWiFiProbe, Description: "Captures 802.11 probe requests for device tracking and movement analysis", Source: ParserSource{Type: "agent", Endpoint: "udp://0.0.0.0:5140"}, Status: "active", CreatedAt: now.Add(-168 * time.Hour), UpdatedAt: now, ParsedCount: 284729, Tags: []string{"wireless", "tracking", "production"}},
		{ID: "parser-ip-flow", Name: "NetFlow v9 Collector", LogType: LogIPFlow, Description: "Collects NetFlow v9 records from core routers", Source: ParserSource{Type: "stream", Endpoint: "udp://0.0.0.0:2055"}, Status: "active", CreatedAt: now.Add(-120 * time.Hour), UpdatedAt: now, ParsedCount: 1523840, Tags: []string{"network", "flows"}},
		{ID: "parser-syslog", Name: "Syslog Receiver", LogType: LogSyslog, Description: "Central syslog aggregation from all network devices", Source: ParserSource{Type: "syslog_receiver", Endpoint: "udp://0.0.0.0:514"}, Status: "active", CreatedAt: now.Add(-240 * time.Hour), UpdatedAt: now, ParsedCount: 945210, ErrorCount: 142, Tags: []string{"infrastructure", "logging"}},
		{ID: "parser-pcap", Name: "Packet Capture Agent", LogType: LogPcap, Description: "Captures raw packets on mirror ports for deep inspection", Source: ParserSource{Type: "agent", Endpoint: "agent://mirror-tap-1"}, Status: "active", CreatedAt: now.Add(-96 * time.Hour), UpdatedAt: now, ParsedCount: 4500000, Tags: []string{"network", "security", "deep-inspection"}},
		{ID: "parser-csi", Name: "WiFi CSI Collector", LogType: LogCSI, Description: "Channel State Information from ESP32 mesh for movement sensing", Source: ParserSource{Type: "stream", Endpoint: "ws://esp32-mesh:8080/csi"}, Status: "active", CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now, ParsedCount: 892100, Tags: []string{"wireless", "csi", "movement"}},
		{ID: "parser-dns", Name: "DNS Query Logger", LogType: LogDNS, Description: "Logs all DNS queries and responses from resolvers", Source: ParserSource{Type: "stream", Endpoint: "udp://dns-resolver:5353"}, Status: "active", CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now, ParsedCount: 2150430, Tags: []string{"network", "dns"}},
		{ID: "parser-firewall", Name: "Firewall Log Parser", LogType: LogFirewall, Description: "Parses firewall accept/deny/drop logs for threat detection", Source: ParserSource{Type: "file", Endpoint: "/var/log/firewall.log"}, Status: "active", CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now, ParsedCount: 780300, ErrorCount: 23, Tags: []string{"security", "firewall"}},
		{ID: "parser-radius", Name: "RADIUS Auth Logger", LogType: LogRadius, Description: "802.1X authentication events from RADIUS server", Source: ParserSource{Type: "syslog_receiver", Endpoint: "udp://radius:1813"}, Status: "paused", CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now, ParsedCount: 45200, Tags: []string{"security", "auth"}},
	}
	for _, c := range configs {
		pe.parsers[c.ID] = c
	}
}

func (pe *ParserEngine) seedEntries() {
	now := time.Now()
	baseLat, baseLng := 23.8103, 90.4125 // Dhaka
	ssids := []string{"AxiomNizam-Corp", "Guest-WiFi", "IoT-Network", "Lab-5GHz", "Eduroam", ""}
	severities := []string{"info", "info", "info", "warn", "error", "info", "debug", "info", "critical", "info"}
	protocols := []string{"TCP", "UDP", "HTTP", "HTTPS", "DNS", "ICMP", "SSH", "TLS", "MQTT", "gRPC"}
	devices := make([]string, 20)
	for i := range devices {
		devices[i] = randMAC()
	}

	for i := 0; i < 500; i++ {
		ts := now.Add(-time.Duration(500-i) * time.Minute)
		devMAC := devices[i%len(devices)]
		lat := baseLat + (float64(i%15)-7)*0.001
		lng := baseLng + (float64(i%12)-6)*0.001

		// Wi-Fi probe entries
		pe.entries = append(pe.entries, ParsedEntry{
			ID: fmt.Sprintf("e-wifi-%d", i), Timestamp: ts, LogType: LogWiFiProbe,
			Source: "AP-Floor-" + fmt.Sprintf("%d", (i%3)+1), Severity: "info",
			Message:   fmt.Sprintf("Probe request from %s for %s", devMAC, ssids[i%len(ssids)]),
			DeviceMAC: devMAC, SSID: ssids[i%len(ssids)],
			Signal: -30 - (i % 50), Location: &GeoPoint{Lat: lat, Lng: lng},
			Fields: map[string]interface{}{"channel": (i%11)*1 + 1, "vendor": "Vendor-" + fmt.Sprintf("%d", i%5)},
		})

		// IP flow entries
		pe.entries = append(pe.entries, ParsedEntry{
			ID: fmt.Sprintf("e-flow-%d", i), Timestamp: ts, LogType: LogIPFlow,
			Source: "router-core-1", Severity: severities[i%len(severities)],
			Message: fmt.Sprintf("Flow %s → %s:%d %s", randIP(), randIP(), 80+(i%9920), protocols[i%len(protocols)]),
			SrcIP:   randIP(), DstIP: randIP(), Protocol: protocols[i%len(protocols)],
			Port: 80 + (i % 9920), Bytes: int64(100 + (i%100)*1024),
			Location: &GeoPoint{Lat: lat + 0.0005, Lng: lng - 0.0003},
		})

		// Syslog
		if i%3 == 0 {
			pe.entries = append(pe.entries, ParsedEntry{
				ID: fmt.Sprintf("e-sys-%d", i), Timestamp: ts, LogType: LogSyslog,
				Source: "switch-core-" + fmt.Sprintf("%d", i%4+1), Severity: severities[i%len(severities)],
				Message: fmt.Sprintf("Interface Gi0/%d link status changed to UP", i%24),
				Fields:  map[string]interface{}{"facility": "local0", "hostname": "sw-core-" + fmt.Sprintf("%d", i%4+1)},
			})
		}

		// CSI
		if i%5 == 0 {
			pe.entries = append(pe.entries, ParsedEntry{
				ID: fmt.Sprintf("e-csi-%d", i), Timestamp: ts, LogType: LogCSI,
				Source: "esp32-node-" + fmt.Sprintf("%d", i%6+1), Severity: "info",
				DeviceMAC: devMAC, Signal: -25 - (i % 40),
				Location: &GeoPoint{Lat: lat - 0.0002, Lng: lng + 0.0004},
				Message:  fmt.Sprintf("CSI frame: 56 subcarriers, amplitude_var=%.3f", 0.01+float64(i%100)*0.005),
				Fields:   map[string]interface{}{"subcarrier_count": 56, "amplitude_variance": 0.01 + float64(i%100)*0.005, "phase_shift": float64(i % 360)},
			})
		}

		// Firewall
		if i%4 == 0 {
			action := "ACCEPT"
			sev := "info"
			if i%12 == 0 {
				action = "DROP"
				sev = "warn"
			}
			if i%20 == 0 {
				action = "DENY"
				sev = "error"
			}
			pe.entries = append(pe.entries, ParsedEntry{
				ID: fmt.Sprintf("e-fw-%d", i), Timestamp: ts, LogType: LogFirewall,
				Source: "fw-edge-1", Severity: sev,
				Message: fmt.Sprintf("%s src=%s dst=%s proto=%s dport=%d", action, randIP(), randIP(), protocols[i%len(protocols)], 22+(i%443)),
				SrcIP:   randIP(), DstIP: randIP(), Protocol: protocols[i%len(protocols)], Port: 22 + (i % 443),
				Fields: map[string]interface{}{"action": action, "rule_id": fmt.Sprintf("rule-%d", i%50+1)},
			})
		}

		// DNS
		if i%2 == 0 {
			domains := []string{"api.axiomnizam.io", "cdn.example.com", "db.internal", "auth.keycloak.local", "grafana.monitoring", "suspicious-domain.xyz"}
			pe.entries = append(pe.entries, ParsedEntry{
				ID: fmt.Sprintf("e-dns-%d", i), Timestamp: ts, LogType: LogDNS,
				Source: "dns-resolver-1", Severity: "info",
				Message: fmt.Sprintf("Query: %s A → %s", domains[i%len(domains)], randIP()),
				SrcIP:   randIP(), Protocol: "DNS",
				Fields: map[string]interface{}{"query": domains[i%len(domains)], "type": "A", "latency_ms": float64(1 + i%50)},
			})
		}
	}
}

func (pe *ParserEngine) seedTracks() {
	now := time.Now()
	baseLat, baseLng := 23.8103, 90.4125
	zones := []string{"Lobby", "Floor-1", "Floor-2", "Server-Room", "Cafeteria", "Lab-A", "Lab-B", "Exit"}

	for d := 0; d < 12; d++ {
		mac := randMAC()
		nPoints := 15 + d*3
		track := &MovementTrack{
			DeviceMAC:  mac,
			DeviceName: fmt.Sprintf("Device-%d", d+1),
			FirstSeen:  now.Add(-time.Duration(nPoints*5) * time.Minute),
			Points:     make([]TrackPoint, 0, nPoints),
			Zones:      make([]string, 0),
		}

		lat, lng := baseLat+float64(d)*0.0003, baseLng-float64(d)*0.0002
		for p := 0; p < nPoints; p++ {
			lat += (float64(p%5) - 2) * 0.0002
			lng += (float64(p%4) - 1.5) * 0.00015
			zone := zones[(d+p)%len(zones)]
			track.Points = append(track.Points, TrackPoint{
				Lat: lat, Lng: lng,
				Timestamp: now.Add(-time.Duration(nPoints-p) * 5 * time.Minute),
				Signal:    -30 - p%35,
				Zone:      zone,
			})
			found := false
			for _, z := range track.Zones {
				if z == zone {
					found = true
					break
				}
			}
			if !found {
				track.Zones = append(track.Zones, zone)
			}
		}
		track.LastSeen = track.Points[len(track.Points)-1].Timestamp

		// Compute distance
		for i := 1; i < len(track.Points); i++ {
			prev := track.Points[i-1]
			cur := track.Points[i]
			track.Distance += haversine(prev.Lat, prev.Lng, cur.Lat, cur.Lng)
		}
		elapsed := track.LastSeen.Sub(track.FirstSeen).Seconds()
		if elapsed > 0 {
			track.AvgSpeed = track.Distance / elapsed
		}

		// Generate predicted path (3 future points)
		if len(track.Points) >= 2 {
			last := track.Points[len(track.Points)-1]
			prev := track.Points[len(track.Points)-2]
			dLat := last.Lat - prev.Lat
			dLng := last.Lng - prev.Lng
			for f := 1; f <= 3; f++ {
				track.Predicted = append(track.Predicted, TrackPoint{
					Lat:       last.Lat + dLat*float64(f),
					Lng:       last.Lng + dLng*float64(f),
					Timestamp: now.Add(time.Duration(f*5) * time.Minute),
					Zone:      "predicted",
				})
			}
		}

		pe.tracks[mac] = track
	}
}
