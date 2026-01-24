package main

import (
	"fmt"

	"example.com/axiomnizam/internal/output"
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "View, switch contexts, and manage clusters",
}

var ConfigViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		handleConfigView()
	},
}

var ConfigCurrentContextCmd = &cobra.Command{
	Use:   "current-context",
	Short: "Show current context",
	Run: func(cmd *cobra.Command, args []string) {
		handleConfigCurrentContext()
	},
}

var ConfigUseContextCmd = &cobra.Command{
	Use:   "use-context [context-name]",
	Short: "Switch to a different context",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handleConfigUseContext(args[0])
	},
}

var ConfigGetClustersCmd = &cobra.Command{
	Use:   "get-clusters",
	Short: "List all available clusters",
	Run: func(cmd *cobra.Command, args []string) {
		handleConfigGetClusters()
	},
}

func handleConfigView() {
	config := configManager.GetCurrentContext()
	if config == nil {
		fmt.Println("❌ No context configured")
		return
	}

	formatter := output.NewFormatter(outputFormat)
	formatter.Print(config)
}

func handleConfigCurrentContext() {
	config := configManager.GetCurrentContext()
	if config == nil {
		fmt.Println("❌ No context configured")
		return
	}

	fmt.Printf("Current context: %s\n", config.Name)
	fmt.Printf("Cluster: %s\n", config.Cluster.Name)
	fmt.Printf("Server: %s\n", config.Cluster.Server)
}

func handleConfigUseContext(contextName string) {
	contexts := configManager.ListContexts()
	found := false
	for _, ctx := range contexts {
		if ctx.Name == contextName {
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("❌ Context '%s' not found\n", contextName)
		return
	}

	if err := configManager.SetCurrentContext(contextName); err != nil {
		fmt.Printf("❌ Failed to switch context: %v\n", err)
		return
	}

	fmt.Printf("✅ Switched to context '%s'\n", contextName)
}

func handleConfigGetClusters() {
	contexts := configManager.ListContexts()

	headers := []string{"NAME", "CLUSTER", "SERVER"}
	rows := make([][]string, 0)

	for _, ctx := range contexts {
		rows = append(rows, []string{
			ctx.Name,
			ctx.Cluster.Name,
			ctx.Cluster.Server,
		})
	}

	if len(rows) == 0 {
		fmt.Println("No clusters configured")
		return
	}

	printTable(headers, rows)
}

func init() {
	ConfigCmd.AddCommand(ConfigViewCmd)
	ConfigCmd.AddCommand(ConfigCurrentContextCmd)
	ConfigCmd.AddCommand(ConfigUseContextCmd)
	ConfigCmd.AddCommand(ConfigGetClustersCmd)
}
