package distributed

import (
	"fmt"
	"os/exec"
	"strings"
)

type DistributedChecker struct {
	etcdEndpoints []string
}

func NewDistributedChecker(endpoints ...string) *DistributedChecker {
	if len(endpoints) == 0 {
		endpoints = []string{"localhost:2379"}
	}
	return &DistributedChecker{
		etcdEndpoints: endpoints,
	}
}

type DistributedStatus struct {
	IsDistributed bool
	Members       []string
	Leader        string
	Healthy       bool
	Error         string
}

func (dc *DistributedChecker) CheckDistributedStatus() *DistributedStatus {
	status := &DistributedStatus{
		IsDistributed: false,
		Members:       []string{},
	}

	cmd := exec.Command("etcdctl", "--endpoints="+strings.Join(dc.etcdEndpoints, ","), "member", "list")
	output, err := cmd.CombinedOutput()

	if err != nil {
		status.Error = fmt.Sprintf("Failed to check etcd status: %v", err)
		return status
	}

	members := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(members) > 0 && members[0] != "" {
		status.IsDistributed = true
		status.Members = members
	}

	leaderCmd := exec.Command("etcdctl", "--endpoints="+strings.Join(dc.etcdEndpoints, ","), "endpoint", "health")
	leaderOutput, leaderErr := leaderCmd.CombinedOutput()

	if leaderErr == nil {
		status.Healthy = true
		status.Leader = strings.TrimSpace(string(leaderOutput))
	}

	return status
}

func (dc *DistributedChecker) GetClusterInfo() map[string]interface{} {
	info := make(map[string]interface{})

	statusCmd := exec.Command("etcdctl", "--endpoints="+strings.Join(dc.etcdEndpoints, ","), "member", "list")
	statusOutput, statusErr := statusCmd.CombinedOutput()

	if statusErr != nil {
		info["error"] = "Failed to get cluster info"
		info["endpoints"] = dc.etcdEndpoints
		return info
	}

	members := []string{}
	for _, line := range strings.Split(strings.TrimSpace(string(statusOutput)), "\n") {
		if line != "" {
			members = append(members, line)
		}
	}

	info["member_count"] = len(members)
	info["members"] = members
	info["endpoints"] = dc.etcdEndpoints

	healthCmd := exec.Command("etcdctl", "--endpoints="+strings.Join(dc.etcdEndpoints, ","), "endpoint", "health")
	healthOutput, _ := healthCmd.CombinedOutput()
	info["health"] = strings.TrimSpace(string(healthOutput))

	return info
}
