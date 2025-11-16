package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

/*

// response represents a response body format
type Response struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Success"`
	Data    any    `json:"data,omitempty"`
}

// newResponse is a helper function to create a response body
func newResponse(success bool, message string, data any) Response {
	return Response{
		Success: success,
		Message: message,
		Data:    data,
	}
}

// errorResponse represents an error response body format
type errorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"Error message"`
}


type MsgApplicationResponse struct {
	ApplicationID   uint64 `json:"application_id"`
	ApplicationName string `json:"application_name"`
	//ProductIDs      string    `json:"product_ids"`
	RequestType string `json:"request_type"`
	//CallbackUrl     string    `json:"callback_url"`
	//WhitelistedIp   string    `json:"whitelisted_ip"`
	SecretKey   string    `json:"secret_key"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
	Status      int       `json:"status" db:"status_cd"`
}

type EditMsgApplicationResponse struct {
	ApplicationID   uint64 `json:"application_id"`
	ApplicationName string `json:"application_name"`
	//ProductIDs      string    `json:"product_ids"`
	RequestType string `json:"request_type"`
	//CallbackUrl     string    `json:"callback_url"`
	//WhitelistedIp   string    `json:"whitelisted_ip"`
	UpdatedDate time.Time `json:"updated_date"`
}
type MsgProviderResponse struct {
	ProviderID        uint64          `json:"provider_id"`
	ProviderName      string          `json:"provider_name"`
	ShortName         string          `json:"short_name"`
	Services          string          `json:"services"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys"`
	//ConfigurationKeys interface{} `json:"configuration_keys"`
	Status int `json:"status"`
}

type EditMsgProviderResponse struct {
	ProviderID        uint64          `json:"provider_id"`
	ProviderName      string          `json:"provider_name"`
	Services          string          `json:"services"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys"`
	//ConfigurationKeys interface{} `json:"configuration_keys"`
	Status int `json:"status"`
}
*/

// type MaintainTemplateResponse struct {
// 	TemplateLocalID uint64 `json:"template_local_id"`
// 	ApplicationID   string `json:"application_id"`
// 	TemplateName    string `json:"template_name"`
// 	TemplateFormat  string `json:"template_format"`
// 	SenderID        string `json:"sender_id"`
// 	EntityID        string `json:"entity_id"`
// 	TemplateID           string `json:"template_id"`
// 	Gateway         string `json:"gateway"`
// 	Status          bool   `json:"status"`
// }

// //new office response is a helper function to create a response body for handling office details data

//	func newMaintainTemplate(mtemplate *domain.MaintainTemplate) MaintainTemplateResponse {
//		return MaintainTemplateResponse{
//			ApplicationID:  mtemplate.ApplicationID,
//			TemplateFormat: mtemplate.TemplateFormat,
//			EntityID:       mtemplate.EntityID,
//			TemplateID:          mtemplate.TemplateID,
//			Gateway:        mtemplate.Gateway,
//		}
//	}
/*
func newMsgApplicationResponse(msgappres *domain.MsgApplications) MsgApplicationResponse {
	return MsgApplicationResponse{
		ApplicationID:   msgappres.ApplicationID,
		ApplicationName: msgappres.ApplicationName,

		RequestType: msgappres.RequestType,

		SecretKey:   msgappres.SecretKey,
		CreatedDate: msgappres.CreatedDate,
		UpdatedDate: msgappres.UpdatedDate,
		Status:      msgappres.Status,
	}
}

func newMsgProviderResponse(msgproviderres *domain.MsgProvider) MsgProviderResponse {
	return MsgProviderResponse{
		ProviderID:   msgproviderres.ProviderID,
		ProviderName: msgproviderres.ProviderName,
		ShortName:    msgproviderres.ShortName,

		Services:          msgproviderres.Services,
		ConfigurationKeys: msgproviderres.ConfigurationKeys,
		Status:            msgproviderres.Status,
	}
}
func editMsgProviderResponse(msgproviderres *domain.MsgProvider) EditMsgProviderResponse {
	return EditMsgProviderResponse{
		ProviderID:   msgproviderres.ProviderID,
		ProviderName: msgproviderres.ProviderName,
		Services:     msgproviderres.Services,
		Status:       msgproviderres.Status,
	}
}
func EditMsgApplicationsResponse(msgappres *domain.EditApplication) EditMsgApplicationResponse {
	return EditMsgApplicationResponse{
		ApplicationID:   msgappres.ApplicationID,
		ApplicationName: msgappres.ApplicationName,

		RequestType: msgappres.RequestType,

		UpdatedDate: msgappres.UpdatedDate,
	}
}


// newErrorResponse is a helper function to create an error response body
func newErrorResponse(message string) errorResponse {
	return errorResponse{
		Success: false,
		Message: message,
	}
}


// meta represents metadata for a paginated response
type meta struct {
	Total uint64 `json:"total" example:"100"`
	Limit uint64 `json:"limit" example:"10"`
	Skip  uint64 `json:"skip" example:"0"`
}

// newMeta is a helper function to create metadata for a paginated response
func newMeta(total, limit, skip uint64) meta {
	return meta{
		Total: total,
		Limit: limit,
		Skip:  skip,
	}
}

*/
// MgApplication Response represents a MgApplication response body

type MgApplicationResponse struct {
	ApplicationID   uint64 `json:"APPLICATION_ID"`
	ApplicationName string `json:"APPLICATION_NAME"`
	ProductIDs      string `json:"PRODUCT_IDS"`
	RequestType     string `json:"REQUEST_TYPE"`
	//CallbackUrl     string    `json:"CALLBACK_URL"`
	//WhitelistedIp   string    `json:"WHITELISTED_IP"`
	SecretKey   string    `json:"SECRET_KEY"`
	CreatedDate time.Time `json:"CREATED_DATE"`
	UpdatedDate time.Time `json:"UPDATED_DATE"`
	Status      int       `json:"STATUS"`
}
type EditMgApplicationResponse struct {
	ApplicationID   uint64 `json:"application_id"`
	ApplicationName string `json:"application_name"`
	ProductIDs      string `json:"product_ids"`
	RequestType     string `json:"request_type"`
}

/*
func EditMgApplicationsResponse(pgappres *domain.EditApplication) EditMgApplicationResponse {
	return EditMgApplicationResponse{
		ApplicationID:   pgappres.ApplicationID,
		ApplicationName: pgappres.ApplicationName,

		RequestType: pgappres.RequestType,
	}
}


// errorStatusMap is a map of defined error messages and their corresponding http status codes
// var errorStatusMap = map[error]int{
// 	port.ErrDataNotFound:               http.StatusNotFound,
// 	port.ErrConflictingData:            http.StatusConflict,
// 	port.ErrInvalidCredentials:         http.StatusUnauthorized,
// 	port.ErrUnauthorized:               http.StatusUnauthorized,
// 	port.ErrEmptyAuthorizationHeader:   http.StatusUnauthorized,
// 	port.ErrInvalidAuthorizationHeader: http.StatusUnauthorized,
// 	port.ErrInvalidAuthorizationType:   http.StatusUnauthorized,
// 	port.ErrInvalidToken:               http.StatusUnauthorized,
// 	port.ErrExpiredToken:               http.StatusUnauthorized,
// 	port.ErrForbidden:                  http.StatusForbidden,
// 	port.ErrNoUpdatedData:              http.StatusBadRequest,
// 	port.ErrInsufficientStock:          http.StatusBadRequest,
// 	port.ErrInsufficientPayment:        http.StatusBadRequest,
// }

// validationError sends an error response for some specific request validation error
func validationError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, err)
}
*/
// handleError determines the status code of an error and returns a JSON response with the error message and status code
// func handleError(ctx *gin.Context, err error) {
// 	statusCode, ok := errorStatusMap[err]
// 	if !ok {
// 		statusCode = http.StatusInternalServerError
// 	}

// 	errRsp := newErrorResponse(err.Error())

// 	ctx.JSON(statusCode, errRsp)
// }

// handleAbort sends an error response and aborts the request with the specified status code and error message

// func handleAbort(ctx *gin.Context, err error) {
// 	statusCode, ok := errorStatusMap[err]
// 	if !ok {
// 		statusCode = http.StatusInternalServerError
// 	}

// 	rsp := newErrorResponse(err.Error())
// 	ctx.AbortWithStatusJSON(statusCode, rsp)
// }

// handleSuccess sends a success response with the specified status code and optional data
func handleSuccess(ctx *gin.Context, data any) {
	// rsp := newResponse(true, "Success", data)
	ctx.JSON(http.StatusOK, data)
}

func handleCreateSuccess(ctx *gin.Context, data any) {
	// rsp := newResponse(true, "Success", data)
	ctx.JSON(http.StatusCreated, data)
}

/*
func handleError(ctx *gin.Context, message string) {
	rsp := newResponse(false, message, nil)
	ctx.JSON(http.StatusInternalServerError, rsp)
}
*/

// type SMSReportResponse struct {
// 	SerialNo        uint64  `json:"serialno"`
// 	Date            string  `json:"date"`
// 	Time            string  `json:"time"`
// 	CommunicationID *string `json:"commid"`
// 	ApplicationID   *string `json:"applicationid"`
// 	FacilityID      *string `json:"facilityid"`
// 	MessageType     *int64  `json:"messagetype"`
// 	MessageText     *string `json:"messagetext"`
// 	MobileNumber    *int64  `json:"mobilenumber"`
// 	GatewayID       *string `json:"gatewayid"`
// 	Status          string  `json:"status"`
// }

/*
type SMSReportResponse2 struct {
	SerialNo        uint64  `json:"serial_no"`
	Date            string  `json:"date"`
	Time            string  `json:"time"`
	CommunicationID *string `json:"comm_id"`
	ApplicationID   *string `json:"application_id"`
	FacilityID      *string `json:"facility_id"`
	MessagePriority *int64  `json:"message_priority"`
	MessageText     *string `json:"message_text"`
	MobileNumber    *int64  `json:"mobile_number"`
	GatewayID       *string `json:"gateway_id"`
	Status          string  `json:"status"`
}

func NewSMSReportResponse(sr []domain.SMSReport) []SMSReportResponse2 {
	rsp := make([]SMSReportResponse2, len(sr))
	for i, value := range sr {
		date := value.CreatedDate.Format("02-01-2006")
		time := value.CreatedDate.Format("15:04:05")
		var status string
		if value.Status == "submitted" {
			status = "Success"
		} else {
			status = "Failed"
		}
		rsp[i] = SMSReportResponse2{
			SerialNo:        value.SerialNo,
			Date:            date,
			Time:            time,
			CommunicationID: value.CommunicationID,
			ApplicationID:   value.ApplicationID,
			FacilityID:      value.FacilityID,
			MessagePriority: value.MessagePriority,
			MessageText:     value.MessageText,
			MobileNumber:    value.MobileNumber,
			GatewayID:       value.GatewayID,
			Status:          status,
		}
	}
	return rsp
}

type SMSAggregateReportResponse struct {
	SerialNo        uint64 `json:"serial_no"`
	ApplicationName string `json:"application_name"`
	TemplateName    string `json:"template_name"`
	ProviderName    string `json:"provider_name"`
	Date            string `json:"date"`
	TotalSMS        int64  `json:"total_sms"`
	Success         int64  `json:"success"`
	Failed          int64  `json:"failed"`
	SuccessPercent  string `json:"success_percent"`
	FailurePercent  string `json:"failure_percent"`
}

func NewSMSAggregateReportResponse(sr []domain.SMSAggregateReport, reportType int8) []SMSAggregateReportResponse {
	rsp := make([]SMSAggregateReportResponse, len(sr))
	for i, value := range sr {
		date := value.CreatedDate.Format("02-01-2006")
		Spercent := float64(value.Success) * 100 / float64(value.TotalSMS)
		Fpercent := float64(value.Failed) * 100 / float64(value.TotalSMS)
		rsp[i] = SMSAggregateReportResponse{
			SerialNo:       value.SerialNo,
			Date:           date,
			TotalSMS:       value.TotalSMS,
			Success:        value.Success,
			Failed:         value.Failed,
			SuccessPercent: fmt.Sprintf("%.2f", Spercent),
			FailurePercent: fmt.Sprintf("%.2f", Fpercent),
		}
		switch reportType {
		case 1:
			rsp[i].ApplicationName = value.ApplicationName
		case 2:
			rsp[i].TemplateName = value.TemplateName
		case 3:
			rsp[i].ProviderName = value.ProviderName
		}
	}
	return rsp
}
*/
