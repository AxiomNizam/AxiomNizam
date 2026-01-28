package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show API server status",
	Long:  "Display the status of the AxiomNizam API server and connected resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleStatus()
	},
}

func handleStatus() error {
	if err := validateServerConnection(); err != nil {
		printWarningMessage("Not connected to server")
		fmt.Println("\nℹ️  Run 'axiomnizamctl login' to connect")
		return nil
	}

	fmt.Println("\n📊 AxiomNizam Status")
	fmt.Println("─────────────────────────────────────")

	context := configManager.GetCurrentContext()
	if context == nil {
		return NewCommandError(ErrConfigError, "No context configured")
	}

	fmt.Printf("✓ Connected to: %s\n", context.Cluster.Server)
	fmt.Printf("✓ Context: %s\n", context.Name)
	fmt.Printf("✓ User: %s\n", context.User)
	fmt.Printf("✓ Namespace: %s\n", context.Namespace)

	// Try to check server health
	if apiClient != nil {
		response, err := apiClient.GetSimple("/api/v1/health")
		if err == nil && response.StatusCode == 200 {
			var health struct {
				Status  string `json:"status"`
				Version string `json:"version"`
			}
			if response.JSON(&health) == nil {
				fmt.Printf("✓ Server Status: %s\n", health.Status)
				if health.Version != "" {
					fmt.Printf("✓ Server Version: %s\n", health.Version)
				}
			}
		}

		// Check distributed status
		distResp, err := apiClient.GetSimple("/distributed")
		if err == nil && distResp.StatusCode == 200 {
			var distStatus struct {
				Data struct {
					IsDistributed bool     `json:"is_distributed"`
					Members       []string `json:"members"`
					Healthy       bool     `json:"healthy"`
				} `json:"data"`
			}
			if distResp.JSON(&distStatus) == nil {
				if distStatus.Data.IsDistributed {
					fmt.Printf("✓ Distributed Mode: ENABLED (%d members)\n", len(distStatus.Data.Members))
				} else {
					fmt.Println("✓ Distributed Mode: DISABLED (single instance)")
				}
			}
		}
	}

	fmt.Println("\n💡 Ready to use. Try 'axiomnizamctl api list'")
	return nil
}
