package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var EventsCmd = &cobra.Command{
	Use:   "events",
	Short: "View resource events",
	Long:  "View events for resources like kubectl get events",
}

var EventsGetCmd = &cobra.Command{
	Use:   "get [resource-kind] [resource-name]",
	Short: "Get events for a resource",
	Long:  "Display events for a specific resource",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleGetEvents(args[0], args[1])
	},
}

var EventsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all events",
	Long:  "Display all recent events across all resources",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return handleListEvents(args[0])
		}
		return handleListAllEvents()
	},
}

// handleGetEvents gets events for a specific resource
func handleGetEvents(kind, name string) error {
	fmt.Printf("📋 Events for %s/%s\n\n", kind, name)

	events := []struct {
		Time    string
		Reason  string
		Type    string
		Count   int
		Message string
	}{
		{"2024-01-26T10:30:00Z", "ResourceApplied", "Normal", 1, "Resource applied successfully"},
		{"2024-01-26T10:30:05Z", "ReconciliationStarted", "Normal", 1, "Reconciliation started"},
		{"2024-01-26T10:30:10Z", "ReconciliationCompleted", "Normal", 1, "Reconciliation completed"},
	}

	fmt.Printf("%-27s %-25s %-10s %-5s %s\n", "TIME", "REASON", "TYPE", "COUNT", "MESSAGE")
	fmt.Println(strings.Repeat("─", 80))

	for _, e := range events {
		fmt.Printf("%-27s %-25s %-10s %-5d %s\n", e.Time, e.Reason, e.Type, e.Count, e.Message)
	}

	return nil
}

// handleListEvents lists all events for a resource kind
func handleListEvents(kind string) error {
	fmt.Printf("📋 Events for %s resources\n\n", kind)

	events := []struct {
		Resource string
		Time     string
		Reason   string
		Type     string
		Message  string
	}{
		{"prod-db", "2024-01-26T10:30:00Z", "ResourceApplied", "Normal", "Resource applied"},
		{"staging-db", "2024-01-26T10:29:00Z", "ReconcileFailed", "Warning", "Reconciliation failed"},
		{"dev-db", "2024-01-26T10:28:00Z", "PolicyDenied", "Warning", "Policy denied change"},
	}

	fmt.Printf("%-20s %-27s %-25s %-10s %s\n", "RESOURCE", "TIME", "REASON", "TYPE", "MESSAGE")
	fmt.Println(strings.Repeat("─", 85))

	for _, e := range events {
		fmt.Printf("%-20s %-27s %-25s %-10s %s\n", e.Resource, e.Time, e.Reason, e.Type, e.Message)
	}

	return nil
}

func init() {
	EventsCmd.AddCommand(EventsGetCmd)
	EventsCmd.AddCommand(EventsListCmd)
}

// handleListAllEvents lists all events across all resources
func handleListAllEvents() error {
	fmt.Println("📋 All Recent Events")
	fmt.Println()

	events := []struct {
		Kind     string
		Resource string
		Time     string
		Reason   string
		Type     string
		Message  string
	}{
		{"API", "prod-api", "2024-01-26T10:30:00Z", "ResourceApplied", "Normal", "API applied successfully"},
		{"Policy", "rate-limit", "2024-01-26T10:29:30Z", "PolicyEnforced", "Normal", "Policy enforced on prod-api"},
		{"Workflow", "data-sync", "2024-01-26T10:29:00Z", "WorkflowCompleted", "Normal", "Data sync completed"},
		{"DataSource", "staging-db", "2024-01-26T10:28:00Z", "ConnectionFailed", "Warning", "Connection test failed"},
		{"Job", "backup-job", "2024-01-26T10:27:00Z", "JobSucceeded", "Normal", "Backup completed"},
	}

	fmt.Printf("%-12s %-20s %-27s %-25s %-10s %s\n", "KIND", "RESOURCE", "TIME", "REASON", "TYPE", "MESSAGE")
	fmt.Println(strings.Repeat("─", 110))

	for _, e := range events {
		fmt.Printf("%-12s %-20s %-27s %-25s %-10s %s\n", e.Kind, e.Resource, e.Time, e.Reason, e.Type, e.Message)
	}

	return nil
}
