package log

import (
	"math/rand"
	"time"

	"github.com/rs/zerolog"
)

// SamplingConfig configures log sampling to reduce log volume.
// Sampling is useful for high-traffic services where logging every event
// would be too expensive or generate too much data.
type SamplingConfig struct {
	// GlobalRate is the global sampling rate (0.0 to 1.0)
	// 0.0 = log nothing, 0.5 = log 50%, 1.0 = log 100%
	// Default: 1.0 (no sampling)
	GlobalRate float64

	// LevelRates contains per-level sampling rates
	// Example: {zerolog.DebugLevel: 0.1} means sample 10% of debug logs
	LevelRates map[zerolog.Level]float64

	// TagRates contains per-tag sampling rates
	// Example: {"database": 0.5} means sample 50% of logs with "database" tag
	TagRates map[string]float64

	// DisabledLevels contains log levels that should never be logged
	// Useful for completely disabling debug logs in production
	DisabledLevels []zerolog.Level

	// Rand is the random source used for sampling decisions
	// Can be set to a custom source for deterministic testing
	// If nil, a default source is created
	Rand *rand.Rand
}

// DefaultSamplingConfig returns a SamplingConfig with no sampling.
// All logs are enabled by default.
func DefaultSamplingConfig() *SamplingConfig {
	return &SamplingConfig{
		GlobalRate:     1.0, // 100% - no sampling
		LevelRates:     make(map[zerolog.Level]float64),
		TagRates:       make(map[string]float64),
		DisabledLevels: []zerolog.Level{},
		Rand:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldLog determines if a log entry should be emitted based on sampling rules.
// It checks in this order:
// 1. DisabledLevels - if level is disabled, return false immediately
// 2. GlobalRate - apply global sampling rate
// 3. LevelRates - apply level-specific sampling rate
// 4. TagRates - apply tag-specific sampling rates
//
// Returns true if the log should be emitted, false if it should be skipped.
func (s *SamplingConfig) ShouldLog(level zerolog.Level, tags []string) bool {
	if s == nil {
		return true // No sampling config = log everything
	}

	// Ensure we have a random source
	if s.Rand == nil {
		s.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	// Check if level is disabled
	for _, disabledLevel := range s.DisabledLevels {
		if level == disabledLevel {
			return false
		}
	}

	// Apply global sampling rate
	if s.GlobalRate < 1.0 {
		if s.Rand.Float64() > s.GlobalRate {
			return false
		}
	}

	// Apply level-specific sampling rate
	if rate, ok := s.LevelRates[level]; ok && rate < 1.0 {
		if s.Rand.Float64() > rate {
			return false
		}
	}

	// Apply tag-specific sampling rates
	// If any tag has a sampling rate, check it
	for _, tag := range tags {
		if rate, ok := s.TagRates[tag]; ok && rate < 1.0 {
			if s.Rand.Float64() > rate {
				return false
			}
		}
	}

	return true
}
