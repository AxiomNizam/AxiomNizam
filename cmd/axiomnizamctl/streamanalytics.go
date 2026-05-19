package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var streamAnalyticsCmd = &cobra.Command{
	Use:   "stream-analytics",
	Short: "Manage stream processing jobs",
}

var saListCmd = &cobra.Command{
	Use:   "list",
	Short: "List stream analytics jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/stream-jobs")
	},
}

var saGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get stream job details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(fmt.Sprintf("/api/v1/stream-jobs/%s", args[0]))
	},
}

var saCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a stream analytics job",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		query, _ := cmd.Flags().GetString("query")
		return postAndPrint("/api/v1/stream-jobs", map[string]interface{}{
			"name": name, "query": query,
		})
	},
}

var saStartCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Start a stream job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return postAndPrint(fmt.Sprintf("/api/v1/stream-jobs/%s/start", args[0]), nil)
	},
}

var saStopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a stream job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return postAndPrint(fmt.Sprintf("/api/v1/stream-jobs/%s/stop", args[0]), nil)
	},
}

var saDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a stream job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteAndPrint(fmt.Sprintf("/api/v1/stream-jobs/%s", args[0]))
	},
}

func init() {
	saCreateCmd.Flags().String("name", "", "Job name")
	saCreateCmd.Flags().String("query", "", "SQL query")
	streamAnalyticsCmd.AddCommand(saListCmd, saGetCmd, saCreateCmd, saStartCmd, saStopCmd, saDeleteCmd)
	RootCmd.AddCommand(streamAnalyticsCmd)
}
