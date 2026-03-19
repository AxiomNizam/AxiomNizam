package waitx

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ControlledWait4xWrapper runs wait4x with a restricted command allowlist.
type ControlledWait4xWrapper struct {
	BinaryPath string
	allowed    map[string]struct{}
}

// NewControlledWait4xWrapper creates a wrapper with safe defaults.
func NewControlledWait4xWrapper(binaryPath string) *ControlledWait4xWrapper {
	if strings.TrimSpace(binaryPath) == "" {
		binaryPath = "wait4x"
	}
	allowed := map[string]struct{}{
		"tcp":        {},
		"http":       {},
		"dns":        {},
		"exec":       {},
		"redis":      {},
		"mysql":      {},
		"postgresql": {},
		"mongodb":    {},
		"kafka":      {},
		"rabbitmq":   {},
		"influxdb":   {},
		"temporal":   {},
	}
	return &ControlledWait4xWrapper{BinaryPath: binaryPath, allowed: allowed}
}

// Allow adds command names to the wrapper allowlist.
func (w *ControlledWait4xWrapper) Allow(commands ...string) {
	if w.allowed == nil {
		w.allowed = make(map[string]struct{})
	}
	for _, cmd := range commands {
		normalized := strings.ToLower(strings.TrimSpace(cmd))
		if normalized != "" {
			w.allowed[normalized] = struct{}{}
		}
	}
}

// Run executes wait4x with allowlist enforcement.
func (w *ControlledWait4xWrapper) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("wait4x arguments are required")
	}

	subCommand := strings.ToLower(strings.TrimSpace(args[0]))
	if _, ok := w.allowed[subCommand]; !ok {
		return fmt.Errorf("subcommand %q is not allowed by controlled wrapper", subCommand)
	}

	cmd := exec.CommandContext(ctx, w.BinaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wait4x execution failed: %w", err)
	}
	return nil
}
