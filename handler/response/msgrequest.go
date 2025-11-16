package response

import (
	"MgApplication/core/domain"
	"MgApplication/core/port"
)

type createSMSResponse struct {
	CommunicationID  string `json:"communication_id"`
	CompleteResponse string `json:"complete_response"`
	ReferenceID      string `json:"reference_id"`
	ResponseCode     string `json:"status"`
	ResponseText     string `json:"response_text"`
}

func NewCreateSMSResponse(msg *domain.MsgResponse) *createSMSResponse {
	response := createSMSResponse{
		CommunicationID:  msg.CommunicationID,
		CompleteResponse: msg.CompleteResponse,
		ReferenceID:      msg.ReferenceID,
		ResponseCode:     msg.ResponseCode,
		ResponseText:     msg.ResponseText,
	}
	return &response
}

type CreateSMSAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      *createSMSResponse `json:"data"`
}
type CreateSMSAPIResponseKafka struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      map[string]interface{} `json:"data"`
}
type TestSMSAPIResponse struct {
	//port.StatusCodeAndMessage `json:",inline"`
	Data map[string]interface{} `json:"data"`
}

type FetchCDACSMSDeliveryStatusResponse struct {
	MobileNumber string `json:"mobile_number" validate:"required" example:"919999999999"`
	SMSStatus    string `json:"sms_status" validate:"required" example:"DELIVRD"`
	TimeStamp    string `json:"timestamp" validate:"required" example:"2022-02-25 17:40:50.0435482"`
}

func NewFetchCDACSMSDeliveryStatusResponse(msg []*domain.CDACSMSDeliveryStatusResponse) []*FetchCDACSMSDeliveryStatusResponse {
	var response []*FetchCDACSMSDeliveryStatusResponse
	for _, msg := range msg {
	cdacresponse := &FetchCDACSMSDeliveryStatusResponse{
		MobileNumber: msg.MobileNumber,
		SMSStatus:    msg.SMSStatus,
		TimeStamp:    msg.TimeStamp,
	}
	response = append(response, cdacresponse)}
	return response
}


type FetchCDACSMSDeliveryStatusAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      []*FetchCDACSMSDeliveryStatusResponse `json:"data"`
}