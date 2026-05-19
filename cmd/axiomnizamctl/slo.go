package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sloCmd = &cobra.Command{
	Use:   "slo",
	Short: "Manage Service Level Objectives",
}

var sloListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all SLOs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/slos")
	},
}

var sloGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get SLO details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(fmt.Sprintf("/api/v1/slos/%s", args[0]))
	},
}

var sloCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new SLO",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		target, _ := cmd.Flags().GetFloat64("target")
		window, _ := cmd.Flags().GetString("window")
		return postAndPrint("/api/v1/slos", map[string]interface{}{
			"name": name, "target": target, "window": window,
		})
	},
}

var sloDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete an SLO",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteAndPrint(fmt.Sprintf("/api/v1/slos/%s", args[0]))
	},
}

var sloErrorBudgetCmd = &cobra.Command{
	Use:   "error-budget [name]",
	Short: "Show error budget for an SLO",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(fmt.Sprintf("/api/v1/slos/%s/error-budget", args[0]))
	},
}

func init() {
	sloCreateCmd.Flags().String("name", "", "SLO name")
	sloCreateCmd.Flags().Float64("target", 99.9, "Target percentage (e.g. 99.9)")
	sloCreateCmd.Flags().String("window", "30d", "Time window (e.g. 30d, 7d)")
	sloCmd.AddCommand(sloListCmd, sloGetCmd, sloCreateCmd, sloDeleteCmd, sloErrorBudgetCmd)
	RootCmd.AddCommand(sloCmd)
}
