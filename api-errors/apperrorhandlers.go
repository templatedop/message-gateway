package apierrors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// HandleNoRouteError handles requests to non-existent routes.
// It creates an application error with a 404 status code and a message indicating
// that the requested path does not exist. The error is then wrapped in an HTTP API
// error response and returned as a JSON response.
//
// Parameters:
//   - ctx: The context of the HTTP request.
//
// Returns:
//   - HTTP 404 Not Found
func HandleNoRouteError(ctx *gin.Context) {
	appError := NewAppError("The requested path does not exist", "404", nil)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorNotFound, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleNoMethodError handles HTTP requests with unsupported methods.
// It creates an application error indicating that the requested HTTP method is not allowed for the specified path,
// and sends an appropriate JSON response with a 404 status code.
//
// Parameters:
//   - ctx: The Gin context for the current request.
//
// Returns:
//   - HTTP 405 HTTP method not supported
func HandleNoMethodError(ctx *gin.Context) {
	appError := NewAppError("The requested HTTP method is not allowed for this path", "405", nil)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorMethodNotAllowed, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleBindingError handles errors that occur during the binding process in a Gin context.
// It checks if the error is a validation error and creates an appropriate AppError with field-specific errors.
// If the error is not a validation error, it handles it as a generic binding error.
// The function then creates a structured HTTP response for the error and sends it as a JSON response.
//
// Parameters:
//   - ctx: The Gin context in which the error occurred.
//   - err: The error that occurred during the binding process.
//
// Returns:
//   - HTTP 400 Bad Request
func HandleBindingError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

	// Check if the error is of type AppError.
	var appErr *AppError
	if appErr, ok := Find[*AppError](err); ok {

		// if errors.As(err, &appErr) {
		apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorBadRequest, *appErr)
		ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
		return
	}

	// Check if the error is a validator.ValidationErrors.
	// var ve validator.ValidationErrors
	// if errors.As(err, &ve) {
	if ve, ok := Find[validator.ValidationErrors](err); ok {

		// Create a new AppError for a binding error.
		newAppErr := NewAppError("Binding error", strconv.Itoa(http.StatusBadRequest), err)
		appErr = &newAppErr // Get a pointer to the newly created AppError

		// Extract field-specific errors.
		var fieldErrors []FieldError
		for _, err := range ve {
			fieldError := appErr.NewFieldError(
				err.Field(),
				err.Value(),
				fmt.Sprintf("Validation failed for '%s' field", err.Field()),
				err.Tag(),
			)
			fieldErrors = append(fieldErrors, fieldError)
		}

		appErr.SetFieldErrors(fieldErrors)
	} else {
		// var syntaxError *json.SyntaxError
		// var unmarshalTypeError *json.UnmarshalTypeError
		// var invalidUnmarshalError *json.InvalidUnmarshalError

		var errMsg string

		_, isSyntaxError := Find[*json.SyntaxError](err)
		unmarshalTypeError, isUnmarshalTypeError := Find[*json.UnmarshalTypeError](err)
		_, isInvalidUnmarshalError := Find[*json.InvalidUnmarshalError](err)

		switch {
		case isSyntaxError, errors.Is(err, io.ErrUnexpectedEOF), isInvalidUnmarshalError:
			errMsg = "Malformed JSON: Check for missing or extra braces, commas, or quotes."
		case errors.Is(err, io.EOF):
			errMsg = "Body cannot be empty"
		case isUnmarshalTypeError:
			if unmarshalTypeError.Field != "" {
				errMsg = fmt.Sprintf(
					"Incorrect JSON type for field '%s' expected '%s' got '%s'",
					unmarshalTypeError.Field,
					unmarshalTypeError.Type,
					unmarshalTypeError.Value,
				)
			} else {
				errMsg = "Malformed JSON or type mismatch at root level"
			}

		default:
			errMsg = "Malformed request"
		}

		er := errors.New(errMsg)
		newAppErr := NewAppError(fmt.Sprintf("Binding error: %v", er), strconv.Itoa(http.StatusBadRequest), er)
		appErr = &newAppErr
	}

	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorBadRequest, *appErr)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleValidationError handles validation errors by checking if the error is of type AppError.
// If it is, it creates an HTTP API error response with a bad request status code and sends it as a JSON response.
// If the error is not of type AppError, it delegates the error handling to the HandleError function.
//
// Parameters:
//   - ctx: The Gin context for the current request.
//   - err: The error to be handled.
//
// Returns:
//   - HTTP 422 Unprocessable Entity
func HandleValidationError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}
	// Assert that the error is of the custom type that contains app error
	// appError, ok := err.(*AppError)
	appError, ok := Find[*AppError](err)
	if !ok {
		apperror := NewAppError(err.Error(), strconv.Itoa(http.StatusUnprocessableEntity), err)
		apiErrorResponse := NewHTTPAPIErrorResponse(AppErrorValidationError, apperror)
		ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
		return
	}
	apiErrorResponse := NewHTTPAPIErrorResponse(AppErrorValidationError, *appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleDBError handles database-related errors and maps them to appropriate HTTP responses.
// It uses the Gin context to send JSON responses based on the type of error encountered.
//
// Parameters:
//   - ctx: The Gin context used to send the JSON response.
//   - err: The error encountered during database operations.
//
// Returns:
//   - HTTP 500 Internal Server Error: For generic server errors or PostgreSQL errors that
//     do not match specific cases.
//   - HTTP 404 Not Found: For "no rows" errors indicating a missing database record.
//   - HTTP 400 Bad Request: For data-related issues, such as invalid input or data exceptions.
//   - HTTP 503 Service Unavailable: For database connection exceptions or service unavailability.
//   - HTTP 409 Conflict: For integrity constraint violations or duplicate records.
//
// The function distinguishes between different types of PostgreSQL errors and maps them to
// corresponding HTTP status codes and error messages. It also handles non-database-related
// errors or unknown errors by sending a generic server error response.
func HandleDBError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

	// Check if the error is of type AppError.
	var appErr *AppError
	if errors.As(err, &appErr) {
		statusCode, convErr := strconv.Atoi(appErr.Code)
		if convErr != nil {
			apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, *appErr)
			ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
			return
		}

		statusCodeAndMessage := mapErrorToHTTP(statusCode)

		apiErrorResponse := NewHTTPAPIErrorResponse(statusCodeAndMessage, *appErr)
		ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
		return
	}

	var appError AppError

	// Handle specific PostgreSQL error types using a switch statement.
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		appError = NewAppError(DBConnectionException.Message, DBConnectionException.Code, err)
		apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
		ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

	case errors.Is(err, pgx.ErrNoRows):
		appError = NewAppError(DBNoData.Message, DBNoData.Code, err)
		apiErrorResponse := NewHTTPAPIErrorResponse(DBErrorRecordNotFound, appError)
		ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

	default:
		// Check if the error is a PostgreSQL error.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Map PostgreSQL error codes to custom dbError codes and messages.
			switch {

			case pgErr.Code == "42P01": // SQLSTATE for "relation does not exist"
				appError = NewAppError(DBSyntaxErrororAccessRuleViolation.Message, DBSyntaxErrororAccessRuleViolation.Code, err)
				apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsCardinalityViolation(pgErr.Code):
				appError = NewAppError(DBCardinalityViolation.Message, DBCardinalityViolation.Code, err)
				apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsWarning(pgErr.Code):
				appError = NewAppError(DBWarning.Message, DBWarning.Code, err)
				apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsNoData(pgErr.Code):
				appError = NewAppError(DBNoData.Message, DBNoData.Code, err)
				apiErrorResponse := NewHTTPAPIErrorResponse(DBErrorRecordNotFound, appError)
				ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsIntegrityConstraintViolation(pgErr.Code):
				appError = NewAppError(DBIntegrityConstraintViolation.Message, DBIntegrityConstraintViolation.Code, err)
				apiErrorResponse := NewHTTPAPIErrorResponse(DBErrorDuplicateRecord, appError)
				ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsSQLStatementNotYetComplete(pgErr.Code):
				appError = NewAppError(DBSQLStatementNotYetComplete.Message, DBSQLStatementNotYetComplete.Code, err)
				apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsConnectionException(pgErr.Code):
				appError = NewAppError(DBConnectionException.Message, DBConnectionException.Code, err)
				apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServiceUnavailable, appError)
				ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsDataException(pgErr.Code):
				appError = NewAppError(DBDataException.Message, DBDataException.Code, err)
				apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorBadRequest, appError)
				ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsTransactionRollback(pgErr.Code):
				appError = NewAppError(DBTransactionRollback.Message, DBTransactionRollback.Code, err)
				apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsSyntaxErrororAccessRuleViolation(pgErr.Code):
				appError = NewAppError(DBSyntaxErrororAccessRuleViolation.Message, DBSyntaxErrororAccessRuleViolation.Code, err)
				apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsInsufficientResources(pgErr.Code):
				appError = NewAppError(DBInsufficientResources.Message, DBInsufficientResources.Code, err)
				apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			// Catch any other PostgreSQL-related errors with a generic message.
			default:
				appError = NewAppError(DBGenericError.Message, DBGenericError.Code, err)
				apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
			}
		} else {
			// Handle non-database-related errors or unknown errors.
			appError = NewAppError(err.Error(), strconv.Itoa(http.StatusInternalServerError), err)
			apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
			ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
		}
	}
}

// HandleError handles errors by creating an application error and an API error response,
// then sends a JSON response with the appropriate status code and error details.
//
// Parameters:
//   - ctx: The Gin context to send the JSON response.
//   - err: The error to be handled and converted into an application error and API error response.
//
// Returns:
//
//	HTTP 500 Internal Server Error
//
// If the provided error is nil, the function returns immediately without doing anything.
func HandleError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

	// Initialize a variable for the appError to store the error details.
	var appErr *AppError

	// Check if the error is of type AppError.
	if errors.As(err, &appErr) {
		// Create a structured HTTP response using the AppError.
		apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, *appErr)
		ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
		return
	}

	// Handle other types of errors generically.
	// Here you can log the error if needed.
	appError := NewAppError(err.Error(), strconv.Itoa(http.StatusInternalServerError), err)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleErrorWithCustomMessage handles an error by creating a custom application error
// with a provided message and the original error. It then constructs an HTTP API error
// response and sends it as a JSON response with the appropriate status code.
//
// Parameters:
//   - ctx: The Gin context to send the JSON response.
//   - message: A custom message to include in the application error.
//   - err: The original error to be handled.
//
// Returns:
//
//	HTTP 500 Internal Server Error
//
// If the provided error is nil, the function returns immediately without doing anything.
func HandleErrorWithCustomMessage(ctx *gin.Context, message string, err error) {
	if err == nil {
		return
	}

	// Initialize a variable for the appError to store the error details.
	// var appErr *AppError
	if appErr, ok := Find[*AppError](err); ok {
		// }
		// Check if the error is of type AppError.
		// if errors.As(err, &appErr) {
		// Create a structured HTTP response using the AppError.
		apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, *appErr)
		ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
		return
	}

	appError := NewAppError(message, strconv.Itoa(http.StatusInternalServerError), err)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleWithMessage handles an error by creating an application error with a given message,
// then constructs an HTTP API error response and sends it as a JSON response with the appropriate status code.
//
// Parameters:
//   - ctx: The Gin context for the current request.
//   - message: The error message to be included in the application error.
//
// Returns:
//
//	HTTP 500 Internal Server Error
//
// The function creates an application error with the provided message and an internal server error status code.
// It then creates an HTTP API error response using this application error and sends it as a JSON response
// with the status code from the API error response.
func HandleWithMessage(ctx *gin.Context, message string) {
	appError := NewAppError(message, strconv.Itoa(http.StatusInternalServerError), nil)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleMarshalError handles errors that occur during marshaling.
// If the provided error is not nil, it creates an application error and an API error response,
// then sends a JSON response with the appropriate status code and error details.
//
// Parameters:
//   - ctx: The Gin context to send the JSON response.
//   - err: The error that occurred during marshaling.
//
// Returns:
//
//	HTTP 400 Bad Request
func HandleMarshalError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

	appError := NewAppError(err.Error(), strconv.Itoa(http.StatusBadRequest), err)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorBadRequest, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// ValidateContentType is a middleware function for the Gin framework that checks if the request's
// "Accept" header matches any of the allowed content types. If the content type is not allowed,
// it returns a structured error response and aborts further request handling.
//
// Parameters:
// - allowedTypes ([]string): A slice of strings representing the allowed content types.
//
// Returns:
//   - gin.HandlerFunc: A Gin handler function that performs the content type validation.
//   - HTTP 415 Unsupported Media Type
func ValidateContentType(allowedTypes []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		contentType := ctx.GetHeader("Accept")

		// Check if the contentType is in the allowedTypes.
		validContentType := false
		for _, allowedType := range allowedTypes {
			if contentType == allowedType {
				validContentType = true
				break
			}
		}

		// If not valid, return a structured error response.
		if !validContentType {
			appError := NewAppError(fmt.Sprintf("Supported types are: %v", allowedTypes), strconv.Itoa(http.StatusUnsupportedMediaType), nil)
			apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorInvalidContentType, appError)
			ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
			ctx.Abort() // Prevent further handling of the request.
			return
		}

		ctx.Next() // Proceed to the next handler if content type is valid.
	}
}

// HandleSizeError handles errors related to payload size exceeding the allowed limit.
// It creates a new application error with a "Payload too large" message and a "413" status code.
// The function then constructs an HTTP API error response and sends it as a JSON response.
//
// Parameters:
// - ctx: The Gin context for the current request.
//
// Returns:
//
//	HTTP 413 File Too Large
func HandleSizeError(ctx *gin.Context) {
	appError := NewAppError("Payload too large.", "413", nil)
	apiErrorResponse := NewHTTPAPIErrorResponse(FileErrorTooLarge, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleRateLimitingError handles rate limiting errors by creating an application error
// with a "Too many requests" message and a 429 status code. It then constructs an HTTP
// API error response and sends it as a JSON response with the appropriate status code.
//
// Parameters:
// - ctx: The Gin context for the current request.
// Returns:
//
//	HTTP 429 Too Many Requests
func HandleRateLimitingError(ctx *gin.Context) {
	appError := NewAppError("Too many requests. Please try again later.", "429", nil)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorTooManyRequests, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleDuplicateEntryError handles errors related to duplicate entries in the application.
// It creates a new application error with a message indicating that the resource already exists,
// sets the HTTP status code to 409 (Conflict), and sends a JSON response with the error details.
//
// Parameters:
//   - ctx: The Gin context for the current request.
//
// Returns:
//
//	HTTP 409 Conflict
func HandleDuplicateEntryError(ctx *gin.Context) {
	appError := NewAppError("Data conflict occurred while adding/updating. Resource already exists.", "409", nil)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorConflict, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleConnectionError handles connection errors by creating an application error
// and sending an appropriate HTTP API error response.
//
// Parameters:
//   - ctx: The Gin context for the current request.
//   - err: The error that occurred during the connection attempt.
//
// Returns:
//
//	HTTP 503 Service Unavailable
//
// The function creates a new application error with a message indicating that the
// service is unavailable and an HTTP status code of 503. It then creates an HTTP
// API error response using this application error and sends it as a JSON response
// with the appropriate status code.
func HandleConnectionError(ctx *gin.Context, err error) {
	appError := NewAppError("Service unavailable. Please try again later.", "503", err)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServiceUnavailable, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleFileTypeError handles errors related to unsupported file types.
// It creates an application error with a specific message and status code,
// then constructs an HTTP API error response and sends it as a JSON response.
//
// Parameters:
//   - ctx: The Gin context for the current request.
//
// Returns:
//
//	HTTP 415 Invalid Content Type
func HandleFileTypeError(ctx *gin.Context) {
	appError := NewAppError("Unsupported file type.", "415", nil)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorInvalidContentType, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleUnauthorizedError handles unauthorized access errors by creating an
// application error with a 401 status code and sending an appropriate JSON
// response to the client.
//
// Parameters:
//   - ctx: The Gin context for the current request.
//
// Returns:
//
//	HTTP 401 Unauthorized
//
// This function is typically used as a middleware or error handler to ensure
// that unauthorized access attempts are properly reported to the client.
func HandleUnauthorizedError(ctx *gin.Context) {
	appError := NewAppError("Unauthorized access. Authentication is required.", "401", nil)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorUnauthorized, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleUnauthorizedErrorWithDetail handles unauthorized errors by creating an
// application-specific error and sending an HTTP response with the appropriate
// status code and error details.
//
// Parameters:
//   - ctx: The Gin context for the current request.
//   - detail: The error detail to be included in the application error.
//
// Returns:
//
//	HTTP 401 Unauthorized
//
// This function constructs an application error using the provided detail,
// creates an HTTP API error response with a 401 Unauthorized status code, and
// sends the response as JSON.
func HandleUnauthorizedErrorWithDetail(ctx *gin.Context, err error) {
	appError := NewAppError(err.Error(), "401", err)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorUnauthorized, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleForbiddenError handles forbidden access errors by creating an
// application-specific error and sending an HTTP 403 Forbidden response
// with a JSON payload containing the error details.
//
// Parameters:
//   - ctx: The Gin context for the current request.
//
// Returns:
//   - HTTP 403 Forbidden
func HandleForbiddenError(ctx *gin.Context) {
	appError := NewAppError("Access to this resource is forbidden. Insufficient permissions.", "403", nil)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorForbidden, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleRequestTimeoutError handles request timeout errors by creating an application error
// with a "Request timed out." message and a "408" status code. It then creates an
// HTTP API error response with the appropriate status code and sends it as a JSON
// response.
//
// Parameters:
//   - ctx: The Gin context for the current request.
//
// Returns:
//   - HTTP 408 Request Timeout
func HandleRequestTimeoutError(ctx *gin.Context) {
	appError := NewAppError("Request timed out.", "408", nil)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorRequestTimeout, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleServiceUnavailableError handles server timeout errors by creating an
// application error with a 503 status code and sending an appropriate
// HTTP API error response.
//
// Parameters:
//   - ctx: The Gin context for the current request.
//
// Returns:
//   - HTTP 503 Service Unavailable
func HandleServiceUnavailableError(ctx *gin.Context) {
	appError := NewAppError("Server took too long to respond.", "503", nil)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServiceUnavailable, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleGatewayTimeoutError handles the Gateway Timeout error (HTTP 504) by creating an
// appropriate application error and sending a JSON response with the error details.
//
// Parameters:
//   - ctx: The Gin context for the current request.
//
// Returns:
//   - HTTP 504 Gateway Timeout
func HandleGatewayTimeoutError(ctx *gin.Context) {
	appError := NewAppError("Server/Gateway timeout occurred.", "504", nil)
	apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorGatewayTimeout, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleBulkErrors processes a slice of AppError and sends a JSON response with the appropriate HTTP status code.
//
// Parameters:
//   - ctx: The Gin context to send the JSON response.
//   - err: A slice of AppError containing the errors to be handled.
//
// Returns:
//   - HTTP 400 Bad Request: If the bulk errors pertain to client-side issues (default behavior).
//   - The status code may vary if different error mapping logic is used in the implementation.
func HandleBulkErrors(ctx *gin.Context, err []AppError) {
	apiErrorResponse := NewHTTPAPIBulkErrorResponse(HTTPErrorBadRequest, err)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// HandleErrorWithStatusCodeAndMessage handles an error by creating an AppError and an HTTPAPIErrorResponse,
// then sends a JSON response with the appropriate status code and error message.
//
// Parameters:
//   - statusCodeAndMessage: A struct containing the status code and message to be used in the response.
//   - message: A custom message to be included in the AppError.
//   - err: The original error that occurred.
//
// Returns:
//   - AppError.
func HandleErrorWithStatusCodeAndMessage(statusCodeAndMessage statusCodeAndMessage, message string, err error) *AppError {
	appError := NewAppError(message, strconv.Itoa(statusCodeAndMessage.StatusCode), err)
	return &appError
}

func checkDBError(err error) APIErrorResponse {

	var appError AppError
	var apiErrorResponse APIErrorResponse

	// Handle specific PostgreSQL error types using a switch statement.
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		appError = NewAppError(DBConnectionException.Message, DBConnectionException.Code, err)
		apiErrorResponse = NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
		// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

	case errors.Is(err, pgx.ErrNoRows):
		appError = NewAppError(DBNoData.Message, DBNoData.Code, err)
		apiErrorResponse = NewHTTPAPIErrorResponse(DBErrorRecordNotFound, appError)
		// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

	default:
		// Check if the error is a PostgreSQL error.
		// var pgErr *pgconn.PgError
		if pgErr, ok := Find[*pgconn.PgError](err); ok {
			// }
			// if errors.As(err, &pgErr) {
			// Map PostgreSQL error codes to custom dbError codes and messages.
			switch {

			case pgErr.Code == "42P01": // SQLSTATE for "relation does not exist"
				appError = NewAppError(DBSyntaxErrororAccessRuleViolation.Message, DBSyntaxErrororAccessRuleViolation.Code, err)
				apiErrorResponse = NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsCardinalityViolation(pgErr.Code):
				appError = NewAppError(DBCardinalityViolation.Message, DBCardinalityViolation.Code, err)
				apiErrorResponse = NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsWarning(pgErr.Code):
				appError = NewAppError(DBWarning.Message, DBWarning.Code, err)
				apiErrorResponse = NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsNoData(pgErr.Code):
				appError = NewAppError(DBNoData.Message, DBNoData.Code, err)
				apiErrorResponse = NewHTTPAPIErrorResponse(DBErrorRecordNotFound, appError)
				// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsIntegrityConstraintViolation(pgErr.Code):
				appError = NewAppError(DBIntegrityConstraintViolation.Message, DBIntegrityConstraintViolation.Code, err)
				apiErrorResponse = NewHTTPAPIErrorResponse(DBErrorDuplicateRecord, appError)
				// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsSQLStatementNotYetComplete(pgErr.Code):
				appError = NewAppError(DBSQLStatementNotYetComplete.Message, DBSQLStatementNotYetComplete.Code, err)
				apiErrorResponse = NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsConnectionException(pgErr.Code):
				appError = NewAppError(DBConnectionException.Message, DBConnectionException.Code, err)
				apiErrorResponse = NewHTTPAPIErrorResponse(HTTPErrorServiceUnavailable, appError)
				// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsDataException(pgErr.Code):
				appError = NewAppError(DBDataException.Message, DBDataException.Code, err)
				apiErrorResponse = NewHTTPAPIErrorResponse(HTTPErrorBadRequest, appError)
				// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsTransactionRollback(pgErr.Code):
				appError = NewAppError(DBTransactionRollback.Message, DBTransactionRollback.Code, err)
				apiErrorResponse = NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsSyntaxErrororAccessRuleViolation(pgErr.Code):
				appError = NewAppError(DBSyntaxErrororAccessRuleViolation.Message, DBSyntaxErrororAccessRuleViolation.Code, err)
				apiErrorResponse = NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			case pgerrcode.IsInsufficientResources(pgErr.Code):
				appError = NewAppError(DBInsufficientResources.Message, DBInsufficientResources.Code, err)
				apiErrorResponse = NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)

			// Catch any other PostgreSQL-related errors with a generic message.
			default:
				appError = NewAppError(DBGenericError.Message, DBGenericError.Code, err)
				apiErrorResponse = NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
				// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
			}
		} else {
			// Handle non-database-related errors or unknown errors.
			appError = NewAppError(HTTPErrorServerError.Message, strconv.Itoa(http.StatusInternalServerError), err)
			apiErrorResponse = NewHTTPAPIErrorResponse(HTTPErrorServerError, appError)
			// ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
		}
	}

	return apiErrorResponse

}

func HandleCommonError(ctx *gin.Context, err error) {
	if err == nil {
		return
	}

	// Check if the error is of type AppError.
	// var appErr *AppError
	if appErr, ok := Find[*AppError](err); ok {
		// }
		// if errors.As(err, &appErr) {

		if len(appErr.FieldErrors) > 0 {
			HandleValidationError(ctx, err)
			return
		}

		statusCode, convErr := strconv.Atoi(appErr.Code)
		if convErr != nil {
			apiErrorResponse := NewHTTPAPIErrorResponse(HTTPErrorServerError, *appErr)
			ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
			return
		}

		statusCodeAndMessage := mapErrorToHTTP(statusCode)

		apiErrorResponse := NewHTTPAPIErrorResponse(statusCodeAndMessage, *appErr)
		ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
		return
	}

	apiErrorResponse := checkDBError(err)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}

// ErrorResponseWithStatusCodeAndMessage handles an error by creating an AppError and an HTTPAPIErrorResponse,
// then sends a JSON response with the appropriate status code and error message.
//
// Parameters:
//   - ctx: The Gin context to send the JSON response.
//   - statusCodeAndMessage: A struct containing the status code and message to be used in the response.
//   - message: A custom message to be included in the AppError.
//   - err: The original error that occurred.
//
// Returns:
//   - HTTP <statusCode>: The HTTP status code defined in the statusCodeAndMessage struct.
func ErrorResponseWithStatusCodeAndMessage(ctx *gin.Context, statusCodeAndMessage statusCodeAndMessage, message string, err error) {
	appError := NewAppError(message, strconv.Itoa(statusCodeAndMessage.StatusCode), err)
	apiErrorResponse := NewHTTPAPIErrorResponse(statusCodeAndMessage, appError)
	ctx.JSON(apiErrorResponse.StatusCode, apiErrorResponse)
}
