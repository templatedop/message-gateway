package validation

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	// Personal Identification Patterns
	panNumberPattern = regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]$`)
	//employeeIDPattern       = regexp.MustCompile(`^\d{8}$`) not required as string validation of employee id is removed
	pranPattern             = regexp.MustCompile(`^\d{12}$`)
	aadharPattern           = regexp.MustCompile(`^\d{12}$`)
	drivingLicenseNoPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]{9,19}$`)
	passportNoPattern       = regexp.MustCompile(`^[A-Za-z][0-9]{7}$`) //passport no is 8 digit  G1234567
	voterIDPattern          = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]{8}[0-9]$`)

	// Customer Identification Patterns
	customerIDPattern           = regexp.MustCompile(`^\d{10}$`)
	officeCustomerIDPattern     = regexp.MustCompile(`^[a-zA-Z0-9\-]{1,50}$`)
	officeCustomerPattern       = regexp.MustCompile(`^[a-zA-Z0-9.\s]+$`)
	receiverKYCReferencePattern = regexp.MustCompile(`^KYCREF[A-Z0-9]{0,44}$`)
	gstINPattern                = regexp.MustCompile(`^[0-9]{2}[A-Za-z]{5}[0-9]{4}[A-Za-z]{1}[A-Za-z0-9]{1}[Zz]{1}[A-Za-z0-9]{1}$`)

	// Bank Identification Patterns
	bankUserIDPattern = regexp.MustCompile(`^[A-Z0-9]{1,50}$`)
	bankIDPattern     = regexp.MustCompile(`^[A-Z0-9]{1,50}$`)

	// Facility Identification Patterns
	csiFacilityIDPattern = regexp.MustCompile(`^[A-Z]{2}\d{11}$`)
	facilityIDPattern    = regexp.MustCompile(`^[A-Z]{2}\d{11}$`)
	batchNamePattern     = regexp.MustCompile(`^BATCH_\d{1,2}$`)
	beatNamePattern      = regexp.MustCompile(`^BEAT_\d{1,2}$`)

	// Application Identification Patterns
	applicationIDPattern = regexp.MustCompile(`^[A-Z]{3}\d{8}-\d{3}$`)

	// Name Patterns
	personnelNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z.\s]{1,48}[A-Za-z]$`)
	soboNamePattern      = regexp.MustCompile(`^[A-Za-z][A-Za-z.\s]{1,48}[A-Za-z]$`)
	officeNamePattern    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.,&'\s()\-_]{1,48}[A-Za-z0-9\-_'().\s]$`)

	onlyNumbers           = regexp.MustCompile(`^\d+$`)
	onlyNumbersWithSpaces = regexp.MustCompile(`^[\d\s\-.,:;_!@#$%^*()\[\]{}|\/\\<>?+=]+$`)

	// Address Patterns
	addressPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9\s,.\/#-]{1,48}[A-Za-z0-9]$`)

	//Cityname Pattern
	cityNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.\s]{1,48}[A-Za-z0-9]$`)

	// Contact Patterns
	emailPattern              = regexp.MustCompile(`^[a-zA-Z0-9._+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneLengthPattern        = regexp.MustCompile(`^\d{10}$`)
	mobileNumberStringPattern = regexp.MustCompile(`^[6-9]\d{9}$`)
	phoneNumberPattern        = regexp.MustCompile(`^\d{5}([- ]?)\d{6}$`)

	// Vehicle Patterns
	vehicleRegistrationNumberPattern = regexp.MustCompile(`^[A-Z]{2}[A-Z\d]{2}[A-Z]{0,3}\d{4,7}$|^\d{2}[A-Z]{2}\d{4}[A-Z]{1,2}$|^\d{2}BH\d{7}$`)

	// Code Patterns
	glCodePattern        = regexp.MustCompile(`^GL\d{11}$`)
	productCodePattern   = regexp.MustCompile(`^[A-Z]{3}\d{12}$`)
	barCodeNumberPattern = regexp.MustCompile(`^[A-Za-z]{2}\d{9}[A-Za-z]{2}$`)
	// Date and Time Patterns
	timeStampPattern    = regexp.MustCompile(`^(0[1-9]|[12][0-9]|3[01])-(0[1-9]|1[0-2])-(\d{4}) ([01]\d|2[0-3]):([0-5]\d):([0-5]\d)$`)
	dateyyyymmddPattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`)
	dateddmmyyyyPattern = regexp.MustCompile(`^(0[1-9]|[12][0-9]|3[01])-(0[1-9]|1[0-2])-\d{4}$`)
	timePattern         = regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d):([0-5]\d)$`)
	monthPattern        = regexp.MustCompile(`^(0[1-9]|1[0-2]|[1-9]|January|February|March|April|May|June|July|August|September|October|November|December|jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec|JAN|FEB|MAR|APR|MAY|JUN|JUL|AUG|SEP|OCT|NOV|DEC|JANUARY|FEBRUARY|MARCH|APRIL|MAY|JUNE|JULY|AUGUST|SEPTEMBER|OCTOBER|NOVEMBER|DECEMBER)$`)
	yearPattern         = regexp.MustCompile(`^\d{4}$`)

	// Order Patterns
	orderNumberPattern           = regexp.MustCompile(`^[A-Z]{2}\d{19}$`)
	bagIdPattern                 = regexp.MustCompile(`^(?:[A-Za-z]{3}\d{10}|[A-Za-z]{15}\d{14})$`)
	posBookingOrderNumberPattern = regexp.MustCompile(`^[A-Z]{2}\d{19}$`)

	// Transport Patterns
	awbNumberPattern   = regexp.MustCompile(`^[A-Z]{4}\d{9}$`)
	pnrNoPattern       = regexp.MustCompile(`^[A-Z]{3}\d{6}$`)
	pliIDPattern       = regexp.MustCompile(`^[A-Za-z]{3}\d{7,10}$`)
	pliOfficeIDPattern = regexp.MustCompile(`^[A-Z]{3}\d{10}$`)
	flightNopattern    = regexp.MustCompile("^[A-Za-z0-9 ]+$")
	trainNoPattern     = regexp.MustCompile(`^\d{5}$`)

	// Payment Patterns
	paymentTransIDPattern = regexp.MustCompile(`^\d{2}[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)

	// Miscellaneous Patterns
	hoaPattern                               = regexp.MustCompile(`^\d{15}$`)
	specialCharPattern                       = regexp.MustCompile(`[!@#$%^&*()<>:;"{}[\]\\]`)
	allZerosRegex                            = regexp.MustCompile("^0+$")
	customValidateAnyStringLengthto50Pattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]{0,48}[a-zA-Z]$`)
	solIdPattern                             = regexp.MustCompile(`^\d{6}\d{2}$`)
	stringFieldPattern                       = regexp.MustCompile(`^[A-Za-z0-9\s,_.\/\-\(\)]{1,50}$`)
)

var statesOfIndia = map[string]struct{}{
	"andhra pradesh":              {},
	"arunachal pradesh":           {},
	"assam":                       {},
	"bihar":                       {},
	"chhattisgarh":                {},
	"delhi":                       {},
	"goa":                         {},
	"gujarat":                     {},
	"haryana":                     {},
	"himachal pradesh":            {},
	"jammu & kashmir":             {},
	"jammu and kashmir":           {},
	"jharkhand":                   {},
	"karnataka":                   {},
	"kerala":                      {},
	"maharashtra":                 {},
	"madhya pradesh":              {},
	"manipur":                     {},
	"meghalaya":                   {},
	"mizoram":                     {},
	"nagaland":                    {},
	"odisha":                      {},
	"punjab":                      {},
	"rajasthan":                   {},
	"sikkim":                      {},
	"tamil nadu":                  {},
	"tripura":                     {},
	"telangana":                   {},
	"uttar pradesh":               {},
	"uttarakhand":                 {},
	"west bengal":                 {},
	"andaman & nicobar":           {},
	"andaman and nicobar":         {},
	"andaman and nicobar islands": {},
	"andaman & nicobar islands":   {},
	"chandigarh":                  {},
	"dadra & nagar haveli":        {},
	"dadra and nagar haveli":      {},
	"daman & diu":                 {},
	"daman and diu":               {},
	"lakshadweep":                 {},
	"puducherry":                  {},
	"1cbpo":                       {},
	"2cbpo":                       {},
	"56 apo":                      {},
	"99 apo":                      {},
}

func newValidateBeatNamePatternValidator() validationRule {
	return newRule("beat_name", validateBeatNamePattern, "field %s must be in the format 'BEAT_XX', where 'XX' is a number between 0 and 99, but received %v")
}
func newBatchNumberPatternValidator() validationRule {
	return newRule("batch_number", validateBatchNamePattern, "field %s must be in the format 'BATCH_XX', where 'XX' is a number between 0 and 99, but received %v")
}
func newTimePatternValidator() validationRule {
	return newRule("time", validateTimePattern, "field %s must be in the format 'HH:MM:SS', with a valid hour (00-23), minute (00-59), and second (00-59), but received %v")
}
func newValidateHOAPatternValidator() validationRule {
	return newRule("head_of_account", validateHOAPattern, "field %s must be 15 digits, but received %v")
}
func newPersonnelNameValidator() validationRule {
	return newRule("personnel_name", validatePersonnelNamePattern, "field %s  must start and end with a letter(capital and small letters allowed) and can contain spaces in between. It should be between 3 and 50 characters long, where the middle part can include letters and spaces., but received %v")
}
func newAddressPatternValidator() validationRule {
	return newRule("address", validateAddressPattern, "field %s must start and end with an alphanumeric character, and may contain letters, digits, spaces, commas, periods, and hyphens in between. The total length should be between 3 and 50 characters. , but received %v")
}
func newEmailValidator() validationRule {
	return newRule("simple_email", validateEmailPattern, "field %s must follow the format: local-part@domain.tld, where the local part can include letters, digits, and special characters (._+-), and the domain must contain at least one dot followed by a top-level domain of at least 2 letters, but received %v")
}
func newGValidatePhoneLengthPatternValidator() validationRule {
	return newRule("phone_length", validatePhoneLengthPattern, "field %s must be 10 digits long and should be between (1000000000 9999999999), but received %v")
}
func newGValidateSOBONamePatternValidator() validationRule {
	return newRule("so_bo_name", validateSOBONamePattern, "field %s  must start and end with a letter, contain only letters and spaces, and be between 3 and 50 characters long, but received %v")
}
func newGValidatePANNumberPatternValidator() validationRule {
	return newRule("pan_number", validatePANNumberPattern, "field %s must consist of exactly 5 uppercase letters, followed by 4 digits, and ending with 1 uppercase letter, but received %v")
}
func newGValidateVehicleRegistrationNumberPatternValidator() validationRule {
	return newRule("vehicle_registration_number", validateVehicleRegistrationNumberPattern, "field %s must either be in the format 'XX99XX9999' or '99XX9999XX' where 'XX' is a letter and '99' is a digit, but received %v")
}
func newGValidateBarCodeNumberPatternValidator() validationRule {
	return newRule("bar_code_number", validateBarCodeNumberPattern, "field %s  must consist of 2 uppercase letters, followed by 9 digits, and ending with 2 uppercase letters , but received %v")
}
func newCustomValidateGLCodePatternValidator() validationRule {
	return newRule("gl_code", customValidateGLCodePattern, "field %s must start with 'GL' followed by exactly 11 digits, but received %v")
}
func newTimeStampValidatePatternValidator() validationRule {
	return newRule("date_time_stamp", timeStampValidatePattern, "field %s must be in the format 'DD-MM-YYYY HH:MM:SS', with a valid day (01-31), month (01-12), and time in 24-hour format (00-23:00-59:00-59), but received %v")
}
func newCustomValidateAnyStringLengthto50PatternValidator() validationRule {
	return newRule("customValidateAnyStringLengthto50Pattern", validateAnyStringLengthto50Pattern, "field %s must start and end with a letter and can contain up to 50 characters total, including letters and numbers, but received %v")
}
func newDateyyyymmddPatternValidator() validationRule {
	return newRule("date_yyyy_mm_dd", validatedateyyyymmddPattern, "field %s must be in the format 'YYYY-MM-DD', where YYYY is the year, MM is the month (01-12), and DD is the day (01-31), but received %v")
}
func newDateyyyymmddPatternValidatorWithddmmyyyMessage() validationRule {
	return newRule("date_yyyy_mm_dd_inverse", validatedateyyyymmddPattern, "field %s must be in the format 'DD-MM-YYYY', where YYYY is the year, MM is the month (01-12), and DD is the day (01-31), but received %v")
}
func newDateddmmyyyyPatternValidator() validationRule {
	return newRule("date_dd_mm_yyyy", validatedateddmmyyyyPattern, "field %s must be in the format 'DD-MM-YYYY', where DD is the day (01-31), MM is the month (01-12), and YYYY is the year (4 digits), but received %v")
}
func newValidateEmployeeIDPatternValidator() validationRule {
	return newRule("employee_id", validateEmployeeIDPattern, "field %s must be exactly 8 digits , but received %v")
}
func newValidateValidateGSTINPatternValidator() validationRule {
	return newRule("gst_in", validateGSTINPattern, "field %s must be a GST number in the format 'XXYYYYYZZZZABZC', where XX is the state code (2 digits), YYYYY is the business name (5 letters), ZZZZ is the registration number (4 digits), A is the entity type (1 letter), B is an alphanumeric character (1), Z is a fixed character, and C is a checksum digit (1 digit), but received %v")
}
func newPhoneNumberValidator() validationRule {
	return newRule("phone_number", validatePhoneNumberPattern, "field %s must be a valid 10-digit phone number, but received %v")
}

// ********************************

func newValidateBankUserIDPatternValidator() validationRule {
	return newRule("bank_user_id", validateBankUserIDPattern, "field %s must contain between 1 and 50 characters, consisting of uppercase letters and digits only, but received %v")
}
func newValidateOrderNumberPatternValidator() validationRule {
	return newRule("order_number", validateOrderNumberPattern, "field %s must be in the format 'LLDDDDDDDDDDDDDDDDDD', where 'LL' are 2 uppercase letters and 'DDDDDDDDDDDDDDDDDDD' are 19 digits, but received %v")
}
func newValidateAWBNumberPatternValidator() validationRule {
	return newRule("awb_number", validateAWBNumberPattern, "field %s must be in the format 'LLLLDDDDDDDDD', where 'LLLL' are 4 uppercase letters and 'DDDDDDDDD' are 9 digits, but received %v")
}
func newValidatePNRNoPatternValidator() validationRule {
	return newRule("pnr_no", validatePNRNoPattern, "field %s must be in the format 'LLLDDDDDD', where 'LLL' are 3 uppercase letters and 'DDDDDD' are 6 digits, but received %v")
}
func newValidatePLIIDPatternValidator() validationRule {
	return newRule("pli_id", validatePLIIDPattern, "field %s must be in the format 'LLLDDDDDDDD', where 'LLL' are 3 uppercase letters and 'DDDDDDDDDD' are 10 digits, but received %v")
}
func newValidatePaymentTransIDPatternValidator() validationRule {
	return newRule("payment_trans_id", validatePaymentTransIDPattern, "field %s  must be in the format 'XXYYYYYYYY-YYYY-4YYY-ZZZZ-YYYYYYYYYYYY', where 'XX' are 2 digits, 'Y' are hexadecimal characters, and 'Z' are hexadecimal characters with specific rules for version and variant, but received %v")
}
func newValidateOfficeCustomerIDPatternValidator() validationRule {
	return newRule("office_customer_id", validateOfficeCustomerIDPattern, "field %s must contain between 1 and 50 characters, consisting of letters, digits, and hyphens only, but received %v")
}
func newValidateBankIDPatternValidator() validationRule {
	return newRule("bank_id", validateBankIDPattern, "field %s must contain between 1 and 50 characters, consisting of uppercase letters and digits only, but received %v")
}
func newValidateCSIFacilityIDPatternValidator() validationRule {
	return newRule("csi_facility_id", validateCSIFacilityIDPattern, "field %s must be in the format 'LLDDDDDDDDDDD', where 'LL' are 2 uppercase letters and 'DDDDDDDDDDDDD' are 11 digits, but received %v")
}
func newValidatePosBookingOrderNumberPatternValidator() validationRule {
	return newRule("pos_booking_order_number", validatePosBookingOrderNumberPattern, "field %s must be in the format 'LLDDDDDDDDDDDDDDDDD', where 'LL' are 2 uppercase letters and 'DDDDDDDDDDDDDDDDDDD' are 19 digits, but received %v, but received %v")
}
func newValidateSOLIDPatternValidator() validationRule {
	return newRule("sol_id", validateSOLIDPattern, "field %s must be exactly 8 digits, but received %v")
}
func newValidatePLIOfficeIDPatternValidator() validationRule {
	return newRule("pli_office_id", validatePLIOfficeIDPattern, "field %s must be in the format 'LLLDDDDDDDD', where 'LLL' are 3 uppercase letters and 'DDDDDDDDDD' are 10 digits, but received %v")
}
func newValidateProductCodePatternValidator() validationRule {
	return newRule("product_code", validateProductCodePattern, "field %s must be in the format 'LLLDDDDDDDDDD', where 'LLL' are 3 uppercase letters and 'DDDDDDDDDDDD' are 12 digits, but received %v")
}
func newValidateCustomerIDPatternValidator() validationRule {
	return newRule("customer_id", validateCustomerIDPattern, "field %s must be exactly 10 digits in length, but received %v")
}
func newValidateFacilityIDPatternValidator() validationRule {
	return newRule("facility_id", validateFacilityIDPattern, "field %s must be in the format 'LLDDDDDDDDDDD', where 'LL' are 2 uppercase letters and 'DDDDDDDDDDD' are 11 digits, but received %v")
}
func newValidateApplicationIDPatternValidator() validationRule {
	return newRule("application_id", validateApplicationIDPattern, "field %s must be in the format 'LLLDDDDDDDD-DDD', where 'LLL' are 3 uppercase letters, 'DDDDDDDD' are 8 digits, and 'DDD' are 3 digits after the hyphen, but received %v")
}
func newValidateReceiverKYCReferencePatternValidator() validationRule {
	return newRule("receiver_kyc_reference", validateReceiverKYCReferencePattern, "field %s must start with 'KYCREF' followed by up to 44 alphanumeric characters, but received %v")
}
func newValidateOfficeCustomerPatternValidator() validationRule {
	return newRule("office_customer", validateOfficeCustomerPattern, "field %s must consist of letters, numbers, and spaces only, and cannot be empty(special characters are not allowed), but received %v")
}
func newValidatePRANPatternValidator() validationRule {
	return newRule("pran_no", validatePRANPattern, "field %s must be exactly 12 digits, but received %v")
}
func newvalidateCustomFlightNoValidator() validationRule {
	return newRule("flight_no", validateCustomFlightNo, "field %s must contain only letters, digits, and spaces, but received %v")
}
func newvalidatePinCodeGlobalValidator() validationRule {
	return newRule("pincode", validatePinCodeGlobal, "field %s must be 6 digits. The first digit must be 1-9, last five digits cant be all zeros and also last three digits cant be all zeros, but received %v")
}
func newValidateMobileNumberPatternValidator() validationRule {
	return newRule("mobile_number", validateMobileNumberStringPattern, "field %s must be a valid 10-digit phone number starting with a digit between 6 and 9, but received %v")
}
func newCustomofficeidGlobalValidator() validationRule {
	return newRule("office_id", customofficeidGlobal, "field %s must be between 10000000 & 99999999, but received %v")
}
func newvalidateBagIdPatternValidator() validationRule {
	return newRule("bag_id", validateBagIdPattern, "field %s must be a valid bag ID with either 3 letters followed by 10 digits(domestic), or 15 letters followed by 14 digits(international), but received %v")
}
func newCustomTrainNoGlobalValidator() validationRule {
	return newRule("train_no", customTrainNoGlobal, "field %s must be a valid train number between 10000 & 99999, but received %v")
}
func newCustomSCSGlobalValidator() validationRule {
	return newRule("seating_capacity", customSCSGlobal, "field %s must be a number between 1 and 9999, but received %v")
}
func newvalidateCircleIDGlobalValidator() validationRule {
	return newRule("circle_id", validateCircleIDGlobal, "field %s must be a number between 1 and 9999, but received %v")
}
func newvalidateTariffIDGlobalValidator() validationRule {
	return newRule("tariff_id", validateTariffIDGlobal, "field %s must be a number between 1000000000 & 9999999999, but received %v")
}
func newvalidateCIFNumGlobalValidator() validationRule {
	return newRule("cif_number", validateCIFNumGlobal, "field %s must be a number between 100000000 & 999999999, but received %v")
}
func newvalidateContractNumGlobalValidator() validationRule {
	return newRule("contract_number", validateContractNumGlobal, "field %s must be a number between 10000000 & 99999999, but received %v")
}
func newvalidateRegionIDGlobalValidator() validationRule {
	return newRule("region_id", validateRegionIDGlobal, "field %s must be a number between 1000000 & 9999999, but received %v")
}
func newvalidateVasIDGlobalValidator() validationRule {
	return newRule("vas_id", validateVasIDGlobal, "field %s must be a number between 1000000 & 9999999, but received %v")
}
func newvalidateUserCodeGlobalValidator() validationRule {
	return newRule("user_code", validateUserCodeGlobal, "field %s must be a number between 10000000 & 99999999, but received %v")
}
func newvalidateHONamePatternValidator() validationRule {
	return newRule("ho_name", validateHONamePattern, "field %s must not contain any special characters, but received %v")
}
func newvalidateHOIDGlobalValidator() validationRule {
	return newRule("ho_id", validateHOIDGlobal, "field %s must be a number between 1000000 &  9999999, but received %v")
}
func newvalidateAccountNoGlobalValidator() validationRule {
	return newRule("account_no", validateAccountNoGlobal, "field %s must be a number between 1000000000 & 9999999999, but received %v")
}
func newIsValidTimestampGlobalValidator() validationRule {
	return newRule("time_stamp", isValidTimestampGlobal, "field %s must be a valid timestamp in string (RFC3339, e.g., '2023-10-05T14:48:00Z'), but received %v")
}
func newIsValidStateValidator() validationRule {
	return newRule("state", validatedStateGlobal, "field %s must be a valid Indian state, but received %v")
}

func newvalidateCityNameValidator() validationRule {
	return newRule("city_name", validateCityNamePattern, "field %s must start and end with  character, and may contain letters,digits and  spaces, commas, periods, and hyphens in between. The total length should be between 3 and 50 characters. , but received %v")
}
func newvalidateAadharValidator() validationRule {
	return newRule("aadhaar_no", validateAadharPattern, "field %s must be exactly 12 digits, but received %v")
}

func newvalidateDrivingLicenseNoValidator() validationRule {
	return newRule("driving_license", validateDrivingLicenseNoPattern, "field %s must be between 10 and 20 alpanumericcharacters, but received %v")
}

func newvalidatePassportNoValidator() validationRule {
	return newRule("passport_no", validatePassportNoPattern, "field %s must be exactly 8 characters in format G1234567, but received %v")
}

func newvalidateVoterIDValidator() validationRule {
	return newRule("voter_id", validateVoterIDPattern, "field %s must be exactly 10 characters in format XYZ3300779 , but received %v")
}

func newvalidateOfficeNameValidator() validationRule {
	return newRule("office_name", validateOfficeNamePattern, "field %s  must start and end with a letter or number only, may include special characters such as . and (), can contain only letters,numbers and spaces, and must be between 3 and 50 characters long, but received %v")
}

func newvalidateMonthValidator() validationRule {
	return newRule("month", validateMonthPattern, "field %s  must start 0X,X0,XX,MAR,March,Mar,MARCH, but received %v")
}

func newvalidateYearValidator() validationRule {

	return newRule("year", validateYearPattern, "field %s must 4 digit year XXXX and in range of 1900 to 3000, but received %v")
}

func newOptionalFieldValidator() validationRule {
	return newRule("optional", optionalField, "field %s must be empty or a valid value, but received %v")
}

func newStringFieldValidator() validationRule {
	return newRule("string_field", validateStringField, "field %s must start and end with an alphanumeric character, and may contain letters, digits, spaces, commas, periods, parentheses and hyphens. The total length should be between 1 and 50 characters, but received %v")
}

func validateStringField(fl validator.FieldLevel) bool {
	return validateWithGlobalRegex(fl, stringFieldPattern)
}

func optionalField(fl validator.FieldLevel) bool {
	return true
}

// validate time stamp in format:2024-01-01T00:00:00Z
func isValidTimestampGlobal(fl validator.FieldLevel) bool {
	_, err := time.Parse(time.RFC3339, fl.Field().String())
	return err == nil
}

func validateWithGlobalRegex(fl validator.FieldLevel, regex *regexp.Regexp) bool {
	fieldValue := fl.Field().String()
	return regex.MatchString(fieldValue)
}
func validateBeatNamePattern(f1 validator.FieldLevel) bool {
	return validateWithGlobalRegex(f1, beatNamePattern)
}
func validateBatchNamePattern(f1 validator.FieldLevel) bool {
	return validateWithGlobalRegex(f1, batchNamePattern)
}
func validateTimePattern(f1 validator.FieldLevel) bool {
	return validateWithGlobalRegex(f1, timePattern)
}
func validateHOAPattern(fl validator.FieldLevel) bool {
	//pattern := `^\d{15}$`
	return validateWithGlobalRegex(fl, hoaPattern)
}
func validatePersonnelNamePattern(fl validator.FieldLevel) bool {
	return validateWithGlobalRegex(fl, personnelNamePattern)
}
func validateAddressPattern(fl validator.FieldLevel) bool {
	return validateWithGlobalRegex(fl, addressPattern)
}
func validateEmailPattern(fl validator.FieldLevel) bool {
	return validateWithGlobalRegex(fl, emailPattern)
}
func validatePhoneNumberPattern(fl validator.FieldLevel) bool {
	return validateWithGlobalRegex(fl, phoneNumberPattern)
}
func validatePhoneLengthPattern(fl validator.FieldLevel) bool {
	// Handle the case where the phone number is a string
	if _, ok := fl.Field().Interface().(string); ok {
		// Validate using a regular expression for exactly 10 digits
		// pattern := `^\d{10}$`
		// return ValidateWithRegex(fl, pattern)
		return validateWithGlobalRegex(fl, phoneLengthPattern)
	}

	// Handle the case where the phone number is a uint64
	if phoneNumber, ok := fl.Field().Interface().(uint64); ok {
		// Check if the phone number has exactly 10 digits
		return phoneNumber >= 1000000000 && phoneNumber <= 9999999999
	}
	//works only for 64 bit system
	// Handle the case where the phone number is an int
	if phoneNumber, ok := fl.Field().Interface().(int); ok {
		// Check if the phone number has exactly 10 digits
		return phoneNumber >= 1000000000 && phoneNumber <= 9999999999
	}

	// If the field is neither a string, uint64, nor int, the validation fails
	return false
}

func validateSOBONamePattern(f1 validator.FieldLevel) bool {
	// Define the regex pattern
	// ^[A-Za-z] -> Start with a letter
	// [A-Za-z\s]{1,48} -> 1 to 48 letters or spaces
	// [A-Za-z]$ -> End with a letter
	//pattern := `^[A-Za-z][A-Za-z\s]{1,48}[A-Za-z]$`

	return validateWithGlobalRegex(f1, soboNamePattern)
}
func validatePANNumberPattern(fl validator.FieldLevel) bool {
	// regex pattern for PAN number (5 letters followed by 4 digits followed by 1 letter)
	//pattern := `^[A-Z]{5}[0-9]{4}[A-Z]$`

	return validateWithGlobalRegex(fl, panNumberPattern)
}
func validateVehicleRegistrationNumberPattern(fl validator.FieldLevel) bool {
	// Define the regex pattern for vehicle registration number
	//pattern := `^[A-Z]{2}\d{2}[A-Z]{1,2}\d{4,7}$`
	return validateWithGlobalRegex(fl, vehicleRegistrationNumberPattern)
}
func validateBarCodeNumberPattern(fl validator.FieldLevel) bool {

	// Define the regex pattern for vehicle registration number
	//pattern := `^[A-Z]{2}\d{6,12}[A-Z]{2}$`
	return validateWithGlobalRegex(fl, barCodeNumberPattern)
}
func customValidateGLCodePattern(fl validator.FieldLevel) bool {
	//pattern := `^GL\d{11}$`
	return validateWithGlobalRegex(fl, glCodePattern)
}
func timeStampValidatePattern(f1 validator.FieldLevel) bool {
	//dateTimeRegex := regexp.MustCompile(`^(0[1-9]|[12][0-9]|3[01])-(0[1-9]|1[0-2])-(\d{4}) ([01]\d|2[0-3]):([0-5]\d):([0-5]\d)$`)
	return validateWithGlobalRegex(f1, timeStampPattern)
}
func validateAnyStringLengthto50Pattern(fl validator.FieldLevel) bool {
	//pattern := `^[a-zA-Z][a-zA-Z0-9]{0,48}[a-zA-Z]$`
	// Check if the string matches the regex pattern
	return validateWithGlobalRegex(fl, customValidateAnyStringLengthto50Pattern)
}
func validatedateyyyymmddPattern(fl validator.FieldLevel) bool {
	//pattern := `^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`
	// Check if the date matches the regex pattern
	return validateWithGlobalRegex(fl, dateyyyymmddPattern)

}
func validatedateddmmyyyyPattern(fl validator.FieldLevel) bool {
	//pattern := `^(0[1-9]|[12][0-9]|3[01])-(0[1-9]|1[0-2])-\d{4}$`

	// Check if the date matches the regex pattern
	return validateWithGlobalRegex(fl, dateddmmyyyyPattern)

}
func isEmployeeID(employeeId int) bool {
	return employeeId >= 10000000 && employeeId <= 99999999
}
func validateEmployeeIDPattern(fl validator.FieldLevel) bool {
	if employeeId, ok := fl.Field().Interface().(uint64); ok {
		return isEmployeeID(int(employeeId))
	}
	if employeeId, ok := fl.Field().Interface().(int64); ok {
		return isEmployeeID(int(employeeId))
	}
	if employeeId, ok := fl.Field().Interface().(int); ok {
		return isEmployeeID(employeeId)
	}
	/**
	* !validation of string is removed as it is not required
	if _, ok := fl.Field().Interface().(string); ok {

		return validateWithGlobalRegex(fl, employeeIDPattern)
	}
	*/
	return false
}
func validateGSTINPattern(fl validator.FieldLevel) bool {

	// Define the regex pattern for GSTIN validation
	//pattern := `^[0-9]{2}[A-Z]{5}[0-9]{4}[A-Z]{1}[A-Z0-9]{1}[Z]{1}[0-9]{1}$`

	return validateWithGlobalRegex(fl, gstINPattern)
}

func validateMobileNumberStringPattern(fl validator.FieldLevel) bool {
	return validateWithGlobalRegex(fl, mobileNumberStringPattern)
}
func validateBagIdPattern(fl validator.FieldLevel) bool {
	return validateWithGlobalRegex(fl, bagIdPattern)
}
func validateCityNamePattern(fl validator.FieldLevel) bool {
	return validateWithGlobalRegex(fl, cityNamePattern)
}
func validateAadharPattern(fl validator.FieldLevel) bool {
	return validateWithGlobalRegex(fl, aadharPattern)
}

func validateDrivingLicenseNoPattern(fl validator.FieldLevel) bool {
	return validateWithGlobalRegex(fl, drivingLicenseNoPattern)
}

func validatePassportNoPattern(fl validator.FieldLevel) bool {
	return validateWithGlobalRegex(fl, passportNoPattern)
}

func validateVoterIDPattern(fl validator.FieldLevel) bool {

	return validateWithGlobalRegex(fl, voterIDPattern)
}

func validateOfficeNamePattern(fl validator.FieldLevel) bool {
	officeName := fl.Field().String()
	if onlyNumbersWithSpaces.MatchString(officeName) {
		return false
	}
	return officeNamePattern.MatchString(officeName)

}

func validateMonthPattern(fl validator.FieldLevel) bool {
	return validateWithGlobalRegex(fl, monthPattern)
}

// Helper function to check if the year is within the valid range
func isValidYear(year int) bool {
	return year >= 1900 && year <= 3000
}
func validateYearPattern(fl validator.FieldLevel) bool {
	if year, ok := fl.Field().Interface().(uint64); ok {
		// Check if the year has exactly 4 digits
		//return year >= 1000 && year <= 9999
		return isValidYear(int(year))
	}
	if year, ok := fl.Field().Interface().(int64); ok {
		// Check if the year has exactly 4 digits
		//return year >= 1000 && year <= 9999
		return isValidYear(int(year))
	}
	if year, ok := fl.Field().Interface().(int); ok {
		// Check if the year has exactly 4 digits
		//return year >= 1000 && year <= 9999
		return isValidYear(year)
	}
	if _, ok := fl.Field().Interface().(string); ok {
		return validateWithGlobalRegex(fl, yearPattern)
	}
	return false

}

//***********************************************

func validatePRANPattern(fl validator.FieldLevel) bool {
	// Handle the case where the PRAN is a string
	if _, ok := fl.Field().Interface().(string); ok {
		// Define a regex pattern to match exactly 12 digits
		//pattern := `^\d{12}$`

		return validateWithGlobalRegex(fl, pranPattern)
	}

	// Handle the case where the PRAN is an int64
	if pranInt, ok := fl.Field().Interface().(int64); ok {
		// Check if the int64 falls within the 12-digit range
		return pranInt >= 100000000000 && pranInt <= 999999999999
	}

	// If the field is neither a valid string nor a valid integer, the validation fails
	return false
}

func validateOfficeCustomerPattern(fl validator.FieldLevel) bool {
	// Regular expression to allow only alphanumeric characters and spaces
	// This will disallow special characters like @, #, $, %, etc.
	//pattern := `^[a-zA-Z0-9\s]+$`

	// Get the field value and convert it to a string
	officeCustomer, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	// Check if the length of the string is within 50 characters
	if len(officeCustomer) > 50 {
		return false
	}

	// Check if the office_customer string matches the allowed pattern
	return validateWithGlobalRegex(fl, officeCustomerPattern)
}

func validateReceiverKYCReferencePattern(fl validator.FieldLevel) bool {

	// Define a regex pattern to match the format KYCREF followed by up to 44 alphanumeric characters
	//pattern := `^KYCREF[A-Z0-9]{0,44}$`
	// Check if the string matches the pattern
	return validateWithGlobalRegex(fl, receiverKYCReferencePattern)
}

func validateApplicationIDPattern(fl validator.FieldLevel) bool {
	// Define a regex pattern to match the format <3 uppercase letters><12 digits with hyphen>
	//pattern := `^[A-Z]{3}\d{8}-\d{3}$`
	// Check if the string matches the pattern
	return validateWithGlobalRegex(fl, applicationIDPattern)
}

func validateFacilityIDPattern(fl validator.FieldLevel) bool {

	// Define a regex pattern to match the format <2 uppercase letters><11 digits>
	//pattern := `^[A-Z]{2}\d{11}$`
	// Check if the string matches the pattern
	return validateWithGlobalRegex(fl, facilityIDPattern)
}

func validateCustomerIDPattern(fl validator.FieldLevel) bool {
	// Handle the case where the value is a string
	if _, ok := fl.Field().Interface().(string); ok {
		return validateWithGlobalRegex(fl, customerIDPattern)
	}

	// Handle the case where the value is an int
	if customerIDInt, ok := fl.Field().Interface().(int); ok {
		return customerIDInt >= 1000000000 && customerIDInt <= 9999999999
	}

	// Handle the case where the value is an int64
	if customerIDInt64, ok := fl.Field().Interface().(int64); ok {
		return customerIDInt64 >= 1000000000 && customerIDInt64 <= 9999999999
	}

	// Handle the case where the value is a uint64
	if customerIDUint64, ok := fl.Field().Interface().(uint64); ok {
		return customerIDUint64 >= 1000000000 && customerIDUint64 <= 9999999999
	}

	// If the field is neither a string nor an integer, validation fails
	return false
}

func validateProductCodePattern(fl validator.FieldLevel) bool {
	// Assume the fl value is always a string

	// Define a regex pattern to match the format <3 uppercase letters><12 digits>
	//pattern := `^[A-Z]{3}\d{12}$`

	// Check if the string matches the pattern
	return validateWithGlobalRegex(fl, productCodePattern)
}

func validatePLIOfficeIDPattern(fl validator.FieldLevel) bool {
	// Assume the fl value is always a string

	// Define a regex pattern to match the format <3 uppercase letters><10 digits>
	//pattern := `^[A-Z]{3}\d{10}$`
	// Check if the string matches the pattern
	return validateWithGlobalRegex(fl, pliOfficeIDPattern)
}

func validateSOLIDPattern(fl validator.FieldLevel) bool {
	// Assume the fl value is always a string

	// Define a regex pattern to match the format <6 digits pincode><2 digits office type number>
	//pattern := `^\d{6}\d{2}$`
	// Check if the string matches the pattern
	return validateWithGlobalRegex(fl, solIdPattern)
}

func validatePosBookingOrderNumberPattern(fl validator.FieldLevel) bool {
	// Assume the fl value is always a string

	// Define a regex pattern to match the format <2 uppercase letters><19 digits>
	//pattern := `^[A-Z]{2}\d{19}$`
	// Check if the string matches the pattern
	return validateWithGlobalRegex(fl, posBookingOrderNumberPattern)
}

func validateCSIFacilityIDPattern(fl validator.FieldLevel) bool {
	// Handle the case where the csi_facility_id is a string
	if _, ok := fl.Field().Interface().(string); ok {
		// Define a regex pattern that matches the format <2 uppercase letters><11 digit numeric>
		//pattern := `^[A-Z]{2}\d{11}$`
		// Check if the csi_facility_id matches the pattern
		return validateWithGlobalRegex(fl, csiFacilityIDPattern)
	}

	// If the field is not a string, the validation fails
	return false
}

func validateBankIDPattern(fl validator.FieldLevel) bool {
	// Handle the case where the value is a string
	if _, ok := fl.Field().Interface().(string); ok {
		// Define a regex pattern to match a string with up to 50 characters consisting of uppercase letters and digits
		//pattern := `^[A-Z0-9]{1,50}$`
		// Check if the string matches the pattern
		return validateWithGlobalRegex(fl, bankIDPattern)
	}

	// If the field is not a string, validation fails
	return false
}

func validateOfficeCustomerIDPattern(fl validator.FieldLevel) bool {
	// Handle the case where the value is a string
	if _, ok := fl.Field().Interface().(string); ok {
		// Define a regex pattern to match any string with up to 50 characters

		//pattern := `^[a-zA-Z0-9\-]{1,50}$`
		// Check if the string matches the pattern
		return validateWithGlobalRegex(fl, officeCustomerIDPattern)
	}

	// If the field is not a string, validation fails
	return false
}

func validatePaymentTransIDPattern(fl validator.FieldLevel) bool {
	// Handle the case where the payment_trans_id is a string
	if _, ok := fl.Field().Interface().(string); ok {
		// Define a regex pattern that matches the format <2digit><uuid v4>
		//pattern := `^\d{2}[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`

		// Check if the payment_trans_id matches the pattern
		return validateWithGlobalRegex(fl, paymentTransIDPattern)
	}

	// If the field is not a string, the validation fails
	return false
}

func validatePLIIDPattern(fl validator.FieldLevel) bool {
	// Handle the case where the pli_id is a string
	if _, ok := fl.Field().Interface().(string); ok {
		// Define a regex pattern that matches the format <3 uppercase letters><10 digit numeric>
		//pattern := `^[A-Z]{3}\d{10}$`
		// Check if the awbnumber matches the pattern
		return validateWithGlobalRegex(fl, pliIDPattern)
	}

	// If the field is not a string, the validation fails
	return false
}

func validatePNRNoPattern(fl validator.FieldLevel) bool {

	// Handle the case where the pnr_no is a string
	if _, ok := fl.Field().Interface().(string); ok {
		// Define a regex pattern that matches the format
		//pattern := `^[A-Z]{3}\d{6}$`
		// Check if the pnr_no matches the pattern
		return validateWithGlobalRegex(fl, pnrNoPattern)
	}

	// If the field is not a string, the validation fails
	return false
}

func validateAWBNumberPattern(fl validator.FieldLevel) bool {
	// Handle the case where the awbnumber is a string
	if _, ok := fl.Field().Interface().(string); ok {
		// Define a regex pattern that matches the format <4 uppercase letters><9 digit numeric>
		//pattern := `^[A-Z]{4}\d{9}$`

		// Check if the awbnumber matches the pattern
		return validateWithGlobalRegex(fl, awbNumberPattern)
	}

	// If the field is not a string, the validation fails
	return false
}

func validateOrderNumberPattern(fl validator.FieldLevel) bool {
	// Handle the case where the order_number is a string
	if _, ok := fl.Field().Interface().(string); ok {
		// Define a regex pattern that matches the format <2 uppercase letters><19 digit numeric>
		//pattern := `^[A-Z]{2}\d{19}$`
		// Check if the order_number matches the pattern
		return validateWithGlobalRegex(fl, orderNumberPattern)
	}

	// If the field is not a string, the validation fails
	return false
}

func validateBankUserIDPattern(fl validator.FieldLevel) bool {
	// Handle the case where the bank_user_id is a string
	if _, ok := fl.Field().Interface().(string); ok {
		// Define a regex pattern that ensures the bank_user_id is alphanumeric and between 1 to 50 characters
		//pattern := `^[A-Z0-9]{1,50}$`

		// Check if the bank_user_id matches the pattern
		return validateWithGlobalRegex(fl, bankUserIDPattern)
	}

	// If the field is not a string, the validation fails
	return false
}

func validateHONamePattern(fl validator.FieldLevel) bool {
	// Handle the case where the ho_name is a string
	if hoName, ok := fl.Field().Interface().(string); ok {
		// Check if the ho_name is not empty and has a maximum length of 50 characters
		if len(hoName) == 0 || len(hoName) > 50 {
			return false
		}

		// Define a regex pattern that disallows special characters @,#/$%!^&*()<>:;"{}[]
		// specialCharPattern := `[!@#$%^&*()<>:;"{}[\]\\]`
		// regex := regexp.MustCompile(specialCharPattern)

		// Check if the ho_name contains any special characters
		if specialCharPattern.MatchString(hoName) {
			return false
		}

		// If all checks pass, return true
		return true
	}

	// If the field is not a string, the validation fails
	return false
}

func validatePinCodeGlobal(fl validator.FieldLevel) bool {
	field := fl.Field()

	var zipCode string

	switch field.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32:
		zipCode = strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint64, reflect.Uint32:
		zipCode = strconv.FormatUint(field.Uint(), 10)
	default:
		return false
	}

	// Check if the length is 6
	if len(zipCode) != 6 {
		return false
	}

	// Check if the pin code contains only digits
	if _, err := strconv.Atoi(zipCode); err != nil {
		return false
	}

	// Check if the first digit is in the range 1 to 9
	firstDigit, _ := strconv.Atoi(string(zipCode[0]))
	if firstDigit < 1 || firstDigit > 9 {
		return false
	}

	// Check if the last five digits are not all zeros
	lastFiveDigits := zipCode[1:6]
	if lastFiveDigits == "00000" {
		return false
	}

	// Check if the last three digits are not all zeros
	lastThreeDigits := zipCode[3:6]
	if lastThreeDigits == "000" {
		return false
	}

	return true
}

func validateCustomFlightNo(fl validator.FieldLevel) bool {
	if _, ok := fl.Field().Interface().(string); ok {
		// Define a regex pattern that matches the format
		//pattern := `^[A-Z]{3}\d{6}$`
		// Check if the pnr_no matches the pattern
		return validateWithGlobalRegex(fl, flightNopattern)
	}

	// If the field is not a string, the validation fails
	return false

}

// //////////////////////////////////////////////////without regex functions
func customofficeidGlobal(fl validator.FieldLevel) bool {
	// Handle the case where the officeId is an int
	if officeId, ok := fl.Field().Interface().(int); ok {
		return officeId >= 10000000 && officeId <= 99999999
	}
	if officeId, ok := fl.Field().Interface().(int64); ok {
		return officeId >= 10000000 && officeId <= 99999999
	}

	// Handle the case where the officeId is a uint64
	if officeId, ok := fl.Field().Interface().(uint64); ok {
		return officeId >= 10000000 && officeId <= 99999999
	}
	/**
	* !!! changed the validation of office id from 7 to 8 digits only and as string validation is not required removed the code below

	// Handle the case where the officeId is a string// changed the validation of office id from 7 to 8 digits only
	if officeIdStr, ok := fl.Field().Interface().(string); ok {
		// Check if the string is not empty and contains only digits
		//if len(officeIdStr) >= 7 && len(officeIdStr) <= 8 {
		if len(officeIdStr) == 8 {
			if _, err := strconv.ParseUint(officeIdStr, 10, 64); err == nil {
				return true
			}
		}
	}
	*/
	// If the field is neither an int, uint64, nor a valid string, the validation fails
	return false
}
func customTrainNoGlobal(fl validator.FieldLevel) bool {
	// Attempt to get the train number as a uint64
	if trainNo, ok := fl.Field().Interface().(uint64); ok {
		// Check if the train number has exactly 5 digits
		return trainNo >= 10000 && trainNo <= 99999
	}

	// Attempt to get the train number as a string
	if trainNoStr, ok := fl.Field().Interface().(string); ok {
		// Define a regex pattern to match exactly 5 digits
		//regex := regexp.MustCompile(`^\d{5}$`)
		// Check if the string matches the regex pattern
		return trainNoPattern.MatchString(trainNoStr)
	}

	// If the value is neither a 5-digit uint64 nor a 5-digit string, validation fails
	return false
}

// seating capacity in a train
func customSCSGlobal(fl validator.FieldLevel) bool {
	// Get the train number from the field
	seating, ok := fl.Field().Interface().(uint64)
	if !ok {
		// If it's not a uint64, the validation fails
		return false
	}
	// Check if the strength  has exactly  1 to 4 digits
	return seating >= 1 && seating <= 9999
}

// Circle_id is validation for integer . example: 90000013 starting with 7 digit
func validateCircleIDGlobal(fl validator.FieldLevel) bool {

	// Handle the case where the Circle_id is an integer
	if usercode, ok := fl.Field().Interface().(int); ok {

		return usercode >= 9000001 && usercode <= 9999999999
	}
	// Check if the value is a string and attempt to parse it as an integer
	if usercodeStr, ok := fl.Field().Interface().(string); ok {
		// Convert the string to an integer
		if usercode, err := strconv.ParseInt(usercodeStr, 10, 64); err == nil {
			return usercode >= 9000001 && usercode <= 9999999999
		}
	}
	// If the field is neither an integer nor a string, the validation fails
	return false
}

// tariff_id  is validation for integer . example: 1234567890 10 digit
func validateTariffIDGlobal(fl validator.FieldLevel) bool {

	// Handle the case where the tariff_id is an integer
	if usercode, ok := fl.Field().Interface().(int); ok {

		return usercode >= 1000000000 && usercode <= 9999999999
	}
	// Check if the value is a string and attempt to parse it as an integer
	if usercodeStr, ok := fl.Field().Interface().(string); ok {
		// Convert the string to an integer
		if usercode, err := strconv.ParseInt(usercodeStr, 10, 64); err == nil {
			return usercode >= 1000000000 && usercode <= 9999999999
		}
	}
	// If the field is neither an integer nor a string, the validation fails
	return false
}

// CIF Number is validation for integer . example: 327711299
func validateCIFNumGlobal(fl validator.FieldLevel) bool {

	// Handle the case where the CIF is an integer
	if usercode, ok := fl.Field().Interface().(int); ok {

		return usercode >= 100000000 && usercode <= 999999999
	}
	// Check if the value is a string and attempt to parse it as an integer
	if usercodeStr, ok := fl.Field().Interface().(string); ok {
		// Convert the string to an integer
		if usercode, err := strconv.ParseInt(usercodeStr, 10, 64); err == nil {
			return usercode >= 100000000 && usercode <= 999999999
		}
	}
	// If the field is neither an integer nor a string, the validation fails
	return false
}

// Contract Number is validation for integer(8) . example: 40057692
func validateContractNumGlobal(fl validator.FieldLevel) bool {

	// Handle the case where the ContractNum is an integer
	if usercode, ok := fl.Field().Interface().(int); ok {

		return usercode >= 10000000 && usercode <= 99999999
	}
	// Check if the value is a string and attempt to parse it as an integer
	if usercodeStr, ok := fl.Field().Interface().(string); ok {
		// Convert the string to an integer
		if usercode, err := strconv.ParseInt(usercodeStr, 10, 64); err == nil {
			return usercode >= 10000000 && usercode <= 99999999
		}
	}
	// If the field is neither an integer nor a string, the validation fails
	return false
}

// region_id is validation for integer(10) . example: 9000001
func validateRegionIDGlobal(fl validator.FieldLevel) bool {

	// Handle the case where the region_id is an integer
	if usercode, ok := fl.Field().Interface().(int); ok {

		return usercode >= 1000000 && usercode <= 9999999
	}
	// Check if the value is a string and attempt to parse it as an integer
	if usercodeStr, ok := fl.Field().Interface().(string); ok {
		// Convert the string to an integer
		if usercode, err := strconv.ParseInt(usercodeStr, 10, 64); err == nil {
			return usercode >= 1000000 && usercode <= 9999999
		}
	}
	// If the field is neither an integer nor a string, the validation fails
	return false
}

// vas_id is validation for integer(10) . example: 1234567
func validateVasIDGlobal(fl validator.FieldLevel) bool {

	// Handle the case where the vas_id is an integer
	if usercode, ok := fl.Field().Interface().(int); ok {
		// Check if the vas_id has exactly 10 digits
		return usercode >= 1000000 && usercode <= 9999999
	}
	// Check if the value is a string and attempt to parse it as an integer
	if usercodeStr, ok := fl.Field().Interface().(string); ok {
		// Convert the string to an integer
		if usercode, err := strconv.ParseInt(usercodeStr, 10, 64); err == nil {
			return usercode >= 1000000 && usercode <= 9999999
		}
	}
	// If the field is neither an integer nor a string, the validation fails
	return false
}

// usercode is validation for integer(10) . example: 10181686
func validateUserCodeGlobal(fl validator.FieldLevel) bool {

	// Handle the case where the usercode is an integer
	if usercode, ok := fl.Field().Interface().(int); ok {
		// Check if the usercode has exactly 10 digits
		return usercode >= 10000000 && usercode <= 99999999
	}
	// Check if the value is a string and attempt to parse it as an integer
	if usercodeStr, ok := fl.Field().Interface().(string); ok {
		// Convert the string to an integer
		if usercode, err := strconv.ParseInt(usercodeStr, 10, 64); err == nil {
			return usercode >= 10000000 && usercode <= 99999999
		}
	}
	// If the field is neither an integer nor a string, the validation fails
	return false
}

// ho_id validation for 7 digit integer or 7 digit string
func validateHOIDGlobal(fl validator.FieldLevel) bool {

	// Handle the case where the ValidateHOID  is a uint64
	if gNumber, ok := fl.Field().Interface().(uint64); ok {
		// Check if the ho_id  has exactly 7 digits
		return gNumber >= 1000000 && gNumber <= 9999999
	}

	// Handle the case where the ValidateHOID  is an int
	if gNumber, ok := fl.Field().Interface().(int); ok {
		// Check if the ho_id has exactly 7 digits
		return gNumber >= 1000000 && gNumber <= 9999999
	}
	// Check if the value is a string and attempt to parse it as an integer
	if gNumberStr, ok := fl.Field().Interface().(string); ok {
		// Convert the string to an integer
		if gNumber, err := strconv.ParseInt(gNumberStr, 10, 64); err == nil {
			return gNumber >= 1000000 && gNumber <= 9999999
		}
	}
	// If the field is neither a string, uint64, nor int, the validation fails
	return false
}

// account_no validation for 10 digit integer or 10 digit string
func validateAccountNoGlobal(fl validator.FieldLevel) bool {

	// Handle the case where the ValidateAccountNo  is a uint64
	if gNumber, ok := fl.Field().Interface().(uint64); ok {
		// Check if the account_no  has exactly 10 digits
		return gNumber >= 1000000000 && gNumber <= 9999999999
	}

	// Handle the case where the ValidateAccountNo  is an int
	if gNumber, ok := fl.Field().Interface().(int); ok {
		// Check if the account_no has exactly 10 digits
		return gNumber >= 1000000000 && gNumber <= 9999999999
	}
	// Check if the value is a string and attempt to parse it as an integer
	if gNumberStr, ok := fl.Field().Interface().(string); ok {
		// Convert the string to an integer
		if gNumber, err := strconv.ParseInt(gNumberStr, 10, 64); err == nil {
			return gNumber >= 1000000000 && gNumber <= 9999999999
		}
	}
	// If the field is neither a string, uint64, nor int, the validation fails
	return false
}

func validatedStateGlobal(fl validator.FieldLevel) bool {
	// Get the state from the field
	state := fl.Field().String()
	state = strings.ToLower(state)

	// Check if the state is a valid Indian state
	_, ok := statesOfIndia[state]
	return ok
}

func generateDynamicStringValidationPattern(minLength, maxLength uint, additionalChars ...rune) (*regexp.Regexp, error) {
	// Base pattern with existing allowed characters (properly escaped)
	basePattern := `A-Za-z0-9\s,_.\/\-\(\)`

	// Add additional characters to the pattern
	for _, char := range additionalChars {
		// Skip characters already present in the base pattern
		if strings.Contains(basePattern, string(char)) || (char == ' ' && strings.Contains(basePattern, `\s`)) {
			continue
		}
		basePattern += regexp.QuoteMeta(string(char))
	}

	// Construct the final regex pattern with dynamic length
	finalPattern := fmt.Sprintf("^[%s]{%d,%d}$", basePattern, minLength, maxLength)

	// Compile the regex pattern
	return regexp.Compile(finalPattern)
}

func validateArgs(tagSuffix string, minLength, maxLength uint, char ...rune) error {
	if tagSuffix == "" {
		return errors.New("validation tag cannot be empty")
	}

	if minLength == 0 && maxLength == 0 {
		return errors.New("both minLength and maxLength cannot be zero")
	}

	if minLength > maxLength {
		return errors.New("minLength cannot be greater than maxLength")
	}

	if len(char) > 0 {
		for _, c := range char {
			if c < 0 || c > 127 {
				return errors.New("additional characters must be valid ASCII characters")
			}
		}
	}

	return nil

}
