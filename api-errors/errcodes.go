package apierrors

// statusCodeAndMessage represents a structured error with an associated HTTP status code and message.
// This struct is used to standardize error responses across the application.
type statusCodeAndMessage struct {
	StatusCode int    `json:"status_code"` // HTTP status code representing the error type (e.g., 404 for Not Found).
	Message    string `json:"message"`     // Descriptive message providing details about the error.
	Success    bool   `json:"success"`     // Success flag to indicate if the operation was successful.
}

// ConstructHTTPError creates a new instance of HTTPError with the provided status code and message.
func NewHTTPStatsuCodeAndMessage(statusCode int, message string) statusCodeAndMessage {
	return statusCodeAndMessage{
		StatusCode: statusCode,
		Message:    message,
		Success:    false,
	}
}

var (
	// Client-side errors (400 range).
	HTTPErrorBadRequest         statusCodeAndMessage = statusCodeAndMessage{StatusCode: 400, Message: "Bad Request", Success: false}          // 400 - General client-side error.
	HTTPErrorUnauthorized       statusCodeAndMessage = statusCodeAndMessage{StatusCode: 401, Message: "Unauthorized", Success: false}         // 401 - Authentication is required and has failed or has not yet been provided.
	HTTPErrorForbidden          statusCodeAndMessage = statusCodeAndMessage{StatusCode: 403, Message: "Forbidden", Success: false}            // 403 - Client does not have permission to access this resource.
	HTTPErrorNotFound           statusCodeAndMessage = statusCodeAndMessage{StatusCode: 404, Message: "Not Found", Success: false}            // 404 - Requested resource could not be found.
	HTTPErrorMethodNotAllowed   statusCodeAndMessage = statusCodeAndMessage{StatusCode: 405, Message: "Method Not Allowed", Success: false}   // 405 - HTTP method not supported.
	HTTPErrorRequestTimeout     statusCodeAndMessage = statusCodeAndMessage{StatusCode: 408, Message: "Request Timeout", Success: false}      // 408 - Request took too long.
	HTTPErrorConflict           statusCodeAndMessage = statusCodeAndMessage{StatusCode: 409, Message: "Conflict", Success: false}             // 409 - Resource conflict, like duplicate data.
	HTTPErrorGone               statusCodeAndMessage = statusCodeAndMessage{StatusCode: 410, Message: "Gone", Success: false}                 // 410 - Resource is no longer available.
	HTTPErrorTooManyRequests    statusCodeAndMessage = statusCodeAndMessage{StatusCode: 429, Message: "Too Many Requests", Success: false}    // 429 - Rate limiting error.
	HTTPErrorInvalidContentType statusCodeAndMessage = statusCodeAndMessage{StatusCode: 415, Message: "Invalid Content Type", Success: false} // 415 - Unsupported content type in request.

	// Server-side errors (500 range).
	HTTPErrorServerError        statusCodeAndMessage = statusCodeAndMessage{StatusCode: 500, Message: "Internal Server Error", Success: false} // 500 - Generic server error.
	HTTPErrorNotImplemented     statusCodeAndMessage = statusCodeAndMessage{StatusCode: 501, Message: "Not Implemented", Success: false}       // 501 - Functionality not implemented.
	HTTPErrorBadGateway         statusCodeAndMessage = statusCodeAndMessage{StatusCode: 502, Message: "Bad Gateway", Success: false}           // 502 - Received an invalid response from upstream server.
	HTTPErrorServiceUnavailable statusCodeAndMessage = statusCodeAndMessage{StatusCode: 503, Message: "Service Unavailable", Success: false}   // 503 - Service is temporarily unavailable.
	HTTPErrorGatewayTimeout     statusCodeAndMessage = statusCodeAndMessage{StatusCode: 504, Message: "Gateway Timeout", Success: false}       // 504 - Timeout in gateway or proxy.

	// Application-specific error codes.
	AppErrorValidationError    statusCodeAndMessage = statusCodeAndMessage{StatusCode: 422, Message: "Validation Error", Success: false}       // Input validation failed.
	AppErrorBindingError       statusCodeAndMessage = statusCodeAndMessage{StatusCode: 400, Message: "Binding Error", Success: false}          // Error binding request parameters.
	AppErrorResourceExhausted  statusCodeAndMessage = statusCodeAndMessage{StatusCode: 429, Message: "Resource Exhausted", Success: false}     // Resource limits have been exceeded, e.g., storage quota.
	AppErrorBusinessRule       statusCodeAndMessage = statusCodeAndMessage{StatusCode: 400, Message: "Business Rule Error", Success: false}    // Business logic condition not met.
	AppErrorDeprecationWarning statusCodeAndMessage = statusCodeAndMessage{StatusCode: 299, Message: "Deprecation Warning", Success: false}    // Using deprecated features or API versions.
	AppErrorDataConsistency    statusCodeAndMessage = statusCodeAndMessage{StatusCode: 500, Message: "Data Consistency Error", Success: false} // Data integrity issues like inconsistent data states.

	// Database-related error codes.
	DBErrorGeneral             statusCodeAndMessage = statusCodeAndMessage{StatusCode: 500, Message: "DB Error", Success: false}             // General database error.
	DBErrorRecordNotFound      statusCodeAndMessage = statusCodeAndMessage{StatusCode: 404, Message: "Record Not Found", Success: false}     // Specific database record could not be found.
	DBErrorDuplicateRecord     statusCodeAndMessage = statusCodeAndMessage{StatusCode: 409, Message: "Duplicate Record", Success: false}     // Duplicate record insertion.
	DBErrorTransactionFailure  statusCodeAndMessage = statusCodeAndMessage{StatusCode: 500, Message: "Transaction Failure", Success: false}  // Database transaction failed.
	DBErrorConstraintViolation statusCodeAndMessage = statusCodeAndMessage{StatusCode: 400, Message: "Constraint Violation", Success: false} // Constraint check failed (e.g., foreign key).

	// Integration and API-to-API communication error codes.
	IntegrationErrorTimeout           statusCodeAndMessage = statusCodeAndMessage{StatusCode: 504, Message: "Timeout Error", Success: false}       // Timeout during API call to external service.
	IntegrationErrorRateLimitExceeded statusCodeAndMessage = statusCodeAndMessage{StatusCode: 429, Message: "Rate Limit Exceeded", Success: false} // Rate-limiting error when interacting with external services.
	IntegrationErrorNetworkError      statusCodeAndMessage = statusCodeAndMessage{StatusCode: 503, Message: "Network Error", Success: false}       // General network error (e.g., DNS failure).
	IntegrationErrorDependencyFailure statusCodeAndMessage = statusCodeAndMessage{StatusCode: 424, Message: "Dependency Failure", Success: false}  // Failed due to external service dependency.
	IntegrationErrorInvalidResponse   statusCodeAndMessage = statusCodeAndMessage{StatusCode: 502, Message: "Invalid Response", Success: false}    // Received invalid or unexpected response from an external service.

	// Security-related error codes.
	SecurityErrorAuthenticationFailed statusCodeAndMessage = statusCodeAndMessage{StatusCode: 401, Message: "Authentication Failed", Success: false} // Failed authentication (e.g., incorrect password).
	SecurityErrorAuthorizationFailed  statusCodeAndMessage = statusCodeAndMessage{StatusCode: 403, Message: "Authorization Failed", Success: false}  // User lacks required permissions.
	SecurityErrorTokenExpired         statusCodeAndMessage = statusCodeAndMessage{StatusCode: 401, Message: "Token Expired", Success: false}         // Authentication token has expired.
	SecurityErrorTokenInvalid         statusCodeAndMessage = statusCodeAndMessage{StatusCode: 401, Message: "Token Invalid", Success: false}         // Provided token is invalid.
	SecurityErrorCSRFTokenInvalid     statusCodeAndMessage = statusCodeAndMessage{StatusCode: 403, Message: "CSRF Token Invalid", Success: false}    // Invalid CSRF token for request.

	// File-related error codes.
	FileErrorNotFound        statusCodeAndMessage = statusCodeAndMessage{StatusCode: 404, Message: "File Not Found", Success: false}        // Requested file could not be found.
	FileErrorUploadFailed    statusCodeAndMessage = statusCodeAndMessage{StatusCode: 500, Message: "File Upload Failed", Success: false}    // Error during file upload.
	FileErrorTooLarge        statusCodeAndMessage = statusCodeAndMessage{StatusCode: 413, Message: "File Too Large", Success: false}        // Uploaded file size exceeds the allowed limit.
	FileErrorUnsupportedType statusCodeAndMessage = statusCodeAndMessage{StatusCode: 415, Message: "Unsupported File Type", Success: false} // File type is not supported.
	FileErrorReadError       statusCodeAndMessage = statusCodeAndMessage{StatusCode: 500, Message: "File Read Error", Success: false}       // Error reading file content.
	FileErrorWriteError      statusCodeAndMessage = statusCodeAndMessage{StatusCode: 500, Message: "File Write Error", Success: false}      // Error writing to file.

	// Custom or unknown error codes.
	CustomError  statusCodeAndMessage = statusCodeAndMessage{StatusCode: 422, Message: "Custom Error", Success: false}  // For custom or specific use-case errors.
	UnknownError statusCodeAndMessage = statusCodeAndMessage{StatusCode: 520, Message: "Unknown Error", Success: false} // Catch-all for unclassified errors.
)
