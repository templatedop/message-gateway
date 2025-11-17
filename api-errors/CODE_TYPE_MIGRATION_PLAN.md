# AppError.Code Migration Plan: String to Int

## Current Problem

The codebase has **two different types of error codes** that are being conflated:

1. **HTTP Status Codes** (integers): 400, 404, 500, etc.
2. **PostgreSQL SQLSTATE Codes** (strings): "42P01", "08", "23", etc.

Currently, `AppError.Code` is a `string`, which causes:
- **Performance overhead**: Converting int → string (`strconv.Itoa`) and string → int (`strconv.Atoi`)
- **Type confusion**: Mixing HTTP codes with database codes
- **Incorrect data**: Database handlers are passing SQLSTATE codes (e.g., "08") to `AppError.Code` instead of HTTP status codes (e.g., 500)

## Proposed Solution

### 1. Add HTTPStatusCode to dbError

**File**: `api-errors/dberrcodes.go`

**Current**:
```go
type dbError struct {
    Code    string  // PostgreSQL SQLSTATE code
    Message string
}

var (
    DBConnectionException = dbError{Code: "08", Message: "Connection Exception"} // HTTP 500 Internal Server Error
)
```

**Proposed**:
```go
type dbError struct {
    Code           string  // PostgreSQL SQLSTATE code (e.g., "08", "42P01")
    HTTPStatusCode int     // HTTP status code for API responses (e.g., 500, 400)
    Message        string
}

var (
    // Server-side errors (500 range)
    DBConnectionException = dbError{
        Code:           "08",
        HTTPStatusCode: 500,
        Message:        "Connection Exception",
    }

    DBNoData = dbError{
        Code:           "02",
        HTTPStatusCode: 404,
        Message:        "No Data",
    }

    // ... update all 50+ dbError definitions
)
```

### 2. Change AppError.Code to int

**File**: `api-errors/apperrors.go`

**Current**:
```go
type AppError struct {
    ID            string       `json:"id,omitempty"`
    Code          string       `json:"code"`
    Message       string       `json:"message"`
    FieldErrors   []FieldError `json:"field_errors,omitempty"`
    Stack         *stackTrace  `json:"-"`
    OriginalError error        `json:"-"`
}

func NewAppError(message string, code string, originalError error) AppError {
    return AppError{
        Stack:         collectStackTrace(),
        Message:       message,
        Code:          code,
        OriginalError: originalError,
    }
}
```

**Proposed**:
```go
type AppError struct {
    ID            string       `json:"id,omitempty"`
    Code          int          `json:"code"`  // ← Changed to int
    Message       string       `json:"message"`
    FieldErrors   []FieldError `json:"field_errors,omitempty"`
    Stack         *stackTrace  `json:"-"`
    OriginalError error        `json:"-"`
}

func NewAppError(message string, code int, originalError error) AppError {  // ← code is now int
    return AppError{
        Stack:         collectStackTrace(),
        Message:       message,
        Code:          code,  // ← No conversion needed
        OriginalError: originalError,
    }
}

func NewAppErrorWithId(message string, code int, originalError error, id string) AppError {  // ← code is now int
    return AppError{
        Stack:         collectStackTrace(),
        Message:       message,
        Code:          code,
        OriginalError: originalError,
        ID:            id,
    }
}
```

### 3. Update respondWithError Helper

**File**: `api-errors/apperrorhandlers.go:33`

**Current**:
```go
func respondWithError(
    ctx *gin.Context,
    statusCodeAndMessage statusCodeAndMessage,
    message string,
    err error,
) {
    appError := NewAppError(message, strconv.Itoa(statusCodeAndMessage.StatusCode), err)  // ← Converting int to string
    apiErrorResponse := NewHTTPAPIErrorResponse(statusCodeAndMessage, appError)
    ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}
```

**Proposed**:
```go
func respondWithError(
    ctx *gin.Context,
    statusCodeAndMessage statusCodeAndMessage,
    message string,
    err error,
) {
    appError := NewAppError(message, statusCodeAndMessage.StatusCode, err)  // ← Direct int, no conversion
    apiErrorResponse := NewHTTPAPIErrorResponse(statusCodeAndMessage, appError)
    ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}
```

### 4. Update Database Error Handlers

**File**: `api-errors/apperrorhandlers.go`

**Current** (lines 222, 227, etc.):
```go
if ctxErr == context.DeadlineExceeded || ctxErr == context.Canceled {
    appError = NewAppError(DBConnectionException.Message, DBConnectionException.Code, err)  // ← Passing "08" (wrong!)
    // ...
}

if errorsutil.Is(err, sql.ErrNoRows) {
    appError = NewAppError(DBNoData.Message, DBNoData.Code, err)  // ← Passing "02" (wrong!)
    // ...
}
```

**Proposed**:
```go
if ctxErr == context.DeadlineExceeded || ctxErr == context.Canceled {
    appError = NewAppError(DBConnectionException.Message, DBConnectionException.HTTPStatusCode, err)  // ← Passing 500 (correct!)
    // ...
}

if errorsutil.Is(err, sql.ErrNoRows) {
    appError = NewAppError(DBNoData.Message, DBNoData.HTTPStatusCode, err)  // ← Passing 404 (correct!)
    // ...
}
```

### 5. Remove strconv Conversions

**Current locations using strconv.Itoa()** (9 instances):
- Line 33: respondWithError (fixed above)
- Line 94: HandleBindingError
- Line 143: HandleBindingError
- Line 169: HandleBindingError
- Line 300: HandleDBError
- Line 334: HandleDBError
- Line 366: HandleErrorWithMessage
- Line 629: HandleClientError
- Line 719: HandleServerError

**Current locations using strconv.Atoi()** (2 instances):
- Line 203: `statusCode, convErr := strconv.Atoi(appErr.Code)` → becomes `statusCode := appErr.Code`
- Line 741: `statusCode, convErr := strconv.Atoi(appErr.Code)` → becomes `statusCode := appErr.Code`

**All will be updated to pass/use int directly**.

## Migration Steps

### Step 1: Update dbError Structure (dberrcodes.go)
- Add `HTTPStatusCode int` field to `dbError` struct
- Update all 50+ `dbError` variable declarations to include HTTP status codes

### Step 2: Update AppError Structure (apperrors.go)
- Change `Code string` to `Code int`
- Update `NewAppError` signature: `code string` → `code int`
- Update `NewAppErrorWithId` signature: `code string` → `code int`

### Step 3: Update Handler Functions (apperrorhandlers.go)
- Replace all `strconv.Itoa(statusCode)` with direct `statusCode`
- Replace all `DBError.Code` with `DBError.HTTPStatusCode`
- Replace all `strconv.Atoi(appErr.Code)` with direct `appErr.Code`
- Remove error handling for conversion failures

### Step 4: Verify and Test
- Run `go build` to verify compilation
- Check JSON serialization still works correctly
- Verify no breaking changes to API response format

## Impact Analysis

### Files Modified
1. ✅ `api-errors/dberrcodes.go` - Add HTTPStatusCode field, update ~50 variable declarations
2. ✅ `api-errors/apperrors.go` - Change Code type and constructor signatures
3. ✅ `api-errors/apperrorhandlers.go` - Update ~20 call sites

### Breaking Changes
- **None for API consumers**: JSON still serializes as `"code": 500` (just int instead of string)
- **Internal only**: Function signatures change, but all calls are within api-errors package

### Benefits
- ✅ **Eliminate 9 `strconv.Itoa()` calls** (performance improvement)
- ✅ **Eliminate 2 `strconv.Atoi()` calls with error handling** (simpler code)
- ✅ **Fix incorrect SQLSTATE codes in AppError** (correctness improvement)
- ✅ **Type safety**: Compiler enforces int usage
- ✅ **Better semantics**: Code field now accurately represents HTTP status code

### Risks
- **Low risk**: Changes are mostly mechanical
- **JSON compatibility**: Clients expecting `"code": "500"` will get `"code": 500`, but this is valid JSON and should work in most languages

## Example: Before and After

### Before
```go
// Creating error with conversion
appError := NewAppError("Not found", strconv.Itoa(404), err)

// Reading error with conversion
statusCode, convErr := strconv.Atoi(appErr.Code)
if convErr != nil {
    // Handle conversion error
}

// Database error passing SQLSTATE code (wrong!)
appError = NewAppError(DBConnectionException.Message, DBConnectionException.Code, err)  // Code = "08"
```

### After
```go
// Creating error - direct int
appError := NewAppError("Not found", 404, err)

// Reading error - direct int
statusCode := appErr.Code

// Database error passing HTTP status code (correct!)
appError = NewAppError(DBConnectionException.Message, DBConnectionException.HTTPStatusCode, err)  // Code = 500
```

## JSON Response Comparison

### Before (string)
```json
{
  "status_code": 500,
  "message": "Internal Server Error",
  "success": false,
  "error": {
    "code": "500",
    "message": "Connection Exception"
  }
}
```

### After (int)
```json
{
  "status_code": 500,
  "message": "Internal Server Error",
  "success": false,
  "error": {
    "code": 500,
    "message": "Connection Exception"
  }
}
```

**Note**: The only difference is `"code": "500"` → `"code": 500`. Both are valid JSON. Most parsers handle this transparently.
