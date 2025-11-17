package handler

import (
	"MgApplication/core/domain"
	"MgApplication/core/port"
	"MgApplication/handler/response"
	"MgApplication/models"
	repo "MgApplication/repo/postgres"
	"math"

	"encoding/json"

	// _ "time"

	config "MgApplication/api-config"
	apierrors "MgApplication/api-errors"
	log "MgApplication/api-log"

	"github.com/gin-gonic/gin"
)

// MgApplication Handler represents the HTTP handler for MgApplication related requests
type ProviderHandler struct {
	svc *repo.ProviderRepository
	c   *config.Config
}

// MgApplication Handler creates a new MgApplicatPion Handler instance
func NewProviderHandler(svc *repo.ProviderRepository, c *config.Config) *ProviderHandler {
	return &ProviderHandler{
		svc,
		c,
	}
}

type createMessageProviderRequest struct {
	ProviderName      string          `json:"provider_name" validate:"required" example:"Test Provider"`
	ShortName         string          `json:"short_name" validate:"required" example:"TP"`
	Services          string          `json:"services" validate:"required" example:"1,2,3,4"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys" validate:"required" swaggertype:"string" example:"	[{\"keyname\":\"key1\",\"keyvalue\":\"keyvalue1\"}]"`
	//ConfigurationKeys interface{} `json:"configuration_keys" validate:"required"`
	Status bool `json:"status" validate:"required" example:"true"`
}

// CreateMessageProvider godoc
//
//	@Summary		Creates a Message Service Provider
//	@Description	Creates a new Message Service Provider
//	@Tags			Providers
//	@ID				CreateMessageProviderHandler
//	@Accept			json
//	@Produce		json
//	@Param			createMessageProviderRequest	body		createMessageProviderRequest			true	"Create Message Service Provider"
//	@Success		201								{object}	response.CreateSMSProviderAPIResponse	"Message Service Provider is created"
//	@Failure		400								{object}	apierrors.APIErrorResponse				"Bad Request"
//	@Failure		401								{object}	apierrors.APIErrorResponse				"Unauthorized"
//	@Failure		403								{object}	apierrors.APIErrorResponse				"Forbidden"
//	@Failure		404								{object}	apierrors.APIErrorResponse				"Data not found"
//	@Failure		409								{object}	apierrors.APIErrorResponse				"Data conflict errpr"
//	@Failure		422								{object}	apierrors.APIErrorResponse				"Binding or Validation error"
//	@Failure		500								{object}	apierrors.APIErrorResponse				"Internal server error"
//	@Failure		502								{object}	apierrors.APIErrorResponse				"Bad Gateway"
//	@Failure		504								{object}	apierrors.APIErrorResponse				"Gateway Timeout"
//	@Router			/sms-providers [post]
func (ph *ProviderHandler) CreateMessageProviderHandler(ctx *gin.Context) {

	var req models.CreateMessageProviderRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Binding failed for createMessageProviderRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for createMessageProviderRequest: %s", err.Error())
		return
	}

	// if !ph.vs.isValidRequestType(req.Services) {
	// 	apierrors.HandleWithMessage(ctx, "invalid values for the services type")
	// 	log.Error(ctx, "Validation failed for createMessageProviderRequest: %s", "invalid services")
	// 	return
	// }

	var aStatus int
	if req.Status {
		aStatus = 1
	} else {
		aStatus = 0
	}

	msgproviderreq := domain.MsgProvider{
		ProviderName:      req.ProviderName,
		ShortName:         req.ShortName,
		Services:          req.Services,
		ConfigurationKeys: req.ConfigurationKeys,
		Status:            aStatus,
	}

	provider, err := ph.svc.CreateMessageProviderRepo(ctx, &msgproviderreq)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in CreateMsgProviderRepo function: %s", err.Error())
		return
	}

	rsp := response.NewCreateSMSProviderResponse(&provider)
	apiRsp := response.CreateSMSProviderAPIResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data:                 rsp,
	}

	log.Debug(ctx, "CreateMessageProviderHandler response: %v", apiRsp)
	handleCreateSuccess(ctx, apiRsp)
}

type listMessageProviderRequest struct {
	Status bool `form:"status" validate:"omitempty" example:"true"`
	port.MetaDataRequest
}

// ListMessageProviders godoc
//
//	@Summary		Get Message Providers
//	@Description	Lists all message service providers
//	@Tags			Providers
//	@ID				ListMessageProvidersHandler
//	@Produce		json
//	@Param			listMessageProviderRequest	query		listMessageProviderRequest				false	"Get Provider Request (by Query)"
//	@Success		200							{object}	response.ListSMSProvidersAPIResponse	"All Message service providers are retrieved"
//	@Failure		400							{object}	apierrors.APIErrorResponse				"Bad Request"
//	@Failure		401							{object}	apierrors.APIErrorResponse				"Unauthorized"
//	@Failure		403							{object}	apierrors.APIErrorResponse				"Forbidden"
//	@Failure		404							{object}	apierrors.APIErrorResponse				"Data not found"
//	@Failure		409							{object}	apierrors.APIErrorResponse				"Data conflict errpr"
//	@Failure		422							{object}	apierrors.APIErrorResponse				"Binding or Validation error"
//	@Failure		500							{object}	apierrors.APIErrorResponse				"Internal server error"
//	@Failure		502							{object}	apierrors.APIErrorResponse				"Bad Gateway"
//	@Failure		504							{object}	apierrors.APIErrorResponse				"Gateway Timeout"
//	@Router			/sms-providers [get]
func (ph *ProviderHandler) ListMessageProvidersHandler(ctx *gin.Context) {

	var req models.ListMessageProviderRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "Binding failed for listMessageProviderRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for listMessageProviderRequest: %s", err.Error())
		return
	}

	if req.Limit == 0 && req.Skip == 0 {
		req.Limit = math.MaxInt32
	}

	msgproviderreq := domain.ListMessageProviders{
		Status: req.Status,
	}

	providers, err := ph.svc.ListMessageProvidersRepo(ctx, msgproviderreq, req.MetaDataRequest)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in ListProvidersRepo function: %s", err.Error())
		return
	}

	total := len(providers)
	rsp := response.NewListSMSProvidersResponse(providers)
	metadata := port.NewMetaDataResponse(req.Skip, req.Limit, total)

	apiRsp := response.ListSMSProvidersAPIResponse{
		StatusCodeAndMessage: port.ListSuccess,
		MetaDataResponse:     metadata,
		Data:                 rsp,
	}

	log.Debug(ctx, "ListMessageProvidersHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}

type fetchMessageProviderRequest struct {
	ProviderID uint64 `uri:"provider-id" validate:"required,numeric" example:"3"`
}

// FetchMessageProvider godoc
//
//	@Summary		Get Message Service Provider by ProviderID
//	@Description	Fetches Message Service Provider by ProviderID
//	@Tags			Providers
//	@ID				FetchMessageProviderHandler
//	@Accept			json
//	@Produce		json
//	@Param			fetchMessageProviderRequest	path		fetchMessageProviderRequest				true	"Get Provider Request"
//	@Success		200							{object}	response.FetchSMSProviderAPIResponse	"Message Provider is retrieved by ProviderID"
//	@Failure		400							{object}	apierrors.APIErrorResponse				"Bad Request"
//	@Failure		401							{object}	apierrors.APIErrorResponse				"Unauthorized"
//	@Failure		403							{object}	apierrors.APIErrorResponse				"Forbidden"
//	@Failure		404							{object}	apierrors.APIErrorResponse				"Data not found"
//	@Failure		409							{object}	apierrors.APIErrorResponse				"Data conflict errpr"
//	@Failure		422							{object}	apierrors.APIErrorResponse				"Binding or Validation error"
//	@Failure		500							{object}	apierrors.APIErrorResponse				"Internal server error"
//	@Failure		502							{object}	apierrors.APIErrorResponse				"Bad Gateway"
//	@Failure		504							{object}	apierrors.APIErrorResponse				"Gateway Timeout"
//	@Router			/sms-providers/{provider-id} [get]
func (ph *ProviderHandler) FetchMessageProviderHandler(ctx *gin.Context) {

	var req models.FetchMessageProviderRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "Binding failed for fetchMessageProviderRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for fetchMessageProviderRequest: %s", err.Error())
		return
	}

	msgproviderreq := domain.MsgProvider{
		ProviderID: req.ProviderID,
	}

	provider, err := ph.svc.FetchMessageProviderRepo(ctx, &msgproviderreq)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in GetProviderbyIDRepo function: %s", err.Error())
		return
	}

	total := len(provider)
	rsp := response.NewFetchSMSProviderResponse(provider)
	metadata := port.NewMetaDataResponse(0, 0, total)

	apiRsp := response.FetchSMSProviderAPIResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		MetaDataResponse:     metadata,
		Data:                 rsp,
	}

	log.Debug(ctx, "FetchMessageProviderHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}

type updateMessageProviderRequest struct {
	ProviderID        uint64          `uri:"provider-id"  validate:"required,numeric" example:"3" json:"-"`
	ProviderName      string          `json:"provider_name" validate:"required" example:"Test Provider"`
	Services          string          `json:"services" validate:"required,services" example:"1,2,3,4"`
	ConfigurationKeys json.RawMessage `json:"configuration_keys"`
	Status            bool            `json:"status" validate:"required" example:"true"`
}

// UpdateMessageProvider godoc
//
//	@Summary		Edits an existing Message Provider
//	@Description	Allows editing of an existing Message Provider
//	@Tags			Providers
//	@ID				UpdateMessageProviderHandler
//	@Accept			json
//	@Produce		json
//	@Param			provider-id						path		uint64									true	"Edit Message Provider Request"
//	@Param			updateMessageProviderRequest	body		updateMessageProviderRequest			true	"Edit Message Provider Request"
//	@Success		200								{object}	response.UpdateSMSProviderAPIResponse	"Message Provider is modified"
//	@Failure		400								{object}	apierrors.APIErrorResponse				"Bad Request"
//	@Failure		401								{object}	apierrors.APIErrorResponse				"Unauthorized"
//	@Failure		403								{object}	apierrors.APIErrorResponse				"Forbidden"
//	@Failure		404								{object}	apierrors.APIErrorResponse				"Data not found"
//	@Failure		409								{object}	apierrors.APIErrorResponse				"Data conflict errpr"
//	@Failure		422								{object}	apierrors.APIErrorResponse				"Binding or Validation error"
//	@Failure		500								{object}	apierrors.APIErrorResponse				"Internal server error"
//	@Failure		502								{object}	apierrors.APIErrorResponse				"Bad Gateway"
//	@Failure		504								{object}	apierrors.APIErrorResponse				"Gateway Timeout"
//	@Router			/sms-providers/{provider-id} [put]
func (ph *ProviderHandler) UpdateMessageProviderHandler(ctx *gin.Context) {

	var req models.UpdateMessageProviderRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "URI Binding failed for updateMessageProviderRequest: %s", err.Error())
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "JSON Binding failed for updateMessageProviderRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for updateMessageProviderRequest: %s", err.Error())
		return
	}

	var aStatus int
	if req.Status {
		aStatus = 1
	} else {
		aStatus = 0
	}

	msgproviderreq := domain.MsgProvider{
		ProviderID:   req.ProviderID,
		ProviderName: req.ProviderName,
		Services:     req.Services,
		Status:       aStatus,
	}

	provider, err := ph.svc.UpdateMessageProviderRepo(ctx, &msgproviderreq)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in EditMsgProviderRepo function: %s", err.Error())
		return
	}

	rsp := response.NewUpdateSMSProviderResponse(&provider)
	apiRsp := response.UpdateSMSProviderAPIResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data:                 rsp,
	}

	log.Debug(ctx, "UpdateMessageProviderHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}

type toggleMessageProviderStatusRequest struct {
	ProviderID uint64 `uri:"provider-id" validate:"required,numeric" example:"3"`
}

// ToggleMessageProviderStatus godoc
//
//	@Summary		Modifies the status of Message Provider
//	@Description	Toggles the status of Message Provider
//	@Tags			Providers
//	@ID				ToggleMessageProviderStatusHandler
//	@Accept			json
//	@Produce		json
//	@Param			toggleMessageProviderStatusRequest	path		toggleMessageProviderStatusRequest			true	"Message Provider Status Request"
//	@Success		200									{object}	response.ToggleProviderStatusAPIResponse	"Message Provider status is modified"
//	@Failure		400									{object}	apierrors.APIErrorResponse					"Bad Request"
//	@Failure		401									{object}	apierrors.APIErrorResponse					"Unauthorized"
//	@Failure		403									{object}	apierrors.APIErrorResponse					"Forbidden"
//	@Failure		404									{object}	apierrors.APIErrorResponse					"Data not found"
//	@Failure		409									{object}	apierrors.APIErrorResponse					"Data conflict errpr"
//	@Failure		422									{object}	apierrors.APIErrorResponse					"Binding or Validation error"
//	@Failure		500									{object}	apierrors.APIErrorResponse					"Internal server error"
//	@Failure		502									{object}	apierrors.APIErrorResponse					"Bad Gateway"
//	@Failure		504									{object}	apierrors.APIErrorResponse					"Gateway Timeout"
//	@Router			/sms-providers/{provider-id}/status [put]
func (ch *ProviderHandler) ToggleMessageProviderStatusHandler(ctx *gin.Context) {

	var req models.ToggleMessageProviderStatusRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "Binding failed for toggleMessageProviderStatusRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for updateMessageProviderRequest: %s", err.Error())
		return
	}

	msgproviderreq := domain.StatusProvider{
		ProviderID: req.ProviderID,
	}

	rsp, err := ch.svc.ToggleMessageProviderStatusRepo(ctx, &msgproviderreq)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in StatusMsgProviderRepo function: %s", err.Error())
		return
	}

	apiRsp := response.ToggleProviderStatusAPIResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		//MetaDataResponse:     metadata,
		Data: rsp,
	}

	log.Debug(ctx, "ToggleMessageProviderStatusHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}
