package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ratelimitCmd = &cobra.Command{
	Use:   "ratelimit",
	Short: "Manage rate limits and quotas",
}

var rlListCmd = &cobra.Command{
	Use:   "list",
	Short: "List rate limit quotas",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/quotas")
	},
}

var rlGetCmd = &cobra.Command{
	Use:   "get [user-id]",
	Short: "Get quota for a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(fmt.Sprintf("/api/v1/quotas/%s", args[0]))
	},
}

var rlSetCmd = &cobra.Command{
	Use:   "set [user-id]",
	Short: "Set quota for a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		return postAndPrint(fmt.Sprintf("/api/v1/quotas/%s", args[0]), map[string]interface{}{
			"limit": limit,
		})
	},
}

var rlResetCmd = &cobra.Command{
	Use:   "reset [user-id]",
	Short: "Reset quota for a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteAndPrint(fmt.Sprintf("/api/v1/quotas/%s", args[0]))
	},
}

var rlEndpointCmd = &cobra.Command{
	Use:   "endpoint",
	Short: "Set endpoint rate limit",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		limit, _ := cmd.Flags().GetInt("limit")
		return postAndPrint("/api/v1/quotas/endpoint", map[string]interface{}{
			"path": path, "limit": limit,
		})
	},
}

func init() {
	rlSetCmd.Flags().Int("limit", 100, "Request limit per window")
	rlEndpointCmd.Flags().String("path", "", "API endpoint path")
	rlEndpointCmd.Flags().Int("limit", 100, "Request limit per window")
	ratelimitCmd.AddCommand(rlListCmd, rlGetCmd, rlSetCmd, rlResetCmd, rlEndpointCmd)
	RootCmd.AddCommand(ratelimitCmd)
}
