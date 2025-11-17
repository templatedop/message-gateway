package handler

import (
	"MgApplication/core/domain"
	"MgApplication/core/port"
	"MgApplication/handler/response"
	repo "MgApplication/repo/postgres"
	"math"
	"time"

	// _ "time"

	config "MgApplication/api-config"
	apierrors "MgApplication/api-errors"
	log "MgApplication/api-log"
	validation "MgApplication/api-validation"

	"github.com/gin-gonic/gin"
)

type ReportsHandler struct {
	svc *repo.ReportsRepository
	c   *config.Config
}

func NewReportsHandler(svc *repo.ReportsRepository, c *config.Config) *ReportsHandler {
	return &ReportsHandler{
		svc,
		c,
	}
}

// SMSDashboard godoc
//
//	@Summary		Get SMS Dashboard data
//	@Description	Fetches SMS Dashboard data
//	@Tags			Reports
//	@ID				SMSDashboardHandler
//	@Produce		json
//	@Success		200	{object}	response.SMSDashboardAPIResponse	"SMS Dashboard data is retrieved"
//	@Failure		400	{object}	apierrors.APIErrorResponse			"Bad Request"
//	@Failure		401	{object}	apierrors.APIErrorResponse			"Unauthorized"
//	@Failure		403	{object}	apierrors.APIErrorResponse			"Forbidden"
//	@Failure		404	{object}	apierrors.APIErrorResponse			"Data not found"
//	@Failure		409	{object}	apierrors.APIErrorResponse			"Data conflict errpr"
//	@Failure		422	{object}	apierrors.APIErrorResponse			"Binding or Validation error"
//	@Failure		500	{object}	apierrors.APIErrorResponse			"Internal server error"
//	@Failure		502	{object}	apierrors.APIErrorResponse			"Bad Gateway"
//	@Failure		504	{object}	apierrors.APIErrorResponse			"Gateway Timeout"
//	@Router			/sms-dashboard [get]
func (ch *ReportsHandler) SMSDashboardHandler(ctx *gin.Context) {

	data, err := ch.svc.SMSDashboardRepo(ctx)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in SMSDashboardRepo function: %s", err.Error())
		return
	}

	rsp := response.NewSMSDashboardResponse(&data)
	apiRsp := response.SMSDashboardAPIResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		//MetaDataResponse:     metadata,
		Data: rsp,
	}

	log.Debug(ctx, "SMSDashboardHandler Response: %v ", apiRsp)
	handleSuccess(ctx, apiRsp)
}

type sentSMSStatusReportRequest struct {
	// +govalid:required
	// +govalid:date=format:dd-mm-yyyy
	FromDate string `form:"from-date" example:"01-01-2008"`
	// +govalid:required
	// +govalid:date=format:dd-mm-yyyy
	ToDate   string `form:"to-date" example:"18-06-2024"`
	port.MetaDataRequest
}

// SentSMSStatusReport godoc
//
//	@Summary		Get all SMS requests
//	@Description	Fetches all SMS requests
//	@Tags			Reports
//	@ID				SentSMSStatusReportHandler
//	@Accept			json
//	@Produce		json
//	@Param			sentSMSStatusReportRequest	query		sentSMSStatusReportRequest				true	"SMS Report Request"
//	@Success		200							{object}	response.SMSSentStatusReportAPIResponse	"All message requests are retrieved"
//	@Failure		400							{object}	apierrors.APIErrorResponse				"Bad Request"
//	@Failure		401							{object}	apierrors.APIErrorResponse				"Unauthorized"
//	@Failure		403							{object}	apierrors.APIErrorResponse				"Forbidden"
//	@Failure		404							{object}	apierrors.APIErrorResponse				"Data not found"
//	@Failure		409							{object}	apierrors.APIErrorResponse				"Data conflict errpr"
//	@Failure		422							{object}	apierrors.APIErrorResponse				"Binding or Validation error"
//	@Failure		500							{object}	apierrors.APIErrorResponse				"Internal server error"
//	@Failure		502							{object}	apierrors.APIErrorResponse				"Bad Gateway"
//	@Failure		504							{object}	apierrors.APIErrorResponse				"Gateway Timeout"
//	@Router			/sms-sent-status-report [get]
func (ch *ReportsHandler) SentSMSStatusReportHandler(ctx *gin.Context) {

	var req sentSMSStatusReportRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "Binding failed for sentSMSStatusReportRequest: %s", err.Error())
		return
	}

	if err := validation.ValidateStruct(req); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for sentSMSStatusReportRequest: %s", err.Error())
		return
	}

	if req.Limit == 0 && req.Skip == 0 {
		req.Limit = math.MaxInt32
	}

	fromDate, _ := time.Parse("02-01-2006", req.FromDate)
	toDate, _ := time.Parse("02-01-2006", req.ToDate)
	if toDate.Before(fromDate) {
		apierrors.HandleWithMessage(ctx, " to_date should be after from_date")
		log.Error(ctx, "to_date should be after from_date")
		return
	}

	smsreport, err := ch.svc.SMSSentStatusReportRepo(ctx, fromDate, toDate, req.MetaDataRequest)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in SMSSentStatusReportRepo function: %s", err.Error())
		return
	}

	total := len(smsreport)
	rsp := response.NewSMSSentStatusReportResponse(smsreport)
	metadata := port.NewMetaDataResponse(req.Skip, req.Limit, total)

	apiRsp := response.SMSSentStatusReportAPIResponse{
		StatusCodeAndMessage: port.ListSuccess,
		MetaDataResponse:     metadata,
		Data:                 rsp,
	}

	log.Debug(ctx, "SentSMSStatusReportHandler Response: %v ", apiRsp)
	handleSuccess(ctx, apiRsp)
}

type aggregateSMSUsageReportRequest struct {
	// +govalid:required
	// +govalid:date=format:dd-mm-yyyy
	FromDate   string `form:"from-date" example:"01-01-2008"`
	// +govalid:required
	// +govalid:date=format:dd-mm-yyyy
	ToDate     string `form:"to-date" example:"18-06-2024"`
	// +govalid:required
	ReportType int8   `form:"report-type" example:"1"`
	port.MetaDataRequest
}

// AggregateSMSUsageReport godoc
//
//	@Summary		Get Aggregate SMS Usage Report
//	@Description	Fetches SMS Aggregate report Applicationwise, Templatewise and Providerwise
//	@Tags			Reports
//	@ID				AggregateSMSUsageReportHandler
//	@Accept			json
//	@Produce		json
//	@Param			aggregateSMSUsageReportRequest	query		aggregateSMSUsageReportRequest			true	"SMS Aggregate Report Request"
//	@Success		200								{object}	response.AggregateSMSReportAPIResponse	"SMS Aggregate report is retrieved"
//	@Failure		400								{object}	apierrors.APIErrorResponse				"Bad Request"
//	@Failure		401								{object}	apierrors.APIErrorResponse				"Unauthorized"
//	@Failure		403								{object}	apierrors.APIErrorResponse				"Forbidden"
//	@Failure		404								{object}	apierrors.APIErrorResponse				"Data not found"
//	@Failure		409								{object}	apierrors.APIErrorResponse				"Data conflict errpr"
//	@Failure		422								{object}	apierrors.APIErrorResponse				"Binding or Validation error"
//	@Failure		500								{object}	apierrors.APIErrorResponse				"Internal server error"
//	@Failure		502								{object}	apierrors.APIErrorResponse				"Bad Gateway"
//	@Failure		504								{object}	apierrors.APIErrorResponse				"Gateway Timeout"
//	@Router			/aggregate-sms-report [get]
func (ch *ReportsHandler) AggregateSMSUsageReportHandler(ctx *gin.Context) {

	var req aggregateSMSUsageReportRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "Binding failed for aggregateSMSUsageReportRequest: %s", err.Error())
		return
	}

	if err := validation.ValidateStruct(req); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for aggregateSMSUsageReportRequest: %s", err.Error())
		return
	}

	if req.Limit == 0 && req.Skip == 0 {
		req.Limit = math.MaxInt32
	}

	fromDate, _ := time.Parse("02-01-2006", req.FromDate)
	toDate, _ := time.Parse("02-01-2006", req.ToDate)
	if toDate.Before(fromDate) {
		apierrors.HandleWithMessage(ctx, "to_date should be after from_date")
		log.Error(ctx, "to_date should be after from_date")
		return
	}

	var smsreport []domain.SMSAggregateReport

	var err error
	switch req.ReportType {
	case 1:
		smsreport, err = ch.svc.AppwiseSMSUsageReportRepo(ctx, fromDate, toDate, req.MetaDataRequest)
		if err != nil {
			apierrors.HandleDBError(ctx, err)
			log.Error(ctx, "Error in AppwiseSMSUsageReportRepo: %s", err.Error())
			return
		}

	case 2:
		smsreport, err = ch.svc.TemplatewiseSMSUsageReportRepo(ctx, fromDate, toDate, req.MetaDataRequest)
		if err != nil {
			apierrors.HandleDBError(ctx, err)
			log.Error(ctx, "Error in TemplatewiseSMSUsageReportRepo: %s", err.Error())
			return
		}

	case 3:
		smsreport, err = ch.svc.ProviderwiseSMSUsageReportRepo(ctx, fromDate, toDate, req.MetaDataRequest)
		if err != nil {
			apierrors.HandleDBError(ctx, err)
			log.Error(ctx, "Error in ProviderwiseSMSUsageReportRepo: %s", err.Error())
			return
		}
	default:
		apierrors.HandleWithMessage(ctx, "Invalid report type. Must be 1, 2 or 3")
		log.Error(ctx, "Invalid report type: %d", req.ReportType)
		return
	}

	total := uint64(len(smsreport))
	rsp := response.NewAggregateSMSReportResponse(smsreport)
	metadata := port.NewMetaDataResponse(req.Skip, req.Limit, int(total))

	apiRsp := response.AggregateSMSReportAPIResponse{
		StatusCodeAndMessage: port.ListSuccess,
		MetaDataResponse:     metadata,
		Data:                 rsp,
	}

	log.Debug(ctx, "AggregateSMSUsageReportHandler Response: %v ", apiRsp)
	handleSuccess(ctx, apiRsp)
}
