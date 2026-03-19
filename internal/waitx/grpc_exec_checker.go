package waitx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// GRPCHealthChecker waits for gRPC health service SERVING status.
type GRPCHealthChecker struct {
	Target             string
	Service            string
	UseTLS             bool
	TLSServerName      string
	InsecureSkipVerify bool
	DialTimeout        time.Duration
	RequestTimeout     time.Duration
}

func (c GRPCHealthChecker) Name() string {
	return "grpc-health:" + strings.TrimSpace(c.Target)
}

func (c GRPCHealthChecker) Check(ctx context.Context) error {
	target := strings.TrimSpace(c.Target)
	if target == "" {
		return fmt.Errorf("grpc target is required")
	}

	dialTimeout := c.DialTimeout
	if dialTimeout <= 0 {
		dialTimeout = 5 * time.Second
	}
	requestTimeout := c.RequestTimeout
	if requestTimeout <= 0 {
		requestTimeout = 5 * time.Second
	}

	dialCtx, cancelDial := context.WithTimeout(ctx, dialTimeout)
	defer cancelDial()

	dialOptions := []grpc.DialOption{grpc.WithBlock()}
	if c.UseTLS {
		tlsCfg := &tls.Config{
			ServerName:         c.TLSServerName,
			InsecureSkipVerify: c.InsecureSkipVerify,
		}
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	} else {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	}

	conn, err := grpc.DialContext(dialCtx, target, dialOptions...)
	if err != nil {
		return fmt.Errorf("grpc dial failed: %w", err)
	}
	defer conn.Close()

	checkCtx, cancelCheck := context.WithTimeout(ctx, requestTimeout)
	defer cancelCheck()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(checkCtx, &grpc_health_v1.HealthCheckRequest{Service: c.Service})
	if err != nil {
		return fmt.Errorf("grpc health check failed: %w", err)
	}
	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("grpc health status is %s", resp.GetStatus().String())
	}

	return nil
}

// CommandChecker waits for a command to exit with expected code.
type CommandChecker struct {
	Command          []string
	ExpectedExitCode int
	WorkingDirectory string
}

func (c CommandChecker) Name() string {
	if len(c.Command) == 0 {
		return "exec:<empty>"
	}
	return "exec:" + c.Command[0]
}

func (c CommandChecker) Check(ctx context.Context) error {
	if len(c.Command) == 0 {
		return fmt.Errorf("command is required")
	}

	cmd := exec.CommandContext(ctx, c.Command[0], c.Command[1:]...)
	if strings.TrimSpace(c.WorkingDirectory) != "" {
		cmd.Dir = c.WorkingDirectory
	}

	output, err := cmd.CombinedOutput()
	if err == nil {
		if c.ExpectedExitCode == 0 {
			return nil
		}
		return fmt.Errorf("command exited 0, expected %d", c.ExpectedExitCode)
	}

	exitCode := extractExitCode(err)
	if exitCode == c.ExpectedExitCode {
		return nil
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed != "" {
		return fmt.Errorf("command exit code %d, expected %d: %s", exitCode, c.ExpectedExitCode, trimmed)
	}
	return fmt.Errorf("command exit code %d, expected %d", exitCode, c.ExpectedExitCode)
}

func extractExitCode(err error) int {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return -1
	}

	if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
		return status.ExitStatus()
	}
	return -1
}
