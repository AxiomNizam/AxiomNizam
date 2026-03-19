package trivy

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type ExternalWrapper struct {
	BinaryPath string
}

func NewExternalWrapper(binaryPath string) *ExternalWrapper {
	if strings.TrimSpace(binaryPath) == "" {
		binaryPath = "trivy"
	}
	return &ExternalWrapper{BinaryPath: binaryPath}
}

func (w *ExternalWrapper) RunJSON(ctx context.Context, args []string) ([]byte, error) {
	if w == nil {
		return nil, fmt.Errorf("trivy wrapper is not configured")
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("trivy arguments are required")
	}

	cmd := exec.CommandContext(ctx, w.BinaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(output))
		if trimmed == "" {
			return nil, fmt.Errorf("trivy execution failed: %w", err)
		}
		return nil, fmt.Errorf("trivy execution failed: %w: %s", err, trimmed)
	}

	return output, nil
}
