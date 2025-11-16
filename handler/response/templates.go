package response

import (
	"MgApplication/core/domain"
	"MgApplication/core/port"
)

type CreateTemplateAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	// Data                 *CreateSMSProviderResponse `json:"data"`
}

type listTemplatesResponse struct {
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
}

func NewListTemplatesResponse(templates []domain.MaintainTemplate) []listTemplatesResponse {
	var response []listTemplatesResponse
	for _, template := range templates {
		templateResponse := listTemplatesResponse{
			TemplateLocalID: template.TemplateLocalID,
			ApplicationID:   template.ApplicationID,
			TemplateName:    template.TemplateName,
			TemplateFormat:  template.TemplateFormat,
			SenderID:        template.SenderID,
			EntityID:        template.EntityID,
			TemplateID:      template.TemplateID,
			Gateway:         template.Gateway,
			MessageType:     template.MessageType,
			Status:          template.Status,
		}
		response = append(response, templateResponse)
	}
	return response
}

type ListTemplatesAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse     `json:",inline"`
	Data                      []listTemplatesResponse `json:"data"`
}

type fetchTemplateResponse struct {
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

func NewFetchTemplateResponse(templates []domain.MaintainTemplate) []fetchTemplateResponse {
	var response []fetchTemplateResponse
	for _, template := range templates {
		templateResponse := fetchTemplateResponse{
			TemplateLocalID: template.TemplateLocalID,
			ApplicationID:   template.ApplicationID,
			TemplateName:    template.TemplateName,
			TemplateFormat: template.TemplateFormat,
			SenderID:        template.SenderID,
			EntityID:        template.EntityID,
			TemplateID:      template.TemplateID,
			Gateway:         template.Gateway,
			MessageType:     template.MessageType,
			Status:          template.Status,
		}
		response = append(response, templateResponse)
	}
	return response
}

type FetchTemplateAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	//MetaDataResponse     `json:",inline"`
	Data []fetchTemplateResponse `json:"data"`
}

type fetchTemplateNameResponse struct {
	TemplateLocalID uint64 `json:"template_local_id" db:"template_local_id"`
	TemplateName    string `json:"template_name" db:"template_name"`
}

func NewFetchTemplateNameResponse(templateNames []domain.GetTemplatebyAPPID) []fetchTemplateNameResponse {
	var response []fetchTemplateNameResponse
	for _, template := range templateNames {
		templateResponse := fetchTemplateNameResponse{
			TemplateLocalID: template.TemplateLocalID,
			TemplateName:    template.TemplateName,
		}
		response = append(response, templateResponse)
	}
	return response
}

type FetchTemplateNameAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	//MetaDataResponse     `json:",inline"`
	Data []fetchTemplateNameResponse `json:"data"`
}

type fetchTemplateDetailsResponse struct {
	TemplateLocalID uint64 `json:"template_local_id" db:"template_local_id"`
	TemplateName    string `json:"template_name" db:"template_name"`
	TemplateFormat  string `json:"template_format" db:"template_format"`
	TemplateID      string `json:"template_id" db:"template_id"`
	EntityID        string `json:"entity_id" db:"entity_id"`
	SenderID        string `json:"sender_id" db:"sender_id"`
	MessageType     string `json:"message_type" db:"message_type"`
}

func NewFetchTemplateDetailsResponse(templateDetails []domain.GetTemplateformatbyID) []fetchTemplateDetailsResponse {
	var response []fetchTemplateDetailsResponse
	for _, template := range templateDetails {
		templateResponse := fetchTemplateDetailsResponse{
			TemplateLocalID: template.TemplateLocalID,
			TemplateName:    template.TemplateName,
			TemplateFormat:  template.TemplateFormat,
			TemplateID:      template.TemplateID,
			EntityID:        template.EntityID,
			SenderID:        template.SenderID,
			MessageType:     template.MessageType,
		}
		response = append(response, templateResponse)
	}
	return response
}

type FetchTemplateDetailsAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	//MetaDataResponse     `json:",inline"`
	Data []fetchTemplateDetailsResponse `json:"data"`
}

type ToggleTemplateStatusAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      interface{} `json:"data"`
}

// func EditTemplateResponse(provider *domain.MsgProvider) *EditTemplateResponse {

// 	response := EditSMSProviderResponse{
// 		ProviderID:        provider.ProviderID,
// 		ProviderName:      provider.ProviderName,
// 		ShortName:         provider.ShortName,
// 		Services:          provider.Services,
// 		ConfigurationKeys: provider.ConfigurationKeys,
// 		Status:            provider.Status,
// 	}
// 	return &response
// }

type UpdateTemplatesAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	//Data                 *EditTemplateResponse `json:"data"`
}
