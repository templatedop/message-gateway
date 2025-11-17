package models

import (
	"MgApplication/core/port"
	"encoding/json"
)

// SMS Request models
type CreateSMSRequest struct {
	RequestID     uint64 `json:"reqid"`
	ApplicationID string `json:"application_id" validate:"required"`
	FacilityID    string `json:"facility_id" validate:"required"`
	Priority      int    `json:"priority" validate:"required"`
	MessageText   string `json:"message_text" validate:"required"`
	SenderID      string `json:"sender_id" validate:"required"`
	MobileNumbers string `json:"mobile_numbers" validate:"required"`
	EntityId      string `json:"entity_id"`
	TemplateID    string `json:"template_id" validate:"required"`
	MessageType   string `json:"message_type"`
}

type CreateTestSMSRequest struct {
	MobileNumber string `json:"mobile_number" validate:"required"`
}

// Template models
type CreateTemplateRequest struct {
	TemplateLocalID uint64 `json:"template_local_id"`
	ApplicationID   string `json:"application_id" validate:"required"`
	TemplateName    string `json:"template_name" validate:"required"`
	TemplateFormat  string `json:"template_format" validate:"required"`
	SenderID        string `json:"sender_id" validate:"required"`
	EntityID        string `json:"entity_id"`
	TemplateID      string `json:"template_id" validate:"required"`
	Gateway         string `json:"gateway" validate:"required"`
	Status          bool   `json:"status" validate:"required"`
	MessageType     string `json:"message_type" validate:"required"`
}

type ListTemplatesRequest struct {
	port.MetaDataRequest
}

type ToggleTemplateStatusRequest struct {
	TemplateLocalID uint64 `uri:"template-local-id" validate:"required"`
}

type FetchTemplateRequest struct {
	TemplateLocalID uint64 `uri:"template-local-id" validate:"required"`
}

type UpdateTemplateRequest struct {
	TemplateLocalID uint64 `uri:"template-local-id" validate:"required" json:"-"`
	ApplicationID   string `json:"application_id" validate:"required"`
	TemplateName    string `json:"template_name" validate:"required"`
	TemplateFormat  string `json:"template_format" validate:"required"`
	SenderID        string `json:"sender_id" validate:"required"`
	EntityID        string `json:"entity_id"`
	TemplateID      string `json:"template_id" validate:"required"`
	Gateway         string `json:"gateway" validate:"required"`
	MessageType     string `json:"message_type" validate:"required"`
	Status          bool   `json:"status" validate:"required"`
}

type FetchTemplateByApplicationRequest struct {
	ApplicationID string `form:"application-id" validate:"required"`
}

type FetchTemplateDetailsRequest struct {
	TemplateLocalID uint64 `form:"template-local-id"`
	ApplicationID   string `form:"application-id"`
	Templateformat  string `form:"template-format"`
}

// Provider models
type CreateMessageProviderRequest struct {
	ProviderName      string          `json:"provider_name" validate:"required"`
	ShortName         string          `json:"short_name" validate:"required"`
	Services          string          `json:"services" validate:"required"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys" validate:"required"`
	Status            bool            `json:"status" validate:"required"`
}

type ListMessageProviderRequest struct {
	Status bool `form:"status"`
	port.MetaDataRequest
}

type FetchMessageProviderRequest struct {
	ProviderID uint64 `uri:"provider-id" validate:"required"`
}

type UpdateMessageProviderRequest struct {
	ProviderID        uint64          `uri:"provider-id" validate:"required" json:"-"`
	ProviderName      string          `json:"provider_name" validate:"required"`
	Services          string          `json:"services" validate:"required"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys"`
	Status            bool            `json:"status" validate:"required"`
}

type ToggleMessageProviderStatusRequest struct {
	ProviderID uint64 `uri:"provider-id" validate:"required"`
}

// Report models
type SentSMSStatusReportRequest struct {
	FromDate string `form:"from-date" validate:"required"`
	ToDate   string `form:"to-date" validate:"required"`
	port.MetaDataRequest
}

type AggregateSMSUsageReportRequest struct {
	FromDate   string `form:"from-date" validate:"required"`
	ToDate     string `form:"to-date" validate:"required"`
	ReportType int8   `form:"report-type" validate:"required"`
	port.MetaDataRequest
}

// Bulk SMS models
type InitiateBulkSMSRequest struct {
	File          uint64 `json:"file_id"`
	ReferenceID   string `json:"reference_id"`
	ApplicationID string `json:"application_id" validate:"required"`
	TemplateName  string `json:"template_name" validate:"required"`
	TemplateID    string `json:"template_id"`
	EntityID      string `json:"entity_id"`
	SenderID      string `json:"sender_id"`
	MobileNo      string `json:"mobile_no" validate:"required"`
	TestMessage   string `json:"test_msg" validate:"required"`
	MessageType   string `json:"messsage_type"`
	IsVerified    int    `json:"isverified"`
}

type ValidateTestSMSRequest struct {
	ReferenceID string `form:"reference-id" validate:"required"`
	TestString  string `form:"test-string" validate:"required"`
}

type SendBulkSMSRequest struct {
	SenderID     string `json:"sender_id" validate:"required"`
	MobileNumber string `json:"mobile_number" validate:"required"`
	MessageType  string `json:"message_type" validate:"required"`
	MessageText  string `json:"message_text" validate:"required"`
	TemplateID   string `json:"template_id" validate:"required"`
	EntityID     string `json:"entity_id" validate:"required"`
}

// CDAC SMS Delivery Status
type FetchCDACSMSDeliveryStatusRequest struct {
	ReferenceID string `form:"reference_id" validate:"required"`
}
