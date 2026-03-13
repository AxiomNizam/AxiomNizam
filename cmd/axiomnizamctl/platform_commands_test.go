package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"example.com/axiomnizam/internal/client"
)

func setupPlatformCommandTestContext(t *testing.T, serverURL string) {
	t.Helper()

	cfgPath := t.TempDir() + "/config.yaml"
	cm, err := client.NewConfigManagerWithPath(cfgPath)
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}

	err = cm.AddOrUpdateContext(&client.Context{
		Name: "test",
		Cluster: &client.ClusterInfo{
			Server: serverURL,
		},
		User:      "tester",
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("failed to add context: %v", err)
	}
	if err := cm.SetCurrentContext("test"); err != nil {
		t.Fatalf("failed to set context: %v", err)
	}

	configManager = cm
	apiClient = client.NewClient(serverURL)
}

func TestPlatformCLICommandsHitExpectedEndpoints(t *testing.T) {
	var (
		mu    sync.Mutex
		paths []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.RequestURI())
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "path": r.URL.RequestURI()})
	}))
	defer server.Close()

	setupPlatformCommandTestContext(t, server.URL)

	tests := []struct {
		name     string
		run      func() error
		expected string
	}{
		{
			name:     "tenant list",
			run:      func() error { return tenantListCmd.RunE(tenantListCmd, []string{}) },
			expected: "/api/v1/tenants",
		},
		{
			name:     "webhook list",
			run:      func() error { return webhookListCmd.RunE(webhookListCmd, []string{}) },
			expected: "/api/v1/webhooks",
		},
		{
			name:     "stream list",
			run:      func() error { return streamListCmd.RunE(streamListCmd, []string{}) },
			expected: "/api/v1/streams",
		},
		{
			name:     "export list",
			run:      func() error { return exportListCmd.RunE(exportListCmd, []string{}) },
			expected: "/api/v1/exports",
		},
		{
			name:     "bulk list",
			run:      func() error { return bulkListCmd.RunE(bulkListCmd, []string{}) },
			expected: "/api/v1/bulk/operations",
		},
		{
			name:     "version history",
			run:      func() error { return versionHistoryCmd.RunE(versionHistoryCmd, []string{"apis", "orders"}) },
			expected: "/api/v1/versioning/history/apis/orders",
		},
		{
			name:     "trace get",
			run:      func() error { return traceGetCmd.RunE(traceGetCmd, []string{"trace-1"}) },
			expected: "/api/v1/tracing/traces/trace-1",
		},
		{
			name:     "lineage graph",
			run:      func() error { return lineageGraphCmd.RunE(lineageGraphCmd, []string{"apis", "orders"}) },
			expected: "/api/v1/lineage/apis/orders",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mu.Lock()
			before := len(paths)
			mu.Unlock()

			if err := tc.run(); err != nil {
				t.Fatalf("command failed: %v", err)
			}

			mu.Lock()
			defer mu.Unlock()
			if len(paths) != before+1 {
				t.Fatalf("expected one request, got before=%d after=%d", before, len(paths))
			}
			actual := paths[len(paths)-1]
			if actual != tc.expected {
				t.Fatalf("expected path %s, got %s", tc.expected, actual)
			}
		})
	}
}
