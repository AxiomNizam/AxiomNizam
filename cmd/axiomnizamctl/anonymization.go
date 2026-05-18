package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var anonymizationCmd = &cobra.Command{
	Use:   "anonymize",
	Short: "Manage data anonymization policies",
}

var anonymListCmd = &cobra.Command{
	Use:   "list",
	Short: "List anonymization policies",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/anonymization-policies")
	},
}

var anonymGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get anonymization policy details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(fmt.Sprintf("/api/v1/anonymization-policies/%s", args[0]))
	},
}

var anonymCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an anonymization policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		table, _ := cmd.Flags().GetString("table")
		method, _ := cmd.Flags().GetString("method")
		return postAndPrint("/api/v1/anonymization-policies", map[string]interface{}{
			"name": name, "table": table, "method": method,
		})
	},
}

var anonymDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete an anonymization policy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteAndPrint(fmt.Sprintf("/api/v1/anonymization-policies/%s", args[0]))
	},
}

var anonymApplyCmd = &cobra.Command{
	Use:   "apply [name]",
	Short: "Apply anonymization to a table",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return postAndPrint(fmt.Sprintf("/api/v1/anonymization-policies/%s/apply", args[0]), nil)
	},
}

func init() {
	anonymCreateCmd.Flags().String("name", "", "Policy name")
	anonymCreateCmd.Flags().String("table", "", "Target table")
	anonymCreateCmd.Flags().String("method", "mask", "Anonymization method: mask, hash, synthetic")
	anonymizationCmd.AddCommand(anonymListCmd, anonymGetCmd, anonymCreateCmd, anonymDeleteCmd, anonymApplyCmd)
	RootCmd.AddCommand(anonymizationCmd)
}
