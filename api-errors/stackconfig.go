package apierrors

import (
	"os"
	"strconv"
	"sync"
)

// StackTraceConfig holds configuration for stack trace collection.
type StackTraceConfig struct {
	// Enabled determines whether stack traces are collected at all.
	// When false, no stack traces are collected (zero overhead).
	// Default: true in development, false in production.
	Enabled bool

	// MaxDepth specifies the maximum number of stack frames to collect.
	// Lower values improve performance but provide less detail.
	// Default: 32 in development, 10 in production (when enabled).
	MaxDepth int

	// CollectFor5xxOnly determines whether to collect stack traces only for server errors (5xx).
	// When true, stack traces are collected only when the error code is >= 500.
	// When false, stack traces are collected for all errors (if Enabled is true).
	// Default: false (collect for all errors when enabled).
	CollectFor5xxOnly bool
}

var (
	// globalConfig is the global stack trace configuration.
	// Access should be done through GetStackTraceConfig() and SetStackTraceConfig().
	globalConfig = StackTraceConfig{
		Enabled:           true, // Default: enabled
		MaxDepth:          32,   // Default: full stack
		CollectFor5xxOnly: false,
	}

	// configMutex protects concurrent access to globalConfig.
	configMutex sync.RWMutex
)

// init initializes the stack trace configuration from environment variables.
func init() {
	loadConfigFromEnv()
}

// loadConfigFromEnv loads stack trace configuration from environment variables.
// Environment variables:
//   - STACKTRACE_ENABLED: "true" or "false" (default: "true")
//   - STACKTRACE_MAX_DEPTH: integer (default: 32)
//   - STACKTRACE_5XX_ONLY: "true" or "false" (default: "false")
//   - APP_ENV: "production", "staging", "development" (sets defaults)
func loadConfigFromEnv() {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Set environment-specific defaults first
	appEnv := os.Getenv("APP_ENV")
	switch appEnv {
	case "production":
		// Production: Disable stack traces by default for performance
		globalConfig.Enabled = false
		globalConfig.MaxDepth = 0
		globalConfig.CollectFor5xxOnly = true

	case "staging":
		// Staging: Enabled but limited depth
		globalConfig.Enabled = true
		globalConfig.MaxDepth = 10
		globalConfig.CollectFor5xxOnly = true

	case "development", "dev", "local":
		// Development: Full stack traces
		globalConfig.Enabled = true
		globalConfig.MaxDepth = 32
		globalConfig.CollectFor5xxOnly = false

	default:
		// Unknown environment: Conservative defaults (limited collection)
		globalConfig.Enabled = true
		globalConfig.MaxDepth = 10
		globalConfig.CollectFor5xxOnly = false
	}

	// Override with explicit environment variables (if set)
	if enabledStr := os.Getenv("STACKTRACE_ENABLED"); enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			globalConfig.Enabled = enabled
		}
	}

	if maxDepthStr := os.Getenv("STACKTRACE_MAX_DEPTH"); maxDepthStr != "" {
		if maxDepth, err := strconv.Atoi(maxDepthStr); err == nil && maxDepth >= 0 {
			globalConfig.MaxDepth = maxDepth
		}
	}

	if only5xxStr := os.Getenv("STACKTRACE_5XX_ONLY"); only5xxStr != "" {
		if only5xx, err := strconv.ParseBool(only5xxStr); err == nil {
			globalConfig.CollectFor5xxOnly = only5xx
		}
	}
}

// GetStackTraceConfig returns a copy of the current global stack trace configuration.
// This is thread-safe.
func GetStackTraceConfig() StackTraceConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}

// SetStackTraceConfig updates the global stack trace configuration.
// This is thread-safe.
//
// Example:
//
//	apierrors.SetStackTraceConfig(apierrors.StackTraceConfig{
//	    Enabled:           false,
//	    MaxDepth:          0,
//	    CollectFor5xxOnly: false,
//	})
func SetStackTraceConfig(config StackTraceConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = config
}

// DisableStackTraces is a convenience function to completely disable stack trace collection.
// Equivalent to SetStackTraceConfig with Enabled=false.
func DisableStackTraces() {
	SetStackTraceConfig(StackTraceConfig{
		Enabled:           false,
		MaxDepth:          0,
		CollectFor5xxOnly: false,
	})
}

// EnableStackTraces is a convenience function to enable stack trace collection with specified depth.
// If maxDepth is 0 or negative, defaults to 32.
func EnableStackTraces(maxDepth int) {
	if maxDepth <= 0 {
		maxDepth = 32
	}
	SetStackTraceConfig(StackTraceConfig{
		Enabled:           true,
		MaxDepth:          maxDepth,
		CollectFor5xxOnly: false,
	})
}

// EnableStackTracesFor5xxOnly enables stack trace collection only for server errors (5xx).
// This is useful in production when you want to capture stack traces for unexpected server errors
// but not for expected client errors (4xx).
func EnableStackTracesFor5xxOnly(maxDepth int) {
	if maxDepth <= 0 {
		maxDepth = 10
	}
	SetStackTraceConfig(StackTraceConfig{
		Enabled:           true,
		MaxDepth:          maxDepth,
		CollectFor5xxOnly: true,
	})
}

// shouldCollectStackTrace determines whether a stack trace should be collected
// for an error with the given HTTP status code.
// This is an internal helper function.
func shouldCollectStackTrace(statusCode int) bool {
	config := GetStackTraceConfig()

	// If stack traces are globally disabled, don't collect
	if !config.Enabled {
		return false
	}

	// If configured to collect only for 5xx errors, check the status code
	if config.CollectFor5xxOnly {
		return statusCode >= 500
	}

	// Otherwise, collect for all errors (when enabled)
	return true
}
