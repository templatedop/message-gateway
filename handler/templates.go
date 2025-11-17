package handler

import (
	"MgApplication/core/domain"
	"MgApplication/core/port"
	"MgApplication/handler/response"
	"MgApplication/models"
	repo "MgApplication/repo/postgres"
	"math"

	// _ "time"

	config "MgApplication/api-config"
	apierrors "MgApplication/api-errors"
	log "MgApplication/api-log"

	"github.com/gin-gonic/gin"
)

// MgApplication Handler represents the HTTP handler for MgApplication related requests
type TemplateHandler struct {
	svc *repo.TemplateRepository
	c   *config.Config
}

// MgApplication Handler creates a new MgApplicatPion Handler instance
func NewTemplateHandler(svc *repo.TemplateRepository, c *config.Config) *TemplateHandler {
	return &TemplateHandler{
		svc,
		c,
	}
}

type createTemplateRequest struct {
	TemplateLocalID uint64 `json:"template_local_id"`
	ApplicationID   string `json:"application_id" validate:"required,numeric" example:"4"`
	TemplateName    string `json:"template_name" validate:"required" example:"Test Template"`
	TemplateFormat  string `json:"template_format" validate:"required" example:"Dear {#var#}, Greetings from India Post on the occasion of {#var#} - Indiapost"`
	SenderID        string `json:"sender_id" validate:"required" example:"INPOST"`
	EntityID        string `json:"entity_id" example:"1001051725995192803"`
	TemplateID      string `json:"template_id" validate:"required,numeric" example:"1007188452935484904"`
	Gateway         string `json:"gateway" validate:"required" example:"1"`
	Status          bool   `json:"status" validate:"required" example:"true"`
	MessageType     string `json:"message_type" validate:"required" example:"PM"`
}

// CreateTemplateHandler godoc
//
//	@Summary		Creates a new message template
//	@Description	Creates a new Message template for message applications
//	@Tags			Templates
//	@ID				CreateTemplateHandler
//	@Accept			json
//	@Produce		json
//	@Param			createTemplateRequest	body		createTemplateRequest				true	"Create new Message Template"
//	@Success		201						{object}	response.CreateTemplateAPIResponse	"Message Template is created"
//	@Failure		400						{object}	apierrors.APIErrorResponse			"Bad Request"
//	@Failure		401						{object}	apierrors.APIErrorResponse			"Unauthorized"
//	@Failure		403						{object}	apierrors.APIErrorResponse			"Forbidden"
//	@Failure		404						{object}	apierrors.APIErrorResponse			"Data not found"
//	@Failure		409						{object}	apierrors.APIErrorResponse			"Data conflict errpr"
//	@Failure		422						{object}	apierrors.APIErrorResponse			"Binding or Validation error"
//	@Failure		500						{object}	apierrors.APIErrorResponse			"Internal server error"
//	@Failure		502						{object}	apierrors.APIErrorResponse			"Bad Gateway"
//	@Failure		504						{object}	apierrors.APIErrorResponse			"Gateway Timeout"
//	@Router			/sms-templates [post]
func (ch *TemplateHandler) CreateTemplateHandler(ctx *gin.Context) {

	var req models.CreateTemplateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "Binding failed for createTemplateRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for createTemplateRequest: %s", err.Error())
		return
	}

	var aStatus int
	if req.Status {
		aStatus = 1
	} else {
		aStatus = 0
	}

	maintaintemplate := domain.MaintainTemplate{
		ApplicationID:  req.ApplicationID,
		TemplateName:   req.TemplateName,
		TemplateFormat: req.TemplateFormat,
		SenderID:       req.SenderID,
		EntityID:       req.EntityID,
		TemplateID:     req.TemplateID,
		Gateway:        req.Gateway,
		MessageType:    req.MessageType,
		Status:         aStatus,
	}

	err := ch.svc.CreateTemplateRepo(ctx, &maintaintemplate)
	if err != nil {
		if err.Error() == "given template_id and template already exists, cannot continue" {
			apierrors.HandleDuplicateEntryError(ctx)
			log.Warn(ctx, "given template_id and template already exists, cannot continue")
			return
		} else {
			apierrors.HandleDBError(ctx, err)
			log.Error(ctx, "Error in CreateTemplateRepo function: %s", err.Error())
			return
		}
	}

	apiRsp := response.CreateTemplateAPIResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		// Data:                 rsp,
	}

	log.Debug(ctx, "CreateTemplateHandler response: %v", apiRsp)
	handleCreateSuccess(ctx, apiRsp)
}

type listTemplatesRequest struct {
	port.MetaDataRequest
}

// ListTemplates godoc
//
//	@Summary		Get all Message Templates
//	@Description	Lists all message templates
//	@Tags			Templates
//	@ID				ListTemplatesHandler
//	@Accept			json
//	@Produce		json
//	@Param			listTemplatesRequest	query		listTemplatesRequest				true	"Fetch Templates"
//	@Success		200						{object}	response.ListTemplatesAPIResponse	"All Message Templates are retrieved"
//	@Failure		400						{object}	apierrors.APIErrorResponse			"Bad Request"
//	@Failure		401						{object}	apierrors.APIErrorResponse			"Unauthorized"
//	@Failure		403						{object}	apierrors.APIErrorResponse			"Forbidden"
//	@Failure		404						{object}	apierrors.APIErrorResponse			"Data not found"
//	@Failure		409						{object}	apierrors.APIErrorResponse			"Data conflict errpr"
//	@Failure		422						{object}	apierrors.APIErrorResponse			"Binding or Validation error"
//	@Failure		500						{object}	apierrors.APIErrorResponse			"Internal server error"
//	@Failure		502						{object}	apierrors.APIErrorResponse			"Bad Gateway"
//	@Failure		504						{object}	apierrors.APIErrorResponse			"Gateway Timeout"
//	@Router			/sms-templates [get]
func (ch *TemplateHandler) ListTemplatesHandler(ctx *gin.Context) {
	var req models.ListTemplatesRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "Binding failed for ListTemplatesRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for ListTemplatesRequest: %s", err.Error())
		return
	}

	if req.Limit == 0 && req.Skip == 0 {
		req.Limit = math.MaxInt32
	}

	listTemplate := domain.Meta{
		Skip:  req.Skip,
		Limit: req.Limit,
	}

	templates, totalCount, err := ch.svc.ListTemplatesRepo(ctx, &listTemplate)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in ListTemplatesRepo function: %s", err.Error())
		return
	}

	rsp := response.NewListTemplatesResponse(templates)
	metadata := port.NewMetaDataResponse(req.Skip, req.Limit, int(totalCount))
	apiRsp := response.ListTemplatesAPIResponse{
		StatusCodeAndMessage: port.ListSuccess,
		MetaDataResponse:     metadata,
		Data:                 rsp,
	}

	log.Debug(ctx, "ListTemplatesHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}

type toggleTemplateStatusRequest struct {
	TemplateLocalID uint64 `uri:"template-local-id" validate:"required,numeric" example:"355"`
}

// ToggleTemplateStatus godoc
//
//	@Summary		Modifies the status of Message Template
//	@Description	Modifies the status of Message Template
//	@Tags			Templates
//	@ID				ToggleTemplateStatusHandler
//	@Accept			json
//	@Produce		json
//	@Param			toggleTemplateStatusRequest	path		toggleTemplateStatusRequest					true	"Message Template Status Request"
//	@Success		200							{object}	response.ToggleTemplateStatusAPIResponse	"Message Template status is modified"
//	@Failure		400							{object}	apierrors.APIErrorResponse					"Bad Request"
//	@Failure		401							{object}	apierrors.APIErrorResponse					"Unauthorized"
//	@Failure		403							{object}	apierrors.APIErrorResponse					"Forbidden"
//	@Failure		404							{object}	apierrors.APIErrorResponse					"Data not found"
//	@Failure		409							{object}	apierrors.APIErrorResponse					"Data conflict errpr"
//	@Failure		422							{object}	apierrors.APIErrorResponse					"Binding or Validation error"
//	@Failure		500							{object}	apierrors.APIErrorResponse					"Internal server error"
//	@Failure		502							{object}	apierrors.APIErrorResponse					"Bad Gateway"
//	@Failure		504							{object}	apierrors.APIErrorResponse					"Gateway Timeout"
//	@Router			/sms-templates/{template-local-id}/status [put]
func (ch *TemplateHandler) ToggleTemplateStatusHandler(ctx *gin.Context) {

	var req models.ToggleTemplateStatusRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "Binding failed for toggleTemplateStatusRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for toggleTemplateStatusRequest: %s", err.Error())
		return
	}

	msgtemplatereq := domain.StatusTemplate{
		TemplateLocalID: req.TemplateLocalID,
	}

	rsp, err := ch.svc.ToggleTemplateStatusRepo(ctx, &msgtemplatereq)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in StatusTemplateRepo function: %s", err.Error())
		return
	}

	apiRsp := response.ToggleTemplateStatusAPIResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		//MetaDataResponse:     metadata,
		Data: rsp,
	}

	log.Debug(ctx, "ToggleTemplateStatusHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}

type fetchTemplateRequest struct {
	TemplateLocalID uint64 `uri:"template-local-id" validate:"required" example:"355"`
}

// FetchTemplate godoc
//
//	@Summary		Get Message Template by TemplateLocalID
//	@Description	Fetches Message Template by TemplateLocalID
//	@Tags			Templates
//	@ID				FetchTemplateHandler
//	@Accept			json
//	@Produce		json
//	@Param			fetchTemplateRequest	path		fetchTemplateRequest				true	"Get Message Template Request"
//	@Success		200						{object}	response.FetchTemplateAPIResponse	"Message Template is retrieved by TemplateLocalID"
//	@Failure		400						{object}	apierrors.APIErrorResponse			"Bad Request"
//	@Failure		401						{object}	apierrors.APIErrorResponse			"Unauthorized"
//	@Failure		403						{object}	apierrors.APIErrorResponse			"Forbidden"
//	@Failure		404						{object}	apierrors.APIErrorResponse			"Data not found"
//	@Failure		409						{object}	apierrors.APIErrorResponse			"Data conflict errpr"
//	@Failure		422						{object}	apierrors.APIErrorResponse			"Binding or Validation error"
//	@Failure		500						{object}	apierrors.APIErrorResponse			"Internal server error"
//	@Failure		502						{object}	apierrors.APIErrorResponse			"Bad Gateway"
//	@Failure		504						{object}	apierrors.APIErrorResponse			"Gateway Timeout"
//	@Router			/sms-templates/{template-local-id} [get]
func (ch *TemplateHandler) FetchTemplateHandler(ctx *gin.Context) {

	var req models.FetchTemplateRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "Binding failed for fetchTemplateRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for fetchTemplateRequest: %s", err.Error())
		return
	}

	msgtemplatereq := domain.MaintainTemplate{
		TemplateLocalID: req.TemplateLocalID,
	}

	template, err := ch.svc.FetchTemplateRepo(ctx, &msgtemplatereq)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in GetTemplatebyIDRepo function: %s", err.Error())
		return
	}

	rsp := response.NewFetchTemplateResponse(template)
	apiRsp := response.FetchTemplateAPIResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		//MetaDataResponse:     metadata,
		Data: rsp,
	}

	log.Debug(ctx, "FetchTemplateHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}

type updateTemplateRequest struct {
	TemplateLocalID uint64 `uri:"template-local-id" validate:"required" example:"355" json:"-"`
	ApplicationID   string `json:"application_id" validate:"required" example:"4"`
	TemplateName    string `json:"template_name" validate:"required" example:"Std. Instruction CANCELLATION"`
	TemplateFormat  string `json:"template_format" validate:"required" example:"Standing Instruction {#var#} on Account No {#var#} was cancelled."`
	SenderID        string `json:"sender_id" validate:"required" example:"INPOST"`
	EntityID        string `json:"entity_id"`
	TemplateID      string `json:"template_id" validate:"required" example:"1007002656392643880"`
	Gateway         string `json:"gateway" validate:"required" example:"1"`
	MessageType     string `json:"message_type" validate:"required" example:"PM"`
	Status          bool   `json:"status" validate:"required" example:"true"`
}

// UpdateTemplate godoc
//
//	@Summary		Edits an existing Message Template
//	@Description	Allows editing of an existing Message Template
//	@Tags			Templates
//	@ID				UpdateTemplateHandler
//	@Accept			json
//	@Produce		json
//	@Param			template-local-id		path		uint64								true	"Edit Message Template Request"
//	@Param			updateTemplateRequest	body		updateTemplateRequest				true	"Edit Message Template Request"
//	@Success		200						{object}	response.UpdateTemplatesAPIResponse	"Message Template is modified"
//	@Failure		400						{object}	apierrors.APIErrorResponse			"Bad Request"
//	@Failure		401						{object}	apierrors.APIErrorResponse			"Unauthorized"
//	@Failure		403						{object}	apierrors.APIErrorResponse			"Forbidden"
//	@Failure		404						{object}	apierrors.APIErrorResponse			"Data not found"
//	@Failure		409						{object}	apierrors.APIErrorResponse			"Data conflict errpr"
//	@Failure		422						{object}	apierrors.APIErrorResponse			"Binding or Validation error"
//	@Failure		500						{object}	apierrors.APIErrorResponse			"Internal server error"
//	@Failure		502						{object}	apierrors.APIErrorResponse			"Bad Gateway"
//	@Failure		504						{object}	apierrors.APIErrorResponse			"Gateway Timeout"
//	@Router			/sms-templates/{template-local-id} [put]
func (ch *TemplateHandler) UpdateTemplateHandler(ctx *gin.Context) {

	var req models.UpdateTemplateRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "URI Binding failed for updateTemplateRequest: %s", err.Error())
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "JSON Binding failed for updateTemplateRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for updateTemplateRequest: %s", err.Error())
		return
	}

	var aStatus int
	if req.Status {
		aStatus = 1
	} else {
		aStatus = 0
	}

	msgtemplatereq := domain.MaintainTemplate{
		TemplateLocalID: req.TemplateLocalID,
		ApplicationID:   req.ApplicationID,
		TemplateName:    req.TemplateName,
		TemplateFormat:  req.TemplateFormat,
		SenderID:        req.SenderID,
		EntityID:        req.EntityID,
		TemplateID:      req.TemplateID,
		Gateway:         req.Gateway,
		MessageType:     req.MessageType,
		Status:          aStatus,
	}

	err := ch.svc.UpdateTemplateRepo(ctx, &msgtemplatereq)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in EditTemplateRepo function: %s", err.Error())
		return
	}

	apiRsp := response.UpdateTemplatesAPIResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		//Data:                 rsp,
	}

	log.Debug(ctx, "UpdateTemplateHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}

type fetchTemplateByApplicationRequest struct {
	ApplicationID string `form:"application-id" validate:"required,numeric" example:"4"`
}

// FetchTemplateByApplication godoc
//
//	@Summary		Get Message Template names by ApplicationID
//	@Description	Fetches Message Template names by ApplicationID
//	@Tags			Templates
//	@ID				FetchTemplateByApplicationHandler
//	@Accept			json
//	@Produce		json
//	@Param			fetchTemplateByApplicationRequest	query		fetchTemplateByApplicationRequest		true	"Get Message Template names Request"
//	@Success		200									{object}	response.FetchTemplateNameAPIResponse	"Message Template names are retrieved by ApplicationID"
//	@Failure		400									{object}	apierrors.APIErrorResponse				"Bad Request"
//	@Failure		401									{object}	apierrors.APIErrorResponse				"Unauthorized"
//	@Failure		403									{object}	apierrors.APIErrorResponse				"Forbidden"
//	@Failure		404									{object}	apierrors.APIErrorResponse				"Data not found"
//	@Failure		409									{object}	apierrors.APIErrorResponse				"Data conflict errpr"
//	@Failure		422									{object}	apierrors.APIErrorResponse				"Binding or Validation error"
//	@Failure		500									{object}	apierrors.APIErrorResponse				"Internal server error"
//	@Failure		502									{object}	apierrors.APIErrorResponse				"Bad Gateway"
//	@Failure		504									{object}	apierrors.APIErrorResponse				"Gateway Timeout"
//	@Router			/sms-templates/name [get]
func (ch *TemplateHandler) FetchTemplateByApplicationHandler(ctx *gin.Context) {

	var req models.FetchTemplateByApplicationRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "Binding failed for fetchTemplateByApplicationRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for fetchTemplateByApplicationRequest: %s", err.Error())
		return
	}

	msgtemplatereq := domain.MaintainTemplate{
		ApplicationID: req.ApplicationID,
	}

	template, err := ch.svc.FetchTemplateByApplicationRepo(ctx, &msgtemplatereq)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in GetTemplatenamesbyIDRepo function: %s", err.Error())
		return
	}

	rsp := response.NewFetchTemplateNameResponse(template)
	apiRsp := response.FetchTemplateNameAPIResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		//MetaDataResponse:     metadata,
		Data: rsp,
	}

	log.Debug(ctx, "FetchTemplateByApplicationHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}

type fetchTemplateDetailsRequest struct {
	TemplateLocalID uint64 `form:"template-local-id" example:"355"`
	ApplicationID   string `form:"application-id"  example:"4"`
	Templateformat  string `form:"template-format" example:"Dear {#var#}, your {#var#} is scheduled on {#var#}. Pl check details over {#var#} - INDPOST"`
}

// FetchTemplateDetails godoc
//
//	@Summary		Get Template Details
//	@Description	Fetch template details based on the provided query parameters such as TemplateLocalID, ApplicationID, and Templateformat.
//	@Tags			Templates
//	@ID				FetchTemplateDetailsHandler
//	@Accept			json
//	@Produce		json
//	@Param			fetchTemplateDetailsRequest	query		fetchTemplateDetailsRequest					false	"Query Parameters"
//	@Success		200							{object}	response.FetchTemplateDetailsAPIResponse	"Fetched successfully"
//	@Failure		400							{object}	apierrors.APIErrorResponse					"Bad Request"
//	@Failure		401							{object}	apierrors.APIErrorResponse					"Unauthorized"
//	@Failure		403							{object}	apierrors.APIErrorResponse					"Forbidden"
//	@Failure		404							{object}	apierrors.APIErrorResponse					"Data not found"
//	@Failure		409							{object}	apierrors.APIErrorResponse					"Data conflict errpr"
//	@Failure		422							{object}	apierrors.APIErrorResponse					"Binding or Validation error"
//	@Failure		500							{object}	apierrors.APIErrorResponse					"Internal server error"
//	@Failure		502							{object}	apierrors.APIErrorResponse					"Bad Gateway"
//	@Failure		504							{object}	apierrors.APIErrorResponse					"Gateway Timeout"
//	@Router			/sms-templates/details [get]
func (ch *TemplateHandler) FetchTemplateDetailsHandler(ctx *gin.Context) {

	var req models.FetchTemplateDetailsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "Binding failed for fetchTemplateDetailsRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for fetchTemplateDetailsRequest: %s", err.Error())
		return
	}

	if req.TemplateLocalID == 0 && req.ApplicationID == "" && req.Templateformat == "" {
		apierrors.HandleWithMessage(ctx, "at least one query param must be provided")
		log.Warn(ctx, "at least one query param must be provided")
		return
	}

	msgtemplatereq := domain.MaintainTemplate{
		TemplateLocalID: req.TemplateLocalID,
		ApplicationID:   req.ApplicationID,
		TemplateFormat:  req.Templateformat,
	}

	template, err := ch.svc.FetchTemplateDetailsRepo(ctx, &msgtemplatereq)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in GetTemplateDetailsRepo function: %s", err.Error())
		return
	}

	rsp := response.NewFetchTemplateDetailsResponse(template)
	apiRsp := response.FetchTemplateDetailsAPIResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		//MetaDataResponse:     metadata,
		Data: rsp,
	}

	log.Debug(ctx, "FetchTemplateDetailsHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}
