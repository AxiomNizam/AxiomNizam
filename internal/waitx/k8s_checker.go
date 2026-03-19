package waitx

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// KubernetesPodReadinessChecker waits for pod readiness via kubectl.
type KubernetesPodReadinessChecker struct {
	KubectlBinary string
	PodName       string
	LabelSelector string
	Namespace     string
	Kubeconfig    string
	Context       string
	MinReady      int
}

func (c KubernetesPodReadinessChecker) Name() string {
	if strings.TrimSpace(c.PodName) != "" {
		return "k8s-pod:" + strings.TrimSpace(c.PodName)
	}
	return "k8s-selector:" + strings.TrimSpace(c.LabelSelector)
}

func (c KubernetesPodReadinessChecker) Check(ctx context.Context) error {
	podName := strings.TrimSpace(c.PodName)
	selector := strings.TrimSpace(c.LabelSelector)
	if podName == "" && selector == "" {
		return fmt.Errorf("either pod name or label selector is required")
	}

	kubectl := strings.TrimSpace(c.KubectlBinary)
	if kubectl == "" {
		kubectl = "kubectl"
	}

	namespace := strings.TrimSpace(c.Namespace)
	if namespace == "" {
		namespace = "default"
	}

	args := make([]string, 0, 16)
	if podName != "" {
		args = append(args, "get", "pod", podName)
	} else {
		args = append(args, "get", "pods", "-l", selector)
	}
	args = append(args, "-n", namespace, "-o", "json")

	if v := strings.TrimSpace(c.Kubeconfig); v != "" {
		args = append(args, "--kubeconfig", v)
	}
	if v := strings.TrimSpace(c.Context); v != "" {
		args = append(args, "--context", v)
	}

	cmd := exec.CommandContext(ctx, kubectl, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl get pod failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	readyCount, totalCount, podNames, parseErr := parsePodReadinessData(out)
	if parseErr != nil {
		return parseErr
	}

	requiredReady := c.MinReady
	if requiredReady <= 0 {
		requiredReady = 1
	}
	if totalCount == 0 {
		return fmt.Errorf("no pods returned for readiness check")
	}
	if readyCount < requiredReady {
		return fmt.Errorf("ready pods %d/%d (required %d): %s", readyCount, totalCount, requiredReady, strings.Join(podNames, ","))
	}

	return nil
}

type podCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

type podStatus struct {
	Phase      string         `json:"phase"`
	Conditions []podCondition `json:"conditions"`
}

type podMetadata struct {
	Name string `json:"name"`
}

type podObject struct {
	Metadata podMetadata `json:"metadata"`
	Status   podStatus   `json:"status"`
}

type podList struct {
	Items []podObject `json:"items"`
}

func parsePodReadinessData(raw []byte) (readyCount int, totalCount int, podNames []string, err error) {
	pods := make([]podObject, 0)

	var list podList
	if unmarshalErr := json.Unmarshal(raw, &list); unmarshalErr == nil && len(list.Items) > 0 {
		pods = append(pods, list.Items...)
	} else {
		var single podObject
		if singleErr := json.Unmarshal(raw, &single); singleErr != nil {
			return 0, 0, nil, fmt.Errorf("failed to parse kubectl pod JSON: %w", singleErr)
		}
		if strings.TrimSpace(single.Metadata.Name) == "" {
			return 0, 0, nil, fmt.Errorf("kubectl output did not include pod metadata")
		}
		pods = append(pods, single)
	}

	totalCount = len(pods)
	podNames = make([]string, 0, len(pods))

	for _, pod := range pods {
		podNames = append(podNames, strings.TrimSpace(pod.Metadata.Name))
		if isPodReady(pod) {
			readyCount++
		}
	}

	return readyCount, totalCount, podNames, nil
}

func isPodReady(pod podObject) bool {
	if strings.ToLower(strings.TrimSpace(pod.Status.Phase)) != "running" {
		return false
	}
	for _, cond := range pod.Status.Conditions {
		if strings.EqualFold(strings.TrimSpace(cond.Type), "ready") && strings.EqualFold(strings.TrimSpace(cond.Status), "true") {
			return true
		}
	}
	return false
}
