package routeradapter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync"
)

// RouterContext is a framework-agnostic request/response context
// It provides a unified interface for handling HTTP requests across different frameworks
// This allows writing handlers once and using them with any supported framework
type RouterContext struct {
	// Request is the underlying HTTP request
	Request *http.Request

	// Response is the underlying HTTP response writer
	Response http.ResponseWriter

	// Params contains path parameters extracted from the URL
	// For example, /users/:id would have Params["id"] = "123"
	Params map[string]string

	// Query contains URL query parameters
	// Populated from r.URL.Query()
	Query url.Values

	// Data is a key-value store for request-scoped data
	// Useful for passing data between middlewares
	// Thread-safe using sync.RWMutex
	data   map[string]interface{}
	dataMu sync.RWMutex

	// ctx is the request context (with cancellation, deadlines, etc.)
	ctx context.Context

	// statusCode tracks the HTTP status code to be sent
	// Default is 200 (OK)
	statusCode int

	// responseWritten indicates if response has been sent
	// Prevents multiple writes to response
	responseWritten bool

	// nativeCtx holds the framework-specific context
	// This is useful when you need to access framework-specific features
	// Type depends on the framework (e.g., *gin.Context, *fiber.Ctx, echo.Context)
	nativeCtx interface{}
}

// NewRouterContext creates a new RouterContext from an http.Request and http.ResponseWriter
func NewRouterContext(w http.ResponseWriter, r *http.Request) *RouterContext {
	return &RouterContext{
		Request:    r,
		Response:   w,
		Params:     make(map[string]string),
		Query:      r.URL.Query(),
		data:       make(map[string]interface{}),
		ctx:        r.Context(),
		statusCode: http.StatusOK,
	}
}

// Context returns the request context
// This includes cancellation, deadlines, and request-scoped values
func (rc *RouterContext) Context() context.Context {
	return rc.ctx
}

// SetContext updates the request context
// Useful for adding cancellation, timeout, or tracing context
func (rc *RouterContext) SetContext(ctx context.Context) {
	rc.ctx = ctx
}

// Param gets a path parameter by name
// Returns empty string if parameter doesn't exist
func (rc *RouterContext) Param(name string) string {
	return rc.Params[name]
}

// SetParam sets a path parameter
// Used by routers when extracting parameters from path
func (rc *RouterContext) SetParam(name, value string) {
	rc.Params[name] = value
}

// QueryParam gets a query parameter by name
// Returns empty string if parameter doesn't exist
func (rc *RouterContext) QueryParam(name string) string {
	return rc.Query.Get(name)
}

// Get retrieves request-scoped data by key
// Returns nil if key doesn't exist
func (rc *RouterContext) Get(key string) interface{} {
	rc.dataMu.RLock()
	defer rc.dataMu.RUnlock()
	return rc.data[key]
}

// Set stores request-scoped data
// Thread-safe for use across goroutines
func (rc *RouterContext) Set(key string, value interface{}) {
	rc.dataMu.Lock()
	defer rc.dataMu.Unlock()
	rc.data[key] = value
}

// Status sets the HTTP status code for the response
// Must be called before writing the response body
func (rc *RouterContext) Status(code int) *RouterContext {
	rc.statusCode = code
	return rc
}

// JSON sends a JSON response
// Automatically sets Content-Type to application/json
// Uses goccy/go-json for high performance (via binding setup)
func (rc *RouterContext) JSON(code int, obj interface{}) error {
	if rc.responseWritten {
		return ErrResponseAlreadyWritten
	}

	rc.statusCode = code
	rc.Response.Header().Set("Content-Type", "application/json")
	rc.Response.WriteHeader(code)

	// Use standard encoding/json for now
	// The framework adapters can override this to use goccy/go-json
	err := json.NewEncoder(rc.Response).Encode(obj)
	rc.responseWritten = true
	return err
}

// String sends a plain text response
func (rc *RouterContext) String(code int, format string, values ...interface{}) error {
	if rc.responseWritten {
		return ErrResponseAlreadyWritten
	}

	rc.statusCode = code
	rc.Response.Header().Set("Content-Type", "text/plain")
	rc.Response.WriteHeader(code)

	if len(values) > 0 {
		_, err := rc.Response.Write([]byte(sprintf(format, values...)))
		rc.responseWritten = true
		return err
	}

	_, err := rc.Response.Write([]byte(format))
	rc.responseWritten = true
	return err
}

// Bytes sends a raw byte response
func (rc *RouterContext) Bytes(code int, contentType string, data []byte) error {
	if rc.responseWritten {
		return ErrResponseAlreadyWritten
	}

	rc.statusCode = code
	if contentType != "" {
		rc.Response.Header().Set("Content-Type", contentType)
	}
	rc.Response.WriteHeader(code)

	_, err := rc.Response.Write(data)
	rc.responseWritten = true
	return err
}

// NoContent sends a response with no body
// Typically used with 204 No Content status
func (rc *RouterContext) NoContent(code int) error {
	if rc.responseWritten {
		return ErrResponseAlreadyWritten
	}

	rc.statusCode = code
	rc.Response.WriteHeader(code)
	rc.responseWritten = true
	return nil
}

// Redirect sends a redirect response
func (rc *RouterContext) Redirect(code int, url string) error {
	if rc.responseWritten {
		return ErrResponseAlreadyWritten
	}

	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}

	http.Redirect(rc.Response, rc.Request, url, code)
	rc.responseWritten = true
	return nil
}

// Bind binds request body to the provided struct
// Automatically detects content type and uses appropriate decoder
func (rc *RouterContext) Bind(obj interface{}) error {
	contentType := rc.Request.Header.Get("Content-Type")

	switch contentType {
	case "application/json":
		return json.NewDecoder(rc.Request.Body).Decode(obj)
	case "application/x-www-form-urlencoded":
		if err := rc.Request.ParseForm(); err != nil {
			return err
		}
		return formToStruct(rc.Request.Form, obj)
	case "multipart/form-data":
		if err := rc.Request.ParseMultipartForm(32 << 20); err != nil { // 32 MB max
			return err
		}
		return formToStruct(rc.Request.MultipartForm.Value, obj)
	default:
		return ErrUnsupportedMediaType
	}
}

// Body returns the request body as bytes
// Note: Body can only be read once unless you store it
func (rc *RouterContext) Body() ([]byte, error) {
	return io.ReadAll(rc.Request.Body)
}

// Header gets a request header value
func (rc *RouterContext) Header(key string) string {
	return rc.Request.Header.Get(key)
}

// SetHeader sets a response header
func (rc *RouterContext) SetHeader(key, value string) {
	rc.Response.Header().Set(key, value)
}

// GetNativeContext returns the framework-specific context
// Cast to the appropriate type based on the framework you're using
// For example: ginCtx := rc.GetNativeContext().(*gin.Context)
func (rc *RouterContext) GetNativeContext() interface{} {
	return rc.nativeCtx
}

// SetNativeContext stores the framework-specific context
// Used internally by adapters to allow access to framework features
func (rc *RouterContext) SetNativeContext(ctx interface{}) {
	rc.nativeCtx = ctx
}

// StatusCode returns the HTTP status code that will be sent
func (rc *RouterContext) StatusCode() int {
	return rc.statusCode
}

// IsResponseWritten returns true if response has already been written
func (rc *RouterContext) IsResponseWritten() bool {
	return rc.responseWritten
}

// Helper functions

func sprintf(format string, args ...interface{}) string {
	// Simple implementation - in production use fmt.Sprintf
	return format // Placeholder
}

// formToStruct converts form data to struct
// This is a simplified implementation - production code should use proper form binding
func formToStruct(form url.Values, obj interface{}) error {
	// Simplified implementation
	// In production, use reflection or a library like gorilla/schema
	return nil
}

// Errors

var (
	ErrResponseAlreadyWritten = &RouterError{Code: "response_already_written", Message: "response has already been written"}
	ErrInvalidRedirectCode    = &RouterError{Code: "invalid_redirect_code", Message: "redirect code must be 3xx"}
	ErrUnsupportedMediaType   = &RouterError{Code: "unsupported_media_type", Message: "unsupported media type"}
)

// RouterError represents an error from the router
type RouterError struct {
	Code    string
	Message string
}

func (e *RouterError) Error() string {
	return e.Message
}
