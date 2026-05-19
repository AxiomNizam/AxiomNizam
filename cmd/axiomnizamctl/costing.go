package main

import (
	"github.com/spf13/cobra"
)

var costingCmd = &cobra.Command{
	Use:   "cost",
	Short: "Manage cost policies and usage tracking",
}

var costingPolicyListCmd = &cobra.Command{
	Use:   "policy list",
	Short: "List cost policies",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/cost-policies")
	},
}

var costingPolicyCreateCmd = &cobra.Command{
	Use:   "policy create",
	Short: "Create a cost policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		limit, _ := cmd.Flags().GetFloat64("limit")
		return postAndPrint("/api/v1/cost-policies", map[string]interface{}{
			"name": name, "limit": limit,
		})
	},
}

var costingUsageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Show usage records",
	RunE: func(cmd *cobra.Command, args []string) error {
		tenant, _ := cmd.Flags().GetString("tenant")
		url := "/api/v1/usage-records"
		if tenant != "" {
			url += "?tenant=" + tenant
		}
		return getAndPrint(url)
	},
}

var costingReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate cost report",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/cost-policies/report")
	},
}

func init() {
	costingPolicyCreateCmd.Flags().String("name", "", "Policy name")
	costingPolicyCreateCmd.Flags().Float64("limit", 0, "Cost limit")
	costingUsageCmd.Flags().String("tenant", "", "Filter by tenant")
	costingCmd.AddCommand(costingPolicyListCmd, costingPolicyCreateCmd, costingUsageCmd, costingReportCmd)
	RootCmd.AddCommand(costingCmd)
}
