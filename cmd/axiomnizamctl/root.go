package main

import (
	"fmt"
	"os"

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
	Version: "1.0.0",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error

		// Skip config initialization for commands that don't need it
		if skipConfigInit(cmd.Name()) {
			return
		}

		// Initialize config manager
		configPath := kubeconfig
		if configPath == "" {
			configPath = client.DefaultConfigPath()
		}

		configManager, err = client.NewConfigManagerWithPath(configPath)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "⚠️  Config initialization: %v\n", err)
			}
			configManager, _ = client.NewConfigManager()
		}

		if err := configManager.Load(); err != nil && verbose {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to load config: %v\n", err)
		}

		// Initialize API client
		config := configManager.GetCurrentContext()
		if config != nil {
			apiClient = client.NewClient(config.Cluster.Server)
			token := configManager.GetToken()
			if token != "" {
				apiClient.SetToken(token)
			}

			// Override context if specified
			if contextName != "" && contextName != config.Name {
				if err := configManager.SetCurrentContext(contextName); err != nil && verbose {
					fmt.Fprintf(os.Stderr, "⚠️  Failed to switch context: %v\n", err)
				}
				config = configManager.GetCurrentContext()
				if config != nil {
					apiClient.SetBaseURL(config.Cluster.Server)
				}
			}
		} else if verbose {
			fmt.Fprintf(os.Stderr, "⚠️  No context configured\n")
		}
	},
}

// skipConfigInit returns true for commands that don't need config initialization
func skipConfigInit(cmdName string) bool {
	noConfigCmds := map[string]bool{
		"login":   true,
		"logout":  true,
		"version": true,
		"config":  true,
		"help":    true,
		"--help":  true,
		"-h":      true,
	}
	return noConfigCmds[cmdName]
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to AxiomNizam",
	Long:  "Authenticate with username and password, save token to ~/.axiomnizam/token",
	Run: func(cmd *cobra.Command, args []string) {
		handleLogin()
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from AxiomNizam",
	Long:  "Delete authentication token",
	Run: func(cmd *cobra.Command, args []string) {
		handleLogout()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("AxiomNizam CLI v1.0.0")
	},
}

func handleLogin() {
	fmt.Println("🔐 AxiomNizam Login")
	username := promptInput("Username")
	password := promptPassword("Password")

	if apiClient == nil {
		fmt.Println("❌ API server not configured")
		return
	}

	// Make login request
	loginPayload := map[string]string{
		"username": username,
		"password": password,
	}

	response, err := apiClient.Post("/api/v1/auth/login", loginPayload)
	if err != nil {
		fmt.Printf("❌ Login failed: %v\n", err)
		return
	}

	if response.StatusCode != 200 {
		fmt.Printf("❌ Login failed: %s\n", response.Status)
		return
	}

	// Extract token from response
	var result map[string]interface{}
	if err := response.JSON(&result); err != nil {
		fmt.Printf("❌ Failed to parse response: %v\n", err)
		return
	}

	token, ok := result["token"].(string)
	if !ok {
		fmt.Println("❌ No token in response")
		return
	}

	// Save token
	if err := configManager.SetToken(token); err != nil {
		fmt.Printf("❌ Failed to save token: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully logged in as %s\n", username)
}

func handleLogout() {
	if err := configManager.DeleteToken(); err != nil {
		fmt.Printf("❌ Failed to logout: %v\n", err)
		return
	}
	fmt.Println("✅ Successfully logged out")
}

func init() {
	// Global persistent flags
	RootCmd.PersistentFlags().StringVar(&outputFormat, "output", "table", "Output format: table, json, yaml, wide")
	RootCmd.PersistentFlags().StringVar(&namespace, "namespace", "default", "Kubernetes namespace")
	RootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (default: ~/.axiomnizam/config)")
	RootCmd.PersistentFlags().StringVar(&contextName, "context", "", "Context to use (overrides current context)")
	RootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	RootCmd.PersistentFlags().BoolVar(&dry, "dry-run", false, "Show what would be done without making changes")

	// Auth commands
	RootCmd.AddCommand(loginCmd)
	RootCmd.AddCommand(logoutCmd)
	RootCmd.AddCommand(currentUserCmd)

	// Resource commands
	RootCmd.AddCommand(APICmd)
	RootCmd.AddCommand(PolicyCmd)
	RootCmd.AddCommand(WorkflowCmd)
	RootCmd.AddCommand(DataSourceCmd)
	RootCmd.AddCommand(JobCmd)

	// Data platform commands
	RootCmd.AddCommand(ApiBankCmd)
	RootCmd.AddCommand(MeshCmd)

	// Integration & monitoring commands
	RootCmd.AddCommand(healthCmd)
	RootCmd.AddCommand(alertsCmd)
	RootCmd.AddCommand(metricsCmd)
	RootCmd.AddCommand(catalogCmd)
	RootCmd.AddCommand(complianceCmd)
	RootCmd.AddCommand(qualityCmd)
	RootCmd.AddCommand(lineageCmd)

	// Admin commands
	RootCmd.AddCommand(ConfigCmd)
	RootCmd.AddCommand(StatusCmd)
	RootCmd.AddCommand(EventsCmd)
	RootCmd.AddCommand(DiffCmd)

	// Utility commands
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(completionCmd)
}
