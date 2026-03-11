package netintel

import (
	"fmt"
	"sync"
	"time"
)

// =====================================================
// Network Intelligence — Topology Graph Engine
// Network topology inference and graph representation
// =====================================================

type NodeType string

const (
	NodeRouter   NodeType = "router"
	NodeSwitch   NodeType = "switch"
	NodeAP       NodeType = "access_point"
	NodeFirewall NodeType = "firewall"
	NodeServer   NodeType = "server"
	NodeDevice   NodeType = "device"
	NodeSensor   NodeType = "sensor"
	NodeCloud    NodeType = "cloud"
	NodeGateway  NodeType = "gateway"
)

type TopologyNode struct {
	ID       string                 `json:"id"`
	Label    string                 `json:"label"`
	Type     NodeType               `json:"type"`
	IP       string                 `json:"ip,omitempty"`
	MAC      string                 `json:"mac,omitempty"`
	Status   string                 `json:"status"` // online, offline, degraded, unknown
	Location *GeoPoint              `json:"location,omitempty"`
	Metrics  map[string]interface{} `json:"metrics,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
	X        float64                `json:"x"` // layout x
	Y        float64                `json:"y"` // layout y
}

type TopologyEdge struct {
	ID        string  `json:"id"`
	Source    string  `json:"source"` // node ID
	Target    string  `json:"target"` // node ID
	Label     string  `json:"label,omitempty"`
	Bandwidth string  `json:"bandwidth,omitempty"` // e.g. "10Gbps"
	Latency   float64 `json:"latency_ms,omitempty"`
	Status    string  `json:"status"` // active, degraded, down
	Traffic   float64 `json:"traffic_mbps,omitempty"`
	Protocol  string  `json:"protocol,omitempty"`
}

type TopologyGraph struct {
	Nodes     []TopologyNode `json:"nodes"`
	Edges     []TopologyEdge `json:"edges"`
	Generated time.Time      `json:"generated_at"`
	Stats     TopologyStats  `json:"stats"`
}

type TopologyStats struct {
	TotalNodes    int            `json:"total_nodes"`
	TotalEdges    int            `json:"total_edges"`
	OnlineNodes   int            `json:"online_nodes"`
	OfflineNodes  int            `json:"offline_nodes"`
	DegradedEdges int            `json:"degraded_edges"`
	ByType        map[string]int `json:"by_type"`
}

type TopologyEngine struct {
	mu    sync.RWMutex
	nodes map[string]*TopologyNode
	edges map[string]*TopologyEdge
}

func NewTopologyEngine() *TopologyEngine {
	te := &TopologyEngine{
		nodes: make(map[string]*TopologyNode),
		edges: make(map[string]*TopologyEdge),
	}
	te.seedTopology()
	return te
}

func (te *TopologyEngine) GetGraph() *TopologyGraph {
	te.mu.RLock()
	defer te.mu.RUnlock()

	nodes := make([]TopologyNode, 0, len(te.nodes))
	edges := make([]TopologyEdge, 0, len(te.edges))
	stats := TopologyStats{ByType: make(map[string]int)}

	for _, n := range te.nodes {
		nodes = append(nodes, *n)
		stats.TotalNodes++
		stats.ByType[string(n.Type)]++
		switch n.Status {
		case "online":
			stats.OnlineNodes++
		case "offline":
			stats.OfflineNodes++
		}
	}
	for _, e := range te.edges {
		edges = append(edges, *e)
		stats.TotalEdges++
		if e.Status == "degraded" {
			stats.DegradedEdges++
		}
	}

	return &TopologyGraph{
		Nodes:     nodes,
		Edges:     edges,
		Generated: time.Now(),
		Stats:     stats,
	}
}

func (te *TopologyEngine) GetNode(id string) (*TopologyNode, bool) {
	te.mu.RLock()
	defer te.mu.RUnlock()
	n, ok := te.nodes[id]
	return n, ok
}

func (te *TopologyEngine) UpdateNodeStatus(id, status string) error {
	te.mu.Lock()
	defer te.mu.Unlock()
	n, ok := te.nodes[id]
	if !ok {
		return fmt.Errorf("node not found: %s", id)
	}
	n.Status = status
	return nil
}

func (te *TopologyEngine) seedTopology() {
	// Core infrastructure
	te.nodes["cloud-gw"] = &TopologyNode{ID: "cloud-gw", Label: "Cloud Gateway", Type: NodeCloud, IP: "203.0.113.1", Status: "online", X: 400, Y: 30, Metrics: map[string]interface{}{"uptime_hours": 2160, "region": "ap-southeast-1"}, Tags: []string{"cloud", "wan"}}
	te.nodes["fw-edge-1"] = &TopologyNode{ID: "fw-edge-1", Label: "Edge Firewall", Type: NodeFirewall, IP: "10.0.0.1", Status: "online", X: 400, Y: 110, Metrics: map[string]interface{}{"rules_active": 847, "throughput_mbps": 940, "connections": 12450}, Tags: []string{"security", "perimeter"}}
	te.nodes["router-core-1"] = &TopologyNode{ID: "router-core-1", Label: "Core Router", Type: NodeRouter, IP: "10.0.0.2", Status: "online", X: 400, Y: 190, Metrics: map[string]interface{}{"interfaces": 24, "bgp_peers": 3, "routes": 4521, "cpu_pct": 34}, Tags: []string{"core", "routing"}}

	// Distribution switches
	te.nodes["sw-dist-1"] = &TopologyNode{ID: "sw-dist-1", Label: "Dist Switch 1", Type: NodeSwitch, IP: "10.0.1.1", Status: "online", X: 200, Y: 280, Metrics: map[string]interface{}{"ports": 48, "vlans": 12, "stp_root": true}, Tags: []string{"distribution", "layer2"}}
	te.nodes["sw-dist-2"] = &TopologyNode{ID: "sw-dist-2", Label: "Dist Switch 2", Type: NodeSwitch, IP: "10.0.2.1", Status: "online", X: 600, Y: 280, Metrics: map[string]interface{}{"ports": 48, "vlans": 12, "stp_root": false}, Tags: []string{"distribution", "layer2"}}

	// Access points
	te.nodes["ap-floor-1"] = &TopologyNode{ID: "ap-floor-1", Label: "AP Floor 1", Type: NodeAP, IP: "10.0.1.10", Status: "online", X: 80, Y: 380, Metrics: map[string]interface{}{"clients": 32, "channel": 6, "tx_power_dbm": 20, "band": "2.4GHz"}, Tags: []string{"wireless", "floor-1"}}
	te.nodes["ap-floor-2"] = &TopologyNode{ID: "ap-floor-2", Label: "AP Floor 2", Type: NodeAP, IP: "10.0.1.11", Status: "online", X: 200, Y: 380, Metrics: map[string]interface{}{"clients": 28, "channel": 36, "tx_power_dbm": 23, "band": "5GHz"}, Tags: []string{"wireless", "floor-2"}}
	te.nodes["ap-floor-3"] = &TopologyNode{ID: "ap-floor-3", Label: "AP Floor 3", Type: NodeAP, IP: "10.0.1.12", Status: "degraded", X: 320, Y: 380, Metrics: map[string]interface{}{"clients": 8, "channel": 44, "tx_power_dbm": 20, "band": "5GHz", "error_rate": 2.3}, Tags: []string{"wireless", "floor-3", "degraded"}}

	// Servers
	te.nodes["srv-app-1"] = &TopologyNode{ID: "srv-app-1", Label: "App Server 1", Type: NodeServer, IP: "10.0.2.10", Status: "online", X: 480, Y: 380, Metrics: map[string]interface{}{"cpu_pct": 45, "memory_pct": 62, "connections": 340}, Tags: []string{"compute", "app"}}
	te.nodes["srv-db-1"] = &TopologyNode{ID: "srv-db-1", Label: "Database Server", Type: NodeServer, IP: "10.0.2.11", Status: "online", X: 600, Y: 380, Metrics: map[string]interface{}{"cpu_pct": 28, "memory_pct": 78, "iops": 4500, "replication_lag_ms": 2}, Tags: []string{"compute", "database"}}
	te.nodes["srv-dns-1"] = &TopologyNode{ID: "srv-dns-1", Label: "DNS Resolver", Type: NodeServer, IP: "10.0.2.12", Status: "online", X: 720, Y: 380, Metrics: map[string]interface{}{"queries_sec": 450, "cache_hit_pct": 89}, Tags: []string{"infrastructure", "dns"}}

	// ESP32 CSI sensors
	te.nodes["esp32-1"] = &TopologyNode{ID: "esp32-1", Label: "CSI Sensor 1", Type: NodeSensor, IP: "10.0.3.1", MAC: "aa:bb:cc:01:01:01", Status: "online", X: 80, Y: 480, Metrics: map[string]interface{}{"subcarriers": 56, "sample_rate_hz": 100, "battery_pct": 92}, Tags: []string{"csi", "sensor", "floor-1"}}
	te.nodes["esp32-2"] = &TopologyNode{ID: "esp32-2", Label: "CSI Sensor 2", Type: NodeSensor, IP: "10.0.3.2", MAC: "aa:bb:cc:01:01:02", Status: "online", X: 200, Y: 480, Metrics: map[string]interface{}{"subcarriers": 56, "sample_rate_hz": 100, "battery_pct": 78}, Tags: []string{"csi", "sensor", "floor-1"}}
	te.nodes["esp32-3"] = &TopologyNode{ID: "esp32-3", Label: "CSI Sensor 3", Type: NodeSensor, IP: "10.0.3.3", MAC: "aa:bb:cc:01:01:03", Status: "online", X: 320, Y: 480, Metrics: map[string]interface{}{"subcarriers": 56, "sample_rate_hz": 100, "battery_pct": 65}, Tags: []string{"csi", "sensor", "floor-2"}}
	te.nodes["esp32-4"] = &TopologyNode{ID: "esp32-4", Label: "CSI Sensor 4", Type: NodeSensor, IP: "10.0.3.4", MAC: "aa:bb:cc:01:01:04", Status: "offline", X: 440, Y: 480, Metrics: map[string]interface{}{"subcarriers": 56, "sample_rate_hz": 0, "battery_pct": 0, "last_seen": "12 minutes ago"}, Tags: []string{"csi", "sensor", "floor-2", "offline"}}

	// Gateway
	te.nodes["gw-iot"] = &TopologyNode{ID: "gw-iot", Label: "IoT Gateway", Type: NodeGateway, IP: "10.0.3.254", Status: "online", X: 260, Y: 430, Metrics: map[string]interface{}{"connected_sensors": 3, "mqtt_topics": 12, "buffer_pct": 15}, Tags: []string{"iot", "gateway"}}

	// Client devices
	te.nodes["dev-laptop-1"] = &TopologyNode{ID: "dev-laptop-1", Label: "Admin Laptop", Type: NodeDevice, IP: "10.0.1.101", Status: "online", X: 80, Y: 560, Metrics: map[string]interface{}{"signal_dbm": -42, "ap": "ap-floor-1"}, Tags: []string{"client", "wireless"}}
	te.nodes["dev-phone-1"] = &TopologyNode{ID: "dev-phone-1", Label: "Mobile Device", Type: NodeDevice, IP: "10.0.1.102", Status: "online", X: 200, Y: 560, Metrics: map[string]interface{}{"signal_dbm": -55, "ap": "ap-floor-2"}, Tags: []string{"client", "wireless", "mobile"}}

	// --- Edges ---
	te.edges["e-cloud-fw"] = &TopologyEdge{ID: "e-cloud-fw", Source: "cloud-gw", Target: "fw-edge-1", Label: "WAN", Bandwidth: "1Gbps", Latency: 12.5, Status: "active", Traffic: 340, Protocol: "IPSec"}
	te.edges["e-fw-router"] = &TopologyEdge{ID: "e-fw-router", Source: "fw-edge-1", Target: "router-core-1", Label: "DMZ→Core", Bandwidth: "10Gbps", Latency: 0.3, Status: "active", Traffic: 520, Protocol: "Ethernet"}
	te.edges["e-router-sw1"] = &TopologyEdge{ID: "e-router-sw1", Source: "router-core-1", Target: "sw-dist-1", Label: "Core→Dist1", Bandwidth: "10Gbps", Latency: 0.1, Status: "active", Traffic: 280, Protocol: "LACP"}
	te.edges["e-router-sw2"] = &TopologyEdge{ID: "e-router-sw2", Source: "router-core-1", Target: "sw-dist-2", Label: "Core→Dist2", Bandwidth: "10Gbps", Latency: 0.1, Status: "active", Traffic: 240, Protocol: "LACP"}

	te.edges["e-sw1-ap1"] = &TopologyEdge{ID: "e-sw1-ap1", Source: "sw-dist-1", Target: "ap-floor-1", Label: "PoE", Bandwidth: "1Gbps", Latency: 0.2, Status: "active", Traffic: 85, Protocol: "802.3af"}
	te.edges["e-sw1-ap2"] = &TopologyEdge{ID: "e-sw1-ap2", Source: "sw-dist-1", Target: "ap-floor-2", Label: "PoE", Bandwidth: "1Gbps", Latency: 0.2, Status: "active", Traffic: 72, Protocol: "802.3af"}
	te.edges["e-sw1-ap3"] = &TopologyEdge{ID: "e-sw1-ap3", Source: "sw-dist-1", Target: "ap-floor-3", Label: "PoE", Bandwidth: "1Gbps", Latency: 1.8, Status: "degraded", Traffic: 15, Protocol: "802.3af"}

	te.edges["e-sw2-app"] = &TopologyEdge{ID: "e-sw2-app", Source: "sw-dist-2", Target: "srv-app-1", Label: "10G", Bandwidth: "10Gbps", Latency: 0.05, Status: "active", Traffic: 180, Protocol: "Ethernet"}
	te.edges["e-sw2-db"] = &TopologyEdge{ID: "e-sw2-db", Source: "sw-dist-2", Target: "srv-db-1", Label: "10G", Bandwidth: "10Gbps", Latency: 0.05, Status: "active", Traffic: 95, Protocol: "Ethernet"}
	te.edges["e-sw2-dns"] = &TopologyEdge{ID: "e-sw2-dns", Source: "sw-dist-2", Target: "srv-dns-1", Label: "1G", Bandwidth: "1Gbps", Latency: 0.1, Status: "active", Traffic: 12, Protocol: "Ethernet"}

	te.edges["e-sw1-gw"] = &TopologyEdge{ID: "e-sw1-gw", Source: "sw-dist-1", Target: "gw-iot", Label: "IoT VLAN", Bandwidth: "1Gbps", Latency: 0.3, Status: "active", Traffic: 5, Protocol: "802.1Q"}
	te.edges["e-gw-esp1"] = &TopologyEdge{ID: "e-gw-esp1", Source: "gw-iot", Target: "esp32-1", Label: "MQTT", Bandwidth: "11Mbps", Latency: 2.1, Status: "active", Traffic: 0.8, Protocol: "WiFi/MQTT"}
	te.edges["e-gw-esp2"] = &TopologyEdge{ID: "e-gw-esp2", Source: "gw-iot", Target: "esp32-2", Label: "MQTT", Bandwidth: "11Mbps", Latency: 1.8, Status: "active", Traffic: 0.7, Protocol: "WiFi/MQTT"}
	te.edges["e-gw-esp3"] = &TopologyEdge{ID: "e-gw-esp3", Source: "gw-iot", Target: "esp32-3", Label: "MQTT", Bandwidth: "11Mbps", Latency: 2.5, Status: "active", Traffic: 0.9, Protocol: "WiFi/MQTT"}
	te.edges["e-gw-esp4"] = &TopologyEdge{ID: "e-gw-esp4", Source: "gw-iot", Target: "esp32-4", Label: "MQTT", Bandwidth: "11Mbps", Latency: 0, Status: "down", Traffic: 0, Protocol: "WiFi/MQTT"}

	te.edges["e-ap1-laptop"] = &TopologyEdge{ID: "e-ap1-laptop", Source: "ap-floor-1", Target: "dev-laptop-1", Label: "WiFi", Bandwidth: "866Mbps", Latency: 3.2, Status: "active", Traffic: 22, Protocol: "802.11ac"}
	te.edges["e-ap2-phone"] = &TopologyEdge{ID: "e-ap2-phone", Source: "ap-floor-2", Target: "dev-phone-1", Label: "WiFi", Bandwidth: "433Mbps", Latency: 5.1, Status: "active", Traffic: 8, Protocol: "802.11ac"}
}
