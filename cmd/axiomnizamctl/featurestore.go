package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var featureStoreCmd = &cobra.Command{
	Use:   "feature-store",
	Short: "Manage ML feature store",
}

var fsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List feature groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/feature-groups")
	},
}

var fsGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get feature group details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(fmt.Sprintf("/api/v1/feature-groups/%s", args[0]))
	},
}

var fsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a feature group",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		source, _ := cmd.Flags().GetString("source")
		return postAndPrint("/api/v1/feature-groups", map[string]interface{}{
			"name": name, "source": source,
		})
	},
}

var fsDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a feature group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteAndPrint(fmt.Sprintf("/api/v1/feature-groups/%s", args[0]))
	},
}

var fsMaterializeCmd = &cobra.Command{
	Use:   "materialize [name]",
	Short: "Trigger feature materialization",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return postAndPrint(fmt.Sprintf("/api/v1/feature-groups/%s/materialize", args[0]), nil)
	},
}

func init() {
	fsCreateCmd.Flags().String("name", "", "Feature group name")
	fsCreateCmd.Flags().String("source", "", "Data source")
	featureStoreCmd.AddCommand(fsListCmd, fsGetCmd, fsCreateCmd, fsDeleteCmd, fsMaterializeCmd)
	RootCmd.AddCommand(featureStoreCmd)
}
