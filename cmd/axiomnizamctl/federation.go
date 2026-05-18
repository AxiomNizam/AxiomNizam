package main

import (
	"github.com/spf13/cobra"
)

var federationCmd = &cobra.Command{
	Use:   "federation",
	Short: "Manage federated queries across data sources",
}

var fedTableListCmd = &cobra.Command{
	Use:   "tables",
	Short: "List virtual tables",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/virtual-tables")
	},
}

var fedTableCreateCmd = &cobra.Command{
	Use:   "table create",
	Short: "Create a virtual table",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		source, _ := cmd.Flags().GetString("source")
		query, _ := cmd.Flags().GetString("query")
		return postAndPrint("/api/v1/virtual-tables", map[string]interface{}{
			"name": name, "source": source, "query": query,
		})
	},
}

var fedQueryCmd = &cobra.Command{
	Use:   "query [sql]",
	Short: "Execute a federated query",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return postAndPrint("/api/v1/federated-queries", map[string]interface{}{
			"sql": args[0],
		})
	},
}

var fedExplainCmd = &cobra.Command{
	Use:   "explain [sql]",
	Short: "Explain a federated query plan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return postAndPrint("/api/v1/federated-queries/explain", map[string]interface{}{
			"sql": args[0],
		})
	},
}

var fedQueriesCmd = &cobra.Command{
	Use:   "history",
	Short: "List federated query history",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/federated-queries")
	},
}

func init() {
	fedTableCreateCmd.Flags().String("name", "", "Table name")
	fedTableCreateCmd.Flags().String("source", "", "Data source")
	fedTableCreateCmd.Flags().String("query", "", "Source query")
	federationCmd.AddCommand(fedTableListCmd, fedTableCreateCmd, fedQueryCmd, fedExplainCmd, fedQueriesCmd)
	RootCmd.AddCommand(federationCmd)
}
