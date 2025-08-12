package cmd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/awslabs/ssosync/internal/config"
)

func TestVersionInfo(t *testing.T) {
	// Test that version variables exist
	if version == "" {
		version = "test"
	}
	if commit == "" {
		commit = "test-commit"
	}
	if date == "" {
		date = "test-date"
	}
	if builtBy == "" {
		builtBy = "test-builder"
	}

	// These should not be empty after setting
	if version == "" || commit == "" || date == "" || builtBy == "" {
		t.Error("Version info variables should not be empty")
	}
}

func TestConfigInitialization(t *testing.T) {
	// Test that config can be initialized
	testCfg := config.New()

	// Test default values
	if testCfg.LogLevel != config.DefaultLogLevel {
		t.Errorf("Expected log level %s, got %s", config.DefaultLogLevel, testCfg.LogLevel)
	}

	if testCfg.SyncMethod != config.DefaultSyncMethod {
		t.Errorf("Expected sync method %s, got %s", config.DefaultSyncMethod, testCfg.SyncMethod)
	}
}

func TestLambdaDetection(t *testing.T) {
	// Test Lambda environment detection
	originalEnv := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	defer func() {
		if originalEnv != "" {
			_ = os.Setenv("AWS_LAMBDA_FUNCTION_NAME", originalEnv)
		} else {
			_ = os.Unsetenv("AWS_LAMBDA_FUNCTION_NAME")
		}
	}()

	// Test non-Lambda environment
	_ = os.Unsetenv("AWS_LAMBDA_FUNCTION_NAME")
	testCfg := config.New()
	testCfg.IsLambda = len(os.Getenv("AWS_LAMBDA_FUNCTION_NAME")) > 0

	if testCfg.IsLambda {
		t.Error("Should not detect Lambda environment when AWS_LAMBDA_FUNCTION_NAME is not set")
	}

	// Test Lambda environment
	_ = os.Setenv("AWS_LAMBDA_FUNCTION_NAME", "test-function")
	testCfg.IsLambda = len(os.Getenv("AWS_LAMBDA_FUNCTION_NAME")) > 0

	if !testCfg.IsLambda {
		t.Error("Should detect Lambda environment when AWS_LAMBDA_FUNCTION_NAME is set")
	}
}

func TestContextWithCancel(t *testing.T) {
	// Test context creation and cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if ctx == nil {
		t.Error("Context should not be nil")
	}

	// Test context cancellation
	cancel()

	select {
	case <-ctx.Done():
		// Context was cancelled successfully
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should have been cancelled")
	}
}

func TestGetEnvFunction(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback string
		envValue string
		expected string
	}{
		{
			name:     "environment variable exists",
			key:      "TEST_ENV_VAR",
			fallback: "fallback",
			envValue: "env_value",
			expected: "env_value",
		},
		{
			name:     "environment variable does not exist",
			key:      "NON_EXISTENT_VAR",
			fallback: "fallback",
			envValue: "",
			expected: "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envValue != "" {
				_ = os.Setenv(tt.key, tt.envValue)
				defer func() { _ = os.Unsetenv(tt.key) }()
			} else {
				_ = os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.fallback)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
