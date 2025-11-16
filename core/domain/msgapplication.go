package domain

import (
	"encoding/json"
	"time"
)

type Meta struct {
	Skip  uint64 `form:"skip" validate:"required,min=0" example:"0"`
	Limit uint64 `form:"limit" validate:"required,min=1" example:"5"`
}

type MsgApplications struct {
	ApplicationID   uint64    `json:"application_id" db:"application_id"`
	ApplicationName string    `json:"application_name" db:"application_name"`
	RequestType     string    `json:"request_type" db:"request_type"`
	SecretKey       string    `json:"secret_key" db:"secret_key"`
	CreatedDate     time.Time `json:"created_date" db:"created_date"`
	UpdatedDate     time.Time `json:"updated_date" db:"updated_date"`
	Status          int       `json:"status" db:"status_cd"`
}

type MsgProvider struct {
	ProviderID        uint64          `json:"provider_id" db:"provider_id"`
	ProviderName      string          `json:"provider_name" db:"provider_name"`
	ShortName         string          `json:"short_name" db:"short_name"`
	Services          string          `json:"services" db:"services"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys" db:"configuration_key"`
	//ConfigurationKeys interface{} `json:"configuration_keys" db:"configuration_key"`
	Status int `json:"status" db:"status_cd"`
}

type MaintainTemplate struct {
	TemplateLocalID uint64 `json:"template_local_id" db:"template_local_id"`
	ApplicationID   string `json:"application_id" db:"application_id"`
	TemplateName    string `json:"template_name" db:"template_name"`
	TemplateFormat  string `json:"template_format" db:"template_format"`
	SenderID        string `json:"sender_id" db:"sender_id"`
	EntityID        string `json:"entity_id" db:"entity_id"`
	TemplateID      string `json:"template_id" db:"template_id"`
	Gateway         string `json:"gateway" db:"gateway"`
	MessageType     string `json:"message_type" db:"message_type"`
	Status          int    `json:"status" db:"status_cd"`
	TotalCount      uint64
}

type InitiateBulkSMS struct {
	File          uint64 `json:"file_id"`
	ReferenceID   string `json:"reference_id" db:"reference_id"`
	ApplicationID string `json:"application_id"`
	TemplateName  string `json:"template_name"`
	TemplateID    string `json:"template_id" db:"template_id"`
	EntityID      string `json:"entity_id" db:"entity_id"`
	SenderID      string `json:"sender_id" db:"sender_id"`
	MessageType   string `json:"message_type" db:"message_type"`
	MobileNo      string `json:"mobile_no"`
	TestMessage   string `json:"test_msg"`
	IsVerified    int    `json:"isverified"`
}
type GetTemplatebyAPPID struct {
	TemplateLocalID uint64 `json:"template_local_id" db:"template_local_id"`
	TemplateName    string `json:"template_name" db:"template_name"`
}
type GetTemplateformatbyID struct {
	TemplateLocalID uint64 `json:"template_local_id" db:"template_local_id"`
	TemplateName    string `json:"template_name" db:"template_name"`
	TemplateFormat  string `json:"template_format" db:"template_format"`
	TemplateID      string `json:"template_id" db:"template_id"`
	EntityID        string `json:"entity_id" db:"entity_id"`
	SenderID        string `json:"sender_id" db:"sender_id"`
	MessageType     string `json:"message_type" db:"message_type"`
}

type TemplateFormat struct {
	TemplateFormat string
}

type GetTemplateIDbyformat struct {
	TemplateID string `json:"template_id" db:"template_id"`
}

type GetSenderIDbyTemplateformat struct {
	SenderID string `json:"sender_id" db:"sender_id"`
}

type GetApplicationDet struct {
	ApplicationID   uint64    `json:"application_id"`
	ApplicationName string    `json:"application_name"`
	RequestType     string    `json:"request_type"`
	SecretKey       string    `json:"secret_key"`
	CreatedDate     time.Time `json:"created_date"`
	UpdatedDate     time.Time `json:"updated_date"`
	Status          int       `json:"status"`
}
type MsgApplicationsGet struct {
	ApplicationID   uint64 `json:"application_id" db:"application_id"`
	ApplicationName string `json:"application_name" db:"application_name"`
	RequestType     string `json:"request_type" db:"request_type"`
	Status          int    `json:"status" db:"status_cd"`
}

type MsgRequest struct {
	RequestID       uint64 `json:"reqid" db:"request_id"`
	ApplicationID   string `json:"application_id" db:"application_id"`
	FacilityID      string `json:"facility_id" db:"facility_id"`
	Priority        int    `json:"priority" db:"priority"`
	MessageText     string `json:"message_text" db:"message_text"`
	SenderID        string `json:"sender_id" db:"sender_id"`
	MobileNumbers   string `json:"mobile_numbers" db:"mobile_number"`
	EntityId        string `json:"entity_id" db:"entity_id"`
	TemplateID      string `json:"template_id" db:"template_id"`
	CommunicationID string `json:"communication_id" db:"communication_id"`
	Gateway         string `json:"gateway" db:"gateway"`
	MessageType     string `json:"message_type" db:"message_type"`
}

type MsgResponse struct {
	CommunicationID  string `json:"communication_id"`
	CompleteResponse string `json:"complete_response"`
	ReferenceID      string `jsong:"reference_id"`
	ResponseCode     string `json:"status"`
	ResponseText     string `json:"response_text"`
}

type CDACSMSDeliveryStatusRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
	MessageID string `json:"message_id"`
	IsPwdEncrypted bool `json:"pwd_encrypted"`
}

type CDACSMSDeliveryStatusResponse struct {
	MobileNumber string `json:"mobile_number"`
	SMSStatus    string `json:"sms_status"`
	TimeStamp    string `json:"timestamp"`
}

type EditApplication struct {
	ApplicationID   uint64    `json:"application_id" db:"application_id"`
	ApplicationName string    `json:"application_name" db:"application_name"`
	RequestType     string    `json:"request_type" db:"request_type"`
	UpdatedDate     time.Time `json:"updated_date" db:"updated_date"`
	Status          int       `json:"status" db:"status_cd"`
}
type StatusApplication struct {
	ApplicationID uint64 `json:"application_id"`
}
type StatusProvider struct {
	ProviderID uint64 `json:"provider_id"`
}
type StatusTemplate struct {
	TemplateLocalID uint64 `json:"template_local_id"`
}

type ValidateTestSMS struct {
	ReferenceID string `json:"reference_id"`
	TestString  string `json:"test_string"`
}

// type SMSReport struct {
// 	SerialNo        uint64    `json:"serialno" db:"serial_number"`
// 	CreatedDate     time.Time `json:"createddate" db:"created_date"`
// 	CommunicationID *string   `json:"commid" db:"communication_id"`
// 	ApplicationID   *string   `json:"applicationid" db:"application_id"`
// 	FacilityID      *string   `json:"facilityid" db:"facility_id"`
// 	MessageType     *int64    `json:"messagetype" db:"priority"`
// 	MessageText     *string   `json:"messagetext" db:"message_text"`
// 	MobileNumber    *int64    `json:"mobilenumber" db:"mobile_number"`
// 	GatewayID       *string   `json:"gatewayid" db:"gateway"`
// 	Status          string    `json:"status" db:"status"`
// }

type SMSReport struct {
	SerialNo        uint64    `json:"serial_no" db:"serial_number"`
	CreatedDate     time.Time `json:"created_date" db:"created_date"`
	CommunicationID *string    `json:"comm_id" db:"communication_id"`
	ApplicationID   *string    `json:"application_id" db:"application_id"`
	FacilityID      *string    `json:"facility_id" db:"facility_id"`
	MessagePriority *int64     `json:"message_priority" db:"priority"`
	MessageText     *string    `json:"message_text" db:"message_text"`
	MobileNumber    *int64     `json:"mobile_number" db:"mobile_number"`
	GatewayID       *string    `json:"gateway_id" db:"gateway"`
	Status          string    `json:"status" db:"status"`
}

type SMSAggregateReport struct {
	SerialNo        uint64    `json:"serial_no" db:"serial_number"`
	ApplicationName string    `json:"application_name" db:"application_name"`
	TemplateName    string    `json:"template_name" db:"template_name"`
	ProviderName    string    `json:"provider_name" db:"provider_name"`
	CreatedDate     time.Time `json:"created_date" db:"created_date"`
	TotalSMS        int64     `json:"total_sms" db:"total_sms"`
	Success         int64     `json:"success" db:"success"`
	Failed          int64     `json:"failed" db:"failed"`
}

type SMSDashboard struct {
	TotalSMSSent       int64 `json:"total_sms_sent" db:"total_sms_sent"`
	TotalOTPMessages   int64 `json:"total_otp_messages" db:"total_otps"`
	TotalTransMessages int64 `json:"total_trans_messages" db:"total_transactions"`
	TotalBulkMessages  int64 `json:"total_bulk_messages" db:"total_bulk_sms"`
	TotalPromMessages  int64 `json:"total_prom_messages" db:"total_promotional_sms"`
	TotalTemplates     int64 `json:"total_templates" db:"total_templates"`
	TotalProviders     int64 `json:"total_providers" db:"total_providers"`
	TotalApplications  int64 `json:"total_applications" db:"total_applications"`
}

type Counter struct {
	Count int `json:"count" db:"count"`
}

type CurrentStatus struct {
	Status int `json:"status" db:"status_cd"`
}

type TransformedData struct {
	MobileNumber string `json:"mobile_number"`
	Message      string `json:"message"`
}

type NicResponse struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id"`
	Code      string `json:"code"`
	Info      string `json:"info"`
}
type NicResponseXml struct {
	Timestamp string `xml:"timestamp"`
	RequestID string `xml:"request_id"`
	Code      string `xml:"code"`
	Info      string `xml:"info"`
}

type ListApplications struct {
	ApplicationID   uint64    `json:"application_id" db:"application_id"`
	ApplicationName string    `json:"application_name" db:"application_name"`
	RequestType     string    `json:"request_type" db:"request_type"`
	SecretKey       string    `json:"secret_key" db:"secret_key"`
	CreatedDate     time.Time `json:"created_date" db:"created_date"`
	UpdatedDate     time.Time `json:"updated_date" db:"updated_date"`
	Status          bool      `json:"status" db:"status_cd"`
}

type ListMessageProviders struct {
	ProviderID        uint64          `json:"provider_id" db:"provider_id"`
	ProviderName      string          `json:"provider_name" db:"provider_name"`
	ShortName         string          `json:"short_name" db:"short_name"`
	Services          string          `json:"services" db:"services"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys" db:"configuration_key"`
	//ConfigurationKeys interface{} `json:"configuration_keys" db:"configuration_key"`
	Status bool `json:"status" db:"status_cd"`
}
