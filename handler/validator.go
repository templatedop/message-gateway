package handler

import (
	"regexp"


	"github.com/go-playground/validator/v10"
)

func ServiceRequestType(f1 validator.FieldLevel) bool {
	// Define the regex pattern for a comma-separated list of numbers between 1 and 4
	requestTypePattern := "^[1-4](,[1-4])*$"
	re := regexp.MustCompile(requestTypePattern)
	return re.MatchString(f1.Field().String())
}

func NewValidatorService() error {

// 	err := validation.Create()
// 	if err != nil {
// 		return err
// 	}
// 	// add the custom validator here
// 	// err = validation.RegisterCustomValidation("validateEmail", ValidateEmail, "Incorrect email Format")
// 	// if err != nil {
// 	// 	return err
// 	// }
// 
// 	// err = validation.RegisterCustomValidation("hourvalidate", HourValidate, "field %s must be in the format 'HH:MM:SS', with a valid hour (00-23), minute (00-59), and second (00-59), but received %v")
// 	// if err != nil {
// 	// 	return err
// 	// }
// 
// 	err = validation.RegisterCustomValidation("request_type", ServiceRequestType, "invalid values for %s, must be 1-4 but received %v")
// 	if err != nil {
// 		return err
// 	}
// 
// 	err = validation.RegisterCustomValidation("services", ServiceRequestType, "invalid values for %s, must be 1-4 but received %v")
	// 	if err != nil {
	// 		return err
	// 	}

	return nil
}

// errorValidResponse represents a JSON response for validation errors.
// type errorValidResponse struct {
// 	Success bool     `json:"success" example:"false"`
// 	Message []string `json:"message" example:"Error message"`
// 	Errorno []string `json:"errorno"`
// }

// // newErrorValidResponse creates a new error response body.
// func newErrorValidResponse(message []string, errorno []string) errorValidResponse {
// 	return errorValidResponse{
// 		Success: false,
// 		Message: message,
// 		Errorno: errorno,
// 	}
// }
