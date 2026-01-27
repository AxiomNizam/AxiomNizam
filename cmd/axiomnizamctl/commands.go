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
	fmt.Println("\n📊 AxiomNizam Status")
	fmt.Println("─" + string([]rune{}))

	context := configManager.GetCurrentContext()
	if context == nil {
		fmt.Println("❌ Not configured. Run 'axiomnizamctl login'")
		return nil
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
	}

	fmt.Println("\n💡 Ready to use. Try 'axiomnizamctl api list'")
	return nil
}
