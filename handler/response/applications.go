package response

import (
	"MgApplication/core/domain"
	"MgApplication/core/port"
	"time"

	"github.com/volatiletech/null/v9"
)

type CreateMsgApplicationResponse struct {
	ApplicationID   null.Uint64 `json:"application_id" db:"application_id"`
	ApplicationName null.String `json:"application_name" db:"application_name"`
	RequestType     null.String `json:"request_type" db:"request_type"`
	SecretKey       null.String `json:"secret_key" db:"secret_key"`
	CreatedDate     null.Time   `json:"created_date" db:"created_date"`
	UpdatedDate     null.Time   `json:"updated_date" db:"updated_date"`
	Status          int         `json:"status" db:"status_cd"`
}

func NewCreateMsgApplicationResponse(appln *domain.MsgApplications) *CreateMsgApplicationResponse {
	response := CreateMsgApplicationResponse{
		ApplicationID:   null.Uint64From(appln.ApplicationID),
		ApplicationName: null.StringFrom(appln.ApplicationName),
		RequestType:     null.StringFrom(appln.RequestType),
		SecretKey:       null.StringFrom(appln.SecretKey),
		CreatedDate:     null.TimeFrom(appln.CreatedDate),
		UpdatedDate:     null.TimeFrom(appln.UpdatedDate),
		Status:          appln.Status,
	}
	return &response
}

type CreateMsgApplicationAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      *CreateMsgApplicationResponse `json:"data"`
}

type listMsgApplicationsResponse struct {
	ApplicationID   uint64 `json:"application_id" db:"application_id"`
	ApplicationName string `json:"application_name" db:"application_name"`
	RequestType     string `json:"request_type" db:"request_type"`
	Status          int    `json:"status" db:"status_cd"`
}

func NewListMsgApplicationsResponse(applications []domain.MsgApplicationsGet) []listMsgApplicationsResponse {
	var response []listMsgApplicationsResponse
	for _, application := range applications {
		applicationResponse := listMsgApplicationsResponse{
			ApplicationID:   application.ApplicationID,
			ApplicationName: application.ApplicationName,
			RequestType:     application.RequestType,
			Status:          application.Status,
		}
		response = append(response, applicationResponse)

	}
	return response
}

type ListMsgApplicationsAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse     `json:",inline"`
	Data                      []listMsgApplicationsResponse `json:"data"`
}

type fetchMsgApplicationResponse struct {
	ApplicationID   uint64 `json:"application_id" db:"application_id"`
	ApplicationName string `json:"application_name" db:"application_name"`
	RequestType     string `json:"request_type" db:"request_type"`
	Status          int    `json:"status" db:"status_cd"`
}

func NewFetchMsgApplicationResponse(applications []domain.MsgApplicationsGet) []fetchMsgApplicationResponse {
	var response []fetchMsgApplicationResponse
	for _, application := range applications {
		applicationResponse := fetchMsgApplicationResponse{
			ApplicationID:   application.ApplicationID,
			ApplicationName: application.ApplicationName,
			RequestType:     application.RequestType,
			Status:          application.Status,
		}
		response = append(response, applicationResponse)

	}
	return response
}

type FetchMsgApplicationAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	// port.MetaDataResponse     `json:",inline"`
	Data []fetchMsgApplicationResponse `json:"data"`
}

/*
type fetchActiveMsgApplicationResponse struct {
	ApplicationID   uint64 `json:"application_id" db:"application_id"`
	ApplicationName string `json:"application_name" db:"application_name"`
	RequestType     string `json:"request_type" db:"request_type"`
	Status          int    `json:"status" db:"status_cd"`
}

func NewFetchActiveMsgApplicationResponse(applications []domain.MsgApplicationsGet) []fetchActiveMsgApplicationResponse {
	var response []fetchActiveMsgApplicationResponse
	for _, application := range applications {
		applicationResponse := fetchActiveMsgApplicationResponse{
			ApplicationID:   application.ApplicationID,
			ApplicationName: application.ApplicationName,
			RequestType:     application.RequestType,
			Status:          application.Status,
		}
		response = append(response, applicationResponse)

	}
	return response
}

type FetchActiveMsgApplicationAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse     `json:",inline"`
	Data                 []fetchActiveMsgApplicationResponse `json:"data"`
}
*/

type updateMsgApplicationResponse struct {
	ApplicationID   uint64    `json:"application_id" db:"application_id"`
	ApplicationName string    `json:"application_name" db:"application_name"`
	RequestType     string    `json:"request_type" db:"request_type"`
	UpdatedDate     time.Time `json:"updated_date" db:"updated_date"`
	Status          int       `json:"status" db:"status_cd"`
}

func NewUpdateMsgApplicationResponse(appln *domain.EditApplication) *updateMsgApplicationResponse {
	response := updateMsgApplicationResponse{
		ApplicationID:   appln.ApplicationID,
		ApplicationName: appln.ApplicationName,
		RequestType:     appln.RequestType,
		UpdatedDate:     appln.UpdatedDate,
		Status:          appln.Status,
	}
	return &response
}

type UpdateMsgApplicationAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      *updateMsgApplicationResponse `json:"data"`
}

// func FetchApplicationStatus(interface{}) {

// }

type ToggleAppStatusAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      interface{} `json:"data"`
}

/*
type getMsgApplicationResponse struct {
	ApplicationID   uint64 `json:"application_id" db:"application_id"`
	ApplicationName string `json:"application_name" db:"application_name"`
	RequestType     string `json:"request_type" db:"request_type"`
	Status          int    `json:"status" db:"status_cd"`
}

func NewGetMsgApplicationResponse(applications []domain.MsgApplicationsGet) []getMsgApplicationResponse {
	var response []getMsgApplicationResponse
	for _, application := range applications {
		applicationResponse := getMsgApplicationResponse{
			ApplicationID:   application.ApplicationID,
			ApplicationName: application.ApplicationName,
			RequestType:     application.RequestType,
			Status:          application.Status,
		}
		response = append(response, applicationResponse)

	}
	return response
}

type GetMsgApplicationAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse          `json:",inline"`
	Data                      []getMsgApplicationResponse `json:"data"`
}
*/
