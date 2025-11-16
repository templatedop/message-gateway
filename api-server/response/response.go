package response

import "io"

// import validation "MgApplication/api-validation"

type Stature interface {
	Status() int
	GetContentType() string
	GetContentDisposition() string
	ResponseType() string
	Object() []byte
}
type Streamer interface {
	Stream(w io.Writer) error
}

type Response[T any] struct {
	Success       bool   `json:"success"`
	Message       string `json:"message,omitempty"`
	Data          T      `json:"data,omitempty"`
	Page          *int   `json:"page,omitempty"`
	Size          *int   `json:"size,omitempty"`
	TotalElements *int   `json:"totalElements,omitempty"`
	TotalPages    *int   `json:"totalPages,omitempty"`

	// Error fields

	// ValidationErrors []validation.FieldError `json:"validationErrors,omitempty"`
	Errors []Errors `json:"errors,omitempty"`

	// Internal fields for response handling
	status int // HTTP status code
}

// Status implements Stature interface
func (r Response[T]) Status() int {
	if r.status == 0 {
		return 200 // Default to OK
	}
	return r.status
}

// GetContentType implements Stature interface
func (r Response[T]) GetContentType() string {
	return "application/json"
}

// GetContentDisposition implements Stature interface
func (r Response[T]) GetContentDisposition() string {
	return ""
}

// ResponseType implements Stature interface
func (r Response[T]) ResponseType() string {
	return "json"
}

// Object implements Stature interface
func (r Response[T]) Object() []byte {
	return nil // Not used for JSON responses
}

// WithStatus sets the HTTP status code for the response
func (r Response[T]) WithStatus(status int) Response[T] {
	r.status = status
	return r
}

type Errors struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func Success(data any) Response[any] {
	resp := Response[any]{
		Success: true,
		Message: "success",
		Data:    data,
	}

	return resp
}

// func Error(msg string, Errors []Errors, fieldErrs ...validation.FieldError) Response[any] {
// 	return Response[any]{
// 		Success:          false,
// 		Message:          msg,
// 		ValidationErrors: fieldErrs,
// 		Errors:           Errors,
// 	}
// }

// type ResponseError struct {
// 	Success          bool                    `json:"success"`
// 	Message          string                  `json:"message,omitempty"`
// 	ValidationErrors []validation.FieldError `json:"validationErrors,omitempty"`
// 	Errors           []Errors                `json:"errors,omitempty"`
// }
