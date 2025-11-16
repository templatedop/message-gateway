# API-Errors Module - Comprehensive Architecture Review

## Executive Summary

The api-errors module demonstrates **sophisticated error handling** with:
- ✅ **Generics-based error inspection** (replacing reflection-based errors.As/errors.Is)
- ✅ **Rich error context** (stack traces, field errors, HTTP mapping)
- ✅ **Comprehensive error categorization** (HTTP, DB, security, file, integration)
- ✅ **Type-safe error handling** without runtime reflection overhead

**Overall Assessment**: **EXCELLENT** - Modern, well-architected error handling system with room for minor refinements.

---

## Architecture Analysis

### 1. Core Components

#### 1.1 AppError (Central Error Type)

**File**: `apperrors.go`

```go
type AppError struct {
    ID            string       // Unique error identifier
    Code          string       // HTTP status code as string
    Message       string       // User-facing message
    FieldErrors   []FieldError // Validation errors
    Stack         *stackTrace  // Stack trace
    OriginalError error        // Wrapped error
}
```

**Strengths**:
- ✅ Implements standard `error` interface
- ✅ Implements `Unwrap()` for error chain traversal
- ✅ Rich context (stack trace, field errors, original error)
- ✅ Supports error IDs for tracking

**Issues**:
1. ⚠️ **Code stored as `string` instead of `int`**
   - Requires `strconv.Atoi()` conversion everywhere
   - Error-prone (what if Code is not a number?)
   - Performance: string-to-int conversion overhead

2. ⚠️ **Stack trace always collected**
   - `collectStackTrace()` called in `NewAppError()` constructor
   - Performance cost even when stack trace not needed
   - No way to disable for production

3. ⚠️ **FieldError.Tag not exposed in JSON**
   - `Tag string \`json:"-"\`` is hidden but could be useful for clients

---

### 2. Generics-Based Error Inspection

**File**: `errutil.go`

#### 2.1 `As[T error]()` Function

```go
func As[T error](err error, target *T) bool {
    // Type-safe error assertion without reflection
}
```

**Benefits over `errors.As()`**:
- ✅ **Type safety**: Compile-time type checking
- ✅ **Performance**: No reflection (direct type assertion)
- ✅ **Ergonomics**: Cleaner API with generics

**Usage Example**:
```go
// Old (reflection-based)
var appErr *AppError
if errors.As(err, &appErr) {
    // ...
}

// New (generics-based)
if appErr, ok := Find[*AppError](err); ok {
    // ...
}
```

#### 2.2 `Find[T error]()` Function

```go
func Find[T error](err error) (T, bool) {
    // Returns value directly instead of pointer
}
```

**Benefits**:
- ✅ **More ergonomic**: Returns value and boolean
- ✅ **Cleaner code**: No pointer declaration needed
- ✅ **Better for Go patterns**: Matches `map[key]` idiom

**Implementation Quality**: 🟢 **EXCELLENT**
- Handles multi-error unwrapping (Unwrap() []error)
- Supports custom As() methods on error types
- Zero allocations in hot path

---

### 3. Error Code System

**Files**: `errcodes.go`, `dberrcodes.go`

#### 3.1 HTTP Error Codes

```go
type statusCodeAndMessage struct {
    StatusCode int    // HTTP status code
    Message    string // Default message
    Success    bool   // Always false for errors
}
```

**Comprehensive Coverage**:
- ✅ HTTP 4xx errors (400, 401, 403, 404, 409, 422, 429, etc.)
- ✅ HTTP 5xx errors (500, 501, 502, 503, 504)
- ✅ Application-specific (validation, binding, business rules)
- ✅ Database-specific (record not found, duplicate, transaction failure)
- ✅ Security-specific (auth, authorization, tokens, CSRF)
- ✅ File-specific (upload, size, type, read/write)
- ✅ Integration-specific (timeout, rate limit, network, dependency)

**Issues**:
1. ⚠️ **No error code constants**
   - Status codes hardcoded everywhere as strings
   - Example: `"404"`, `"500"`, `"422"`
   - Should use const declarations

2. ⚠️ **Inconsistent Success field**
   - Always `false` for all error types
   - Redundant in error responses
   - Could be removed

3. ⚠️ **Typo in function name**
   - `NewHTTPStatsuCodeAndMessage` should be `NewHTTPStatusCodeAndMessage`

---

### 4. Database Error Handling

**File**: `dberrcodes.go`

**Strengths**:
- ✅ **Comprehensive PostgreSQL error coverage**
- ✅ Maps SQLSTATE codes to HTTP status codes
- ✅ Handles all major error classes:
  - Connection exceptions (08)
  - Data exceptions (22)
  - Integrity constraints (23)
  - Transaction rollback (40)
  - Syntax errors (42)
  - Insufficient resources (53)

**Issues**:
1. ⚠️ **Mixed code formats**
   ```go
   DBGenericError = dbError{Code: "500", Message: "DB Error"}  // HTTP code
   DBNoData       = dbError{Code: "02", Message: "No Data"}    // SQLSTATE code
   ```
   - Inconsistent: mixing HTTP codes with SQLSTATE codes
   - Confusing: same Code field, different meaning

2. ⚠️ **Unused error types**
   - Many dbError definitions never referenced
   - Example: DBLocatorException, DBDiagnosticsException
   - Should remove or document why they exist

---

### 5. Error Handlers

**File**: `apperrorhandlers.go` (807 lines)

#### 5.1 Handler Functions

**Comprehensive Coverage**:
- ✅ Route errors (404, 405)
- ✅ Binding errors (validation, JSON parsing)
- ✅ Database errors (PostgreSQL error mapping)
- ✅ Security errors (401, 403, tokens)
- ✅ File errors (upload, size, type)
- ✅ Network errors (timeout, rate limit, connection)
- ✅ Common error handler (auto-detection)

**Issues**:

1. 🔴 **Massive code duplication**
   - Same error handling pattern repeated 20+ times:
   ```go
   appError := NewAppError(message, code, err)
   apiErrorResponse := NewHTTPAPIErrorResponse(statusCodeAndMessage, appError)
   ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
   ```
   - Violates DRY principle
   - Hard to maintain

2. ⚠️ **Mixed use of generics vs errors.As**
   - Some functions use `Find[T]()`:
     ```go
     if appErr, ok := Find[*AppError](err); ok { // Line 68, 155, 353, 764
     ```
   - Some functions use `errors.As()`:
     ```go
     if errors.As(err, &appErr) { // Line 191, 223, 318
     ```
   - Inconsistent migration to generics

3. ⚠️ **HandleCommonError complexity**
   - Lines 757-789: Tries to handle all error types
   - Duplicates logic from `HandleDBError()`
   - Should be the primary handler, others specialized

4. ⚠️ **Commented code**
   - Lines 70-78, 99-101, 352-356, 677-680, 763-766
   - Old `errors.As()` calls left commented
   - Should be removed

5. ⚠️ **Function without ctx.JSON() call**
   - `HandleErrorWithStatusCodeAndMessage()` (line 653)
   - Returns `*AppError` instead of sending response
   - Inconsistent with other handlers

---

### 6. Stack Traces

**File**: `stacktrace.go`

**Implementation**:
```go
func collectStackTrace() *stackTrace {
    var pcs [32]uintptr  // Fixed size: max 32 frames
    n := runtime.Callers(3, pcs[:])  // Skip 3 frames
    // Convert to stackFrame structs
}
```

**Strengths**:
- ✅ Captures function name, file, line number
- ✅ Includes pointer addresses (debugging)
- ✅ Pretty-print formatting

**Issues**:
1. ⚠️ **Always collected**
   - Performance overhead for every error
   - No production vs development toggle
   - Should be configurable

2. ⚠️ **Fixed stack depth (32 frames)**
   - Deep call stacks will be truncated
   - No indication of truncation
   - Should make configurable or document limit

3. ⚠️ **Pointer addresses in JSON**
   - `Pointer` and `FunctionPointer` fields included
   - Not useful for clients
   - Security concern: leaks memory layout (ASLR bypass)
   - Should exclude from JSON

---

## Improvement Recommendations

### Priority 1: Critical Fixes

#### 1.1 Change AppError.Code from string to int

**Current**:
```go
type AppError struct {
    Code string `json:"code"`  // ❌ String requiring conversion
}
```

**Recommended**:
```go
type AppError struct {
    Code int `json:"code"`     // ✅ Direct int type
}
```

**Benefits**:
- Eliminates `strconv.Atoi()` calls everywhere
- Type safety (compile-time checking)
- Better performance
- Cleaner code

**Migration**:
```go
// Update NewAppError signature
func NewAppError(message string, code int, originalError error) AppError {
    return AppError{
        Stack:         collectStackTrace(),
        Message:       message,
        Code:          code,  // Direct int
        OriginalError: originalError,
    }
}

// Update all call sites
// Old: NewAppError("msg", "500", err)
// New: NewAppError("msg", http.StatusInternalServerError, err)
```

---

#### 1.2 Eliminate Code Duplication in Handlers

**Current** (repeated ~20 times):
```go
func HandleSomeError(ctx *gin.Context, err error) {
    appError := NewAppError(message, code, err)
    apiErrorResponse := NewHTTPAPIErrorResponse(statusCode, appError)
    ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}
```

**Recommended**: Create helper function
```go
// Helper to reduce duplication
func respondWithError(
    ctx *gin.Context,
    statusCodeAndMessage statusCodeAndMessage,
    message string,
    err error,
) {
    appError := NewAppError(message, statusCodeAndMessage.StatusCode, err)
    apiErrorResponse := NewHTTPAPIErrorResponse(statusCodeAndMessage, appError)
    ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// Simplified handler
func HandleUnauthorizedError(ctx *gin.Context) {
    respondWithError(
        ctx,
        HTTPErrorUnauthorized,
        "Unauthorized access. Authentication is required.",
        nil,
    )
}
```

**Benefits**:
- 70% code reduction in handlers
- Single point of change for response format
- Easier to maintain and test
- Consistent behavior

---

#### 1.3 Complete Migration to Generics

**Current**: Mixed usage
```go
// Some places use generics
if appErr, ok := Find[*AppError](err); ok { // ✅ Modern

// Some places use reflection
if errors.As(err, &appErr) { // ❌ Old
```

**Recommended**: Use generics everywhere
```go
// Replace all errors.As() with Find[T]()
if appErr, ok := Find[*AppError](err); ok {
    // ...
}

// Replace all errors.Is() with explicit checks or custom Is[T]() function
func Is[T error](err error, target T) bool {
    for err != nil {
        if errors.Is(err, target) {
            return true
        }
        err = errors.Unwrap(err)
    }
    return false
}
```

**Files to update**:
- `apperrorhandlers.go`: Lines 191, 223, 318
- Remove all commented `errors.As()` calls

---

### Priority 2: High-Value Improvements

#### 2.1 Error Code Constants

**Create constants file**: `errorcodes_const.go`

```go
package apierrors

const (
    // HTTP Status Codes
    CodeOK                  = 200
    CodeBadRequest          = 400
    CodeUnauthorized        = 401
    CodeForbidden           = 403
    CodeNotFound            = 404
    CodeConflict            = 409
    CodeUnprocessableEntity = 422
    CodeTooManyRequests     = 429
    CodeInternalServerError = 500
    CodeNotImplemented      = 501
    CodeBadGateway          = 502
    CodeServiceUnavailable  = 503
    CodeGatewayTimeout      = 504

    // Application Codes
    CodeValidationError    = 422
    CodeBindingError       = 400
    CodeBusinessRuleError  = 400

    // Database Codes
    CodeDBRecordNotFound   = 404
    CodeDBDuplicateRecord  = 409
    CodeDBTransactionFail  = 500
)
```

**Usage**:
```go
// Instead of:
NewAppError("msg", "500", err)

// Use:
NewAppError("msg", CodeInternalServerError, err)
```

---

#### 2.2 Configurable Stack Traces

**Add global configuration**:

```go
// config.go
package apierrors

var (
    EnableStackTraces = true  // Toggle for stack traces
    MaxStackDepth     = 32    // Configurable depth
)

// Updated collectStackTrace
func collectStackTrace() *stackTrace {
    if !EnableStackTraces {
        return nil  // Skip collection
    }

    var pcs [128]uintptr  // Increased from 32
    maxDepth := MaxStackDepth
    if maxDepth > 128 {
        maxDepth = 128
    }

    n := runtime.Callers(3, pcs[:maxDepth])
    // ... rest of implementation
}
```

**Usage**:
```go
// In main.go or init
func init() {
    if os.Getenv("ENV") == "production" {
        apierrors.EnableStackTraces = false  // Disable in prod
    }
}
```

---

#### 2.3 Remove Redundant Fields

**statusCodeAndMessage.Success**:
```go
// Current
type statusCodeAndMessage struct {
    StatusCode int    `json:"status_code"`
    Message    string `json:"message"`
    Success    bool   `json:"success"`  // ❌ Always false, redundant
}

// Recommended: Remove Success field
type statusCodeAndMessage struct {
    StatusCode int    `json:"status_code"`
    Message    string `json:"message"`
}

// Or make it meaningful
type APIResponse struct {
    Success bool              `json:"success"`  // true for success, false for error
    Data    interface{}       `json:"data,omitempty"`
    Error   *APIErrorResponse `json:"error,omitempty"`
}
```

---

#### 2.4 Security: Remove Pointer Addresses from Stack Traces

**Current**:
```go
type stackFrame struct {
    Pointer         uintptr `json:"pointer"`          // ❌ Security risk
    FunctionPointer uintptr `json:"function_pointer"` // ❌ Security risk
}
```

**Recommended**:
```go
type stackFrame struct {
    Function string `json:"function"`
    File     string `json:"file"`
    Line     int    `json:"line"`
    // Removed: Pointer and FunctionPointer (internal use only)
}

// For debugging, keep in String() output but not JSON
func (sf *stackFrame) String() string {
    return fmt.Sprintf(
        "Function: %s\nLocation: %s:%d\nPointer: 0x%x",
        sf.Function, sf.File, sf.Line, sf.pointer,  // lowercase = private
    )
}
```

---

### Priority 3: Nice-to-Have Enhancements

#### 3.1 Error Builder Pattern

```go
type ErrorBuilder struct {
    message string
    code    int
    err     error
    fields  []FieldError
    id      string
}

func NewErrorBuilder() *ErrorBuilder {
    return &ErrorBuilder{}
}

func (b *ErrorBuilder) WithMessage(msg string) *ErrorBuilder {
    b.message = msg
    return b
}

func (b *ErrorBuilder) WithCode(code int) *ErrorBuilder {
    b.code = code
    return b
}

func (b *ErrorBuilder) WithError(err error) *ErrorBuilder {
    b.err = err
    return b
}

func (b *ErrorBuilder) WithFieldError(field, tag, message string, value interface{}) *ErrorBuilder {
    b.fields = append(b.fields, FieldError{
        Field:   field,
        Tag:     tag,
        Message: message,
        Value:   value,
    })
    return b
}

func (b *ErrorBuilder) Build() *AppError {
    appErr := NewAppError(b.message, b.code, b.err)
    if len(b.fields) > 0 {
        appErr.SetFieldErrors(b.fields)
    }
    if b.id != "" {
        appErr.ID = b.id
    }
    return &appErr
}

// Usage
appErr := apierrors.NewErrorBuilder().
    WithMessage("Validation failed").
    WithCode(CodeValidationError).
    WithFieldError("email", "email", "Invalid email format", "not-an-email").
    WithFieldError("age", "min", "Must be 18 or older", 15).
    Build()
```

---

#### 3.2 Structured Logging Integration

```go
// Add method to convert AppError to structured log fields
func (e *AppError) LogFields() map[string]interface{} {
    fields := map[string]interface{}{
        "error_code":    e.Code,
        "error_message": e.Message,
    }

    if e.ID != "" {
        fields["error_id"] = e.ID
    }

    if len(e.FieldErrors) > 0 {
        fields["field_errors"] = e.FieldErrors
    }

    if e.OriginalError != nil {
        fields["original_error"] = e.OriginalError.Error()
    }

    return fields
}

// Usage with zerolog
logger.Error().
    Fields(appErr.LogFields()).
    Msg("Request failed")
```

---

#### 3.3 Error Metrics

```go
// Track error occurrences for monitoring
type ErrorMetrics struct {
    mu     sync.RWMutex
    counts map[int]int64  // code -> count
}

var metrics = &ErrorMetrics{
    counts: make(map[int]int64),
}

func (e *AppError) Record() {
    metrics.mu.Lock()
    defer metrics.mu.Unlock()
    metrics.counts[e.Code]++
}

func GetErrorMetrics() map[int]int64 {
    metrics.mu.RLock()
    defer metrics.mu.RUnlock()

    result := make(map[int]int64, len(metrics.counts))
    for k, v := range metrics.counts {
        result[k] = v
    }
    return result
}

// Expose as Prometheus metrics
func RegisterPrometheusMetrics(registry *prometheus.Registry) {
    // ...
}
```

---

#### 3.4 Context-Aware Errors

```go
// Store request context in errors
type AppError struct {
    // ... existing fields
    Context map[string]interface{} `json:"context,omitempty"`
}

func (e *AppError) WithContext(key string, value interface{}) *AppError {
    if e.Context == nil {
        e.Context = make(map[string]interface{})
    }
    e.Context[key] = value
    return e
}

// Usage
appErr := NewAppError("DB query failed", CodeInternalServerError, err).
    WithContext("query", "SELECT * FROM users").
    WithContext("user_id", 123).
    WithContext("trace_id", traceID)
```

---

## File-by-File Recommendations

### apperrors.go
- ✅ **Keep**: Core error types well-designed
- 🔧 **Change**: Code from string to int
- 🔧 **Add**: Optional stack trace collection
- 🔧 **Fix**: NewFieldError doesn't need to be on AppError (utility function)

### errutil.go
- ✅ **Keep**: Excellent generics implementation
- 📖 **Document**: Add usage examples in comments
- ➕ **Add**: `Is[T error]()` function for completeness

### errcodes.go
- 🔧 **Fix**: Typo in `NewHTTPStatsuCodeAndMessage`
- 🔧 **Remove**: Success field (always false)
- ➕ **Add**: Error code constants

### dberrcodes.go
- 🔧 **Fix**: Inconsistent Code field (HTTP vs SQLSTATE)
- 🧹 **Clean**: Remove unused error definitions
- 📖 **Document**: Which errors are actually used

### apperrorhandlers.go
- 🔴 **Refactor**: Extract common response logic
- 🔧 **Fix**: Complete migration to generics
- 🧹 **Remove**: All commented code
- 🔧 **Fix**: `HandleErrorWithStatusCodeAndMessage` to call ctx.JSON()

### stacktrace.go
- 🔧 **Add**: Configurable enable/disable
- 🔧 **Add**: Configurable max depth
- 🔒 **Security**: Remove pointers from JSON
- 📖 **Document**: Stack depth limit

### apierrorresponse.go
- ✅ **Keep**: Clean separation of concerns
- 🔧 **Consider**: Unified response format for success + error

### helper.go
- ✅ **Keep**: Simple mapping function
- 📖 **Document**: Should match constants when Code is int

---

## Testing Recommendations

### Unit Tests Needed

1. **errutil_test.go**
   ```go
   func TestAs(t *testing.T) {
       // Test type assertion with generics
   }

   func TestFind(t *testing.T) {
       // Test error finding
   }

   func TestMultiErrorUnwrap(t *testing.T) {
       // Test multi-error unwrapping
   }
   ```

2. **apperrors_test.go**
   ```go
   func TestAppErrorUnwrap(t *testing.T) {
       // Test error chain unwrapping
   }

   func TestFieldErrors(t *testing.T) {
       // Test field error handling
   }
   ```

3. **handlers_test.go**
   ```go
   func TestHandleBindingError(t *testing.T) {
       // Test binding error handling
   }

   func TestHandleDBError(t *testing.T) {
       // Test PostgreSQL error mapping
   }
   ```

4. **stacktrace_test.go**
   ```go
   func TestStackTraceCollection(t *testing.T) {
       // Test stack trace accuracy
   }

   func BenchmarkStackTrace(b *testing.B) {
       // Measure performance impact
   }
   ```

---

## Performance Considerations

### Current Performance Characteristics

**Good**:
- ✅ Generics avoid reflection overhead
- ✅ Pre-defined error codes (no allocation)
- ✅ Efficient error unwrapping

**Concerns**:
- ⚠️ Stack trace collected for every error (runtime.Callers overhead)
- ⚠️ String-to-int conversion for Code field
- ⚠️ Deep error chains with multiple unwraps

### Optimization Opportunities

1. **Lazy stack trace collection**
   ```go
   type AppError struct {
       // ... existing fields
       stackOnce sync.Once
       Stack     *stackTrace
   }

   func (e *AppError) StackTrace() *stackTrace {
       e.stackOnce.Do(func() {
           if EnableStackTraces {
               e.Stack = collectStackTrace()
           }
       })
       return e.Stack
   }
   ```

2. **Error pooling for hot paths**
   ```go
   var appErrorPool = sync.Pool{
       New: func() interface{} {
           return &AppError{}
       },
   }

   func NewAppErrorPooled(msg string, code int, err error) *AppError {
       e := appErrorPool.Get().(*AppError)
       e.Message = msg
       e.Code = code
       e.OriginalError = err
       return e
   }

   func (e *AppError) Release() {
       *e = AppError{}  // Reset
       appErrorPool.Put(e)
   }
   ```

---

## Summary of Recommendations

### Must-Fix (Priority 1)

| Issue | Impact | Effort | Benefit |
|-------|--------|--------|---------|
| Code field: string→int | High | Medium | Type safety, performance, cleaner code |
| Eliminate handler duplication | Medium | Medium | Maintainability, consistency |
| Complete generics migration | Low | Low | Consistency, performance |
| Remove commented code | Low | Low | Code cleanliness |

### Should-Fix (Priority 2)

| Enhancement | Impact | Effort | Benefit |
|-------------|--------|--------|---------|
| Error code constants | Medium | Low | Type safety, autocomplete |
| Configurable stack traces | Medium | Medium | Production performance |
| Remove Success field | Low | Low | Cleaner API |
| Security: Hide pointers | Low | Low | ASLR protection |

### Nice-to-Have (Priority 3)

| Feature | Impact | Effort | Benefit |
|---------|--------|--------|---------|
| Error builder pattern | Medium | Medium | Developer experience |
| Structured logging | Medium | Low | Observability |
| Error metrics | Low | Medium | Monitoring |
| Context-aware errors | Low | Low | Debugging |

---

## Migration Plan

### Phase 1: Non-Breaking Changes (Week 1)
1. Add error code constants
2. Remove commented code
3. Fix typo in function name
4. Complete generics migration
5. Add unit tests

### Phase 2: Breaking Changes (Week 2)
1. Change Code from string to int
2. Refactor handlers to use common function
3. Update all call sites

### Phase 3: Enhancements (Week 3)
1. Add configurable stack traces
2. Remove pointer addresses from stack frames
3. Add builder pattern
4. Add metrics

### Phase 4: Documentation (Week 4)
1. Add usage examples
2. Document best practices
3. Create migration guide
4. Update README

---

## Conclusion

The api-errors module is **well-architected** with excellent use of modern Go features (generics). The main areas for improvement are:

1. **Consistency**: Complete the migration to generics
2. **Type Safety**: Use int for Code instead of string
3. **Maintainability**: Reduce duplication in handlers
4. **Performance**: Make stack traces optional
5. **Security**: Don't expose memory addresses

**Overall Rating**: 8.5/10 - Excellent foundation with clear paths for improvement.
