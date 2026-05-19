package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var contractsCmd = &cobra.Command{
	Use:   "contract",
	Short: "Manage data contracts",
}

var contractsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List data contracts",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/data-contracts")
	},
}

var contractsGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get data contract details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(fmt.Sprintf("/api/v1/data-contracts/%s", args[0]))
	},
}

var contractsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a data contract",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		table, _ := cmd.Flags().GetString("table")
		return postAndPrint("/api/v1/data-contracts", map[string]interface{}{
			"name": name, "table": table,
		})
	},
}

var contractsDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a data contract",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteAndPrint(fmt.Sprintf("/api/v1/data-contracts/%s", args[0]))
	},
}

var contractsValidateCmd = &cobra.Command{
	Use:   "validate [name]",
	Short: "Validate data against contract",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return postAndPrint(fmt.Sprintf("/api/v1/data-contracts/%s/validate", args[0]), nil)
	},
}

func init() {
	contractsCreateCmd.Flags().String("name", "", "Contract name")
	contractsCreateCmd.Flags().String("table", "", "Target table")
	contractsCmd.AddCommand(contractsListCmd, contractsGetCmd, contractsCreateCmd, contractsDeleteCmd, contractsValidateCmd)
	RootCmd.AddCommand(contractsCmd)
}
