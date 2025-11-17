package log

import (
	"context"
	"strings"
	"testing"
)

// BenchmarkStringConcatenation compares old string concatenation vs strings.Builder
func BenchmarkStringConcatenation(b *testing.B) {
	path := "/api/users"
	raw := "limit=10&offset=20&filter=active"

	b.Run("Old_Concatenation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fullPath := path
			if raw != "" {
				fullPath = path + "?" + raw
			}
			_ = fullPath
		}
	})

	b.Run("StringBuilder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.Grow(len(path) + len(raw) + 1)
			sb.WriteString(path)
			if raw != "" {
				sb.WriteByte('?')
				sb.WriteString(raw)
			}
			_ = sb.String()
		}
	})
}

// BenchmarkStringConcatenation_NoQuery benchmarks when there's no query string
func BenchmarkStringConcatenation_NoQuery(b *testing.B) {
	path := "/api/users"
	raw := ""

	b.Run("Old_Concatenation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fullPath := path
			if raw != "" {
				fullPath = path + "?" + raw
			}
			_ = fullPath
		}
	})

	b.Run("StringBuilder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.Grow(len(path) + len(raw) + 1)
			sb.WriteString(path)
			if raw != "" {
				sb.WriteByte('?')
				sb.WriteString(raw)
			}
			_ = sb.String()
		}
	})
}

// BenchmarkStringConcatenation_LongQuery benchmarks with a very long query string
func BenchmarkStringConcatenation_LongQuery(b *testing.B) {
	path := "/api/search"
	raw := "q=test&limit=100&offset=500&sort=name&order=asc&filter=active&category=books&author=smith&year=2024&language=en"

	b.Run("Old_Concatenation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fullPath := path
			if raw != "" {
				fullPath = path + "?" + raw
			}
			_ = fullPath
		}
	})

	b.Run("StringBuilder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.Grow(len(path) + len(raw) + 1)
			sb.WriteString(path)
			if raw != "" {
				sb.WriteByte('?')
				sb.WriteString(raw)
			}
			_ = sb.String()
		}
	})
}

// BenchmarkWithTags compares old WithTags vs optimized version
func BenchmarkWithTags(b *testing.B) {
	ctx := context.Background()

	b.Run("SingleTag", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = WithTags(ctx, "database")
		}
	})

	b.Run("ThreeTags", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = WithTags(ctx, "database", "payment", "critical")
		}
	})

	b.Run("Accumulation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx1 := WithTags(ctx, "api", "v1")
			ctx2 := WithTags(ctx1, "auth")
			_ = WithTags(ctx2, "admin")
		}
	})
}

// BenchmarkWithTags_Old simulates old implementation for comparison
func BenchmarkWithTags_Old(b *testing.B) {
	ctx := context.Background()

	withTagsOld := func(ctx context.Context, tags ...string) context.Context {
		if ctx == nil {
			ctx = context.Background()
		}
		existingTags := GetTags(ctx)
		allTags := append(existingTags, tags...) // No pre-allocation
		return context.WithValue(ctx, logTagsContextKey, allTags)
	}

	b.Run("SingleTag", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = withTagsOld(ctx, "database")
		}
	})

	b.Run("ThreeTags", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = withTagsOld(ctx, "database", "payment", "critical")
		}
	})

	b.Run("Accumulation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx1 := withTagsOld(ctx, "api", "v1")
			ctx2 := withTagsOld(ctx1, "auth")
			_ = withTagsOld(ctx2, "admin")
		}
	})
}

// BenchmarkWithTags_ManyAccumulations benchmarks many tag accumulations
func BenchmarkWithTags_ManyAccumulations(b *testing.B) {
	ctx := context.Background()

	b.Run("10_Accumulations_New", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := ctx
			for j := 0; j < 10; j++ {
				c = WithTags(c, "tag")
			}
		}
	})

	withTagsOld := func(ctx context.Context, tags ...string) context.Context {
		if ctx == nil {
			ctx = context.Background()
		}
		existingTags := GetTags(ctx)
		allTags := append(existingTags, tags...)
		return context.WithValue(ctx, logTagsContextKey, allTags)
	}

	b.Run("10_Accumulations_Old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := ctx
			for j := 0; j < 10; j++ {
				c = withTagsOld(c, "tag")
			}
		}
	})
}

// BenchmarkGetTags benchmarks tag retrieval
func BenchmarkGetTags(b *testing.B) {
	ctx := WithTags(context.Background(), "tag1", "tag2", "tag3", "tag4", "tag5")

	b.Run("GetTags", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GetTags(ctx)
		}
	})
}

// BenchmarkAddFieldsToEvent benchmarks the field addition helper
func BenchmarkAddFieldsToEvent(b *testing.B) {
	setupTestLogger()

	fields := map[string]interface{}{
		"string_field": "value",
		"int_field":    42,
		"float_field":  3.14,
		"bool_field":   true,
		"user_id":      "user-123",
	}

	b.Run("AddFieldsToEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			event := InfoEvent(nil)
			addFieldsToEvent(event, fields)
		}
	})
}

// BenchmarkWithFields compares WithFields vs Event API
func BenchmarkWithFields(b *testing.B) {
	setupTestLogger()
	ctx := context.Background()

	fields := map[string]interface{}{
		"user_id":  "123",
		"action":   "login",
		"duration": 45,
		"success":  true,
	}

	b.Run("InfoWithFields", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			InfoWithFields(ctx, "user action", fields)
		}
	})

	b.Run("InfoEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			InfoEvent(ctx).
				Str("user_id", "123").
				Str("action", "login").
				Int("duration", 45).
				Bool("success", true).
				Msg("user action")
		}
	})
}
