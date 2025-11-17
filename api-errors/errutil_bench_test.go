package apierrors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"
)

// Benchmark errors for testing
var (
	benchSentinel1 = errors.New("benchmark sentinel 1")
	benchSentinel2 = errors.New("benchmark sentinel 2")
)

// benchCustomError for benchmarking Find vs errors.As
type benchCustomError struct {
	msg  string
	code int
}

func (e *benchCustomError) Error() string {
	return fmt.Sprintf("%s (code: %d)", e.msg, e.code)
}

// wrapN wraps an error N times
func wrapN(err error, n int) error {
	for i := 0; i < n; i++ {
		err = fmt.Errorf("wrap%d: %w", i, err)
	}
	return err
}

// =============================================================================
// Is() vs errors.Is() Benchmarks
// =============================================================================

// BenchmarkIs_NilError benchmarks Is() with nil error
func BenchmarkIs_NilError(b *testing.B) {
	b.Run("Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Is(nil, benchSentinel1)
		}
	})

	b.Run("errors.Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = errors.Is(nil, benchSentinel1)
		}
	})
}

// BenchmarkIs_ExactMatch benchmarks Is() with exact match
func BenchmarkIs_ExactMatch(b *testing.B) {
	b.Run("Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Is(benchSentinel1, benchSentinel1)
		}
	})

	b.Run("errors.Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = errors.Is(benchSentinel1, benchSentinel1)
		}
	})
}

// BenchmarkIs_NoMatch benchmarks Is() when target not found
func BenchmarkIs_NoMatch(b *testing.B) {
	b.Run("Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Is(benchSentinel1, benchSentinel2)
		}
	})

	b.Run("errors.Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = errors.Is(benchSentinel1, benchSentinel2)
		}
	})
}

// BenchmarkIs_Wrapped1 benchmarks Is() with 1 level of wrapping
func BenchmarkIs_Wrapped1(b *testing.B) {
	wrapped := wrapN(benchSentinel1, 1)

	b.Run("Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Is(wrapped, benchSentinel1)
		}
	})

	b.Run("errors.Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = errors.Is(wrapped, benchSentinel1)
		}
	})
}

// BenchmarkIs_Wrapped3 benchmarks Is() with 3 levels of wrapping
func BenchmarkIs_Wrapped3(b *testing.B) {
	wrapped := wrapN(benchSentinel1, 3)

	b.Run("Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Is(wrapped, benchSentinel1)
		}
	})

	b.Run("errors.Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = errors.Is(wrapped, benchSentinel1)
		}
	})
}

// BenchmarkIs_Wrapped5 benchmarks Is() with 5 levels of wrapping
func BenchmarkIs_Wrapped5(b *testing.B) {
	wrapped := wrapN(benchSentinel1, 5)

	b.Run("Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Is(wrapped, benchSentinel1)
		}
	})

	b.Run("errors.Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = errors.Is(wrapped, benchSentinel1)
		}
	})
}

// BenchmarkIs_RealWorld benchmarks Is() with real-world errors
func BenchmarkIs_RealWorld_EOF(b *testing.B) {
	wrapped := wrapN(io.EOF, 2)

	b.Run("Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Is(wrapped, io.EOF)
		}
	})

	b.Run("errors.Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = errors.Is(wrapped, io.EOF)
		}
	})
}

func BenchmarkIs_RealWorld_DeadlineExceeded(b *testing.B) {
	wrapped := wrapN(context.DeadlineExceeded, 2)

	b.Run("Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Is(wrapped, context.DeadlineExceeded)
		}
	})

	b.Run("errors.Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = errors.Is(wrapped, context.DeadlineExceeded)
		}
	})
}

// =============================================================================
// Find[T]() vs errors.As() Benchmarks
// =============================================================================

// BenchmarkFind_NilError benchmarks Find() with nil error
func BenchmarkFind_NilError(b *testing.B) {
	b.Run("Find", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = Find[*benchCustomError](nil)
		}
	})

	b.Run("errors.As", func(b *testing.B) {
		b.ReportAllocs()
		var target *benchCustomError
		for i := 0; i < b.N; i++ {
			_ = errors.As(nil, &target)
		}
	})
}

// BenchmarkFind_ExactMatch benchmarks Find() with exact type match
func BenchmarkFind_ExactMatch(b *testing.B) {
	err := &benchCustomError{msg: "test", code: 123}

	b.Run("Find", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = Find[*benchCustomError](err)
		}
	})

	b.Run("errors.As", func(b *testing.B) {
		b.ReportAllocs()
		var target *benchCustomError
		for i := 0; i < b.N; i++ {
			_ = errors.As(err, &target)
		}
	})
}

// BenchmarkFind_NoMatch benchmarks Find() when type not found
func BenchmarkFind_NoMatch(b *testing.B) {
	err := errors.New("plain error")

	b.Run("Find", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = Find[*benchCustomError](err)
		}
	})

	b.Run("errors.As", func(b *testing.B) {
		b.ReportAllocs()
		var target *benchCustomError
		for i := 0; i < b.N; i++ {
			_ = errors.As(err, &target)
		}
	})
}

// BenchmarkFind_Wrapped1 benchmarks Find() with 1 level of wrapping
func BenchmarkFind_Wrapped1(b *testing.B) {
	err := wrapN(&benchCustomError{msg: "test", code: 123}, 1)

	b.Run("Find", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = Find[*benchCustomError](err)
		}
	})

	b.Run("errors.As", func(b *testing.B) {
		b.ReportAllocs()
		var target *benchCustomError
		for i := 0; i < b.N; i++ {
			_ = errors.As(err, &target)
		}
	})
}

// BenchmarkFind_Wrapped3 benchmarks Find() with 3 levels of wrapping
func BenchmarkFind_Wrapped3(b *testing.B) {
	err := wrapN(&benchCustomError{msg: "test", code: 123}, 3)

	b.Run("Find", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = Find[*benchCustomError](err)
		}
	})

	b.Run("errors.As", func(b *testing.B) {
		b.ReportAllocs()
		var target *benchCustomError
		for i := 0; i < b.N; i++ {
			_ = errors.As(err, &target)
		}
	})
}

// BenchmarkFind_Wrapped5 benchmarks Find() with 5 levels of wrapping
func BenchmarkFind_Wrapped5(b *testing.B) {
	err := wrapN(&benchCustomError{msg: "test", code: 123}, 5)

	b.Run("Find", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = Find[*benchCustomError](err)
		}
	})

	b.Run("errors.As", func(b *testing.B) {
		b.ReportAllocs()
		var target *benchCustomError
		for i := 0; i < b.N; i++ {
			_ = errors.As(err, &target)
		}
	})
}

// BenchmarkFind_RealWorld benchmarks Find() with real-world error types
func BenchmarkFind_RealWorld_JSONSyntaxError(b *testing.B) {
	err := wrapN(&json.SyntaxError{Offset: 10}, 2)

	b.Run("Find", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = Find[*json.SyntaxError](err)
		}
	})

	b.Run("errors.As", func(b *testing.B) {
		b.ReportAllocs()
		var target *json.SyntaxError
		for i := 0; i < b.N; i++ {
			_ = errors.As(err, &target)
		}
	})
}

func BenchmarkFind_RealWorld_JSONUnmarshalTypeError(b *testing.B) {
	err := wrapN(&json.UnmarshalTypeError{Value: "string", Type: nil}, 2)

	b.Run("Find", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = Find[*json.UnmarshalTypeError](err)
		}
	})

	b.Run("errors.As", func(b *testing.B) {
		b.ReportAllocs()
		var target *json.UnmarshalTypeError
		for i := 0; i < b.N; i++ {
			_ = errors.As(err, &target)
		}
	})
}

// =============================================================================
// Combined Real-World Scenario Benchmarks
// =============================================================================

// BenchmarkRealWorldScenario_HandleDBError simulates real error handling
func BenchmarkRealWorldScenario_HandleDBError(b *testing.B) {
	// Simulate a database timeout error wrapped by multiple layers
	err := wrapN(context.DeadlineExceeded, 3)

	b.Run("Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Is(err, context.DeadlineExceeded)
		}
	})

	b.Run("errors.Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = errors.Is(err, context.DeadlineExceeded)
		}
	})
}

// BenchmarkRealWorldScenario_HandleJSONError simulates JSON parsing error handling
func BenchmarkRealWorldScenario_HandleJSONError(b *testing.B) {
	// Simulate a JSON syntax error wrapped by API layer
	err := wrapN(&json.SyntaxError{Offset: 42}, 2)

	b.Run("Find", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = Find[*json.SyntaxError](err)
		}
	})

	b.Run("errors.As", func(b *testing.B) {
		b.ReportAllocs()
		var target *json.SyntaxError
		for i := 0; i < b.N; i++ {
			_ = errors.As(err, &target)
		}
	})
}

// BenchmarkRealWorldScenario_MultipleChecks simulates checking multiple error types
func BenchmarkRealWorldScenario_MultipleChecks(b *testing.B) {
	err := wrapN(&json.SyntaxError{Offset: 42}, 2)

	b.Run("Find", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Check multiple types (common pattern in error handlers)
			if _, ok := Find[*json.SyntaxError](err); ok {
				continue
			}
			if _, ok := Find[*json.UnmarshalTypeError](err); ok {
				continue
			}
			if Is(err, io.EOF) {
				continue
			}
		}
	})

	b.Run("errors.As+errors.Is", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var syntaxErr *json.SyntaxError
			if errors.As(err, &syntaxErr) {
				continue
			}
			var unmarshalErr *json.UnmarshalTypeError
			if errors.As(err, &unmarshalErr) {
				continue
			}
			if errors.Is(err, io.EOF) {
				continue
			}
		}
	})
}
