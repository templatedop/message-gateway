# Models Package - Request Validation

This package contains all API request models with type-safe, generated validation using [govalid](https://github.com/templatedop/govalid).

## Quick Start

### Using Validation

```go
import "MgApplication/models"

// In your handler
func CreateSMSHandler(ctx *gin.Context) {
    var req models.CreateSMSRequest

    // Bind JSON
    if err := ctx.ShouldBindJSON(&req); err != nil {
        return handleError(ctx, err)
    }

    // Validate
    if err := req.Validate(); err != nil {
        return handleValidationError(ctx, err)
    }

    // Process valid request
}
```

### Adding New Request Models

1. **Define struct with validation tags:**
```go
type MyRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"gt=0,lt=150"`
}
```

2. **Generate validation code:**
```bash
govalid ./models
```

3. **Use in handler:**
```go
var req models.MyRequest
if err := req.Validate(); err != nil {
    // handle error
}
```

## Available Validation Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Field must not be empty | `validate:"required"` |
| `email` | Valid email format | `validate:"email"` |
| `gt=N` | Greater than N | `validate:"gt=0"` |
| `lt=N` | Less than N | `validate:"lt=100"` |
| `minlength=N` | Min string length | `validate:"minlength=5"` |
| `maxlength=N` | Max string length | `validate:"maxlength=100"` |

[See full list in GOVALID_MIGRATION.md](../docs/GOVALID_MIGRATION.md#supported-validation-tags)

## Generated Files

Generated validation files follow the pattern: `requests_<structname>_validator.go`

These files are auto-generated and should not be manually edited. To regenerate:

```bash
govalid ./models
```

## Performance

Govalid provides:
- **Zero heap allocations** during validation
- **5-44x faster** than reflection-based validators
- **Type-safe** validation at compile time

## Error Handling

Validation errors provide detailed information:

```go
if err := req.Validate(); err != nil {
    if validationErrors, ok := err.(govaliderrors.ValidationErrors); ok {
        for _, ve := range validationErrors {
            fmt.Printf("Field: %s\n", ve.Path)
            fmt.Printf("Reason: %s\n", ve.Reason)
            fmt.Printf("Value: %v\n", ve.Value)
        }
    }
}
```

## Request Models

### SMS Request Models
- `CreateSMSRequest` - Create SMS message request
- `CreateTestSMSRequest` - Test SMS request

### Template Models
- `CreateTemplateRequest` - Create new template
- `UpdateTemplateRequest` - Update existing template
- `FetchTemplateRequest` - Fetch template by ID
- `ListTemplatesRequest` - List all templates

### Provider Models
- `CreateMessageProviderRequest` - Create message provider
- `UpdateMessageProviderRequest` - Update provider
- `FetchMessageProviderRequest` - Fetch provider by ID

### Report Models
- `SentSMSStatusReportRequest` - SMS status report
- `AggregateSMSUsageReportRequest` - Usage aggregation report

### Bulk SMS Models
- `InitiateBulkSMSRequest` - Initiate bulk SMS
- `ValidateTestSMSRequest` - Validate test SMS
- `SendBulkSMSRequest` - Send bulk SMS

## Documentation

For detailed documentation, see:
- [GOVALID_MIGRATION.md](../docs/GOVALID_MIGRATION.md) - Complete migration guide
- [govalid README](https://github.com/templatedop/govalid#readme) - Library documentation

## Regenerating Validation

After modifying struct tags:

```bash
# From project root
govalid ./models
```

## Example

Complete example of a validated request:

```go
// Define request
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,minlength=2"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"gte=18,lte=120"`
    Password string `json:"password" validate:"required,minlength=8"`
}

// In handler
func CreateUser(ctx *gin.Context) {
    var req models.CreateUserRequest

    if err := ctx.ShouldBindJSON(&req); err != nil {
        return apierrors.HandleBindingError(ctx, err)
    }

    if err := req.Validate(); err != nil {
        return apierrors.HandleValidationError(ctx, err)
    }

    // Create user with validated data
    user, err := service.CreateUser(req)
    if err != nil {
        return apierrors.HandleError(ctx, err)
    }

    ctx.JSON(http.StatusCreated, user)
}
```
