package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	appError "MgApplication/api-errors"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	ut "github.com/templatedop/universal-translator-master"
)

var (
	validate                 *validator.Validate
	uni                      *ut.UniversalTranslator
	trans                    ut.Translator
	once                     sync.Once
	customValidationMessages = map[string]func(string, any) string{}
	message                  = "validation error"
	validatorErrorMessage    = "validator not initialized"
	translatorErrorMessage   = "translator not initialized"
	unprocessibleEntityCode  = "422"
	serverErrorCode          = "500"
)

var structFieldTags = []string{"json", "param", "form"}

func getStructFieldName(fld reflect.StructField) string {
	for _, st := range structFieldTags {
		name := strings.SplitN(fld.Tag.Get(st), ",", 2)[0]
		if name == "" {
			continue
		}
		if name == "-" {
			return ""
		}
		return name
	}
	return fld.Name
}

func getDefaultRules() []validationRule {
	return []validationRule{
		newStringFieldValidator(),
		newValidateBeatNamePatternValidator(),
		newValidateHOAPatternValidator(),
		newPersonnelNameValidator(),
		newAddressPatternValidator(),
		newEmailValidator(),
		newGValidatePhoneLengthPatternValidator(),
		newGValidateSOBONamePatternValidator(),
		newGValidatePANNumberPatternValidator(),
		newGValidateVehicleRegistrationNumberPatternValidator(),
		newGValidateBarCodeNumberPatternValidator(),
		newCustomValidateGLCodePatternValidator(),
		newTimeStampValidatePatternValidator(),
		newCustomValidateAnyStringLengthto50PatternValidator(),
		newDateyyyymmddPatternValidator(),
		newDateddmmyyyyPatternValidator(),
		newValidateEmployeeIDPatternValidator(),
		newValidateValidateGSTINPatternValidator(),
		newValidateBankUserIDPatternValidator(),
		newValidateOrderNumberPatternValidator(),
		newValidateAWBNumberPatternValidator(),
		newValidatePNRNoPatternValidator(),
		newValidatePLIIDPatternValidator(),
		newValidatePaymentTransIDPatternValidator(),
		newValidateOfficeCustomerIDPatternValidator(),
		newValidateBankIDPatternValidator(),
		newValidateCSIFacilityIDPatternValidator(),
		newValidatePosBookingOrderNumberPatternValidator(),
		newValidateSOLIDPatternValidator(),
		newValidatePLIOfficeIDPatternValidator(),
		newValidateProductCodePatternValidator(),
		newValidateCustomerIDPatternValidator(),
		newValidateFacilityIDPatternValidator(),
		newValidateApplicationIDPatternValidator(),
		newValidateReceiverKYCReferencePatternValidator(),
		newValidateOfficeCustomerPatternValidator(),
		newValidatePRANPatternValidator(),
		newvalidateCustomFlightNoValidator(),
		newvalidatePinCodeGlobalValidator(),

		newValidateMobileNumberPatternValidator(),
		newBatchNumberPatternValidator(),
		newPhoneNumberValidator(),
		newTimePatternValidator(),
		newCustomofficeidGlobalValidator(),
		newvalidateBagIdPatternValidator(),
		newCustomTrainNoGlobalValidator(),
		newCustomSCSGlobalValidator(),
		newvalidateCircleIDGlobalValidator(),
		newvalidateTariffIDGlobalValidator(),
		newvalidateCIFNumGlobalValidator(),
		newvalidateContractNumGlobalValidator(),
		newvalidateRegionIDGlobalValidator(),
		newvalidateVasIDGlobalValidator(),
		newvalidateUserCodeGlobalValidator(),
		newvalidateHONamePatternValidator(),
		newvalidateHOIDGlobalValidator(),
		newvalidateAccountNoGlobalValidator(),
		newIsValidTimestampGlobalValidator(),
		newIsValidStateValidator(),
		newvalidateCityNameValidator(),
		newvalidateAadharValidator(),
		newvalidateDrivingLicenseNoValidator(),
		newvalidatePassportNoValidator(),
		newvalidateVoterIDValidator(),
		newvalidateOfficeNameValidator(),
		newvalidateMonthValidator(),
		newvalidateYearValidator(),
		newOptionalFieldValidator(),
		newDateyyyymmddPatternValidatorWithddmmyyyMessage(),
	}
}

func registerDefaultRules(rules []validationRule, val *validator.Validate) error {
	for _, r := range rules {
		if err := val.RegisterValidation(r.Name(), r.Apply); err != nil {
			return err
		}
		customValidationMessages[r.Name()] = r.Message
	}
	return nil
}

// ValidateStruct validates the fields of a given struct based on predefined rules.
// It returns an error if the validation fails or if the validator or translator is not initialized.
//
// Parameters:
//   - s: The struct to be validated.
//
// Returns:
//   - error: An error object containing validation errors if any field fails validation,
//     or an error indicating issues with the validator or translator initialization.
//
// The function performs the following steps:
//  1. Checks if the validator is initialized. If not, returns an error indicating the validator is not set up.
//  2. Checks if the translator is initialized. If not, returns an error indicating the translator is not set up.
//  3. Validates the struct fields. If validation errors are found, it constructs a detailed error object
//     containing field-specific error messages and returns it.
//
// Usage:
//
// The ValidateStruct function is used to validate the fields of a struct based on predefined rules.
func ValidateStruct(s interface{}) error {
	var appErr appError.AppError
	// check if the validator is initialized
	if validate == nil {

		appErr := appError.NewAppError(validatorErrorMessage, serverErrorCode, errors.New(validatorErrorMessage))
		return &appErr
	}

	if trans == nil {
		appErr := appError.NewAppError(translatorErrorMessage, serverErrorCode, errors.New(translatorErrorMessage))
		return &appErr
	}
	err := validate.Struct(s)
	if err != nil {
		var validatorErrors validator.ValidationErrors
		appErr = appError.NewAppError(message, unprocessibleEntityCode, err)
		errors.As(err, &validatorErrors)
		var apiFieldErrors []appError.FieldError
		for _, e := range validatorErrors {
			tag := e.Tag()
			if Emsg, ok := customValidationMessages[tag]; ok {
				apiFieldErrors = append(apiFieldErrors, appErr.NewFieldError(e.Field(), e.Value(), Emsg(e.Field(), e.Value()), tag))
			} else {
				apiFieldErrors = append(apiFieldErrors, appErr.NewFieldError(e.Field(), e.Value(), e.Translate(trans), tag))
			}
		}
		appErr.SetFieldErrors(apiFieldErrors)
		return &appErr
	}
	return nil
}

// RegisterCustomValidation registers a custom validation function with a specific tag and error message.
//
// Parameters:
//   - tag: A string representing the validation tag. This tag is used to identify the custom validation rule.
//   - fn: A validator.Func which is the custom validation function to be registered. This function should contain the logic for the custom validation.
//   - message: A string representing the error message to be returned when the validation fails.
//
// Returns:
//   - error: An error if the registration fails. Possible reasons for failure include an empty tag, a nil validation function, or if the tag is already registered.
//
// Usage:
//
//	This function is used to add custom validation rules to the validator. It ensures that the tag and function are valid and not already registered before adding the rule.
func RegisterCustomValidation(tag string, fn validator.Func, message string) error {
	if tag == "" {
		return errors.New("validation tag cannot be empty")
	}
	if fn == nil {
		return errors.New("validation function cannot be nil")
	}

	if _, exists := customValidationMessages[tag]; exists {
		return fmt.Errorf("validation tag '%s' is already registered", tag)
	}

	rule := newRule(tag, fn, message)
	err := registerDefaultRules([]validationRule{rule}, validate)
	if err != nil {
		return err
	}

	return nil
}

// Create initializes the validator with default rules and translations.
//
// It uses a sync.Once to ensure that the initialization is performed only once.
// The function sets up the validator, the universal translator, and registers
// default translations and rules. If any error occurs during the initialization
// process, it returns the error. If the validator is not properly initialized,
// it returns a specific error indicating the failure.
//
// Returns:
//   - error: An error if the initialization fails. Possible reasons for failure include issues with setting up the validator, translator, or registering default rules and translations.
//
// Usage:
//
//	This function is used to initialize the validator with default settings. It ensures that the initialization is performed only once and handles any errors that may occur during the process.
func Create() error {
	var initErr error
	once.Do(func() {
		rules := getDefaultRules()
		validate = validator.New()
		eng := en.New()
		uni = ut.New(eng, eng)
		trans, _ = uni.GetTranslator("en")

		// Register default translations for the validator
		if err := en_translations.RegisterDefaultTranslations(validate, trans); err != nil {
			initErr = err
			return
		}
		validate.RegisterTagNameFunc(getStructFieldName)
		err := registerDefaultRules(rules, validate)
		if err != nil {
			initErr = err
			return
		}
	})
	if initErr != nil {
		return initErr
	}
	//check if the validator is initialized
	if validate == nil {
		return errors.New(validatorErrorMessage)
	}
	return nil
}

// GenericStringValidation validates a string based on the provided parameters.
//
// Parameters:
//   - tagSuffix: A string suffix used to create a unique validation tag. Must not be empty.
//   - minLength: The minimum length of the string to be validated.
//   - maxLength: The maximum length of the string to be validated.
//   - char: Optional variadic parameter specifying additional allowed characters in the string.
//
// The function already allows alphanumeric characters, spaces, commas, periods, parentheses, and hyphens by default.
//
// The final validation tag will be in the format "string_<tagSuffix>".
func GenericStringValidation(tagSuffix string, minLength, maxLength uint, char ...rune) error {

	err := validateArgs(tagSuffix, minLength, maxLength, char...)
	if err != nil {
		return err
	}

	pattern, err := generateDynamicStringValidationPattern(minLength, maxLength, char...)
	if err != nil {
		fmt.Println("Error compiling regex:", err)
		return err
	}

	tag := fmt.Sprintf("string_%s", tagSuffix)

	if _, exists := customValidationMessages[tag]; exists {
		return fmt.Errorf("validation tag '%s' is already registered", tag)
	}
	var msg string

	if len(char) > 0 {
		charStr := strings.Join(strings.Split(string(char), ""), ", ")
		charStr = strings.ReplaceAll(charStr, "%", "%%") // escape percent signs
		msg = fmt.Sprintf(
			"field %%s must contain alphanumeric characters, and may contain letters, digits, spaces, commas, periods, parentheses, hyphens and %s. The total length should be between %d and %d characters, but received %%v",
			charStr, minLength, maxLength,
		)
	} else {
		msg = fmt.Sprintf(
			"field %%s must contain alphanumeric characters, and may contain letters, digits, spaces, commas, periods, parentheses, hyphens. The total length should be between %d and %d characters, but received %%v",
			minLength, maxLength,
		)
	}

	rule := newRule(tag, func(fl validator.FieldLevel) bool {
		field := fl.Field()
		if field.Kind() == reflect.String {
			return validateWithGlobalRegex(fl, pattern)
		}
		return false
	}, msg)
	err = registerDefaultRules([]validationRule{rule}, validate)
	if err != nil {
		return err
	}

	return nil

}
