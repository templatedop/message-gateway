package apierrors

func mapErrorToHTTP(statusCode int) statusCodeAndMessage {
	switch statusCode {

	case 400:
		return HTTPErrorBadRequest
	case 401:
		return HTTPErrorUnauthorized
	case 403:
		return HTTPErrorForbidden
	case 404:
		return HTTPErrorNotFound
	case 409:
		return HTTPErrorConflict
	case 422:
		return AppErrorValidationError
	case 429:
		return HTTPErrorTooManyRequests
	case 500:
		return HTTPErrorServerError
	case 501:
		return HTTPErrorNotImplemented
	case 503:
		return HTTPErrorServiceUnavailable
	case 504:
		return HTTPErrorGatewayTimeout
	default:
		return HTTPErrorServerError
	}
}
