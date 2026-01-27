package main

import (
	"context"
	ctx "context"
	"fmt"
	"strings"

	"example.com/axiomnizam/internal/client"
	"github.com/spf13/cobra"
)

var (
	username     string
	password     string
	serverURL    string
	insecure     bool
	contextAlias string
	orgID        string
	apiKey       string
	loginMethod  string // "password" or "api-key"
)

var loginCmd = &cobra.Command{
	Use:   "login [server-url]",
	Short: "Authenticate with AxiomNizam server",
	Long: `Login to an AxiomNizam server and save authentication token.

Supports multiple authentication methods:
- Interactive login (username/password)
- API key authentication
- Token-based auth

Examples:
  axiomnizamctl login                                    # Interactive login to default server
  axiomnizamctl login https://api.example.com            # Login to specific server
  axiomnizamctl login --username admin --password secret # Non-interactive login
  axiomnizamctl login --api-key my-key --server https://api.example.com`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			serverURL = args[0]
		}
		return handleLogin()
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from AxiomNizam",
	Long:  "Remove authentication token and clear current context",
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleLogout()
	},
}

var currentUserCmd = &cobra.Command{
	Use:   "current-user",
	Short: "Show current logged-in user",
	Long:  "Display information about the currently authenticated user",
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleCurrentUser()
	},
}

func handleLogin() error {
	fmt.Println("\n🔐 AxiomNizam Login")
	fmt.Println(strings.Repeat("─", 40))

	// Determine server URL
	if serverURL == "" {
		// Use current context server if available
		if configManager != nil && configManager.GetCurrentContext() != nil {
			serverURL = configManager.GetCurrentContext().Cluster.Server
		} else {
			serverURL = promptInput("Server URL (default: http://localhost:8000)")
			if serverURL == "" {
				serverURL = "http://localhost:8000"
			}
		}
	}

	// Validate server URL
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}

	fmt.Printf("\n📍 Server: %s\n", serverURL)

	// Determine authentication method
	if apiKey != "" {
		return loginWithAPIKey(serverURL)
	}

	if loginMethod == "" {
		loginMethod = "password"
	}

	if loginMethod == "password" {
		return loginWithPassword(serverURL)
	}

	return fmt.Errorf("unsupported login method: %s", loginMethod)
}

func loginWithPassword(serverURL string) error {
	// Get credentials
	if username == "" {
		username = promptInput("Username")
	}
	if username == "" {
		return fmt.Errorf("username is required")
	}

	if password == "" {
		password = promptPassword("Password")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	// Create temporary client
	tempClient := client.NewClient(serverURL)
	tempClient.SetSkipTLSVerify(insecure)

	// Perform login
	loginReq := map[string]string{
		"username": username,
		"password": password,
	}

	response, err := tempClient.Post(context.Background(), "/api/v1/auth/login", loginReq)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("login failed: %s", response.Status)
	}

	// Parse response
	var result struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expiresAt,omitempty"`
		User      struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"user"`
	}

	if err := response.JSON(&result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Token == "" {
		return fmt.Errorf("no token in response")
	}

	// Save to config
	return saveLoginContext(serverURL, result.Token, username)
}

func loginWithAPIKey(serverURL string) error {
	if apiKey == "" {
		apiKey = promptPassword("API Key")
	}
	if apiKey == "" {
		return fmt.Errorf("API key is required")
	}

	// Create temporary client
	tempClient := client.NewClient(serverURL)
	tempClient.SetSkipTLSVerify(insecure)
	tempClient.SetToken(apiKey)

	// Verify API key
	response, err := tempClient.Get(context.Background(), "/api/v1/auth/verify", nil)
	if err != nil {
		return fmt.Errorf("API key verification failed: %w", err)
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("invalid API key: %s", response.Status)
	}

	// Save to config
	var result struct {
		User struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"user"`
	}
	response.JSON(&result)

	return saveLoginContext(serverURL, apiKey, result.User.Name)
}

func saveLoginContext(serverURL, token, user string) error {
	if configManager == nil {
		configManager = client.NewConfigManager()
	}

	// Determine context name
	ctxName := contextAlias
	if ctxName == "" {
		ctxName = "default"
	}

	// Save token
	if err := configManager.SetToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	// Create or update context
	newContext := &client.Context{
		Name: ctxName,
		Cluster: &client.ClusterInfo{
			Server:                serverURL,
			InsecureSkipTLSVerify: insecure,
		},
		User:      user,
		Namespace: "default",
	}

	if err := configManager.AddOrUpdateContext(newContext); err != nil {
		return fmt.Errorf("failed to save context: %w", err)
	}

	// Set as current context
	if err := configManager.SetCurrentContext(ctxName); err != nil {
		return fmt.Errorf("failed to set current context: %w", err)
	}

	// Success message
	fmt.Println("\n✅ Authentication successful!")
	fmt.Printf("   User: %s\n", user)
	fmt.Printf("   Server: %s\n", serverURL)
	fmt.Printf("   Context: %s\n", ctxName)
	fmt.Println("\n💡 Tip: Use 'axiomnizamctl config view' to see your configuration")

	return nil
}

func handleLogout() error {
	if configManager == nil {
		configManager = client.NewConfigManager()
	}

	if err := configManager.DeleteToken(); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	fmt.Println("✅ Successfully logged out")
	fmt.Println("💡 Tip: Run 'axiomnizamctl login' to authenticate again")

	return nil
}

func handleCurrentUser() error {
	token := configManager.GetToken()
	if token == "" {
		fmt.Println("❌ Not authenticated. Run 'axiomnizamctl login' first.")
		return nil
	}

	context := configManager.GetCurrentContext()
	if context == nil {
		fmt.Println("❌ No context configured")
		return nil
	}

	fmt.Println("\n👤 Current User")
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("   User: %s\n", context.User)
	fmt.Printf("   Context: %s\n", context.Name)
	fmt.Printf("   Server: %s\n", context.Cluster.Server)
	fmt.Printf("   Namespace: %s\n", context.Namespace)

	// Try to fetch user info from server
	if apiClient != nil {
		response, err := apiClient.Get(ctx.Background(), "/api/v1/auth/whoami", nil)
		if err == nil && response.StatusCode == 200 {
			var userInfo struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Email string `json:"email"`
				Role  string `json:"role"`
			}
			if response.JSON(&userInfo) == nil {
				if userInfo.Email != "" {
					fmt.Printf("   Email: %s\n", userInfo.Email)
				}
				if userInfo.Role != "" {
					fmt.Printf("   Role: %s\n", userInfo.Role)
				}
			}
		}
	}

	return nil
}

func init() {
	// Login command flags
	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication")
	loginCmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication (interactive if not provided)")
	loginCmd.Flags().StringVar(&apiKey, "api-key", "", "API key for authentication (alternative to password)")
	loginCmd.Flags().StringVar(&loginMethod, "method", "password", "Authentication method: password or api-key")
	loginCmd.Flags().StringVar(&contextAlias, "context", "", "Name to save this context as (default: default)")
	loginCmd.Flags().StringVar(&serverURL, "server", "", "Server URL (can also be provided as argument)")
	loginCmd.Flags().BoolVar(&insecure, "insecure-skip-tls-verify", false, "Skip TLS certificate verification")

	// Mark password flag as sensitive (hides from help if needed)
	loginCmd.Flags().Lookup("password").NoOptDefVal = "true"
	loginCmd.Flags().Lookup("api-key").NoOptDefVal = "true"
}
