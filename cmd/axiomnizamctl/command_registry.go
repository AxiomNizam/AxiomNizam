package main

import "github.com/spf13/cobra"

func bindPersistentFlags(root *cobra.Command) {
	root.PersistentFlags().StringVar(&outputFormat, "output", "table", "Output format: table, json, yaml, wide")
	root.PersistentFlags().StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	root.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (default: ~/.axiomnizam/config)")
	root.PersistentFlags().StringVar(&contextName, "context", "", "Context to use (overrides current context)")
	root.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	root.PersistentFlags().BoolVar(&dry, "dry-run", false, "Show what would be done without making changes")
}

func registerRootCommands(root *cobra.Command) {
	for _, cmd := range authCommands() {
		root.AddCommand(cmd)
	}
	for _, cmd := range resourceCommands() {
		root.AddCommand(cmd)
	}
	for _, cmd := range platformCommands() {
		root.AddCommand(cmd)
	}
	for _, cmd := range integrationCommands() {
		root.AddCommand(cmd)
	}
	for _, cmd := range adminCommands() {
		root.AddCommand(cmd)
	}
	for _, cmd := range utilityCommands() {
		root.AddCommand(cmd)
	}
}

func authCommands() []*cobra.Command {
	return []*cobra.Command{loginCmd, logoutCmd, currentUserCmd}
}

func resourceCommands() []*cobra.Command {
	return []*cobra.Command{APICmd, PolicyCmd, WorkflowCmd, DataSourceCmd, JobCmd}
}

func platformCommands() []*cobra.Command {
	return []*cobra.Command{ApiBankCmd, MeshCmd, TenantCmd, RBACCmd, EventBusCmd, WebhookCmd, StreamCmd, ExportCmd, BulkCmd, VersioningCmd, TraceCmd, LineageAPICmd, IncidentCmd}
}

func integrationCommands() []*cobra.Command {
	return []*cobra.Command{healthCmd, alertsCmd, metricsCmd, catalogCmd, complianceCmd, qualityCmd, lineageCmd}
}

func adminCommands() []*cobra.Command {
	return []*cobra.Command{ConfigCmd, StatusCmd, EventsCmd, DiffCmd, CertCmd}
}

func utilityCommands() []*cobra.Command {
	return []*cobra.Command{versionCmd, completionCmd, waitCmd, scanCmd, discoverCmd}
}
