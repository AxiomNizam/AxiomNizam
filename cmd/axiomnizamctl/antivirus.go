package main

import (
	"github.com/spf13/cobra"
)

var antivirusCmd = &cobra.Command{
	Use:   "antivirus",
	Short: "Manage antivirus scanning engine",
}

var antivirusStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show antivirus engine status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/antivirus/status")
	},
}

var antivirusScanCmd = &cobra.Command{
	Use:   "scan [file-path]",
	Short: "Submit a file for scanning",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return postAndPrint("/api/v1/antivirus/scan", map[string]string{"path": args[0]})
	},
}

var antivirusPatternsCmd = &cobra.Command{
	Use:   "patterns",
	Short: "List loaded detection patterns",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/antivirus/patterns")
	},
}

var antivirusMetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Show scanning metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint("/api/v1/antivirus/metrics")
	},
}

func init() {
	antivirusCmd.AddCommand(antivirusStatusCmd, antivirusScanCmd, antivirusPatternsCmd, antivirusMetricsCmd)
	RootCmd.AddCommand(antivirusCmd)
}
