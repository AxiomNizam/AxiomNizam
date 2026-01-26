// AxiomNizam CLI Command Tree - Cobra Skeleton
// This file documents the complete command hierarchy and flow

package main

/*
COMMAND TREE STRUCTURE:

axiomnizamctl
│
├─ Auth Commands
│  ├─ login [url]                           # Authenticate with server
│  │  ├─ --username, -u                     # Username
│  │  ├─ --password, -p                     # Password (interactive if omitted)
│  │  ├─ --api-key                          # API key authentication
│  │  ├─ --method                           # Auth method: password|api-key
│  │  ├─ --context                          # Context name to save
│  │  ├─ --server                           # Server URL
│  │  └─ --insecure-skip-tls-verify        # Skip TLS verification
│  │
│  ├─ logout                                # Clear authentication
│  │
│  └─ current-user                          # Show logged-in user info
│
├─ API Resource Commands (kubectl-style)
│  ├─ api create                            # Interactive API creation
│  ├─ api list [options]                    # List APIs in namespace
│  ├─ api get <name>                        # Get specific API
│  ├─ api describe <name>                   # Detailed API info + events
│  ├─ api apply -f file.yaml [options]      # Apply from YAML (triggers reconciliation)
│  │  ├─ -f, --filename                     # Path to YAML file
│  │  ├─ --dry-run                          # Show what would be applied
│  │  ├─ --force                            # Skip reconciliation wait
│  │  └─ --timeout                          # Reconciliation timeout
│  │
│  ├─ api update <name>                     # Update single field
│  ├─ api delete <name>                     # Delete resource
│  ├─ api diff -f file.yaml                 # Show differences
│  └─ api watch [name]                      # Watch for changes
│
├─ Policy Commands
│  ├─ policy create                         # Create new policy
│  ├─ policy list                           # List all policies
│  ├─ policy get <name>                     # Get policy details
│  ├─ policy apply -f file.yaml             # Apply policy from YAML
│  ├─ policy delete <name>                  # Delete policy
│  └─ policy describe <name>                # Show detailed info
│
├─ Workflow Commands
│  ├─ workflow create                       # Create workflow
│  ├─ workflow list                         # List workflows
│  ├─ workflow get <name>                   # Get workflow
│  ├─ workflow apply -f file.yaml           # Apply workflow
│  ├─ workflow delete <name>                # Delete workflow
│  ├─ workflow run <name> [params]          # Trigger workflow execution
│  ├─ workflow describe <name>              # Show details
│  └─ workflow logs <name>                  # Show execution logs
│
├─ DataSource Commands
│  ├─ datasource create                     # Create datasource
│  ├─ datasource list                       # List datasources
│  ├─ datasource get <name>                 # Get datasource
│  ├─ datasource apply -f file.yaml         # Apply from YAML
│  ├─ datasource delete <name>              # Delete datasource
│  ├─ datasource test <name>                # Test connection
│  └─ datasource describe <name>            # Show details
│
├─ Job Commands
│  ├─ job list                              # List jobs
│  ├─ job get <name>                        # Get job details
│  ├─ job logs <name>                       # Show job logs
│  ├─ job describe <name>                   # Show detailed info
│  └─ job delete <name>                     # Delete job
│
├─ Configuration Commands (kubeconfig-style)
│  ├─ config view [options]                 # Display merged kubeconfig
│  │  ├─ --flatten                          # Show flattened output
│  │  └─ --show-token                       # Display token (security warning)
│  │
│  ├─ config current-context                # Show current context name
│  ├─ config use-context <context>          # Switch to context
│  ├─ config get-clusters                   # List all clusters
│  ├─ config set-context <name> [options]   # Create/update context
│  │  ├─ --cluster                          # Cluster name
│  │  ├─ --user                             # User name
│  │  └─ --namespace                        # Default namespace
│  │
│  ├─ config set-cluster <name> [options]   # Configure cluster
│  │  ├─ --server                           # Server URL (required)
│  │  ├─ --certificate-authority            # Path to CA cert
│  │  └─ --insecure-skip-tls-verify        # Skip TLS verification
│  │
│  ├─ config delete-context <name>          # Remove context
│  └─ config rename-context <old> <new>     # Rename context
│
├─ Status & Monitoring
│  ├─ status                                # Show API server status
│  └─ events [options]                      # Display recent events
│     ├─ --limit                            # Number of events to show
│     ├─ --sort                             # Sort order
│     └─ --resource                         # Filter by resource type
│
└─ Utility Commands
   ├─ version                               # Display CLI version
   ├─ completion <shell>                    # Generate shell completion
   │  └─ bash|zsh|fish|powershell          # Shell type
   │
   └─ help [command]                        # Show help


GLOBAL FLAGS (all commands):
  --kubeconfig string                       # Path to kubeconfig file
  --context string                          # Context to use (overrides current)
  --namespace string                        # Default namespace (default: "default")
  --output, -o string                       # Output format: table|json|yaml|wide
  --verbose                                 # Enable verbose output
  --dry-run                                 # Preview without applying
  -h, --help                                # Show help
  --version                                 # Show version


AUTHENTICATION FLOW:
1. User runs: axiomnizamctl login https://api.example.com
2. Prompts for credentials (username/password or API key)
3. Sends credentials to /api/v1/auth/login
4. Receives JWT token
5. Saves token to ~/.axiomnizam/token (secure: 0600)
6. Saves context to ~/.axiomnizam/config (no secrets)
7. All subsequent requests include: Authorization: Bearer <token>


KUBECONFIG-STYLE CONFIG:
~/.axiomnizam/
├─ config                                  # Contexts, clusters, users (no secrets)
├─ token                                   # JWT token (chmod 0600)
└─ ca.crt                                  # CA certificate (optional)

Format:
contexts:
  - name: production
    cluster:
      server: https://api.prod.example.com
      insecure-skip-tls-verify: false
    user: prod-user
    namespace: production
  - name: staging
    cluster:
      server: https://api.staging.example.com
    user: staging-user
    namespace: staging


APPLY → CONTROLLER RECONCILE FLOW:

1. CLI Validation:
   - Parse YAML file
   - Extract metadata (kind, name, namespace)
   - Validate against schema

2. API Server:
   - Send POST /api/v1/namespaces/{ns}/{kind}s
   - Server validates and stores in etcd
   - Returns resource with generation number
   - Generation indicates desired state version

3. Controller Detection:
   - Informer watches for new resources
   - Detects metadata.generation change
   - Enqueues work item in workqueue

4. Resource Controller:
   - Dequeues work item (namespace, name)
   - Fetches latest resource from store
   - Passes to Reconciler with context

5. Reconciliation:
   - Compare desired spec vs actual state
   - Execute actions (create, update, delete)
   - Reconcile until conditions are met

6. Status Update:
   - Update status.phase (Pending, Ready, Failed)
   - Add conditions with reasons
   - Store observedGeneration

7. CLI Feedback:
   - Poll status endpoint while waiting
   - Display reconciliation progress
   - Return success/error to user


EXAMPLE WORKFLOWS:

# Basic login and use
$ axiomnizamctl login
🔐 AxiomNizam Login
Server URL: https://api.example.com
Username: admin
Password: ****
✅ Authentication successful!

# Switch context
$ axiomnizamctl config use-context staging
✅ Switched to context 'staging'

# Apply API with reconciliation
$ axiomnizamctl api apply -f api.yaml
📖 Reading resource from api.yaml...
📦 Resource: API/users-api (generation: 1)
📡 Sending to API server...
✅ Applied successfully!
🔄 Controller Status: Pending
⏳ Waiting for controller reconciliation...
✅ Reconciliation complete: Ready

# Dry-run to preview
$ axiomnizamctl api apply -f api.yaml --dry-run
🔍 Dry-run mode: showing what would be applied
📋 Apply Plan
─────────────
Kind: API
Name: users-api
Namespace: default
Spec:
  database: postgresql
  table: users
  ...

# Watch resource status
$ axiomnizamctl api describe users-api
📋 API: users-api
Status: Ready
Conditions:
  - Initialized (True)
  - Ready (True)
  - Updated (2024-01-26 15:30:45)
Events:
  - Created (2024-01-26 15:30:00)
  - ReconciliationStarted (2024-01-26 15:30:10)
  - ReconciliationCompleted (2024-01-26 15:30:45)
*/
