package apierrors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"
)

// Test errors for testing
var (
	errBase      = errors.New("base error")
	errSentinel1 = errors.New("sentinel error 1")
	errSentinel2 = errors.New("sentinel error 2")
)

// Custom error type for testing Find
type customError struct {
	msg  string
	code int
}

func (e *customError) Error() string {
	return fmt.Sprintf("%s (code: %d)", e.msg, e.code)
}

// Error that implements Is() method
type customIsError struct {
	msg    string
	target error
}

func (e *customIsError) Error() string {
	return e.msg
}

func (e *customIsError) Is(target error) bool {
	return e.target == target
}

// Wrapped error for testing
func wrapError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

// TestIs tests the Is function
func TestIs(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "nil error and nil target",
			err:    nil,
			target: nil,
			want:   true,
		},
		{
			name:   "nil error and non-nil target",
			err:    nil,
			target: errSentinel1,
			want:   false,
		},
		{
			name:   "non-nil error and nil target",
			err:    errSentinel1,
			target: nil,
			want:   false,
		},
		{
			name:   "exact match",
			err:    errSentinel1,
			target: errSentinel1,
			want:   true,
		},
		{
			name:   "no match",
			err:    errSentinel1,
			target: errSentinel2,
			want:   false,
		},
		{
			name:   "wrapped error - match",
			err:    wrapError(errSentinel1, "wrapper"),
			target: errSentinel1,
			want:   true,
		},
		{
			name:   "wrapped error - no match",
			err:    wrapError(errSentinel1, "wrapper"),
			target: errSentinel2,
			want:   false,
		},
		{
			name:   "double wrapped error - match",
			err:    wrapError(wrapError(errSentinel1, "wrapper1"), "wrapper2"),
			target: errSentinel1,
			want:   true,
		},
		{
			name:   "io.EOF match",
			err:    io.EOF,
			target: io.EOF,
			want:   true,
		},
		{
			name:   "wrapped io.EOF",
			err:    wrapError(io.EOF, "read failed"),
			target: io.EOF,
			want:   true,
		},
		{
			name:   "context.DeadlineExceeded",
			err:    context.DeadlineExceeded,
			target: context.DeadlineExceeded,
			want:   true,
		},
		{
			name:   "wrapped context.DeadlineExceeded",
			err:    wrapError(context.DeadlineExceeded, "timeout"),
			target: context.DeadlineExceeded,
			want:   true,
		},
		{
			name:   "custom Is() method - match",
			err:    &customIsError{msg: "custom", target: errSentinel1},
			target: errSentinel1,
			want:   true,
		},
		{
			name:   "custom Is() method - no match",
			err:    &customIsError{msg: "custom", target: errSentinel1},
			target: errSentinel2,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Is(tt.err, tt.target)
			if got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFind tests the Find function
func TestFind(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantOk  bool
		wantVal interface{}
	}{
		{
			name:   "nil error",
			err:    nil,
			wantOk: false,
		},
		{
			name:    "direct match - customError",
			err:     &customError{msg: "test", code: 123},
			wantOk:  true,
			wantVal: &customError{msg: "test", code: 123},
		},
		{
			name:    "wrapped customError",
			err:     wrapError(&customError{msg: "test", code: 456}, "wrapper"),
			wantOk:  true,
			wantVal: &customError{msg: "test", code: 456},
		},
		{
			name:   "no match - different type",
			err:    errors.New("plain error"),
			wantOk: false,
		},
		{
			name:    "json.SyntaxError",
			err:     &json.SyntaxError{Offset: 10},
			wantOk:  true,
			wantVal: &json.SyntaxError{Offset: 10},
		},
		{
			name:    "wrapped json.SyntaxError",
			err:     wrapError(&json.SyntaxError{Offset: 20}, "json parse failed"),
			wantOk:  true,
			wantVal: &json.SyntaxError{Offset: 20},
		},
		{
			name:    "json.UnmarshalTypeError",
			err:     &json.UnmarshalTypeError{Value: "string", Type: nil, Offset: 5},
			wantOk:  true,
			wantVal: &json.UnmarshalTypeError{Value: "string", Type: nil, Offset: 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.wantVal.(type) {
			case *customError:
				got, ok := Find[*customError](tt.err)
				if ok != tt.wantOk {
					t.Errorf("Find() ok = %v, want %v", ok, tt.wantOk)
				}
				if ok && (got.msg != v.msg || got.code != v.code) {
					t.Errorf("Find() = %+v, want %+v", got, v)
				}
			case *json.SyntaxError:
				got, ok := Find[*json.SyntaxError](tt.err)
				if ok != tt.wantOk {
					t.Errorf("Find() ok = %v, want %v", ok, tt.wantOk)
				}
				if ok && got.Offset != v.Offset {
					t.Errorf("Find() offset = %d, want %d", got.Offset, v.Offset)
				}
			case *json.UnmarshalTypeError:
				got, ok := Find[*json.UnmarshalTypeError](tt.err)
				if ok != tt.wantOk {
					t.Errorf("Find() ok = %v, want %v", ok, tt.wantOk)
				}
				if ok && got.Value != v.Value {
					t.Errorf("Find() value = %s, want %s", got.Value, v.Value)
				}
			default:
				// For cases where we don't expect a match
				_, ok := Find[*customError](tt.err)
				if ok != tt.wantOk {
					t.Errorf("Find() ok = %v, want %v", ok, tt.wantOk)
				}
			}
		})
	}
}

// TestAs tests the As function
func TestAs(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target interface{}
		wantOk bool
	}{
		{
			name:   "nil error",
			err:    nil,
			target: new(*customError),
			wantOk: false,
		},
		{
			name:   "direct match",
			err:    &customError{msg: "test", code: 123},
			target: new(*customError),
			wantOk: true,
		},
		{
			name:   "wrapped match",
			err:    wrapError(&customError{msg: "test", code: 456}, "wrapper"),
			target: new(*customError),
			wantOk: true,
		},
		{
			name:   "no match",
			err:    errors.New("plain error"),
			target: new(*customError),
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch target := tt.target.(type) {
			case **customError:
				got := As(tt.err, target)
				if got != tt.wantOk {
					t.Errorf("As() = %v, want %v", got, tt.wantOk)
				}
				if got && *target == nil {
					t.Error("As() returned true but target is nil")
				}
			}
		})
	}
}

// TestAs_Panic tests that As panics when target is nil
func TestAs_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("As() did not panic with nil target")
		}
	}()

	As[*customError](errors.New("test"), nil) // Should panic because target is nil
}

// TestIsVsErrorsIs compares our Is() with errors.Is()
func TestIsVsErrorsIs(t *testing.T) {
	testCases := []struct {
		name   string
		err    error
		target error
	}{
		{"nil/nil", nil, nil},
		{"nil/non-nil", nil, errSentinel1},
		{"non-nil/nil", errSentinel1, nil},
		{"exact match", errSentinel1, errSentinel1},
		{"no match", errSentinel1, errSentinel2},
		{"wrapped match", wrapError(errSentinel1, "wrap"), errSentinel1},
		{"wrapped no match", wrapError(errSentinel1, "wrap"), errSentinel2},
		{"double wrapped", wrapError(wrapError(errSentinel1, "w1"), "w2"), errSentinel1},
		{"io.EOF", io.EOF, io.EOF},
		{"wrapped io.EOF", wrapError(io.EOF, "read"), io.EOF},
		{"context.DeadlineExceeded", context.DeadlineExceeded, context.DeadlineExceeded},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Is(tc.err, tc.target)
			want := errors.Is(tc.err, tc.target)
			if got != want {
				t.Errorf("Is() = %v, errors.Is() = %v, mismatch", got, want)
			}
		})
	}
}

// TestFindVsErrorsAs compares our Find() with errors.As()
func TestFindVsErrorsAs(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{"nil", nil},
		{"direct customError", &customError{msg: "test", code: 123}},
		{"wrapped customError", wrapError(&customError{msg: "test", code: 456}, "wrap")},
		{"plain error", errors.New("plain")},
		{"json.SyntaxError", &json.SyntaxError{Offset: 10}},
		{"wrapped json.SyntaxError", wrapError(&json.SyntaxError{Offset: 20}, "wrap")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test with customError
			var asTarget *customError
			asOk := errors.As(tc.err, &asTarget)
			findResult, findOk := Find[*customError](tc.err)

			if asOk != findOk {
				t.Errorf("errors.As() = %v, Find() = %v, mismatch", asOk, findOk)
			}

			if asOk && findOk {
				if asTarget.msg != findResult.msg || asTarget.code != findResult.code {
					t.Errorf("errors.As() result = %+v, Find() result = %+v, mismatch", asTarget, findResult)
				}
			}
		})
	}
}

// TestMultipleWrapping tests deeply nested error wrapping
func TestMultipleWrapping(t *testing.T) {
	base := &customError{msg: "base", code: 100}
	wrapped := wrapError(base, "level1")
	wrapped = wrapError(wrapped, "level2")
	wrapped = wrapError(wrapped, "level3")
	wrapped = wrapError(wrapped, "level4")
	wrapped = wrapError(wrapped, "level5")

	// Test Is with a sentinel error that's actually in the chain
	sentinelErr := errors.New("sentinel")
	wrappedWithSentinel := wrapError(sentinelErr, "level1")
	wrappedWithSentinel = wrapError(wrappedWithSentinel, "level2")
	wrappedWithSentinel = wrapError(wrappedWithSentinel, "level3")

	if !Is(wrappedWithSentinel, sentinelErr) {
		t.Error("Is() failed to find sentinel error in deeply wrapped chain")
	}

	// Test Find
	found, ok := Find[*customError](wrapped)
	if !ok {
		t.Fatal("Find() failed to find customError in deeply wrapped chain")
	}
	if found.msg != "base" || found.code != 100 {
		t.Errorf("Find() = %+v, want {msg:base, code:100}", found)
	}
}
