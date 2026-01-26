package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"example.com/axiomnizam/internal/integration"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check system health and status",
	Long:  "Check the health of all platform components",
}

var healthCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Perform full system health check",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		health := integration.GlobalHealthMonitor.CheckHealth(ctx)

		fmt.Printf("System Health Status: %s\n", health.Status)
		fmt.Printf("Checked at: %s\n", health.CheckedAt.Format(time.RFC3339))
		fmt.Printf("Uptime: %v\n\n", health.Uptime)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "COMPONENT\tSTATUS\tLAST_CHECKED\tDETAILS")

		for _, comp := range health.Components {
			details := ""
			for k, v := range comp.Details {
				details += fmt.Sprintf("%s=%v ", k, v)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				comp.Component, comp.Status, comp.LastChecked.Format(time.RFC3339), details)
		}

		w.Flush()

		fmt.Printf("\nSummary: %d healthy, %d degraded, %d unhealthy\n",
			health.Summary["healthy"], health.Summary["degraded"], health.Summary["unhealthy"])

		return nil
	},
}

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Manage and view system alerts",
}

var alertsCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for new alerts",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		alerts := integration.GlobalAlertManager.GenerateAlerts(ctx)

		if len(alerts) == 0 {
			fmt.Println("No new alerts")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSEVERITY\tCOMPONENT\tMESSAGE\tTIMESTAMP")

		for _, alert := range alerts {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				alert.ID, alert.Severity, alert.Component, alert.Message, alert.Timestamp.Format(time.RFC3339))
		}

		w.Flush()

		return nil
	},
}

var alertsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active alerts",
	RunE: func(cmd *cobra.Command, args []string) error {
		alerts := integration.GlobalAlertManager.GetActiveAlerts()

		if len(alerts) == 0 {
			fmt.Println("No active alerts")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSEVERITY\tCOMPONENT\tMESSAGE\tTIMESTAMP")

		for _, alert := range alerts {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				alert.ID, alert.Severity, alert.Component, alert.Message, alert.Timestamp.Format(time.RFC3339))
		}

		w.Flush()

		return nil
	},
}

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "View platform metrics",
}

var metricsCollectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collect current platform metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		metrics := integration.GlobalPlatformMetricsCollector.CollectMetrics(ctx)

		fmt.Printf("Platform Metrics (collected at %s)\n\n", metrics.Timestamp.Format(time.RFC3339))

		fmt.Println("=== Data Mesh Metrics ===")
		for k, v := range metrics.DataMeshMetrics {
			fmt.Printf("%s: %v\n", k, v)
		}

		fmt.Println("\n=== API Bank Metrics ===")
		for k, v := range metrics.APIBankMetrics {
			fmt.Printf("%s: %v\n", k, v)
		}

		fmt.Println("\n=== Compliance Metrics ===")
		for k, v := range metrics.ComplianceMetrics {
			fmt.Printf("%s: %v\n", k, v)
		}

		fmt.Println("\n=== Performance Metrics ===")
		for k, v := range metrics.PerformanceMetrics {
			fmt.Printf("%s: %v\n", k, v)
		}

		return nil
	},
}

var catalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Unified data and API catalog",
}

var catalogSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search unified catalog",
	Long:  "Search across API banks and data mesh by tag",
	Flags: cobra.Flag{},
	RunE: func(cmd *cobra.Command, args []string) error {
		tag, _ := cmd.Flags().GetString("tag")
		if tag == "" {
			return fmt.Errorf("--tag is required")
		}

		result := integration.GlobalCatalogIntegration.UnifiedSearch(tag)

		fmt.Printf("Search results for tag: %s\n\n", tag)

		// Print API banks
		if banks, ok := result["apiBanks"]; ok {
			fmt.Println("=== API Banks ===")
			if refs, ok := banks.(interface{}); ok {
				fmt.Printf("%v\n", refs)
			}
		}

		// Print data products
		if products, ok := result["dataProducts"]; ok {
			fmt.Println("\n=== Data Products ===")
			if prods, ok := products.(interface{}); ok {
				fmt.Printf("%v\n", prods)
			}
		}

		return nil
	},
}

var catalogListCmd = &cobra.Command{
	Use:   "list",
	Short: "List complete data catalog",
	RunE: func(cmd *cobra.Command, args []string) error {
		catalog := integration.GlobalCatalogIntegration.GetCompleteDataCatalog()

		fmt.Println("=== Complete Data Catalog ===\n")

		if banks, ok := catalog["apiBanks"].(map[string]interface{}); ok {
			fmt.Printf("API Banks: %v\n", banks["count"])
		}

		if mesh, ok := catalog["dataMesh"].(map[string]interface{}); ok {
			fmt.Printf("Data Mesh Domains: %v\n", mesh["domainCount"])
			fmt.Printf("Total Data Products: %v\n", mesh["totalProducts"])
		}

		return nil
	},
}

var complianceCmd = &cobra.Command{
	Use:   "compliance",
	Short: "Manage compliance and auditing",
}

var complianceReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate compliance report",
	RunE: func(cmd *cobra.Command, args []string) error {
		report := integration.GlobalComplianceAuditor.GenerateReport(integration.AuditFilter{})

		fmt.Println("=== Compliance Report ===\n")
		fmt.Printf("Generated: %s\n", report.GeneratedAt.Format(time.RFC3339))
		fmt.Printf("Total Operations: %d\n", report.TotalOperations)
		fmt.Printf("Successful: %d (%.2f%%)\n", report.SuccessfulOps, float64(report.SuccessfulOps)*100/float64(report.TotalOperations))
		fmt.Printf("Denied: %d (%.2f%%)\n", report.DeniedOps, float64(report.DeniedOps)*100/float64(report.TotalOperations))
		fmt.Printf("Failed: %d (%.2f%%)\n", report.FailedOps, float64(report.FailedOps)*100/float64(report.TotalOperations))

		fmt.Println("\n=== Operations by Type ===")
		for op, count := range report.OperationsByType {
			fmt.Printf("%s: %d\n", op, count)
		}

		fmt.Println("\n=== Risk Assessment ===")
		for k, v := range report.RiskAssessment {
			fmt.Printf("%s: %v\n", k, v)
		}

		return nil
	},
}

var complianceAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View audit log",
	RunE: func(cmd *cobra.Command, args []string) error {
		user, _ := cmd.Flags().GetString("user")
		operation, _ := cmd.Flags().GetString("operation")

		filter := integration.AuditFilter{
			User:      user,
			Operation: operation,
		}

		records := integration.GlobalComplianceAuditor.GetAuditLog(filter)

		if len(records) == 0 {
			fmt.Println("No audit records found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TIMESTAMP\tUSER\tOPERATION\tRESOURCE\tSTATUS\tRESOURCE_TYPE")

		for _, record := range records {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				record.Timestamp.Format(time.RFC3339), record.User, record.Operation, record.Resource, record.Status, record.ResourceType)
		}

		w.Flush()

		return nil
	},
}

var qualityCmd = &cobra.Command{
	Use:   "quality",
	Short: "Monitor data quality",
}

var qualityCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check data quality for domain",
	Long:  "Generate quality report for all products in a domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		if domain == "" {
			return fmt.Errorf("--domain is required")
		}

		report := integration.GlobalDataQualityMonitor.GetQualityReport(domain)

		if errMsg, ok := report["error"]; ok {
			return fmt.Errorf("%v", errMsg)
		}

		fmt.Printf("Quality Report for Domain: %s\n\n", domain)
		fmt.Printf("Total Products: %v\n", report["totalProducts"])
		fmt.Printf("Average Quality Score: %v%%\n", report["averageQualityScore"])

		if scores, ok := report["productScores"].([]map[string]interface{}); ok {
			fmt.Println("\n=== Product Scores ===")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PRODUCT\tQUALITY_SCORE\tSUBSCRIPTIONS\tHAS_SLA\tHAS_PORTS")

			for _, score := range scores {
				fmt.Fprintf(w, "%v\t%v%%\t%v\t%v\t%v\n",
					score["productName"], score["qualityScore"], score["subscriptions"],
					score["hasSLA"], score["hasPorts"])
			}
			w.Flush()
		}

		return nil
	},
}

var lineageCmd = &cobra.Command{
	Use:   "lineage",
	Short: "Analyze data lineage",
}

var lineageTraceCmd = &cobra.Command{
	Use:   "trace",
	Short: "Trace data flow",
	Long:  "Analyze complete data flow (upstream, downstream, related)",
	RunE: func(cmd *cobra.Command, args []string) error {
		domain, _ := cmd.Flags().GetString("domain")
		product, _ := cmd.Flags().GetString("product")

		if domain == "" || product == "" {
			return fmt.Errorf("--domain and --product are required")
		}

		analysis := integration.GlobalDataLineageAnalyzer.AnalyzeDataFlow(domain, product)

		fmt.Printf("Data Lineage Analysis\n")
		fmt.Printf("Data Product: %v\n\n", analysis["dataProduct"])

		if downstream, ok := analysis["downstream"].(map[string]interface{}); ok {
			fmt.Printf("Downstream: %v consumers\n", downstream["count"])
		}

		if upstream, ok := analysis["upstream"].(map[string]interface{}); ok {
			fmt.Printf("Upstream: %v sources\n", upstream["count"])
		}

		if related, ok := analysis["relatedProducts"].(map[string]interface{}); ok {
			fmt.Printf("Related Products: %v\n", related["count"])
		}

		return nil
	},
}

func init() {
	// Health command
	healthCmd.AddCommand(healthCheckCmd)
	rootCmd.AddCommand(healthCmd)

	// Alerts command
	alertsCmd.AddCommand(alertsCheckCmd, alertsListCmd)
	rootCmd.AddCommand(alertsCmd)

	// Metrics command
	metricsCmd.AddCommand(metricsCollectCmd)
	rootCmd.AddCommand(metricsCmd)

	// Catalog command
	catalogSearchCmd.Flags().String("tag", "", "Tag to search for")
	catalogCmd.AddCommand(catalogSearchCmd, catalogListCmd)
	rootCmd.AddCommand(catalogCmd)

	// Compliance command
	complianceAuditCmd.Flags().String("user", "", "Filter by user")
	complianceAuditCmd.Flags().String("operation", "", "Filter by operation")
	complianceCmd.AddCommand(complianceReportCmd, complianceAuditCmd)
	rootCmd.AddCommand(complianceCmd)

	// Quality command
	qualityCheckCmd.Flags().String("domain", "", "Domain name")
	qualityCmd.AddCommand(qualityCheckCmd)
	rootCmd.AddCommand(qualityCmd)

	// Lineage command
	lineageTraceCmd.Flags().String("domain", "", "Domain name")
	lineageTraceCmd.Flags().String("product", "", "Product name")
	lineageCmd.AddCommand(lineageTraceCmd)
	rootCmd.AddCommand(lineageCmd)
}
