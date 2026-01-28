package main

import (
	"fmt"
	"os"
	"strings"
)

type ErrorCode int

const (
	ErrInvalidArgs  ErrorCode = 1
	ErrServerError  ErrorCode = 2
	ErrNotFound     ErrorCode = 3
	ErrUnauthorized ErrorCode = 4
	ErrConflict     ErrorCode = 5
	ErrInvalidInput ErrorCode = 6
	ErrFileNotFound ErrorCode = 7
	ErrYAMLError    ErrorCode = 8
	ErrNetwork      ErrorCode = 9
	ErrTimeout      ErrorCode = 10
	ErrConfigError  ErrorCode = 11
)

type CommandError struct {
	Code    ErrorCode
	Message string
	Details string
}

func (e *CommandError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

func NewCommandError(code ErrorCode, message string, details ...string) *CommandError {
	detail := ""
	if len(details) > 0 {
		detail = strings.Join(details, ": ")
	}
	return &CommandError{
		Code:    code,
		Message: message,
		Details: detail,
	}
}

func handleCommandError(err error) {
	if err == nil {
		return
	}

	cmdErr, ok := err.(*CommandError)
	if !ok {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(int(ErrServerError))
	}

	fmt.Fprintf(os.Stderr, "❌ %s\n", cmdErr.Error())
	if verbose {
		fmt.Fprintf(os.Stderr, "🔍 Error Code: %d\n", cmdErr.Code)
	}
	os.Exit(int(cmdErr.Code))
}

func validateServerConnection() error {
	if apiClient == nil {
		return NewCommandError(
			ErrConfigError,
			"Not connected to server",
			"Run 'axiomnizamctl login' first",
		)
	}

	if configManager == nil || configManager.GetCurrentContext() == nil {
		return NewCommandError(
			ErrConfigError,
			"No context configured",
			"Run 'axiomnizamctl config use-context <name>' to set a context",
		)
	}

	return nil
}

func validateNamespace() error {
	if namespace == "" {
		namespace = "default"
	}
	if !isValidNamespace(namespace) {
		return NewCommandError(
			ErrInvalidInput,
			fmt.Sprintf("Invalid namespace: %s", namespace),
			"Namespace must contain only lowercase letters, numbers, and hyphens",
		)
	}
	return nil
}

func isValidNamespace(ns string) bool {
	if ns == "" {
		return false
	}
	for _, ch := range ns {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return false
		}
	}
	return true
}

func validateResourceName(name string) error {
	if name == "" {
		return NewCommandError(ErrInvalidInput, "Resource name cannot be empty")
	}
	if len(name) > 63 {
		return NewCommandError(ErrInvalidInput, "Resource name too long (max 63 characters)")
	}
	if !isValidResourceName(name) {
		return NewCommandError(
			ErrInvalidInput,
			fmt.Sprintf("Invalid resource name: %s", name),
			"Must start with lowercase letter, contain only lowercase letters, numbers, and hyphens",
		)
	}
	return nil
}

func isValidResourceName(name string) bool {
	if len(name) == 0 {
		return false
	}
	if name[0] < 'a' || name[0] > 'z' {
		return false
	}
	for i, ch := range name {
		if i > 0 && ch == '-' && (i == len(name)-1) {
			return false
		}
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return false
		}
	}
	return true
}

func printSuccessMessage(message string, details ...string) {
	fmt.Printf("✅ %s\n", message)
	for _, detail := range details {
		fmt.Printf("   %s\n", detail)
	}
}

func printWarningMessage(message string, details ...string) {
	fmt.Printf("⚠️  %s\n", message)
	for _, detail := range details {
		fmt.Printf("   %s\n", detail)
	}
}

func printInfoMessage(message string, details ...string) {
	fmt.Printf("ℹ️  %s\n", message)
	for _, detail := range details {
		fmt.Printf("   %s\n", detail)
	}
}

func printErrorMessage(message string, details ...string) {
	fmt.Printf("❌ %s\n", message)
	for _, detail := range details {
		fmt.Printf("   %s\n", detail)
	}
}
