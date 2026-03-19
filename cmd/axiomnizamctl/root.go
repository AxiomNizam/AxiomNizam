package main

import (
	"fmt"

	"example.com/axiomnizam/internal/client"
	"github.com/spf13/cobra"
)

var (
	configManager *client.ConfigManager
	apiClient     *client.Client
	outputFormat  string
	namespace     string
	verbose       bool
	kubeconfig    string
	contextName   string
	dry           bool
)

var RootCmd = &cobra.Command{
	Use:   "axiomnizamctl",
	Short: "AxiomNizam CLI - Control plane for data APIs",
	Long: `AxiomNizam CLI - Kubernetes-style control plane for APIs, policies, workflows, and data sources.

AxiomNizam provides a kubectl-like interface for managing cloud-native data infrastructure.
It supports declarative configuration management, policy enforcement, and automated workflows.

Website: https://axiom-nizam.io
Docs: https://docs.axiom-nizam.io`,
	Version:       "1.0.0",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeCLIContext(cmd)
	},
}

// skipConfigInit returns true for commands that don't need config initialization
func skipConfigInit(cmdName string) bool {
	noConfigCmds := map[string]bool{
		"login":       true,
		"logout":      true,
		"version":     true,
		"completion":  true,
		"config":      true,
		"wait":        true,
		"scan":        true,
		"tcp":         true,
		"dns":         true,
		"http":        true,
		"grpc-health": true,
		"k8s-pod":     true,
		"image":       true,
		"fs":          true,
		"k8s":         true,
		"repo":        true,
		"mysql":       true,
		"postgresql":  true,
		"mongodb":     true,
		"redis":       true,
		"rabbitmq":    true,
		"kafka":       true,
		"influxdb":    true,
		"temporal":    true,
		"custom":      true,
		"external":    true,
		"help":        true,
		"--help":      true,
		"-h":          true,
	}
	return noConfigCmds[cmdName]
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("AxiomNizam CLI v1.0.0")
	},
}

func init() {
	bindPersistentFlags(RootCmd)
	registerRootCommands(RootCmd)
}
