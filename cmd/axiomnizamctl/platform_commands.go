package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var TenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "Manage tenants",
}

var tenantListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tenants",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/tenants")
	},
}

var tenantGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get tenant",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/tenants/" + args[0])
	},
}

var tenantCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create tenant",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := promptInput("Tenant name")
		owner := promptInput("Owner user ID")
		if name == "" || owner == "" {
			return NewCommandError(ErrInvalidInput, "name and owner are required")
		}

		payload := map[string]interface{}{
			"name":           name,
			"displayName":    promptInput("Display name"),
			"owner":          owner,
			"tier":           strings.ToUpper(promptInput("Tier (FREE|PRO|ENTERPRISE)")),
			"isolationLevel": strings.ToUpper(promptInput("Isolation (SHARED|SCHEMA|DATABASE)")),
		}
		return postAndPrint("/api/v1/tenants", payload)
	},
}

var RBACCmd = &cobra.Command{
	Use:   "rbacx",
	Short: "Manage RBAC roles and checks",
}

var rbacRoleListCmd = &cobra.Command{
	Use:   "roles",
	Short: "List roles",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/rbac/roles")
	},
}

var rbacRoleCreateCmd = &cobra.Command{
	Use:   "create-role",
	Short: "Create role",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := promptInput("Role name")
		tenantID := promptInput("Tenant ID")
		if name == "" || tenantID == "" {
			return NewCommandError(ErrInvalidInput, "tenant and role name are required")
		}
		payload := map[string]interface{}{
			"tenantId":    tenantID,
			"name":        name,
			"description": promptInput("Description"),
		}
		return postAndPrint("/api/v1/rbac/roles", payload)
	},
}

var rbacCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check permission",
	RunE: func(cmd *cobra.Command, args []string) error {
		tenantID := promptInput("Tenant ID")
		principalID := promptInput("Principal ID")
		resource := promptInput("Resource")
		action := promptInput("Action")
		if principalID == "" || resource == "" || action == "" {
			return NewCommandError(ErrInvalidInput, "principal, resource and action are required")
		}
		payload := map[string]interface{}{
			"tenantId":    tenantID,
			"principalId": principalID,
			"resource":    resource,
			"action":      action,
		}
		return postAndPrint("/api/v1/rbac/permissions/check", payload)
	},
}

var WebhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Manage webhooks",
}

var webhookListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/webhooks")
	},
}

var webhookCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create webhook",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := promptInput("Webhook name")
		url := promptInput("Webhook URL")
		if name == "" || url == "" {
			return NewCommandError(ErrInvalidInput, "name and url are required")
		}
		payload := map[string]interface{}{
			"name":        name,
			"description": promptInput("Description"),
			"url":         url,
			"secret":      promptInput("Secret"),
			"events":      []string{"job.completed", "etl.failed"},
		}
		return postAndPrint("/api/v1/webhooks", payload)
	},
}

var webhookTestCmd = &cobra.Command{
	Use:   "test [id]",
	Short: "Test webhook delivery",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return postAndPrint("/api/v1/webhooks/"+args[0]+"/test", map[string]interface{}{})
	},
}

var StreamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Manage streams",
}

var streamListCmd = &cobra.Command{
	Use:   "list",
	Short: "List streams",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/streams")
	},
}

var streamCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create stream",
	RunE: func(cmd *cobra.Command, args []string) error {
		payload := map[string]interface{}{
			"tenantId": promptInput("Tenant ID"),
			"topic":    promptInput("Topic"),
			"query":    promptInput("Query (optional)"),
		}
		return postAndPrint("/api/v1/streams", payload)
	},
}

var streamCancelCmd = &cobra.Command{
	Use:   "cancel [id]",
	Short: "Cancel stream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteAndPrint("/api/v1/streams/" + args[0])
	},
}

var ExportCmd = &cobra.Command{
	Use:   "exportx",
	Short: "Manage exports",
}

var exportListCmd = &cobra.Command{
	Use:   "list",
	Short: "List exports",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/exports")
	},
}

var exportCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create export job",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := promptInput("Export name")
		database := promptInput("Database")
		table := promptInput("Table")
		if name == "" || database == "" || table == "" {
			return NewCommandError(ErrInvalidInput, "name, database, table are required")
		}
		payload := map[string]interface{}{
			"name":   name,
			"format": strings.ToUpper(promptInput("Format (CSV|JSON|PARQUET)")),
			"source": map[string]interface{}{
				"type":     "table",
				"database": database,
				"table":    table,
			},
			"destination": map[string]interface{}{
				"type": "local",
				"path": "/tmp",
			},
		}
		return postAndPrint("/api/v1/exports", payload)
	},
}

var exportProgressCmd = &cobra.Command{
	Use:   "progress [id]",
	Short: "Get export progress",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/exports/" + args[0] + "/progress")
	},
}

var BulkCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Manage bulk operations",
}

var bulkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List bulk operations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/bulk/operations")
	},
}

var bulkSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit bulk operation",
	RunE: func(cmd *cobra.Command, args []string) error {
		tenantID := promptInput("Tenant ID")
		opType := promptInput("Operation type")
		if tenantID == "" || opType == "" {
			return NewCommandError(ErrInvalidInput, "tenant and operation type are required")
		}
		payload := map[string]interface{}{
			"tenantId": tenantID,
			"type":     opType,
			"items": []map[string]interface{}{
				{"id": "item-1", "value": "sample"},
			},
			"options": map[string]interface{}{"dryRun": dry},
		}
		return postAndPrint("/api/v1/bulk/operations", payload)
	},
}

var bulkProgressCmd = &cobra.Command{
	Use:   "progress [id]",
	Short: "Get bulk operation progress",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/bulk/operations/" + args[0] + "/progress")
	},
}

var VersioningCmd = &cobra.Command{
	Use:   "versioning",
	Short: "Version history and rollback",
}

var versionHistoryCmd = &cobra.Command{
	Use:   "history [resource-type] [resource-id]",
	Short: "Get version history",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/versioning/history/" + args[0] + "/" + args[1])
	},
}

var versionDiffCmd = &cobra.Command{
	Use:   "diff [resource-type] [resource-id]",
	Short: "Get diff between versions",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetInt64("from")
		to, _ := cmd.Flags().GetInt64("to")
		path := fmt.Sprintf("/api/v1/versioning/diff/%s/%s?from=%d&to=%d", args[0], args[1], from, to)
		return getAndPrint(path)
	},
}

var versionRollbackCmd = &cobra.Command{
	Use:   "rollback [resource-type] [resource-id]",
	Short: "Rollback resource to version",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, _ := cmd.Flags().GetInt64("version")
		if target <= 0 {
			return NewCommandError(ErrInvalidInput, "--version must be > 0")
		}
		payload := map[string]interface{}{
			"targetVersion": target,
			"reason":        promptInput("Rollback reason"),
		}
		return postAndPrint("/api/v1/versioning/versions/"+args[0]+"/"+args[1]+"/rollback", payload)
	},
}

var TraceCmd = &cobra.Command{
	Use:   "trace",
	Short: "Search and inspect traces",
}

var traceSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search traces",
	RunE: func(cmd *cobra.Command, args []string) error {
		service, _ := cmd.Flags().GetString("service")
		limit, _ := cmd.Flags().GetInt("limit")
		path := fmt.Sprintf("/api/v1/tracing/traces/search?service=%s&limit=%d", service, limit)
		return getAndPrint(path)
	},
}

var traceGetCmd = &cobra.Command{
	Use:   "get [trace-id]",
	Short: "Get trace by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/tracing/traces/" + args[0])
	},
}

var LineageAPICmd = &cobra.Command{
	Use:   "lineagex",
	Short: "Explore lineage graph APIs",
}

var lineageGraphCmd = &cobra.Command{
	Use:   "graph [resource-type] [resource-id]",
	Short: "Get lineage graph",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/lineage/" + args[0] + "/" + args[1])
	},
}

var lineageImpactCmd = &cobra.Command{
	Use:   "impact [resource-type] [resource-id]",
	Short: "Get lineage impact analysis",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/lineage/impact/" + args[0] + "/" + args[1])
	},
}

var IncidentCmd = &cobra.Command{
	Use:   "incidents",
	Short: "View alerts and incidents summary",
}

var incidentOverviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Aggregate notifications and netintel incidents",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateServerConnection(); err != nil {
			return err
		}

		paths := []string{
			"/api/notifications/status",
			"/api/v1/netintel/alerts",
			"/api/v1/netintel/anomalies",
		}

		payload := map[string]interface{}{}
		for _, path := range paths {
			resp, err := apiClient.GetSimple(path)
			if err != nil {
				payload[path] = map[string]interface{}{"error": err.Error()}
				continue
			}
			payload[path] = decodeResponseJSON(resp)
		}

		return printPrettyJSON(payload)
	},
}

func getAndPrint(path string) error {
	if err := validateServerConnection(); err != nil {
		return err
	}
	resp, err := apiClient.GetSimple(path)
	if err != nil {
		return NewCommandError(ErrNetwork, "Request failed", err.Error())
	}
	return printPrettyJSON(decodeResponseJSON(resp))
}

func postAndPrint(path string, body interface{}) error {
	if err := validateServerConnection(); err != nil {
		return err
	}
	resp, err := apiClient.PostSimple(path, body)
	if err != nil {
		return NewCommandError(ErrNetwork, "Request failed", err.Error())
	}
	return printPrettyJSON(decodeResponseJSON(resp))
}

func deleteAndPrint(path string) error {
	if err := validateServerConnection(); err != nil {
		return err
	}
	resp, err := apiClient.DeleteSimple(path)
	if err != nil {
		return NewCommandError(ErrNetwork, "Request failed", err.Error())
	}
	return printPrettyJSON(decodeResponseJSON(resp))
}

func decodeResponseJSON(resp interface {
	JSON(v interface{}) error
	String() string
}) interface{} {
	var decoded interface{}
	if err := resp.JSON(&decoded); err != nil {
		return map[string]interface{}{"raw": resp.String()}
	}
	return decoded
}

func printPrettyJSON(data interface{}) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return NewCommandError(ErrInvalidInput, "Failed to encode JSON", err.Error())
	}
	fmt.Println(string(b))
	return nil
}

func init() {
	TenantCmd.AddCommand(tenantListCmd, tenantGetCmd, tenantCreateCmd)

	RBACCmd.AddCommand(rbacRoleListCmd, rbacRoleCreateCmd, rbacCheckCmd)

	WebhookCmd.AddCommand(webhookListCmd, webhookCreateCmd, webhookTestCmd)

	StreamCmd.AddCommand(streamListCmd, streamCreateCmd, streamCancelCmd)

	ExportCmd.AddCommand(exportListCmd, exportCreateCmd, exportProgressCmd)

	BulkCmd.AddCommand(bulkListCmd, bulkSubmitCmd, bulkProgressCmd)

	versionDiffCmd.Flags().Int64("from", 1, "Source version")
	versionDiffCmd.Flags().Int64("to", 2, "Target version")
	versionRollbackCmd.Flags().Int64("version", 0, "Target version number")
	VersioningCmd.AddCommand(versionHistoryCmd, versionDiffCmd, versionRollbackCmd)

	traceSearchCmd.Flags().String("service", "", "Service name")
	traceSearchCmd.Flags().Int("limit", 20, "Maximum results")
	TraceCmd.AddCommand(traceSearchCmd, traceGetCmd)

	LineageAPICmd.AddCommand(lineageGraphCmd, lineageImpactCmd)

	IncidentCmd.AddCommand(incidentOverviewCmd)
}
