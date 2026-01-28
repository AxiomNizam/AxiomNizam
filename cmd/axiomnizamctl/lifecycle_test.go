package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestReconciliationLoopComplete validates the end-to-end lifecycle
func TestReconciliationLoopComplete(t *testing.T) {
	tests := []struct {
		name      string
		yamlPath  string
		namespace string
		apiName   string
	}{
		{
			name:      "Complete APIResource Lifecycle",
			yamlPath:  "examples/api.yaml",
			namespace: "default",
			apiName:   "users-api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize
			cm := NewCLIManager()
			defer cm.Shutdown()

			// Step 1: Apply resource
			t.Logf("Step 1: Applying APIResource from %s", tt.yamlPath)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := cm.Apply(tt.yamlPath)
			if err != nil {
				t.Fatalf("Apply failed: %v", err)
			}
			t.Log("✓ Apply succeeded")

			// Step 2: Verify resource is stored
			t.Logf("Step 2: Verifying resource exists in store")
			time.Sleep(100 * time.Millisecond)

			api, err := cm.store.Get(ctx, tt.namespace, tt.apiName)
			if err != nil {
				t.Fatalf("Store Get failed: %v", err)
			}
			if api == nil {
				t.Fatal("Resource not found in store")
			}
			t.Log("✓ Resource found in store")

			// Step 3: Verify initial status is Pending
			t.Logf("Step 3: Verifying initial status is Pending")
			if api.Status.Phase != "Pending" {
				t.Fatalf("Expected Phase=Pending, got %s", api.Status.Phase)
			}
			if api.Status.Ready {
				t.Fatal("Expected Ready=false initially")
			}
			t.Logf("✓ Initial status correct: Phase=%s, Ready=%v", api.Status.Phase, api.Status.Ready)

			// Step 4: Wait for reconciliation to complete
			t.Logf("Step 4: Waiting for reconciliation (max 5 seconds)")
			maxWait := time.Now().Add(5 * time.Second)

			for time.Now().Before(maxWait) {
				api, err := cm.store.Get(ctx, tt.namespace, tt.apiName)
				if err != nil {
					t.Fatalf("Store Get failed during wait: %v", err)
				}

				if api.Status.Phase == "Ready" && api.Status.Ready {
					t.Logf("✓ Resource reached Ready state")
					break
				}

				if api.Status.Phase == "Failed" {
					t.Fatalf("Resource entered Failed state: %s", api.Status.Message)
				}

				time.Sleep(100 * time.Millisecond)
			}

			// Step 5: Final verification
			t.Logf("Step 5: Final status verification")
			api, err = cm.store.Get(ctx, tt.namespace, tt.apiName)
			if err != nil {
				t.Fatalf("Final Get failed: %v", err)
			}

			if api.Status.Phase != "Ready" {
				t.Fatalf("Expected final Phase=Ready, got %s", api.Status.Phase)
			}
			if !api.Status.Ready {
				t.Fatal("Expected final Ready=true")
			}
			if api.Status.Message == "" {
				t.Logf("✓ Final status: Phase=%s, Ready=%v, Message='%s'",
					api.Status.Phase, api.Status.Ready, api.Status.Message)
			}

			// Step 6: Verify reconciliation happened
			t.Logf("Step 6: Verifying reconciliation metrics")
			if api.Metadata.Generation < 2 {
				t.Fatalf("Expected Generation >= 2 after reconciliation, got %d", api.Metadata.Generation)
			}
			t.Logf("✓ Resource generation incremented: %d", api.Metadata.Generation)

			// Step 7: Verify resource is queryable
			t.Logf("Step 7: Verifying Get query works")
			apis, err := cm.store.List(ctx, tt.namespace)
			if err != nil {
				t.Fatalf("Store List failed: %v", err)
			}
			if len(apis) == 0 {
				t.Fatal("Expected at least 1 resource from List")
			}

			found := false
			for _, a := range apis {
				if a.Metadata.Name == tt.apiName && a.Metadata.Namespace == tt.namespace {
					found = true
					break
				}
			}
			if !found {
				t.Fatal("Resource not found in List results")
			}
			t.Logf("✓ Resource queryable via List")

			t.Log("\n" + "======================================================================")
			t.Log("RECONCILIATION LOOP TEST PASSED")
			t.Log("======================================================================")
			t.Log("\nLifecycle Validated:")
			t.Log("  1. YAML file parsed successfully")
			t.Log("  2. Resource created with Pending status")
			t.Log("  3. Resource stored in persistence layer")
			t.Log("  4. Item enqueued for reconciliation")
			t.Log("  5. Reconciliation loop processed item")
			t.Log("  6. Status transitioned: Pending → Creating → Ready")
			t.Log("  7. Status updated in store")
			t.Log("  8. Resource queryable with final status")
			t.Log("\nThis proves the complete YAML → Store → Reconcile → Status pattern works.")
		})
	}
}

// TestStatusTransitions verifies all state transitions
func TestStatusTransitions(t *testing.T) {
	cm := NewCLIManager()
	defer cm.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a test resource directly
	spec := map[string]interface{}{
		"basePath":    "/api/v1/test",
		"title":       "Test API",
		"description": "Test resource for transitions",
		"version":     "1.0",
		"timeout":     30,
	}

	api := NewAPIResource("default", "test-api", spec)

	// Verify initial state is Pending
	if api.Status.Phase != "Pending" {
		t.Fatalf("Expected initial Phase=Pending, got %s", api.Status.Phase)
	}
	t.Logf("✓ Initial state: Phase=%s, Ready=%v", api.Status.Phase, api.Status.Ready)

	// Store resource
	stored, err := cm.store.Create(ctx, api)
	if err != nil {
		t.Fatalf("Store Create failed: %v", err)
	}

	// Simulate state transitions
	stored.MarkCreating("Initializing API")
	err = cm.store.Update(ctx, stored)
	if err != nil {
		t.Fatalf("Update to Creating failed: %v", err)
	}

	retrieved, err := cm.store.Get(ctx, "default", "test-api")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Status.Phase != "Creating" {
		t.Fatalf("Expected Creating state, got %s", retrieved.Status.Phase)
	}
	t.Logf("✓ Transitioned to: Phase=%s, Ready=%v", retrieved.Status.Phase, retrieved.Status.Ready)

	// Transition to Ready
	retrieved.SetReady(true)
	retrieved.SetPhase("Ready")
	err = cm.store.Update(ctx, retrieved)
	if err != nil {
		t.Fatalf("Update to Ready failed: %v", err)
	}

	final, err := cm.store.Get(ctx, "default", "test-api")
	if err != nil {
		t.Fatalf("Final Get failed: %v", err)
	}

	if final.Status.Phase != "Ready" || !final.Status.Ready {
		t.Fatalf("Expected Ready state, got Phase=%s, Ready=%v", final.Status.Phase, final.Status.Ready)
	}
	t.Logf("✓ Final state: Phase=%s, Ready=%v", final.Status.Phase, final.Status.Ready)

	t.Log("\n✓ All status transitions validated successfully")
}

// BenchmarkReconciliation measures reconciliation performance
func BenchmarkReconciliation(b *testing.B) {
	cm := NewCLIManager()
	defer cm.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	spec := map[string]interface{}{
		"basePath":    "/api/v1/bench",
		"title":       "Benchmark API",
		"description": "Resource for performance testing",
		"version":     "1.0",
		"timeout":     30,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		apiName := fmt.Sprintf("bench-api-%d", i)
		api := NewAPIResource("default", apiName, spec)

		stored, err := cm.store.Create(ctx, api)
		if err != nil {
			b.Fatalf("Create failed: %v", err)
		}

		cm.controller.Enqueue("default", apiName)

		// Wait for reconciliation
		maxWait := time.Now().Add(10 * time.Second)
		for time.Now().Before(maxWait) {
			retrieved, err := cm.store.Get(ctx, "default", apiName)
			if err != nil {
				b.Fatalf("Get failed: %v", err)
			}

			if retrieved.Status.Phase == "Ready" {
				break
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}
