package main

import (
	"fmt"
	"os"
	"time"

	"example.com/axiomnizam/internal/output"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var APICmd = &cobra.Command{
	Use:   "api",
	Short: "Manage APIs",
	Long:  "Create, list, get, update, delete, and apply API resources",
}

var APICreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API",
	Long:  "Interactively create a new API resource",
	Run: func(cmd *cobra.Command, args []string) {
		handleAPICreate()
	},
}

var APIListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all APIs",
	Long:  "List all API resources in namespace",
	Run: func(cmd *cobra.Command, args []string) {
		handleAPIList()
	},
}

var APIGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get API details",
	Long:  "Get detailed information about a specific API",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleAPIGet(args[0])
	},
}

var APIUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update an API",
	Long:  "Update a specific field in an API resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleAPIUpdate(args[0])
	},
}

var APIDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete an API",
	Long:  "Delete an API resource (requires confirmation)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleAPIDelete(args[0])
	},
}

var APIApplyCmd = &cobra.Command{
	Use:   "apply -f [file]",
	Short: "Apply API from YAML (triggers controller reconciliation)",
	Long: `Create or update an API resource from YAML file.

This command uses Kubernetes-style reconciliation:
1. Parses and validates the YAML file
2. Sends to API server with metadata
3. Controller detects change and queues reconciliation
4. Reconciler applies desired state
5. Status is updated with reconciliation result

Examples:
  axiomnizamctl api apply -f api.yaml
  axiomnizamctl api apply -f api.yaml --dry-run
  axiomnizamctl api apply -f api.yaml --namespace prod`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filename, _ := cmd.Flags().GetString("filename")
		if filename == "" {
			return fmt.Errorf("filename flag (-f) is required")
		}

		opts := ApplyOptions{
			Filename:  filename,
			DryRun:    dry,
			Force:     false,
			Namespace: namespace,
			Timeout:   30 * time.Second,
		}

		if force, _ := cmd.Flags().GetBool("force"); force {
			opts.Force = force
		}

		if t, _ := cmd.Flags().GetDuration("timeout"); t > 0 {
			opts.Timeout = t
		}

		return handleApply(opts)
	},
}

var APIDescribeCmd = &cobra.Command{
	Use:   "describe [name]",
	Short: "Show detailed API information",
	Long:  "Show detailed information about an API including status and events",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleAPIDescribe(args[0])
	},
}

var APIDiffCmd = &cobra.Command{
	Use:   "diff -f [file]",
	Short: "Show differences between file and server",
	Long:  "Show what would change if you applied a YAML file",
	Run: func(cmd *cobra.Command, args []string) {
		file, _ := cmd.Flags().GetString("filename")
		if file == "" {
			fmt.Println("❌ filename flag is required")
			return
		}
		handleAPIDiff(file)
	},
}

func handleAPICreate() {
	fmt.Println("📝 Create API Resource")

	name := promptInput("API Name")
	db := promptInput("Database")
	table := promptInput("Table")

	apiResource := map[string]interface{}{
		"apiVersion": "axiom-nizam.io/v1",
		"kind":       "API",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"database": db,
			"table":    table,
			"rateLimit": map[string]interface{}{
				"enabled":             true,
				"requests_per_second": 100,
			},
		},
	}

	// Send to server
	response, err := apiClient.PostSimple("/api/v1/apis", apiResource)
	if err != nil {
		fmt.Printf("❌ Failed to create API: %v\n", err)
		return
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		fmt.Printf("✅ API '%s' created successfully\n", name)
	} else {
		fmt.Printf("❌ Failed: %s\n", response.Status)
	}
}

func handleAPIList() {
	fmt.Println("📋 APIs")

	response, err := apiClient.GetSimple("/api/v1/apis")
	if err != nil {
		fmt.Printf("❌ Failed to list APIs: %v\n", err)
		return
	}

	if response.StatusCode != 200 {
		fmt.Printf("❌ Request failed: %s\n", response.Status)
		return
	}

	var apis []map[string]interface{}
	if err := response.JSON(&apis); err != nil {
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		return
	}

	formatter := output.NewFormatter(outputFormat)
	formatter.Print(apis)
}

func handleAPIGet(name string) {
	response, err := apiClient.GetSimple(fmt.Sprintf("/api/v1/namespaces/%s/apis/%s", namespace, name))
	if err != nil {
		fmt.Printf("❌ Failed to get API: %v\n", err)
		return
	}

	if response.StatusCode != 200 {
		fmt.Printf("❌ API not found\n")
		return
	}

	var api map[string]interface{}
	if err := response.JSON(&api); err != nil {
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		return
	}

	formatter := output.NewFormatter(outputFormat)
	formatter.Print(api)
}

func handleAPIUpdate(name string) {
	field := promptInput("Field to update")
	value := promptInput("New value")

	updatePayload := map[string]interface{}{
		field: value,
	}

	response, err := apiClient.PutSimple(fmt.Sprintf("/api/v1/namespaces/%s/apis/%s", namespace, name), updatePayload)
	if err != nil {
		fmt.Printf("❌ Failed to update API: %v\n", err)
		return
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		fmt.Printf("✅ API '%s' updated successfully\n", name)
	} else {
		fmt.Printf("❌ Failed: %s\n", response.Status)
	}
}

func handleAPIDelete(name string) {
	if !confirmAction("Are you sure you want to delete this API?") {
		fmt.Println("❌ Cancelled")
		return
	}

	response, err := apiClient.DeleteSimple(fmt.Sprintf("/api/v1/namespaces/%s/apis/%s", namespace, name))
	if err != nil {
		fmt.Printf("❌ Failed to delete API: %v\n", err)
		return
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		fmt.Printf("✅ API '%s' deleted successfully\n", name)
	} else {
		fmt.Printf("❌ Failed: %s\n", response.Status)
	}
}

func handleAPIApply(filename string) {
	// Read YAML file
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("❌ Failed to read file: %v\n", err)
		return
	}

	// Parse YAML
	var resource map[string]interface{}
	if err := yaml.Unmarshal(data, &resource); err != nil {
		fmt.Printf("❌ Failed to parse YAML: %v\n", err)
		return
	}

	// Extract name from metadata
	metadata, ok := resource["metadata"].(map[string]interface{})
	if !ok {
		fmt.Println("❌ Invalid resource: missing metadata")
		return
	}

	name, ok := metadata["name"].(string)
	if !ok {
		fmt.Println("❌ Invalid resource: missing metadata.name")
		return
	}

	// Send to server
	response, err := apiClient.PostSimple("/api/v1/apis", resource)
	if err != nil {
		fmt.Printf("❌ Failed to apply API: %v\n", err)
		return
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		fmt.Printf("✅ API '%s' applied successfully\n", name)
	} else {
		fmt.Printf("❌ Failed: %s\n", response.Status)
	}
}

func handleAPIDescribe(name string) {
	response, err := apiClient.GetSimple(fmt.Sprintf("/api/v1/namespaces/%s/apis/%s", namespace, name))
	if err != nil {
		output.PrintError(output.ErrServerError, err.Error())
		return
	}

	if response.StatusCode != 200 {
		output.PrintError(output.ErrNotFound, fmt.Sprintf("API '%s' not found in namespace '%s'", name, namespace))
		return
	}

	var api map[string]interface{}
	if err := response.JSON(&api); err != nil {
		output.PrintError(output.ErrInvalidInput, "Failed to parse response")
		return
	}

	// Print detailed info
	metadata := api["metadata"].(map[string]interface{})
	spec := api["spec"].(map[string]interface{})
	status := api["status"].(map[string]interface{})

	fmt.Printf("\n📋 API: %s\n", name)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Namespace:  %v\n", metadata["namespace"])
	fmt.Printf("Database:   %v\n", spec["database"])
	fmt.Printf("Table:      %v\n", spec["table"])
	if phase, ok := status["phase"]; ok {
		fmt.Printf("Status:     %v\n", phase)
	}
	if age, ok := metadata["age"]; ok {
		fmt.Printf("Age:        %v\n", age)
	}
	fmt.Println()

	// Try to get events
	eventsResp, _ := apiClient.Get(fmt.Sprintf("/api/v1/namespaces/%s/apis/%s/events", namespace, name))
	if eventsResp != nil && eventsResp.StatusCode == 200 {
		var events []map[string]interface{}
		if err := eventsResp.JSON(&events); err == nil && len(events) > 0 {
			fmt.Println("📝 Recent Events:")
			headers := []string{"TYPE", "REASON", "MESSAGE", "AGE"}
			rows := [][]string{}
			for _, event := range events {
				rows = append(rows, []string{
					fmt.Sprintf("%v", event["type"]),
					fmt.Sprintf("%v", event["reason"]),
					fmt.Sprintf("%v", event["message"]),
					fmt.Sprintf("%v", event["age"]),
				})
			}
			printTable(headers, rows)
		}
	}
	fmt.Println()
}

func handleAPIDiff(filename string) {
	// Read YAML file
	data, err := os.ReadFile(filename)
	if err != nil {
		output.PrintError(output.ErrInvalidInput, "Failed to read file: "+err.Error())
		return
	}

	// Parse YAML
	var resource map[string]interface{}
	if err := yaml.Unmarshal(data, &resource); err != nil {
		output.PrintError(output.ErrInvalidYAML, "Failed to parse YAML: "+err.Error())
		return
	}

	metadata := resource["metadata"].(map[string]interface{})
	name := metadata["name"].(string)

	// Get current from server
	response, err := apiClient.Get(fmt.Sprintf("/api/v1/namespaces/%s/apis/%s", namespace, name))
	if err != nil {
		output.PrintError(output.ErrServerError, err.Error())
		return
	}

	if response.StatusCode == 404 {
		fmt.Printf("📄 Resource '%s' does not exist on server\n", name)
		fmt.Printf("⚠️  Applying this file will CREATE a new resource\n\n")
		return
	}

	var current map[string]interface{}
	if err := response.JSON(&current); err != nil {
		output.PrintError(output.ErrInvalidInput, "Failed to parse response")
		return
	}

	fileSpec := resource["spec"].(map[string]interface{})
	currentSpec := current["spec"].(map[string]interface{})

	fmt.Printf("\n📊 Diff: %s\n", name)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	hasDiff := false
	for key := range fileSpec {
		fileVal := fmt.Sprintf("%v", fileSpec[key])
		currVal := fmt.Sprintf("%v", currentSpec[key])
		if fileVal != currVal {
			hasDiff = true
			fmt.Printf("  %s\n", key)
			fmt.Printf("    - %v (current)\n", currVal)
			fmt.Printf("    + %v (new)\n", fileVal)
		}
	}

	if !hasDiff {
		fmt.Println("✅ No differences - file matches server state")
	}
	fmt.Println()
}

func init() {
	APICmd.AddCommand(APICreateCmd)
	APICmd.AddCommand(APIListCmd)
	APICmd.AddCommand(APIGetCmd)
	APICmd.AddCommand(APIUpdateCmd)
	APICmd.AddCommand(APIDeleteCmd)
	APICmd.AddCommand(APIApplyCmd)
	APICmd.AddCommand(APIDescribeCmd)
	APICmd.AddCommand(APIDiffCmd)

	// Apply command flags
	APIApplyCmd.Flags().StringP("filename", "f", "", "YAML file path (required)")
	APIApplyCmd.Flags().BoolP("force", "", false, "Skip waiting for reconciliation")
	APIApplyCmd.Flags().Duration("timeout", 30*time.Second, "Timeout for reconciliation")
	APIApplyCmd.MarkFlagRequired("filename")

	// Diff command flags
	APIDiffCmd.Flags().StringP("filename", "f", "", "YAML file path (required)")
	APIDiffCmd.MarkFlagRequired("filename")
}
