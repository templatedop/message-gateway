package response

import (
	"MgApplication/core/domain"
	"MgApplication/core/port"
	"encoding/json"
)

type createSMSProviderResponse struct {
	ProviderID        uint64          `json:"provider_id" db:"provider_id"`
	ProviderName      string          `json:"provider_name" db:"provider_name"`
	ShortName         string          `json:"short_name" db:"short_name"`
	Services          string          `json:"services" db:"services"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys" db:"configuration_key"`
	//ConfigurationKeys interface{} `json:"configuration_keys" db:"configuration_key"`
	Status int `json:"status" db:"status_cd"`
}

func NewCreateSMSProviderResponse(provider *domain.MsgProvider) *createSMSProviderResponse {

	response := createSMSProviderResponse{
		ProviderID:        provider.ProviderID,
		ProviderName:      provider.ProviderName,
		ShortName:         provider.ShortName,
		Services:          provider.Services,
		ConfigurationKeys: provider.ConfigurationKeys,
		Status:            provider.Status,
	}
	return &response
}

type CreateSMSProviderAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      *createSMSProviderResponse `json:"data"`
}

type listSMSProvidersResponse struct {
	ProviderID        uint64          `json:"provider_id" db:"provider_id"`
	ProviderName      string          `json:"provider_name" db:"provider_name"`
	ShortName         string          `json:"short_name" db:"short_name"`
	Services          string          `json:"services" db:"services"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys" db:"configuration_key"`
	//ConfigurationKeys interface{} `json:"configuration_keys" db:"configuration_key"`
	Status int `json:"status" db:"status_cd"`
}

func NewListSMSProvidersResponse(providers []domain.MsgProvider) []listSMSProvidersResponse {
	var response []listSMSProvidersResponse
	for _, provider := range providers {
		providerResponse := listSMSProvidersResponse{
			ProviderID:        provider.ProviderID,
			ProviderName:      provider.ProviderName,
			ShortName:         provider.ShortName,
			Services:          provider.Services,
			ConfigurationKeys: provider.ConfigurationKeys,
			Status:            provider.Status,
		}
		response = append(response, providerResponse)
	}
	return response
}

type ListSMSProvidersAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse          `json:",inline"`
	Data                      []listSMSProvidersResponse `json:"data"`
}

type fetchSMSProviderResponse struct {
	ProviderID        uint64          `json:"provider_id" db:"provider_id"`
	ProviderName      string          `json:"provider_name" db:"provider_name"`
	ShortName         string          `json:"short_name" db:"short_name"`
	Services          string          `json:"services" db:"services"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys" db:"configuration_key"`
	//ConfigurationKeys interface{} `json:"configuration_keys" db:"configuration_key"`
	Status int `json:"status" db:"status_cd"`
}

func NewFetchSMSProviderResponse(providers []domain.MsgProvider) []fetchSMSProviderResponse {
	var response []fetchSMSProviderResponse
	for _, provider := range providers {
		providerResponse := fetchSMSProviderResponse{
			ProviderID:        provider.ProviderID,
			ProviderName:      provider.ProviderName,
			ShortName:         provider.ShortName,
			Services:          provider.Services,
			ConfigurationKeys: provider.ConfigurationKeys,
			Status:            provider.Status,
		}
		response = append(response, providerResponse)
	}
	return response
}

type FetchSMSProviderAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse          `json:",inline"`
	Data                      []fetchSMSProviderResponse `json:"data"`
}

/*
type fetchActiveSMSProviderResponse struct {
	ProviderID        uint64          `json:"provider_id" db:"provider_id"`
	ProviderName      string          `json:"provider_name" db:"provider_name"`
	ShortName         string          `json:"short_name" db:"short_name"`
	Services          string          `json:"services" db:"services"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys" db:"configuration_key"`
	//ConfigurationKeys interface{} `json:"configuration_keys" db:"configuration_key"`
	Status int `json:"status" db:"status_cd"`
}

func NewFetchActiveSMSProviderResponse(providers []domain.MsgProvider) []fetchActiveSMSProviderResponse {
	var response []fetchActiveSMSProviderResponse
	for _, provider := range providers {
		providerResponse := fetchActiveSMSProviderResponse{
			ProviderID:        provider.ProviderID,
			ProviderName:      provider.ProviderName,
			ShortName:         provider.ShortName,
			Services:          provider.Services,
			ConfigurationKeys: provider.ConfigurationKeys,
			Status:            provider.Status,
		}
		response = append(response, providerResponse)

	}
	return response
}

type FetchActiveSMSProviderAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse          `json:",inline"`
	Data                      []fetchActiveSMSProviderResponse `json:"data"`
}
*/

type updateSMSProviderResponse struct {
	ProviderID        uint64          `json:"provider_id" db:"provider_id"`
	ProviderName      string          `json:"provider_name" db:"provider_name"`
	ShortName         string          `json:"short_name" db:"short_name"`
	Services          string          `json:"services" db:"services"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys" db:"configuration_key"`
	//ConfigurationKeys interface{} `json:"configuration_keys" db:"configuration_key"`
	Status int `json:"status" db:"status_cd"`
}

func NewUpdateSMSProviderResponse(provider *domain.MsgProvider) *updateSMSProviderResponse {

	response := updateSMSProviderResponse{
		ProviderID:        provider.ProviderID,
		ProviderName:      provider.ProviderName,
		ShortName:         provider.ShortName,
		Services:          provider.Services,
		ConfigurationKeys: provider.ConfigurationKeys,
		Status:            provider.Status,
	}
	return &response
}

type UpdateSMSProviderAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      *updateSMSProviderResponse `json:"data"`
}

// func FetchProviderStatus(interface{}) {

// }

type ToggleProviderStatusAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      interface{} `json:"data"`
}

/*
type getSMSProvidersResponse struct {
	ProviderID        uint64          `json:"provider_id" db:"provider_id"`
	ProviderName      string          `json:"provider_name" db:"provider_name"`
	ShortName         string          `json:"short_name" db:"short_name"`
	Services          string          `json:"services" db:"services"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys" db:"configuration_key"`
	//ConfigurationKeys interface{} `json:"configuration_keys" db:"configuration_key"`
	Status int `json:"status" db:"status_cd"`
}

func NewGetSMSProviderResponse(providers []domain.MsgProvider) []getSMSProvidersResponse {
	var response []getSMSProvidersResponse
	for _, provider := range providers {
		providerResponse := getSMSProvidersResponse{
			ProviderID:        provider.ProviderID,
			ProviderName:      provider.ProviderName,
			ShortName:         provider.ShortName,
			Services:          provider.Services,
			ConfigurationKeys: provider.ConfigurationKeys,
			Status:            provider.Status,
		}
		response = append(response, providerResponse)
	}
	return response
}

type GetSMSProvidersAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse     `json:",inline"`
	Data                 []getSMSProvidersResponse `json:"data"`
}
*/