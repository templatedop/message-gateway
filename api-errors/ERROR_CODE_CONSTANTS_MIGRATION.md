# Error Code Constants Migration

## Migration Completed ✅

**Date**: 2025-11-17
**Status**: Complete - Replaced all magic numbers with http.Status constants

## Overview

Replaced all magic number HTTP status codes (400, 404, 500, etc.) in `errcodes.go` with named constants from the `net/http` package.

## Changes Made

### File Modified: `api-errors/errcodes.go`

**Added import**:
```go
import "net/http"
```

**Replaced all magic numbers with constants** (42 instances):

| Category | Variable | Before | After |
|----------|----------|--------|-------|
| **Client Errors (4xx)** | | | |
| | HTTPErrorBadRequest | `StatusCode: 400` | `StatusCode: http.StatusBadRequest` |
| | HTTPErrorUnauthorized | `StatusCode: 401` | `StatusCode: http.StatusUnauthorized` |
| | HTTPErrorForbidden | `StatusCode: 403` | `StatusCode: http.StatusForbidden` |
| | HTTPErrorNotFound | `StatusCode: 404` | `StatusCode: http.StatusNotFound` |
| | HTTPErrorMethodNotAllowed | `StatusCode: 405` | `StatusCode: http.StatusMethodNotAllowed` |
| | HTTPErrorRequestTimeout | `StatusCode: 408` | `StatusCode: http.StatusRequestTimeout` |
| | HTTPErrorConflict | `StatusCode: 409` | `StatusCode: http.StatusConflict` |
| | HTTPErrorGone | `StatusCode: 410` | `StatusCode: http.StatusGone` |
| | HTTPErrorInvalidContentType | `StatusCode: 415` | `StatusCode: http.StatusUnsupportedMediaType` |
| | HTTPErrorTooManyRequests | `StatusCode: 429` | `StatusCode: http.StatusTooManyRequests` |
| **Server Errors (5xx)** | | | |
| | HTTPErrorServerError | `StatusCode: 500` | `StatusCode: http.StatusInternalServerError` |
| | HTTPErrorNotImplemented | `StatusCode: 501` | `StatusCode: http.StatusNotImplemented` |
| | HTTPErrorBadGateway | `StatusCode: 502` | `StatusCode: http.StatusBadGateway` |
| | HTTPErrorServiceUnavailable | `StatusCode: 503` | `StatusCode: http.StatusServiceUnavailable` |
| | HTTPErrorGatewayTimeout | `StatusCode: 504` | `StatusCode: http.StatusGatewayTimeout` |
| **Application Errors** | | | |
| | AppErrorValidationError | `StatusCode: 422` | `StatusCode: http.StatusUnprocessableEntity` |
| | AppErrorBindingError | `StatusCode: 400` | `StatusCode: http.StatusBadRequest` |
| | AppErrorResourceExhausted | `StatusCode: 429` | `StatusCode: http.StatusTooManyRequests` |
| | AppErrorBusinessRule | `StatusCode: 400` | `StatusCode: http.StatusBadRequest` |
| | AppErrorDeprecationWarning | `StatusCode: 299` | (kept as literal - non-standard) |
| | AppErrorDataConsistency | `StatusCode: 500` | `StatusCode: http.StatusInternalServerError` |
| **Database Errors** | | | |
| | DBErrorGeneral | `StatusCode: 500` | `StatusCode: http.StatusInternalServerError` |
| | DBErrorRecordNotFound | `StatusCode: 404` | `StatusCode: http.StatusNotFound` |
| | DBErrorDuplicateRecord | `StatusCode: 409` | `StatusCode: http.StatusConflict` |
| | DBErrorTransactionFailure | `StatusCode: 500` | `StatusCode: http.StatusInternalServerError` |
| | DBErrorConstraintViolation | `StatusCode: 400` | `StatusCode: http.StatusBadRequest` |
| **Integration Errors** | | | |
| | IntegrationErrorTimeout | `StatusCode: 504` | `StatusCode: http.StatusGatewayTimeout` |
| | IntegrationErrorRateLimitExceeded | `StatusCode: 429` | `StatusCode: http.StatusTooManyRequests` |
| | IntegrationErrorNetworkError | `StatusCode: 503` | `StatusCode: http.StatusServiceUnavailable` |
| | IntegrationErrorDependencyFailure | `StatusCode: 424` | `StatusCode: http.StatusFailedDependency` |
| | IntegrationErrorInvalidResponse | `StatusCode: 502` | `StatusCode: http.StatusBadGateway` |
| **Security Errors** | | | |
| | SecurityErrorAuthenticationFailed | `StatusCode: 401` | `StatusCode: http.StatusUnauthorized` |
| | SecurityErrorAuthorizationFailed | `StatusCode: 403` | `StatusCode: http.StatusForbidden` |
| | SecurityErrorTokenExpired | `StatusCode: 401` | `StatusCode: http.StatusUnauthorized` |
| | SecurityErrorTokenInvalid | `StatusCode: 401` | `StatusCode: http.StatusUnauthorized` |
| | SecurityErrorCSRFTokenInvalid | `StatusCode: 403` | `StatusCode: http.StatusForbidden` |
| **File Errors** | | | |
| | FileErrorNotFound | `StatusCode: 404` | `StatusCode: http.StatusNotFound` |
| | FileErrorUploadFailed | `StatusCode: 500` | `StatusCode: http.StatusInternalServerError` |
| | FileErrorTooLarge | `StatusCode: 413` | `StatusCode: http.StatusRequestEntityTooLarge` |
| | FileErrorUnsupportedType | `StatusCode: 415` | `StatusCode: http.StatusUnsupportedMediaType` |
| | FileErrorReadError | `StatusCode: 500` | `StatusCode: http.StatusInternalServerError` |
| | FileErrorWriteError | `StatusCode: 500` | `StatusCode: http.StatusInternalServerError` |
| **Custom Errors** | | | |
| | CustomError | `StatusCode: 422` | `StatusCode: http.StatusUnprocessableEntity` |
| | UnknownError | `StatusCode: 520` | (kept as literal - Cloudflare non-standard) |

## Non-Standard Codes Kept as Literals

Two status codes were intentionally kept as numeric literals with explanatory comments:

1. **299** (AppErrorDeprecationWarning) - Non-standard code for deprecation warnings
2. **520** (UnknownError) - Cloudflare-specific "Unknown Error" code

These are not defined in the `net/http` package and have been documented with comments explaining their non-standard status.

## Benefits

### 1. **Readability**
```go
// Before
HTTPErrorNotFound = statusCodeAndMessage{StatusCode: 404, Message: "Not Found", Success: false}

// After
HTTPErrorNotFound = statusCodeAndMessage{StatusCode: http.StatusNotFound, Message: "Not Found", Success: false}
```
The constant name `http.StatusNotFound` is self-documenting.

### 2. **Maintainability**
- Changes to HTTP status codes (unlikely but possible) only need to be updated in one place (`net/http` package)
- Consistent with Go standard library conventions
- Easier to search and refactor

### 3. **Type Safety**
- Constants are type-checked by the compiler
- Prevents typos (e.g., `404` vs `440`)
- IDE autocomplete support

### 4. **Standards Compliance**
- Uses official Go standard library constants
- Aligns with Go community best practices
- Matches patterns used in well-maintained Go projects

## Example Usage

### Before Migration
```go
var HTTPErrorNotFound = statusCodeAndMessage{
    StatusCode: 404,  // Magic number
    Message:    "Not Found",
    Success:    false,
}
```

### After Migration
```go
var HTTPErrorNotFound = statusCodeAndMessage{
    StatusCode: http.StatusNotFound,  // Named constant
    Message:    "Not Found",
    Success:    false,
}
```

## Standard HTTP Status Constants Used

From `net/http` package:
- `StatusBadRequest` (400)
- `StatusUnauthorized` (401)
- `StatusForbidden` (403)
- `StatusNotFound` (404)
- `StatusMethodNotAllowed` (405)
- `StatusRequestTimeout` (408)
- `StatusConflict` (409)
- `StatusGone` (410)
- `StatusRequestEntityTooLarge` (413)
- `StatusUnsupportedMediaType` (415)
- `StatusUnprocessableEntity` (422)
- `StatusFailedDependency` (424)
- `StatusTooManyRequests` (429)
- `StatusInternalServerError` (500)
- `StatusNotImplemented` (501)
- `StatusBadGateway` (502)
- `StatusServiceUnavailable` (503)
- `StatusGatewayTimeout` (504)

## Verification

### Compilation
```bash
cd api-errors && go build -o /dev/null .
# Exit code: 0 (Success)
```

### No Functional Changes
This migration is purely a refactoring:
- ✅ No API changes
- ✅ No behavior changes
- ✅ Same runtime values
- ✅ Same JSON responses

## Statistics

- **Total replacements**: 42 magic numbers → named constants
- **Non-standard codes preserved**: 2 (with documentation)
- **Files modified**: 1 (`errcodes.go`)
- **Lines changed**: ~42 lines
- **New imports**: 1 (`net/http`)

## Related Improvements

This migration completes the error code improvements series:

1. ✅ **AppError.Code string→int migration** - Type safety for error codes
2. ✅ **Remove strconv conversions** - Performance improvement
3. ✅ **Add dbError.HTTPStatusCode** - Proper HTTP code mapping for database errors
4. ✅ **Replace magic numbers with constants** ← **This migration**

## Next Steps (Future)

Optional future improvements:
1. Create custom constants for non-standard codes (299, 520)
2. Add error code validation in tests
3. Document error code usage patterns

## Conclusion

All HTTP status code magic numbers have been successfully replaced with named constants from the `net/http` package, improving code readability, maintainability, and alignment with Go best practices.
