package apierrors

import "encoding/json"

// APIErrorResponse represents the structure of an API error response.
// It includes the HTTP status code and message, as well as an application-specific error.
type APIErrorResponse struct {
	statusCodeAndMessage `json:",inline"`
	AppError             AppError `json:"error"`
}

// Struct to handle multiple API errors
type APIBulkErrorResponse struct {
	statusCodeAndMessage `json:",inline"`
	Errors               []AppError `json:"errors"`
}

func (r *APIErrorResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *APIErrorResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, r)
}

// NewAPIErrorResponse creates a new instance of APIErrorResponse with the provided parameters.
//
// Parameters:
//   - statusCode: An integer representing the HTTP status code (e.g., 400 for Bad Request).
//   - message: A string containing a user-friendly error message describing the API error.
//   - err: An instance of AppError representing the application-specific error, which may contain additional context about the error.
//
// Returns:
//   An instance of APIErrorResponse containing the provided status code, message, and application error.
//   This can be used to structure the response returned to the client in case of an API error.
func NewAPIErrorResponse(statusCode int, message string, err AppError) APIErrorResponse {
	return APIErrorResponse{
		statusCodeAndMessage: statusCodeAndMessage{StatusCode: statusCode, Message: message},
		AppError:             err,
	}
}

// NewHTTPAPIErrorResponse creates a new APIErrorResponse instance.
// It takes a StatusCodeAndMessage and an AppError as parameters and returns an APIErrorResponse.
//
// Parameters:
//   - httpError: A StatusCodeAndMessage representing the HTTP status code and message.
//   - err: An AppError representing the application-specific error.
//
// Returns:
//   An APIErrorResponse containing the provided HTTP status code and message, and the application-specific error.
func NewHTTPAPIErrorResponse(httpError statusCodeAndMessage, err AppError) APIErrorResponse {
	return APIErrorResponse{
		statusCodeAndMessage: httpError,
		AppError:             err,
	}
}

func NewHTTPAPIBulkErrorResponse(httpError statusCodeAndMessage, errs []AppError) APIBulkErrorResponse {
	return APIBulkErrorResponse{
		statusCodeAndMessage: httpError,
		Errors:               errs,
	}
}