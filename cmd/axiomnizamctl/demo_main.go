package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

// DemoMain shows the complete reconciliation loop:
// YAML → Parse → Store (Pending) → Reconcile → Status Updated (Ready)
func DemoMain() {
	fmt.Println("======================================================================")
	fmt.Println("RECONCILIATION LOOP DEMONSTRATION")
	fmt.Println("======================================================================")
	fmt.Println()

	// Step 1: Initialize CLI Manager
	fmt.Println("[STEP 1] Initializing CLI Manager and Controllers...")
	cm := NewCLIManager()
	defer cm.Shutdown()
	fmt.Println("✓ CLI Manager started")
	fmt.Println("✓ Store initialized")
	fmt.Println("✓ Controller started with 3 workers")
	fmt.Println()

	// Step 2: Apply YAML file
	fmt.Println("[STEP 2] Applying APIResource from YAML...")
	fmt.Println("File: examples/api.yaml")
	yamlPath := "examples/api.yaml"

	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		log.Fatalf("YAML file not found: %s", yamlPath)
	}

	startTime := time.Now()
	err := cm.Apply(yamlPath)
	if err != nil {
		log.Fatalf("Failed to apply resource: %v", err)
	}
	duration := time.Since(startTime)

	fmt.Printf("✓ Resource applied successfully (took %.2fs)\n", duration.Seconds())
	fmt.Println()

	// Step 3: Get resource status
	fmt.Println("[STEP 3] Checking final resource status...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Wait a moment for metrics to update
	time.Sleep(500 * time.Millisecond)
	cm.Get("default")
	fmt.Println()

	// Step 4: Show detailed resource info
	fmt.Println("[STEP 4] Describing resource in detail...")
	cm.Describe("default", "users-api")
	fmt.Println()

	// Step 5: Show controller metrics
	fmt.Println("[STEP 5] Controller Reconciliation Metrics...")
	cm.ShowControllerMetrics()
	fmt.Println()

	// Step 6: Summary
	fmt.Println("======================================================================")
	fmt.Println("RECONCILIATION LOOP COMPLETE")
	fmt.Println("======================================================================")
	fmt.Println()
	fmt.Println("Lifecycle Summary:")
	fmt.Println("  1. YAML file read and parsed")
	fmt.Println("  2. APIResource created with Status.Phase = Pending")
	fmt.Println("  3. Resource stored in in-memory store")
	fmt.Println("  4. Item enqueued for reconciliation")
	fmt.Println("  5. Controller worker picked up item from queue")
	fmt.Println("  6. Pending → Creating state transition")
	fmt.Println("  7. Simulated work (1-2s sleep)")
	fmt.Println("  8. Creating → Ready state transition")
	fmt.Println("  9. Status updated with Phase=Ready, Ready=true")
	fmt.Println(" 10. Resource displayed with final status")
	fmt.Println()
	fmt.Println("This demonstrates the complete Kubernetes-style reconciliation pattern:")
	fmt.Println("  Desired State (YAML) → Store → Work Queue → Worker → Actual State → Status Updated")
	fmt.Println()
}

func main() {
	DemoMain()
}
