package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mlPipelineCmd = &cobra.Command{
	Use:   "ml",
	Short: "Manage ML pipelines and model deployments",
}

var mlListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ML pipelines",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/ml-pipelines")
	},
}

var mlGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get pipeline details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(fmt.Sprintf("/api/v1/ml-pipelines/%s", args[0]))
	},
}

var mlCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an ML pipeline",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		return postAndPrint("/api/v1/ml-pipelines", map[string]interface{}{"name": name})
	},
}

var mlRunCmd = &cobra.Command{
	Use:   "run [name]",
	Short: "Trigger a pipeline run",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return postAndPrint(fmt.Sprintf("/api/v1/ml-pipelines/%s/run", args[0]), nil)
	},
}

var mlDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a pipeline",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteAndPrint(fmt.Sprintf("/api/v1/ml-pipelines/%s", args[0]))
	},
}

var mlDeployCmd = &cobra.Command{
	Use:   "deployments",
	Short: "List model deployments",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/model-deployments")
	},
}

func init() {
	mlCreateCmd.Flags().String("name", "", "Pipeline name")
	mlPipelineCmd.AddCommand(mlListCmd, mlGetCmd, mlCreateCmd, mlRunCmd, mlDeleteCmd, mlDeployCmd)
	RootCmd.AddCommand(mlPipelineCmd)
}
