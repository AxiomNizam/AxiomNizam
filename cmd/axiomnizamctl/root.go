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
)

var RootCmd = &cobra.Command{
	Use:   "axiomnizamctl",
	Short: "AxiomNizam CLI - Control plane for data APIs",
	Long: `AxiomNizam is a Kubernetes-style control plane for managing APIs, policies, workflows, and data sources.

Usage:
  axiomnizamctl [command] [flags]

Examples:
  axiomnizamctl login
  axiomnizamctl api apply -f api.yaml
  axiomnizamctl policy list
  axiomnizamctl workflow run daily-etl`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		var err error

		// Initialize config manager
		configManager, err = client.NewConfigManager()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to load config: %v\n", err)
			return
		}

		// Initialize API client
		config := configManager.GetCurrentContext()
		if config != nil {
			apiClient = client.NewClient(config.Cluster.Server)
			token := configManager.GetToken()
			if token != "" {
				apiClient.SetToken(token)
			}
		}
	},
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
	RootCmd.PersistentFlags().StringVar(&outputFormat, "output", "table", "Output format: table, json, yaml, wide")
	RootCmd.PersistentFlags().StringVar(&namespace, "namespace", "default", "Kubernetes namespace")

	RootCmd.AddCommand(loginCmd)
	RootCmd.AddCommand(logoutCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(APICmd)
	RootCmd.AddCommand(PolicyCmd)
	RootCmd.AddCommand(WorkflowCmd)
	RootCmd.AddCommand(DataSourceCmd)
	RootCmd.AddCommand(JobCmd)
	RootCmd.AddCommand(ConfigCmd)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
