package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var schemaRegistryCmd = &cobra.Command{
	Use:   "schema",
	Short: "Manage schema registry",
}

var schemaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List schemas",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/schemas")
	},
}

var schemaGetCmd = &cobra.Command{
	Use:   "get [subject]",
	Short: "Get schema for a subject",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(fmt.Sprintf("/api/v1/schema-subjects/%s", args[0]))
	},
}

var schemaSubjectsCmd = &cobra.Command{
	Use:   "subjects",
	Short: "List all subjects",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/schema-subjects")
	},
}

var schemaVersionsCmd = &cobra.Command{
	Use:   "versions [subject]",
	Short: "List versions for a subject",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(fmt.Sprintf("/api/v1/schema-subjects/%s/versions", args[0]))
	},
}

var schemaCompatibilityCmd = &cobra.Command{
	Use:   "check [subject]",
	Short: "Check schema compatibility",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return postAndPrint(fmt.Sprintf("/api/v1/schema-subjects/%s/compatibility", args[0]), nil)
	},
}

func init() {
	schemaRegistryCmd.AddCommand(schemaListCmd, schemaGetCmd, schemaSubjectsCmd, schemaVersionsCmd, schemaCompatibilityCmd)
	RootCmd.AddCommand(schemaRegistryCmd)
}
