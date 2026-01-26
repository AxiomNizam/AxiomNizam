package main

import (
	"fmt"
	"os"

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

var EventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Display recent events",
	Long:  "Show recent events across all resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleEvents()
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate a shell completion script for axiomnizamctl.

Examples:
  # Install bash completion
  axiomnizamctl completion bash | sudo tee /etc/bash_completion.d/axiomnizamctl

  # Install zsh completion
  axiomnizamctl completion zsh | sudo tee /usr/share/zsh/site-functions/_axiomnizamctl

  # Install fish completion
  axiomnizamctl completion fish | sudo tee /usr/share/fish/vendor_completions.d/axiomnizamctl.fish`,
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	Args:      cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return RootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return RootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return RootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return RootCmd.GenPowerShellCompletion(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
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

func handleEvents() error {
	if apiClient == nil {
		return fmt.Errorf("not authenticated. Run 'axiomnizamctl login' first")
	}

	response, err := apiClient.GetSimple("/api/v1/events?limit=10&sort=-timestamp")
	if err != nil {
		return fmt.Errorf("failed to fetch events: %w", err)
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("failed to fetch events: %s", response.Status)
	}

	var events struct {
		Items []struct {
			Timestamp string `json:"timestamp"`
			Type      string `json:"type"`
			Reason    string `json:"reason"`
			Message   string `json:"message"`
			Involved  struct {
				Kind string `json:"kind"`
				Name string `json:"name"`
			} `json:"involvedObject"`
		} `json:"items"`
	}

	if err := response.JSON(&events); err != nil {
		return fmt.Errorf("failed to parse events: %w", err)
	}

	if len(events.Items) == 0 {
		fmt.Println("No recent events")
		return nil
	}

	fmt.Println("\n📋 Recent Events")
	fmt.Println("─" + string([]rune{}))

	headers := []string{"TIMESTAMP", "TYPE", "REASON", "OBJECT", "MESSAGE"}
	rows := make([][]string, 0)

	for _, event := range events.Items {
		obj := event.Involved.Kind + "/" + event.Involved.Name
		rows = append(rows, []string{
			event.Timestamp,
			event.Type,
			event.Reason,
			obj,
			event.Message,
		})
	}

	printTable(headers, rows)
	return nil
}
