package handler

import (
	config "MgApplication/api-config"
	apierrors "MgApplication/api-errors"
	log "MgApplication/api-log"
	serverHandler "MgApplication/api-server/handler"
	serverRoute "MgApplication/api-server/route"
	"MgApplication/core/domain"
	"MgApplication/core/port"
	"MgApplication/handler/response"
	repo "MgApplication/repo/postgres"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"reflect"

	"github.com/go-pdf/fpdf"

	"github.com/gin-gonic/gin"
)

// MgApplication Handler represents the HTTP handler for MgApplication related requests
type ApplicationHandler struct {
	*serverHandler.Base
	svc *repo.ApplicationRepository
	c   *config.Config
}

// MgApplication Handler creates a new MgApplicatPion Handler instance
func NewApplicationHandler(svc *repo.ApplicationRepository, c *config.Config) *ApplicationHandler {
	base := serverHandler.New("Applications").SetPrefix("/v1").AddPrefix("/applications")
	return &ApplicationHandler{
		base,
		svc,
		c,
	}
}

func (c *ApplicationHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		//route.GET("/greet", c.greetHandler).Name("Greet Route"),
		//route.GET("/query", c.greetWithQuery).Name("Greet with query"),
		//route.GET("/param/:text", c.greetWithParam).Name("Greet with param"),
		// route.POST("/body", c.greetWithBody).Name("Greet with body"),
		// route.POST("/register", c.register).Name("Register"),

		serverRoute.POST("", c.CreateMessageApplicationHandler).Name("Create Message Application"),
		serverRoute.POST("xml", c.CreateMessageApplicationXMLHandler).Name("Create Message Application XML"),
		serverRoute.GET("", c.ListMessageApplicationsHandler).Name("List all message applications"),
		serverRoute.GET("/:application-id", c.FetchApplicationHandler).Name("Fetch application by id"),
		serverRoute.PUT("/:application-id", c.UpdateMessageApplicationHandler).Name("Fetch application by id"),

		//route.GET("/simulate-error", c.testcustomcode2).Name("Simulate Error"),
	}
}

func (c *ApplicationHandler) Middlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		func(ctx *gin.Context) {
			log.Info(ctx, "Inside ApplicationHandler middleware")
		},
	}
}

// create MgApplication  Request represents a request body for creating a MgApplication Handler
type createMessageApplicationRequest struct {
	ApplicationID   uint64 `json:"application_id"`
	ApplicationName string `json:"application_name" validate:"required" example:"Test Application"`
	RequestType     string `json:"request_type" validate:"required,request_type" example:"1"`
	Status          bool   `json:"status" validate:"required" example:"true"`
}

type createMessageApplicationXMLRequest struct {
	XMLName         xml.Name `xml:"CreateMessageApplicationRequest"`
	ApplicationID   uint64   `xml:"application_id"`
	ApplicationName string   `xml:"application_name" validate:"required" example:"Test Application"`
	RequestType     string   `xml:"request_type" validate:"required,request_type" example:"1"`
	Status          bool     `xml:"status" validate:"required" example:"true"`
}

type createMessageApplicationRequestForm struct {
	ApplicationID   uint64 `form:"application_id"`
	ApplicationName string `form:"application_name" validate:"required" example:"Test Application"`
	RequestType     string `form:"request_type" validate:"required,request_type" example:"1"`
	Status          bool   `form:"status" validate:"required" example:"true"`
	// Single logo file upload (form field name: logo)
	Logo *multipart.FileHeader `form:"logo"`
	// Multiple attachments (repeat field name attachments or use attachments[] depending on client)
	Attachments []*multipart.FileHeader `form:"attachments"`
}
type createMessageApplicationRequestFormTest struct {
	ApplicationName string `form:"application_name" validate:"required" example:"Test Application"`
	RequestType     string `form:"request_type" validate:"required,request_type" example:"1"`
	Status          bool   `form:"status" validate:"required" example:"true"`
}

func (ah *ApplicationHandler) CreateMessageApplicationXMLHandler(sctx *serverRoute.Context, req createMessageApplicationXMLRequest) (*response.CreateMsgApplicationAPIResponse, error) {
	// var req createMessageApplicationRequest
	// if err := ctx.ShouldBindJSON(&req); err != nil {
	// 	apierrors.HandleBindingError(ctx, err)
	// 	log.Error(ctx, "Binding failed for createMessageApplicationRequest: %s", err.Error())
	// 	return
	// }

	// if err := validation.ValidateStruct(req); err != nil {
	// 	apierrors.HandleValidationError(ctx, err)
	// 	log.Error(ctx, "Validation failed for createMessageApplicationRequest: %s", err.Error())
	// 	return
	// }
	fmt.Println("11111111111111111111", req)

	SecretKeyGenerated, errSecret := GenerateRandomString(16)
	if errSecret != nil {
		// apierrors.HandleError(sctx.Ctx, errSecret)
		log.Error(sctx.Ctx, "Error while generating secret key: %s", errSecret.Error())
		return nil, errSecret
	}

	var aStatus int
	if req.Status {
		aStatus = 1
	} else {
		aStatus = 0
	}

	msgappreq := domain.MsgApplications{
		ApplicationName: req.ApplicationName,
		RequestType:     req.RequestType,
		SecretKey:       SecretKeyGenerated,
		Status:          aStatus,
	}

	msg, err := ah.svc.CreateMsgApplicationRepo(sctx.Ctx, &msgappreq)
	if err != nil {
		// apierrors.HandleDBError(sctx.Ctx, err)
		log.Error(sctx.Ctx, "Error in CreateMsgApplicationRepo function: %s", err.Error())
		return nil, err
	}

	rsp := response.NewCreateMsgApplicationResponse(&msg)
	apiRsp := response.CreateMsgApplicationAPIResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data:                 rsp,
	}

	log.Debug(sctx.Ctx, "CreateMessageApplicationHandler response: %v", apiRsp)
	// handleCreateSuccess(sctx.Ctx, apiRsp)

	return &apiRsp, nil
}

func (ah *ApplicationHandler) CreateMessageApplicationHandler(sctx *serverRoute.Context, req createMessageApplicationRequestForm) (*response.CreateMsgApplicationAPIResponse, error) {
	// var req createMessageApplicationRequest
	// if err := ctx.ShouldBindJSON(&req); err != nil {
	// 	apierrors.HandleBindingError(ctx, err)
	// 	log.Error(ctx, "Binding failed for createMessageApplicationRequest: %s", err.Error())
	// 	return
	// }
	// var reqTest createMessageApplicationRequestFormTest
	// if err := validation.ValidateStruct(reqTest); err != nil {
	// 	// apierrors.HandleValidationError(ctx, err)
	// 	// log.Error(ctx, "Validation failed for createMessageApplicationRequest: %s", err.Error())
	// 	return nil, err
	// }

	// var testVal = 0
	// if testVal == 0 {
	// 	// err := apierrors.HandleErrorWithStatusCodeAndMessageNew(apierrors.HTTPErrorConflict, "all lien details are mandatory when lien_status is 'Y'", nil)
	// 	err := apierrors.HandleErrorWithStatusCodeAndMessage(
	// 		apierrors.HTTPErrorConflict,
	// 		"Unable to fetch approver office ID from MDM Office Master",
	// 		errors.New("noRowsAffected"),
	// 	)
	// 	return nil, err
	// }

	// Removed intentional panic that indexed a nil slice
	// fmt.Println(req.)
	if req.Logo != nil {
		f, err := req.Logo.Open()
		if err != nil { /* handle */
			fmt.Println("Error in opening logo file: ", err)
		} else {
			defer f.Close()
			// Read the first 512 bytes for demonstration (or use io.ReadAll for full content)
			buf := make([]byte, 512)
			n, readErr := f.Read(buf)
			if readErr != nil && readErr.Error() != "EOF" {
				fmt.Println("Error reading logo file: ", readErr)
			} else {
				fmt.Println("*******************", buf[:n])
			}
		}
		// io.Copy(dst, f) ...
	}

	fmt.Println("11111111111111111111", req.Logo.Filename, req.Logo.Size)
	fmt.Println("222222222222222222", len(req.Attachments))
	for _, attachment := range req.Attachments {
		fmt.Println("33333333333333333333", attachment.Filename, attachment.Size)
	}

	SecretKeyGenerated, errSecret := GenerateRandomString(16)
	if errSecret != nil {
		// apierrors.HandleError(sctx.Ctx, errSecret)
		log.Error(sctx.Ctx, "Error while generating secret key: %s", errSecret.Error())
		return nil, errSecret
	}

	var aStatus int
	if req.Status {
		aStatus = 1
	} else {
		aStatus = 0
	}

	msgappreq := domain.MsgApplications{
		ApplicationName: req.ApplicationName,
		RequestType:     req.RequestType,
		SecretKey:       SecretKeyGenerated,
		Status:          aStatus,
	}

	msg, err := ah.svc.CreateMsgApplicationRepo(sctx.Ctx, &msgappreq)
	if err != nil {
		// apierrors.HandleDBError(sctx.Ctx, err)
		log.Error(sctx.Ctx, "Error in CreateMsgApplicationRepo function: %s", err.Error())
		return nil, err
	}

	rsp := response.NewCreateMsgApplicationResponse(&msg)
	apiRsp := response.CreateMsgApplicationAPIResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data:                 rsp,
	}

	log.Debug(sctx.Ctx, "CreateMessageApplicationHandler response: %v", apiRsp)
	// handleCreateSuccess(sctx.Ctx, apiRsp)

	return &apiRsp, nil
}

type updateMessageApplicationRequest struct {
	ApplicationID   uint64 `uri:"application-id" validate:"required,numeric" example:"4" json:"-"`
	ApplicationName string `json:"application_name" validate:"required" example:"Test Application"`
	RequestType     string `json:"request_type" validate:"required,request_type" example:"1"`
	Status          bool   `json:"status" validate:"required" example:"true"`
}

// UpdateMessageApplication godoc
//
//	@Summary		Edits an existing Message Application
//	@Description	Allows editing of an existing Message Application
//	@Tags			Applications
//	@ID				UpdateMessageApplicationHandler
//	@Accept			json
//	@Produce		json
//	@Param			application-id					path		uint64										true	"Edit Message Application Request"	SchemaExample(4)
//	@Param			updateMessageApplicationRequest	body		updateMessageApplicationRequest				true	"Edit Message Application Request"
//	@Success		200								{object}	response.UpdateMsgApplicationAPIResponse	"Message Application is modified"
//	@Failure		400								{object}	apierrors.APIErrorResponse					"Bad Request"
//	@Failure		401								{object}	apierrors.APIErrorResponse					"Unauthorized"
//	@Failure		403								{object}	apierrors.APIErrorResponse					"Forbidden"
//	@Failure		404								{object}	apierrors.APIErrorResponse					"Data not found"
//	@Failure		409								{object}	apierrors.APIErrorResponse					"Data conflict errpr"
//	@Failure		422								{object}	apierrors.APIErrorResponse					"Binding or Validation error"
//	@Failure		500								{object}	apierrors.APIErrorResponse					"Internal server error"
//	@Failure		502								{object}	apierrors.APIErrorResponse					"Bad Gateway"
//	@Failure		504								{object}	apierrors.APIErrorResponse					"Gateway Timeout"
//	@Router			/applications/{application-id} [put]
func (ah *ApplicationHandler) UpdateMessageApplicationHandler(sctx *serverRoute.Context, req updateMessageApplicationRequest) (*response.UpdateMsgApplicationAPIResponse, error) {

	// var req updateMessageApplicationRequest
	// if err := ctx.ShouldBindUri(&req); err != nil {
	// 	apierrors.HandleBindingError(ctx, err)
	// 	log.Error(ctx, "URI Binding failed for updateMessageApplicationRequest: %s", err.Error())
	// 	return
	// }
	// if err := ctx.ShouldBindJSON(&req); err != nil {
	// 	apierrors.HandleBindingError(ctx, err)
	// 	log.Error(ctx, "JSON Binding failed for updateMessageApplicationRequest: %s", err.Error())
	// 	return
	// }

	// if err := validation.ValidateStruct(req); err != nil {
	// 	apierrors.HandleValidationError(ctx, err)
	// 	log.Error(ctx, "Validation failed for updateMessageApplicationRequest: %s", err.Error())
	// 	return
	// }

	fmt.Println("*******************", req)

	var aStatus int
	if req.Status {
		aStatus = 1
	} else {
		aStatus = 0
	}

	msgappreq := domain.EditApplication{
		ApplicationID:   req.ApplicationID,
		ApplicationName: req.ApplicationName,
		RequestType:     req.RequestType,
		Status:          aStatus,
	}

	msgApp, err := ah.svc.UpdateMsgApplicationRepo(sctx.Ctx, &msgappreq)
	if err != nil {
		// apierrors.HandleDBError(sctx.Ctx, err)
		log.Error(sctx.Ctx, "Error in EditMsgApplicationRepo function: %s", err.Error())
		return nil, err
	}

	rsp := response.NewUpdateMsgApplicationResponse(&msgApp)
	apiRsp := response.UpdateMsgApplicationAPIResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data:                 rsp,
	}

	log.Debug(sctx.Ctx, "UpdateMessageApplicationHandler response: %v", apiRsp)

	return &apiRsp, nil

	// handleSuccess(ctx, apiRsp)
}

type listMessageApplicationsRequest struct {
	Status bool `form:"status"  example:"true" validate:"omitempty"`
	port.MetaDataRequest
}

// ListMessageApplicationsHandler godoc
//
//	@Summary		Get Message Applications
//	@Description	Lists all message applications
//	@Tags			Applications
//	@ID				ListMessageApplicationsHandler
//	@Produce		json
//	@Param			listMessageApplicationsRequest	query		listMessageApplicationsRequest			false	"Get Applications (by query)"
//	@Success		200								{object}	response.ListMsgApplicationsAPIResponse	"All Message Applications are retrieved"
//	@Failure		400								{object}	apierrors.APIErrorResponse				"Bad Request"
//	@Failure		401								{object}	apierrors.APIErrorResponse				"Unauthorized"
//	@Failure		403								{object}	apierrors.APIErrorResponse				"Forbidden"
//	@Failure		404								{object}	apierrors.APIErrorResponse				"Data not found"
//	@Failure		409								{object}	apierrors.APIErrorResponse				"Data conflict errpr"
//	@Failure		422								{object}	apierrors.APIErrorResponse				"Binding or Validation error"
//	@Failure		500								{object}	apierrors.APIErrorResponse				"Internal server error"
//	@Failure		502								{object}	apierrors.APIErrorResponse				"Bad Gateway"
//	@Failure		504								{object}	apierrors.APIErrorResponse				"Gateway Timeout"
//	@Router			/applications [get]
func (ah *ApplicationHandler) ListMessageApplicationsHandler(sctx *serverRoute.Context, req listMessageApplicationsRequest) (*port.FileResponse, error) {

	// var req listMessageApplicationsRequest

	// if err := ctx.ShouldBindQuery(&req); err != nil {
	// 	apierrors.HandleBindingError(ctx, err)
	// 	log.Error(ctx, "Binding failed for listMessageApplicationsRequest: %s", err.Error())
	// 	return
	// }

	// if err := validation.ValidateStruct(req); err != nil {
	// 	apierrors.HandleValidationError(ctx, err)
	// 	log.Error(ctx, "Validation failed for listMessageApplicationsRequest: %s", err.Error())
	// 	return
	// }

	if req.Limit == 0 && req.Skip == 0 {
		req.Limit = math.MaxInt32
	}

	msgappreq := domain.ListApplications{
		Status: req.Status,
	}

	applications, err := ah.svc.ListApplicationsRepo(sctx.Ctx, msgappreq, req.MetaDataRequest)
	if err != nil {
		// apierrors.HandleDBError(ctx, err)
		log.Error(sctx.Ctx, "Error in ListApplicationsRepo function: %s", err.Error())
		return nil, err
	}

	// total := len(applications)
	// rsp := response.NewListMsgApplicationsResponse(applications)
	// metadata := port.NewMetaDataResponse(req.Skip, req.Limit, total)

	// apiRsp := response.ListMsgApplicationsAPIResponse{
	// 	StatusCodeAndMessage: port.CreateSuccess,
	// 	MetaDataResponse:     metadata,
	// 	Data:                 rsp,
	// }

	// Stream PDF generation via io.Pipe to avoid large memory usage
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		pdf := fpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(0, 10, "Applications List")
		pdf.Ln(12)
		pdf.SetFont("Arial", "", 10)

		pdf.SetFillColor(240, 240, 240)
		pdf.CellFormat(25, 8, "ID", "1", 0, "L", true, 0, "")
		pdf.CellFormat(80, 8, "Name", "1", 0, "L", true, 0, "")
		pdf.CellFormat(35, 8, "RequestType", "1", 0, "L", true, 0, "")
		pdf.CellFormat(25, 8, "Status", "1", 0, "L", true, 0, "")
		pdf.Ln(-1)

		for _, a := range applications {
			var id, name, rtype, status string
			switch v := any(a).(type) {
			case interface {
				GetApplicationID() uint64
				GetApplicationName() string
				GetRequestType() string
				GetStatus() any
			}:
				id = fmt.Sprintf("%d", v.GetApplicationID())
				name = v.GetApplicationName()
				rtype = v.GetRequestType()
				status = fmt.Sprintf("%v", v.GetStatus())
			default:
				id = fmt.Sprintf("%v", getFieldValue(a, "ApplicationID"))
				name = fmt.Sprintf("%v", getFieldValue(a, "ApplicationName"))
				rtype = fmt.Sprintf("%v", getFieldValue(a, "RequestType"))
				status = fmt.Sprintf("%v", getFieldValue(a, "Status"))
			}
			pdf.CellFormat(25, 7, id, "1", 0, "L", false, 0, "")
			pdf.CellFormat(80, 7, name, "1", 0, "L", false, 0, "")
			pdf.CellFormat(35, 7, rtype, "1", 0, "L", false, 0, "")
			pdf.CellFormat(25, 7, status, "1", 0, "L", false, 0, "")
			pdf.Ln(-1)
		}

		if err := pdf.Output(w); err != nil {
			log.Error(sctx.Ctx, "failed to stream PDF: %v", err)
			return
		}
	}()

	fileRes := port.FileResponse{
		ContentType:        "application/octet-stream", // changed from application/pdf per requirement
		ContentDisposition: `attachment; filename="applications.pdf"`,
		Reader:             r,
	}
	return &fileRes, nil
}

// getFieldValue retrieves a named exported field from a struct, else returns empty string
func getFieldValue(item any, field string) any {
	rv := reflect.ValueOf(item)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ""
	}
	fv := rv.FieldByName(field)
	if !fv.IsValid() {
		return ""
	}
	return fv.Interface()
}

type fetchApplicationRequest struct {
	ApplicationID uint64 `uri:"application-id" validate:"required,numeric"  example:"4"`
}

// FetchApplicationHandler godoc
//
//	@Summary		Get Message Application by ApplicationID
//	@Description	Fetches Message Application by ApplicationID
//	@Tags			Applications
//	@ID				FetchApplicationHandler
//	@Accept			json
//	@Produce		json
//	@Param			fetchApplicationRequest	path		fetchApplicationRequest					true	"Get Application Request (example:1)"
//	@Success		200						{object}	response.FetchMsgApplicationAPIResponse	"Message Application is retrieved"
//	@Failure		400						{object}	apierrors.APIErrorResponse				"Bad Request"
//	@Failure		401						{object}	apierrors.APIErrorResponse				"Unauthorized"
//	@Failure		403						{object}	apierrors.APIErrorResponse				"Forbidden"
//	@Failure		404						{object}	apierrors.APIErrorResponse				"Data not found"
//	@Failure		409						{object}	apierrors.APIErrorResponse				"Data conflict errpr"
//	@Failure		422						{object}	apierrors.APIErrorResponse				"Binding or Validation error"
//	@Failure		500						{object}	apierrors.APIErrorResponse				"Internal server error"
//	@Failure		502						{object}	apierrors.APIErrorResponse				"Bad Gateway"
//	@Failure		504						{object}	apierrors.APIErrorResponse				"Gateway Timeout"
//	@Router			/applications/{application-id} [get]
func (ah *ApplicationHandler) FetchApplicationHandler(sctx *serverRoute.Context, req fetchApplicationRequest) (*response.FetchMsgApplicationAPIResponse, error) {

	// var req fetchApplicationRequest
	// if err := ctx.ShouldBindUri(&req); err != nil {
	// 	apierrors.HandleBindingError(ctx, err)
	// 	log.Error(ctx, "Binding failed for fetchApplicationRequest: %s", err.Error())
	// 	return
	// }

	// if err := validation.ValidateStruct(req); err != nil {
	// 	apierrors.HandleValidationError(ctx, err)
	// 	log.Error(ctx, "Validation failed for fetchApplicationRequest: %s", err.Error())
	// 	return
	// }

	msgappreq := domain.MsgApplications{
		ApplicationID: req.ApplicationID,
	}

	applications, err := ah.svc.FetchApplicationRepo(sctx.Ctx, &msgappreq)
	if err != nil {
		// apierrors.HandleDBError(ctx, err)
		log.Error(sctx.Ctx, "Error in GetAppbyIDRepo function: %s", err.Error())
		return nil, err
	}

	// total := len(applications)
	rsp := response.NewFetchMsgApplicationResponse(applications)
	// metadata := response.NewMetaDataResponse(0, 0, total)

	apiRsp := response.FetchMsgApplicationAPIResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		// MetaDataResponse:     metadata,
		Data: rsp,
	}

	log.Debug(sctx.Ctx, "FetchApplicationHandler response: %v", apiRsp)
	// handleSuccess(sctx.Ctx, apiRsp)

	return &apiRsp, nil
}

type toggleApplicationStatusRequest struct {
	ApplicationID uint64 `uri:"application-id" validate:"required,numeric" example:"4"`
}

// ToggleApplicationStatus godoc
//
//	@Summary		Modifies the status of Message Application
//	@Description	Toggles the status of Message Application
//	@Tags			Applications
//	@ID				ToggleApplicationStatusHandler
//	@Accept			json
//	@Produce		json
//	@Param			toggleApplicationStatusRequest	path		toggleApplicationStatusRequest		true	"Application ID (example:5)"
//	@Success		200								{object}	response.ToggleAppStatusAPIResponse	"Message Application status is modified"
//	@Failure		400								{object}	apierrors.APIErrorResponse			"Bad Request"
//	@Failure		401								{object}	apierrors.APIErrorResponse			"Unauthorized"
//	@Failure		403								{object}	apierrors.APIErrorResponse			"Forbidden"
//	@Failure		404								{object}	apierrors.APIErrorResponse			"Data not found"
//	@Failure		409								{object}	apierrors.APIErrorResponse			"Data conflict errpr"
//	@Failure		422								{object}	apierrors.APIErrorResponse			"Binding or Validation error"
//	@Failure		500								{object}	apierrors.APIErrorResponse			"Internal server error"
//	@Failure		502								{object}	apierrors.APIErrorResponse			"Bad Gateway"
//	@Failure		504								{object}	apierrors.APIErrorResponse			"Gateway Timeout"
//	@Router			/applications/{application-id}/status [put]
func (ah *ApplicationHandler) ToggleApplicationStatusHandler(ctx *gin.Context) {

	var req toggleApplicationStatusRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "Binding failed for toggleApplicationStatusRequest: %s", err.Error())
		return
	}

	// if err := validation.ValidateStruct(req); err != nil {
	// 		apierrors.HandleValidationError(ctx, err)
	// 		log.Error(ctx, "Validation failed for toggleApplicationStatusRequest: %s", err.Error())
	// 		return
	// 	}

	msgappreq := domain.StatusApplication{
		ApplicationID: req.ApplicationID,
	}

	applications, err := ah.svc.ToggleApplicationStatusRepo(ctx, &msgappreq)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in StatusMsgApplicationRepo function: %s", err.Error())
		return
	}

	apiRsp := response.ToggleAppStatusAPIResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		//MetaDataResponse:     metadata,
		Data: applications,
	}

	log.Debug(ctx, "ToggleApplicationStatusHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}
