# GoValid Validation Migration

## Overview

This document describes the migration from the custom `api-validation` package to `govalid`, a code generation-based validation library that provides zero-allocation, type-safe validation.

## What Changed

### 1. Validation Library

**Before:** Custom reflection-based validation using `api-validation` package
**After:** Code generation-based validation using `github.com/templatedop/govalid`

### 2. Performance Improvements

- **Zero allocations** during validation
- **5-44x faster** than reflection-based validators
- Type-safe validation functions generated at build time

### 3. Architecture

**Old Architecture:**
```
Handler -> api-validation.ValidateStruct(req) -> Reflection-based validation
```

**New Architecture:**
```
Handler -> req.Validate() -> Generated type-safe validation
```

## Implementation Details

### Request Models

All request models have been moved to the `models` package with proper validation tags:

```go
// models/requests.go
type CreateSMSRequest struct {
    ApplicationID string `json:"application_id" validate:"required"`
    FacilityID    string `json:"facility_id" validate:"required"`
    Priority      int    `json:"priority" validate:"required"`
    MessageText   string `json:"message_text" validate:"required"`
    SenderID      string `json:"sender_id" validate:"required"`
    MobileNumbers string `json:"mobile_numbers" validate:"required"`
    TemplateID    string `json:"template_id" validate:"required"`
}
```

### Generated Validation Code

For each struct with `validate` tags, govalid generates:

1. **Error constants** - Pre-defined errors for each validation rule
2. **Validation function** - `ValidateXxx()` function
3. **Validate method** - Implements `govalid.Validator` interface

Example generated code:

```go
// models/requests_createsmsrequest_validator.go
var (
    ErrNilCreateSMSRequest = errors.New("input CreateSMSRequest is nil")
    ErrCreateSMSRequestApplicationIDRequiredValidation = govaliderrors.ValidationError{
        Reason: "field ApplicationID is required",
        Path:   "CreateSMSRequest.ApplicationID",
        Type:   "required",
    }
    // ... more error constants
)

func ValidateCreateSMSRequest(t *CreateSMSRequest) error {
    if t == nil {
        return ErrNilCreateSMSRequest
    }
    var errs govaliderrors.ValidationErrors
    if t.ApplicationID == "" {
        err := ErrCreateSMSRequestApplicationIDRequiredValidation
        err.Value = t.ApplicationID
        errs = append(errs, err)
    }
    // ... more validation checks
    if len(errs) > 0 {
        return errs
    }
    return nil
}

func (t *CreateSMSRequest) Validate() error {
    return ValidateCreateSMSRequest(t)
}
```

### Handler Updates

All handlers have been updated to use the new validation:

**Before:**
```go
import validation "MgApplication/api-validation"

var req createSMSRequest
if err := ctx.ShouldBindJSON(&req); err != nil {
    apierrors.HandleBindingError(ctx, err)
    return
}

if err := validation.ValidateStruct(req); err != nil {
    apierrors.HandleValidationError(ctx, err)
    return
}
```

**After:**
```go
import "MgApplication/models"

var req models.CreateSMSRequest
if err := ctx.ShouldBindJSON(&req); err != nil {
    apierrors.HandleBindingError(ctx, err)
    return
}

if err := req.Validate(); err != nil {
    apierrors.HandleValidationError(ctx, err)
    return
}
```

## Supported Validation Tags

### Basic Validators
- `required` - Field must not be empty/zero value
- `omitempty` - Skip validation if field is empty

### Numeric Validators
- `gt=N` - Greater than N
- `gte=N` - Greater than or equal to N
- `lt=N` - Less than N
- `lte=N` - Less than or equal to N
- `min=N` - Minimum value N
- `eq=N` - Equal to N
- `ne=N` - Not equal to N

### String Validators
- `maxlength=N` - Maximum string length
- `minlength=N` - Minimum string length
- `length=N` - Exact string length
- `alpha` - Alphabetic characters only
- `alphanum` - Alphanumeric characters only
- `lowercase` - Lowercase characters only
- `oneof=val1 val2` - Must be one of the values
- `contains=text` - Must contain text
- `excludes=text` - Must not contain text

### Format Validators
- `email` - Valid email address
- `url` - Valid URL
- `uuid` - Valid UUID
- `uri` - Valid URI
- `fqdn` - Fully qualified domain name
- `ipv4` - Valid IPv4 address
- `ipv6` - Valid IPv6 address
- `latitude` - Valid latitude
- `longitude` - Valid longitude
- `iscolour` - Valid color code

### Collection Validators
- `maxitems=N` - Maximum items in collection
- `minitems=N` - Minimum items in collection
- `unique` - All items must be unique

### Advanced Validators
- `cel` - Common Expression Language for complex logic
- `required_if` - Required if another field meets criteria
- `required_with` - Required if another field is present
- `excluded_if` - Excluded if another field meets criteria

## Code Generation Workflow

### Manual Generation
```bash
# Install govalid tool
go install github.com/templatedop/govalid/cmd/govalid@latest

# Generate validation code
govalid ./models
```

### Generated Files Location
Generated files are created in the `models/` directory with naming pattern:
```
requests_<structname>_validator.go
```

Example:
- `requests_createsmsrequest_validator.go`
- `requests_createtemplaterequest_validator.go`
- `requests_createmessageproviderrequest_validator.go`

## Migration Checklist

- [x] Install govalid tool
- [x] Update go.mod dependencies
- [x] Create models package with request structs
- [x] Add validation struct tags
- [x] Generate validation code
- [x] Update handler imports
- [x] Replace validation.ValidateStruct calls
- [x] Update utility functions
- [x] Remove api-validation package
- [x] Test compilation
- [x] Document changes

## Benefits

1. **Performance**
   - Zero heap allocations
   - 5-44x faster than reflection-based validation
   - Compile-time type safety

2. **Developer Experience**
   - Clear error messages with field paths
   - Type-safe validation functions
   - IDE autocomplete support

3. **Maintainability**
   - Generated code is readable
   - Easy to add new validation rules
   - Clear separation of concerns

4. **Error Handling**
   - Structured error types
   - Detailed validation errors
   - Aggregated error reporting

## Example Usage

### Basic Validation
```go
req := models.CreateSMSRequest{
    ApplicationID: "123",
    FacilityID:    "facility1",
    Priority:      1,
}

if err := req.Validate(); err != nil {
    // Handle validation error
    log.Printf("Validation failed: %v", err)
}
```

### Error Handling
```go
if err := req.Validate(); err != nil {
    if validationErrors, ok := err.(govaliderrors.ValidationErrors); ok {
        for _, ve := range validationErrors {
            log.Printf("Field: %s, Reason: %s, Value: %v",
                ve.Path, ve.Reason, ve.Value)
        }
    }
}
```

### Using in HTTP Handlers
```go
func CreateHandler(ctx *gin.Context) {
    var req models.CreateSMSRequest

    if err := ctx.ShouldBindJSON(&req); err != nil {
        apierrors.HandleBindingError(ctx, err)
        return
    }

    if err := req.Validate(); err != nil {
        apierrors.HandleValidationError(ctx, err)
        return
    }

    // Process valid request
}
```

## Testing

### Unit Tests
```go
func TestCreateSMSRequestValidation(t *testing.T) {
    tests := []struct {
        name    string
        req     models.CreateSMSRequest
        wantErr bool
    }{
        {
            name: "valid request",
            req: models.CreateSMSRequest{
                ApplicationID: "123",
                FacilityID:    "facility1",
                Priority:      1,
                MessageText:   "Test message",
                SenderID:      "SENDER",
                MobileNumbers: "1234567890",
                TemplateID:    "template1",
            },
            wantErr: false,
        },
        {
            name: "missing application id",
            req: models.CreateSMSRequest{
                FacilityID:    "facility1",
                Priority:      1,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.req.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Troubleshooting

### Regenerating Validation Code
If you modify struct tags, regenerate validation code:
```bash
govalid ./models
```

### Compilation Errors
If you see "undefined: Validate" errors:
1. Ensure validation code is generated
2. Check that struct has `validate` tags
3. Verify govalid is installed

### Import Issues
If imports fail:
```bash
go mod tidy
go mod vendor  # if using vendoring
```

## Future Enhancements

Potential improvements:
1. Add custom validators using CEL expressions
2. Implement validation middleware
3. Add validation benchmarks
4. Create validation test helpers

## References

- [govalid GitHub Repository](https://github.com/templatedop/govalid)
- [govalid Documentation](https://github.com/templatedop/govalid#readme)
- [CEL Specification](https://github.com/google/cel-spec)

## Support

For issues or questions:
1. Check govalid documentation
2. Review generated validation code
3. Test with unit tests
4. Check error messages for details
