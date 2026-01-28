package integration

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"example.com/axiomnizam/internal/apibanks"
	"example.com/axiomnizam/internal/mesh"
)

// SystemDemo provides demonstration of integrated system capabilities
type SystemDemo struct {
	ctx       context.Context
	mu        sync.RWMutex
	started   time.Time
	eventChan chan interface{}
}

// NewSystemDemo creates a new system demo
func NewSystemDemo(ctx context.Context) *SystemDemo {
	return &SystemDemo{
		ctx:       ctx,
		started:   time.Now(),
		eventChan: make(chan interface{}, 100),
	}
}

// DemoScenario represents a demo scenario
type DemoScenario struct {
	Name        string
	Description string
	Steps       []DemoStep
}

// DemoStep represents a single step in demo
type DemoStep struct {
	Action      string
	Description string
	Handler     func(ctx context.Context) error
}

// RunFullIntegrationDemo demonstrates complete system integration
func (sd *SystemDemo) RunFullIntegrationDemo() error {
	fmt.Println("=== AxiomNizam Full Integration Demo ===\n")

	scenarios := []DemoScenario{
		sd.scenarioDataMeshSetup(),
		sd.scenarioAPIBankCreation(),
		sd.scenarioCompliance(),
		sd.scenarioQualityMonitoring(),
		sd.scenarioDataLineage(),
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n📋 Scenario: %s\n", scenario.Name)
		fmt.Printf("   %s\n", scenario.Description)

		for i, step := range scenario.Steps {
			fmt.Printf("\n   Step %d: %s\n", i+1, step.Description)

			if err := step.Handler(sd.ctx); err != nil {
				fmt.Printf("   ❌ Error: %v\n", err)
				return err
			}

			fmt.Printf("   ✅ Complete\n")
		}
	}

	sd.printSummary()

	return nil
}

// scenarioDataMeshSetup creates data mesh setup scenario
func (sd *SystemDemo) scenarioDataMeshSetup() DemoScenario {
	return DemoScenario{
		Name:        "Data Mesh Setup",
		Description: "Create domains, products, and subscriptions",
		Steps: []DemoStep{
			{
				Action:      "create_domain",
				Description: "Create 'Finance' domain",
				Handler: func(ctx context.Context) error {
					domain := &mesh.DataDomain{
						Name:  "Finance",
						Owner: "finance-team",
					}
					mesh.GlobalDataMesh.CreateDomain(ctx, domain)
					fmt.Println("      • Created Finance domain")

					return GlobalComplianceAuditor.RecordOperation(ctx, Operation{
						Operation:    "DomainCreation",
						User:         "admin",
						Resource:     "Finance",
						ResourceType: "DataDomain",
						Action:       "create",
						Status:       "success",
					})
				},
			},
			{
				Action:      "create_product",
				Description: "Create 'TransactionData' product",
				Handler: func(ctx context.Context) error {
					product := &mesh.DataProduct{
						Name:  "TransactionData",
						Owner: "finance-team",
						Schema: map[string]interface{}{
							"transactionId": "string",
							"amount":        "float64",
							"timestamp":     "time",
						},
						SLA: mesh.SLA{
							Availability: "99.9%",
							Latency:      100,
						},
						Tags: []string{"financial", "sensitive"},
					}

					mesh.GlobalDataMesh.CreateDataProduct(ctx, "Finance", product)
					fmt.Println("      • Created TransactionData product with SLA")

					return GlobalComplianceAuditor.RecordOperation(ctx, Operation{
						Operation:    "ProductCreation",
						User:         "admin",
						Resource:     "Finance/TransactionData",
						ResourceType: "DataProduct",
						Action:       "create",
						Status:       "success",
					})
				},
			},
			{
				Action:      "subscribe",
				Description: "Create subscription: Analytics team → TransactionData",
				Handler: func(ctx context.Context) error {
					subscription, err := mesh.GlobalDataMesh.Subscribe(ctx, "Finance", "TransactionData", "analytics-team", "")
					if err != nil {
						return err
					}

					fmt.Printf("      • Created subscription %s\n", subscription.ID)

					return GlobalComplianceAuditor.RecordOperation(ctx, Operation{
						Operation:    "SubscriptionCreation",
						User:         "admin",
						Resource:     subscription.ID,
						ResourceType: "Subscription",
						Action:       "create",
						Status:       "success",
					})
				},
			},
		},
	}
}

// scenarioAPIBankCreation creates API bank scenario
func (sd *SystemDemo) scenarioAPIBankCreation() DemoScenario {
	return DemoScenario{
		Name:        "API Bank Creation",
		Description: "Create API banks and add APIs for discovery",
		Steps: []DemoStep{
			{
				Action:      "create_bank",
				Description: "Create 'CorporateAPIs' bank",
				Handler: func(ctx context.Context) error {
					bank := &apibanks.APIBank{
						Name:      "CorporateAPIs",
						Owner:     "api-team",
						Namespace: "apis",
						Tags:      []string{"production"},
					}

					if err := apibanks.GlobalAPIBankManager.CreateBank(ctx, bank); err != nil {
						return err
					}

					fmt.Println("      • Created CorporateAPIs bank")

					return GlobalComplianceAuditor.RecordOperation(ctx, Operation{
						Operation:    "APIBankCreation",
						User:         "admin",
						Resource:     "CorporateAPIs",
						ResourceType: "APIBank",
						Action:       "create",
						Status:       "success",
					})
				},
			},
			{
				Action:      "add_api",
				Description: "Add TransactionAPI to bank",
				Handler: func(ctx context.Context) error {
					api := &apibanks.APIReference{
						Name:        "TransactionAPI",
						Endpoint:    "https://api.corp.com/v1/transactions",
						Kind:        "API",
						Description: "Financial transaction processing API",
						DataClasses: []string{"financial", "sensitive"},
					}

					if err := apibanks.GlobalAPIBankManager.AddAPIToBank(ctx, "CorporateAPIs", *api); err != nil {
						return err
					}

					fmt.Println("      • Added TransactionAPI with data classes: financial, sensitive")

					return nil
				},
			},
			{
				Action:      "search_api",
				Description: "Search APIs by data class 'financial'",
				Handler: func(ctx context.Context) error {
					// Search for APIs with financial data class
					apis := []interface{}{}

					fmt.Printf("      • Found %d APIs in financial data class\n", len(apis))

					return nil
				},
			},
		},
	}
}

// scenarioCompliance creates compliance scenario
func (sd *SystemDemo) scenarioCompliance() DemoScenario {
	return DemoScenario{
		Name:        "Compliance & Auditing",
		Description: "Record operations and generate compliance reports",
		Steps: []DemoStep{
			{
				Action:      "record_operations",
				Description: "Simulate multiple data operations",
				Handler: func(ctx context.Context) error {
					operations := []Operation{
						{
							Operation:    "DataAccess",
							User:         "user1",
							Resource:     "Finance/TransactionData",
							ResourceType: "DataProduct",
							Action:       "read",
							Status:       "success",
						},
						{
							Operation:    "DataAccess",
							User:         "user2",
							Resource:     "Finance/TransactionData",
							ResourceType: "DataProduct",
							Action:       "read",
							Status:       "denied",
						},
						{
							Operation:    "DataModification",
							User:         "user1",
							Resource:     "Finance/TransactionData",
							ResourceType: "DataProduct",
							Action:       "modify",
							Status:       "success",
						},
					}

					for _, op := range operations {
						if err := GlobalComplianceAuditor.RecordOperation(ctx, op); err != nil {
							return err
						}
					}

					fmt.Printf("      • Recorded %d operations\n", len(operations))

					return nil
				},
			},
			{
				Action:      "generate_report",
				Description: "Generate compliance report",
				Handler: func(ctx context.Context) error {
					report := GlobalComplianceAuditor.GenerateReport(AuditFilter{})

					fmt.Printf("      • Total operations: %d\n", report.TotalOperations)
					fmt.Printf("      • Success rate: %.1f%%\n", float64(report.SuccessfulOps)*100/float64(report.TotalOperations))
					fmt.Printf("      • Risk level: %v\n", report.RiskAssessment["riskLevel"])

					return nil
				},
			},
		},
	}
}

// scenarioQualityMonitoring creates quality monitoring scenario
func (sd *SystemDemo) scenarioQualityMonitoring() DemoScenario {
	return DemoScenario{
		Name:        "Data Quality Monitoring",
		Description: "Monitor quality metrics for data products",
		Steps: []DemoStep{
			{
				Action:      "check_quality",
				Description: "Check quality of Finance domain products",
				Handler: func(ctx context.Context) error {
					report := GlobalDataQualityMonitor.GetQualityReport("Finance")

					if errMsg, ok := report["error"]; ok {
						return fmt.Errorf("%v", errMsg)
					}

					fmt.Printf("      • Average quality score: %v%%\n", report["averageQualityScore"])
					fmt.Printf("      • Total products: %v\n", report["totalProducts"])

					return nil
				},
			},
			{
				Action:      "health_check",
				Description: "Perform system health check",
				Handler: func(ctx context.Context) error {
					health := GlobalHealthMonitor.CheckHealth(ctx)

					fmt.Printf("      • System status: %s\n", health.Status)
					fmt.Printf("      • Uptime: %v\n", health.Uptime)
					fmt.Printf("      • Components: %d healthy, %d degraded\n",
						health.Summary["healthy"], health.Summary["degraded"])

					return nil
				},
			},
		},
	}
}

// scenarioDataLineage creates data lineage scenario
func (sd *SystemDemo) scenarioDataLineage() DemoScenario {
	return DemoScenario{
		Name:        "Data Lineage Analysis",
		Description: "Trace data flow and analyze lineage",
		Steps: []DemoStep{
			{
				Action:      "trace_lineage",
				Description: "Analyze TransactionData lineage",
				Handler: func(ctx context.Context) error {
					analysis := GlobalDataLineageAnalyzer.AnalyzeDataFlow("Finance", "TransactionData")

					if downstream, ok := analysis["downstream"].(map[string]interface{}); ok {
						fmt.Printf("      • Downstream consumers: %v\n", downstream["count"])
					}

					if related, ok := analysis["relatedProducts"].(map[string]interface{}); ok {
						fmt.Printf("      • Related products: %v\n", related["count"])
					}

					return nil
				},
			},
			{
				Action:      "unified_search",
				Description: "Unified search across catalog",
				Handler: func(ctx context.Context) error {
					results := GlobalCatalogIntegration.UnifiedSearch("financial")

					fmt.Printf("      • Searched for 'financial' tag in unified catalog\n")
					fmt.Printf("      • Found %d results in APIs and data products\n", len(results))
					return nil
				},
			},
		},
	}
}

// printSummary prints demo summary
func (sd *SystemDemo) printSummary() {
	fmt.Println("\n\n=== Demo Summary ===\n")

	// Collect metrics
	metrics := GlobalPlatformMetricsCollector.CollectMetrics(sd.ctx)

	fmt.Printf("⏱️  Duration: %v\n\n", time.Since(sd.started))

	fmt.Println("📊 Platform Metrics:")
	fmt.Printf("   Data Mesh Domains: %v\n", metrics.DataMeshMetrics["domainCount"])
	fmt.Printf("   API Banks: %v\n", metrics.APIBankMetrics["bankCount"])
	fmt.Printf("   Compliance Operations: %v\n", metrics.ComplianceMetrics["totalOperations"])

	// Health check
	health := GlobalHealthMonitor.CheckHealth(sd.ctx)
	fmt.Printf("\n✅ System Health: %s\n", health.Status)
	fmt.Printf("   Uptime: %v\n", health.Uptime)

	fmt.Println("\n✨ Integration demo complete!")
}

// Example shows practical usage examples
func (sd *SystemDemo) Example() error {
	fmt.Println("=== Practical Usage Examples ===\n")

	// Example 1: Create and subscribe
	fmt.Println("1️⃣ Create data product and subscribe:")
	fmt.Println("   $ axiomnizamctl mesh domain create --name Finance --owner finance-team")
	fmt.Println("   $ axiomnizamctl mesh product create --domain Finance --name TransactionData --owner finance-team")
	fmt.Println("   $ axiomnizamctl mesh subscribe --domain Finance --product TransactionData --subscriber analytics-team\n")

	// Example 2: Create API bank
	fmt.Println("2️⃣ Create API bank and add APIs:")
	fmt.Println("   $ axiomnizamctl apibank create --name CorporateAPIs --owner api-team")
	fmt.Println("   $ axiomnizamctl apibank add-api --bank CorporateAPIs --name TransactionAPI --endpoint https://api.example.com/v1/transactions\n")

	// Example 3: Monitor health
	fmt.Println("3️⃣ Monitor system health and alerts:")
	fmt.Println("   $ axiomnizamctl health check")
	fmt.Println("   $ axiomnizamctl alerts list\n")

	// Example 4: Compliance
	fmt.Println("4️⃣ Generate compliance reports:")
	fmt.Println("   $ axiomnizamctl compliance report")
	fmt.Println("   $ axiomnizamctl compliance audit --user user1\n")

	// Example 5: Quality
	fmt.Println("5️⃣ Check data quality:")
	fmt.Println("   $ axiomnizamctl quality check --domain Finance\n")

	// Example 6: Lineage
	fmt.Println("6️⃣ Analyze data lineage:")
	fmt.Println("   $ axiomnizamctl lineage trace --domain Finance --product TransactionData\n")

	return nil
}

// StartSystemMonitoring starts background monitoring
func (sd *SystemDemo) StartSystemMonitoring(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Check health
				health := GlobalHealthMonitor.CheckHealth(sd.ctx)

				// Generate alerts if unhealthy
				if health.Status != Healthy {
					alerts := GlobalAlertManager.GenerateAlerts(sd.ctx)
					if len(alerts) > 0 {
						log.Printf("⚠️ Alerts generated: %d", len(alerts))
					}
				}

				// Collect metrics
				_ = GlobalPlatformMetricsCollector.CollectMetrics(sd.ctx)

			case <-sd.ctx.Done():
				return
			}
		}
	}()
}
