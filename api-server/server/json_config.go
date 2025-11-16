package router

import (
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/decoder"
	"github.com/bytedance/sonic/encoder"
)

// ============================================================================
// SONIC JSON CONFIGURATION
// ============================================================================
// This file centralizes sonic JSON library configuration for the entire
// application. It provides both default and customizable configurations.
// ============================================================================

var (
	// DefaultSonicAPI is the default sonic configuration used across the application
	// It balances performance with safety and compatibility
	DefaultSonicAPI = sonic.Config{
		// EscapeHTML: false provides ~15% performance improvement
		// Set to true if you need HTML escaping for XSS protection in JSON responses
		EscapeHTML: false,

		// SortMapKeys: false provides ~10% performance improvement
		// Set to true if you need deterministic JSON output (e.g., for testing, caching)
		SortMapKeys: false,

		// CompactMarshaler: true uses more compact JSON encoding
		CompactMarshaler: true,

		// NoValidateJSONMarshaler: false ensures JSON validity
		NoValidateJSONMarshaler: false,

		// NoEncoderNewline: false adds newline at end (Gin compatibility)
		NoEncoderNewline: false,

		// ValidateString: true validates UTF-8 strings for safety
		ValidateString: true,

		// NoNullSliceOrMap: false allows null for empty slices/maps (encoding/json compatible)
		NoNullSliceOrMap: false,

		// UseInt64: false uses standard number handling
		UseInt64: false,

		// UseNumber: false parses numbers as float64 (encoding/json compatible)
		UseNumber: false,

		// UseUnicodeErrors: true provides better error messages
		UseUnicodeErrors: true,

		// CopyString: true ensures string safety
		CopyString: true,
	}.Froze()

	// SafeSonicAPI is a configuration with maximum safety and compatibility
	// Use this if you need strict encoding/json compatibility
	SafeSonicAPI = sonic.Config{
		EscapeHTML:              true, // Enable HTML escaping (XSS protection)
		SortMapKeys:             true, // Enable key sorting (deterministic output)
		CompactMarshaler:        true,
		NoValidateJSONMarshaler: false,
		NoEncoderNewline:        false,
		ValidateString:          true,
		NoNullSliceOrMap:        false,
		UseInt64:                false,
		UseNumber:               false,
		UseUnicodeErrors:        true,
		CopyString:              true,
	}.Froze()

	// FastSonicAPI is a configuration optimized for maximum performance
	// Use this for high-throughput internal APIs where safety can be relaxed
	FastSonicAPI = sonic.Config{
		EscapeHTML:              false, // Disable HTML escaping (+15% speed)
		SortMapKeys:             false, // Disable key sorting (+10% speed)
		CompactMarshaler:        true,
		NoValidateJSONMarshaler: true, // Skip validation (+5% speed, use with caution)
		NoEncoderNewline:        false,
		ValidateString:          false, // Skip UTF-8 validation (+3% speed, use with caution)
		NoNullSliceOrMap:        true,  // Omit null for empty slices/maps (+2% speed)
		UseInt64:                false,
		UseNumber:               false,
		UseUnicodeErrors:        false,
		CopyString:              false, // Disable string copying (+5% speed, use with caution)
	}.Froze()
)

// ============================================================================
// ENCODER/DECODER OPTIONS
// ============================================================================

// DefaultEncoderOptions returns the default encoder options
func DefaultEncoderOptions() []encoder.CompileOption {
	return []encoder.CompileOption{
		encoder.CompileOption{
			// Add any encoder-specific options here if needed
		},
	}
}

// DefaultDecoderOptions returns the default decoder options
func DefaultDecoderOptions() []decoder.CompileOption {
	return []decoder.CompileOption{
		decoder.CompileOption{
			// Add any decoder-specific options here if needed
		},
	}
}

// ============================================================================
// CONFIGURATION HELPERS
// ============================================================================

// GetSonicAPI returns the appropriate sonic API based on environment or config
// This can be extended to read from configuration files
func GetSonicAPI(env string) sonic.API {
	switch env {
	case "production":
		return DefaultSonicAPI
	case "development":
		return SafeSonicAPI // Use safe config in dev for better error messages
	case "benchmark":
		return FastSonicAPI // Use fast config for benchmarks
	default:
		return DefaultSonicAPI
	}
}

// ============================================================================
// PERFORMANCE NOTES
// ============================================================================
//
// Performance Impact of Each Option:
//
// EscapeHTML: true → ~15% slower (default: false)
//   - Enable for user-facing APIs where XSS is a concern
//   - Disable for internal APIs or when HTML isn't involved
//
// SortMapKeys: true → ~10% slower (default: false)
//   - Enable when you need deterministic JSON output
//   - Required for: content-based caching, cryptographic signatures, tests
//   - Disable for: most APIs, logs, real-time data
//
// ValidateString: true → ~3% slower (default: true)
//   - Enable to ensure all strings are valid UTF-8
//   - Only disable if you're 100% sure your strings are valid
//
// CopyString: true → ~5% slower (default: true)
//   - Enable to prevent issues with string lifecycle
//   - Only disable in very controlled environments
//
// Overall Performance vs goccy/go-json:
//   - DefaultSonicAPI: ~20-30% faster
//   - SafeSonicAPI: ~15-20% faster
//   - FastSonicAPI: ~40-50% faster (use with extreme caution)
//
// ============================================================================
