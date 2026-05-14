// Package distributed provides lightweight health/membership probes for the
// etcd cluster backing the control plane.  Prior revisions shelled out to
// the `etcdctl` binary, which is fragile (depends on an external binary
// being installed, version-matched, and on PATH) and cannot surface
// structured errors.  This implementation uses the native
// go.etcd.io/etcd/client/v3 library already present in go.mod.
package distributed

import (
	"context"
	"fmt"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// DistributedChecker probes the etcd cluster used by the control plane.
// It is safe for concurrent use.
type DistributedChecker struct {
	etcdEndpoints []string
	dialTimeout   time.Duration
	opTimeout     time.Duration
}

// NewDistributedChecker creates a checker targeting the given endpoints.
// If none are supplied, the conventional localhost endpoint is used.
func NewDistributedChecker(endpoints ...string) *DistributedChecker {
	if len(endpoints) == 0 {
		endpoints = []string{"localhost:2379"}
	}
	return &DistributedChecker{
		etcdEndpoints: endpoints,
		dialTimeout:   3 * time.Second,
		opTimeout:     5 * time.Second,
	}
}

// WithDialTimeout overrides the default client dial timeout.
func (dc *DistributedChecker) WithDialTimeout(d time.Duration) *DistributedChecker {
	if d > 0 {
		dc.dialTimeout = d
	}
	return dc
}

// WithOpTimeout overrides the default per-operation timeout.
func (dc *DistributedChecker) WithOpTimeout(d time.Duration) *DistributedChecker {
	if d > 0 {
		dc.opTimeout = d
	}
	return dc
}

// DistributedStatus reports the observed cluster state.
type DistributedStatus struct {
	IsDistributed bool
	Members       []string
	Leader        string
	Healthy       bool
	Error         string
}

func (dc *DistributedChecker) newClient() (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   dc.etcdEndpoints,
		DialTimeout: dc.dialTimeout,
	})
}

// CheckDistributedStatus returns membership, leader and health for the
// configured endpoints.  Errors are surfaced on status.Error rather than
// returned, so callers on the hot health-check path can render the
// partial result.
func (dc *DistributedChecker) CheckDistributedStatus() *DistributedStatus {
	status := &DistributedStatus{Members: []string{}}

	cli, err := dc.newClient()
	if err != nil {
		status.Error = fmt.Sprintf("etcd dial failed: %v", err)
		return status
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), dc.opTimeout)
	defer cancel()

	memberResp, err := cli.MemberList(ctx)
	if err != nil {
		status.Error = fmt.Sprintf("MemberList failed: %v", err)
		return status
	}
	for _, m := range memberResp.Members {
		// Preserve a comma-delimited textual form close to the prior
		// `etcdctl member list` output for dashboards that parse it.
		status.Members = append(status.Members, fmt.Sprintf(
			"%x, started, %s, %s, %s",
			m.ID, m.Name,
			strings.Join(m.PeerURLs, " "),
			strings.Join(m.ClientURLs, " "),
		))
	}
	status.IsDistributed = len(status.Members) > 0

	// A successful round-trip for a trivial Get is sufficient evidence
	// of a healthy quorum on the contacted endpoint.
	healthCtx, healthCancel := context.WithTimeout(context.Background(), dc.opTimeout)
	defer healthCancel()
	if _, err := cli.Get(healthCtx, "health-probe"); err == nil {
		status.Healthy = true
	} else if status.Error == "" {
		status.Error = fmt.Sprintf("health probe failed: %v", err)
	}

	// Identify the current leader via per-endpoint Status, then map the
	// returned member-id back to a human-readable name.
	for _, ep := range dc.etcdEndpoints {
		sCtx, sCancel := context.WithTimeout(context.Background(), dc.opTimeout)
		sResp, err := cli.Status(sCtx, ep)
		sCancel()
		if err != nil || sResp.Leader == 0 {
			continue
		}
		for _, m := range memberResp.Members {
			if m.ID == sResp.Leader {
				status.Leader = m.Name
				break
			}
		}
		if status.Leader != "" {
			break
		}
	}

	return status
}

// GetClusterInfo returns a structured map describing the etcd cluster,
// suitable for JSON rendering by admin APIs.
func (dc *DistributedChecker) GetClusterInfo() map[string]interface{} {
	info := map[string]interface{}{"endpoints": dc.etcdEndpoints}

	cli, err := dc.newClient()
	if err != nil {
		info["error"] = fmt.Sprintf("etcd dial failed: %v", err)
		return info
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), dc.opTimeout)
	defer cancel()

	memberResp, err := cli.MemberList(ctx)
	if err != nil {
		info["error"] = fmt.Sprintf("MemberList failed: %v", err)
		return info
	}

	members := make([]map[string]interface{}, 0, len(memberResp.Members))
	for _, m := range memberResp.Members {
		members = append(members, map[string]interface{}{
			"id":         fmt.Sprintf("%x", m.ID),
			"name":       m.Name,
			"peerURLs":   m.PeerURLs,
			"clientURLs": m.ClientURLs,
			"isLearner":  m.IsLearner,
		})
	}
	info["member_count"] = len(members)
	info["members"] = members

	health := make([]map[string]interface{}, 0, len(dc.etcdEndpoints))
	for _, ep := range dc.etcdEndpoints {
		sCtx, sCancel := context.WithTimeout(context.Background(), dc.opTimeout)
		sResp, err := cli.Status(sCtx, ep)
		sCancel()
		entry := map[string]interface{}{"endpoint": ep}
		if err != nil {
			entry["healthy"] = false
			entry["error"] = err.Error()
		} else {
			entry["healthy"] = true
			entry["version"] = sResp.Version
			entry["dbSize"] = sResp.DbSize
			entry["leader"] = fmt.Sprintf("%x", sResp.Leader)
			entry["raftIndex"] = sResp.RaftIndex
			entry["raftTerm"] = sResp.RaftTerm
		}
		health = append(health, entry)
	}
	info["health"] = health

	return info
}
