package handler

import (
	"MgApplication/core/domain"
	"MgApplication/core/port"
	"MgApplication/handler/response"
	"MgApplication/models"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	apierrors "MgApplication/api-errors"
	log "MgApplication/api-log"
)

type initiateBulkSMSRequest struct {
	File          uint64 `json:"file_id"`
	ReferenceID   string `json:"reference_id"`
	ApplicationID string `json:"application_id" validate:"required" example:"4"`
	TemplateName  string `json:"template_name" validate:"required" example:"355"`
	TemplateID    string `json:"template_id"`
	EntityID      string `json:"entity_id"`
	SenderID      string `json:"sender_id"`
	MobileNo      string `json:"mobile_no" validate:"required" example:"9000000000"`
	TestMessage   string `json:"test_msg" validate:"required" example:"Your Transaction is Successful. Your Order No is 906627, booked on : 180261"`
	MessageType   string `json:"messsage_type"`
	IsVerified    int    `json:"isverified"`
}

// InitiateBulkSMS godoc
//
//	@Summary		Send Bulk SMSes
//	@Description	Initiates sending of Bulk SMSes
//	@Tags			BulkSMS
//	@ID				InitiateBulkSMSHandler
//	@Accept			json
//	@Produce		json
//	@Param			initiateBulkSMSRequest	body		initiateBulkSMSRequest				true	"Initiate Bulk SMS Request"
//	@Success		201						{object}	response.BulkSMSInitiateAPIResponse	"Submitted successfully"
//	@Failure		400						{object}	apierrors.APIErrorResponse			"Bad Request"
//	@Failure		401						{object}	apierrors.APIErrorResponse			"Unauthorized"
//	@Failure		403						{object}	apierrors.APIErrorResponse			"Forbidden"
//	@Failure		404						{object}	apierrors.APIErrorResponse			"Data not found"
//	@Failure		409						{object}	apierrors.APIErrorResponse			"Data conflict errpr"
//	@Failure		422						{object}	apierrors.APIErrorResponse			"Binding or Validation error"
//	@Failure		500						{object}	apierrors.APIErrorResponse			"Internal server error"
//	@Failure		502						{object}	apierrors.APIErrorResponse			"Bad Gateway"
//	@Failure		504						{object}	apierrors.APIErrorResponse			"Gateway Timeout"
//	@Router			/bulk-sms-initiate [post]
func (ch *MgApplicationHandler) InitiateBulkSMSHandler(ctx *gin.Context) {

	var req models.InitiateBulkSMSRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		apierrors.HandleError(ctx, err)
		log.Error(ctx, "Binding failed for initiateBulkSMSRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for initiateBulkSMSRequest: %s", err.Error())
		return
	}

	initiateSMS := domain.InitiateBulkSMS{
		ApplicationID: req.ApplicationID,
		TemplateName:  req.TemplateName,
		MobileNo:      req.MobileNo,
		TestMessage:   req.TestMessage,
		EntityID:      req.EntityID,
		MessageType:   req.MessageType,
	}
	config, err := ch.svc.InitiateBulkSMSRepo(ctx, &initiateSMS)
	if err != nil {
		apierrors.HandleDBError(ctx, err)
		log.Error(ctx, "Error in InitiateBulkSMSRepo function: %s", err.Error())
		return
	}

	dltsender := strings.Split(config, "_")
	//var TemplateID, SenderID, EntityID, reference_id string
	if len(dltsender) == 5 {
		req.TemplateID = dltsender[0]
		req.SenderID = dltsender[1]
		req.EntityID = dltsender[2]
		req.MessageType = dltsender[3]
		req.ReferenceID = dltsender[4]
	} else {
		req.TemplateID = ""
		req.SenderID = ""
		req.EntityID = ""
		req.MessageType = ""
		req.ReferenceID = ""
		apierrors.HandleWithMessage(ctx, "cannot initiate bulk sms")
		return
	}

	if req.MessageType == "UC" {
		req.TestMessage = UnicodemsgConvertCDAC(req.TestMessage)
	} else {
		req.MessageType = "PM"
	}

	if req.TemplateID != "" && req.SenderID != "" {
		Bulkrsp, err := ch.SendSMSCDAC(SMSParams{
			ch.c.GetString("sms.cdac.username"),
			ch.c.GetString("sms.cdac.password"),
			req.TestMessage, req.SenderID,
			req.MobileNo,
			ch.c.GetString("sms.cdac.securekey"),
			req.TemplateID,
			req.MessageType})
		if err != nil {
			log.Error(ctx, "Error sending SMS using SendSMSCDAC: %s", err.Error())
			// ch.vs.handleError(ctx, err)
			apierrors.HandleError(ctx, err)
			return
		}
		// response := map[string]interface{}{
		// 	"msgresponse":  rsp,
		// 	"reference_id": req.ReferenceID,
		// }

		rsp := response.NewBulkSMSInitiateResponse(Bulkrsp, req.ReferenceID)
		apiRsp := response.BulkSMSInitiateAPIResponse{
			StatusCodeAndMessage: port.CreateSuccess,
			Data:                 rsp,
		}

		handleCreateSuccess(ctx, apiRsp)
	} else {
		// ch.vs.handleError(ctx, errors.New("could not initiate bulk sms"))
		// apperror :=apierrors.AppError("could not initiate bulk sms", apierrors.HTTPErrorBadRequest.StatusCode, "")
		apierrors.HandleWithMessage(ctx, "could not initiate bulk sms")
	}
}

type validateTestSMSRequest struct {
	ReferenceID string `form:"reference-id" validate:"required" example:"lnv579ejt2vmaq03i7up"`
	TestString  string `form:"test-string" validate:"required" example:"906627"`
}

// ValidateTestSMS godoc
//
//	@Summary		Validates Test SMS
//	@Description	Validates Test SMS
//	@Tags			BulkSMS
//	@ID				ValidateTestSMSHandler
//	@Accept			json
//	@Produce		json
//	@Param			validateTestSMSRequest	query		validateTestSMSRequest					true	"Validate Test SMS"
//	@Success		200						{object}	response.ValidateBulkSMSOTPAPIResponse	"True if validation succeeds, false otherwise"
//	@Failure		400						{object}	apierrors.APIErrorResponse				"Bad Request"
//	@Failure		401						{object}	apierrors.APIErrorResponse				"Unauthorized"
//	@Failure		403						{object}	apierrors.APIErrorResponse				"Forbidden"
//	@Failure		404						{object}	apierrors.APIErrorResponse				"Data not found"
//	@Failure		409						{object}	apierrors.APIErrorResponse				"Data conflict errpr"
//	@Failure		422						{object}	apierrors.APIErrorResponse				"Binding or Validation error"
//	@Failure		500						{object}	apierrors.APIErrorResponse				"Internal server error"
//	@Failure		502						{object}	apierrors.APIErrorResponse				"Bad Gateway"
//	@Failure		504						{object}	apierrors.APIErrorResponse				"Gateway Timeout"
//	@Router			/bulk-sms-validate-otp [get]
func (ch *MgApplicationHandler) ValidateTestSMSHandler(ctx *gin.Context) {
	var req models.ValidateTestSMSRequest
	if err := ctx.ShouldBind(&req); err != nil {
		log.Error(ctx, "URI Binding Error in ValidateTestSMS: %s", err.Error())
		// ch.vs.handleError(ctx, err)
		apierrors.HandleError(ctx, err)
		return
	}
	// if !ch.vs.handleValidation(ctx, req) {
	// 	return
	// }

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for validateTestSMSRequest: %s", err.Error())
		return
	}

	verifySMS := domain.ValidateTestSMS{
		ReferenceID: req.ReferenceID,
		TestString:  req.TestString,
	}
	isvalid, err := ch.svc.ValidateTestSMSRepo(ctx, &verifySMS)
	if err != nil {
		if err.Error() == "Invalid test string, please refer to the message sent to the mobile and enter one of the test numbers sent" {
			apierrors.HandleValidationError(ctx, err)
			log.Warn(ctx, "Validation error in ValidateTestSMSRepo: %s", err.Error())
			return
		} else {
			apierrors.HandleDBError(ctx, err)
			log.Error(ctx, "Error in ValidateTestSMSRepo function: %s", err.Error())
			return
		}
	}
	//handleSuccess(ctx, isvalid)

	apiRsp := response.ValidateBulkSMSOTPAPIResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data:                 isvalid,
	}

	log.Debug(ctx, "ValidateTestSMSHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}

type FileInfo struct {
	Path         string
	CreationTime time.Time
}

var fileStore = make(map[string]FileInfo)
var mu sync.RWMutex // Use a read-write mutex

// Structs to represent the XML structure
type Request struct {
	XMLName xml.Name      `xml:"a2wml"`
	Version string        `xml:"version,attr"`
	Request RequestDetail `xml:"request"`
}

type RequestDetail struct {
	Username    string        `xml:"username,attr"`
	Pin         string        `xml:"pin,attr"`
	MessageList []MessageList `xml:"messageList"`
}

// type MessageList struct {
// 	FromAddress   string `xml:"fromAddress"`
// 	DestAddress   string `xml:"destAddress"`
// 	MessageType	  string `xml:"messageType"`
// 	MessageTxt    string `xml:"messageTxt"`
// 	DltTemplateID string `xml:"dlt_template_id"`
// 	DltEntityID   string `xml:"dlt_entity_id"`
// }

type MessageList struct {
	FromAddress   string `xml:"fromAddress" json:"from_address"`
	DestAddress   string `xml:"destAddress" json:"dest_address"`
	MessageType   string `xml:"messageType" json:"message_type"`
	MessageTxt    string `xml:"messageTxt" json:"message_txt"`
	DltTemplateID string `xml:"dlt_template_id" json:"dlt_template_id"`
	DltEntityID   string `xml:"dlt_entity_id" json:"dlt_entity_id"`
}

type NICResponse struct {
	XMLName   xml.Name `xml:"a2wml"`
	Version   string   `xml:"response>version"`
	Timestamp string   `xml:"response>timestamp"`
	RequestID string   `xml:"response>request ID"`
	Code      string   `xml:"response>code"`
	Info      string   `xml:"response>info"`
}

type SendBulkSMSRequestOld struct {
	UniqueID    string `json:"unique_id" validate:"required"` // Unique ID to retrieve the file path
	SenderID    string `json:"sender_id" validate:"required" `
	TemplateID  string `json:"template_id" validate:"required" `
	MessageType string `json:"message_type" validate:"required" `
}

type sendBulkSMSRequest struct {
	SenderID     string `json:"sender_id" validate:"required"`
	MobileNumber string `json:"mobile_number" validate:"required"`
	MessageType  string `json:"message_type" validate:"required"`
	MessageText  string `json:"message_text" validate:"required"`
	TemplateID   string `json:"template_id" validate:"required"`
	EntityID     string `json:"entity_id" validate:"required"`
}

/*
func (ch *MgApplicationHandler) SendBulkSMSOld(gctx *gin.Context) {
	var req SendBulkSMSRequestOld
	if err := gctx.BindJSON(&req); err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		apierrors.HandleBindingError(gctx, err)
		return
	}

	// Retrieve the file path from fileStore using the unique ID
	mu.RLock()
	fileInfo, exists := fileStore[req.UniqueID]
	mu.RUnlock()
	if !exists {
		gctx.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	tempFilePath := fileInfo.Path

	// Ensure the file will be cleaned up even if an error occurs
	defer func() {
		mu.Lock()
		if err := os.Remove(tempFilePath); err != nil {
			log.Error(gctx, "Failed to delete output Excel file: %v", err.Error())
		}
		delete(fileStore, req.UniqueID) // Remove the entry from fileStore
		mu.Unlock()
	}()

	// Read the processed output Excel file
	outputRows, err := ReadExcelFile(tempFilePath)
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read output Excel file"})
		return
	}

	var NICUsername, NICPassword string
	if req.SenderID == "INPOST" {
		// NICUsername = ch.c.NICBulkUsername()
		NICUsername = ch.c.GetString("sms.nic.bulk.username")
		// NICPassword = ch.c.NICBulkPassword()
		NICPassword = ch.c.GetString("sms.nic.bulk.password")
	} else if req.SenderID == "DOPBNK" || req.SenderID == "DOPCBS" {
		// NICUsername = ch.c.NICDOPBNKUserName()
		NICUsername = ch.c.GetString("nic.DOPBNKusername")
		// NICPassword = ch.c.NICDOPBNKPassword()
		NICPassword = ch.c.GetString("nic.DOPBNKpassword")
	} else if req.SenderID == "DOPPLI" {
		// NICUsername = ch.c.NICDOPPLIUserName()
		NICUsername = ch.c.GetString("sms.nic.DOPPLIuserName")
		// NICPassword = ch.c.NICDOPPLIPassword()
		NICPassword = ch.c.GetString("sms.nic.DOPPLIpassword")
	}

	// Create the message list, skipping the first row
	var messageList []MessageList
	for i, row := range outputRows {
		if i == 0 {
			continue // Skip the first row
		}
		if req.MessageType == "UC" {
			row[1] = UnicodemsgConvertNIC(row[1])
		} else {
			req.MessageType = "PM"
		}

		messageList = append(messageList, MessageList{
			FromAddress:   req.SenderID,
			DestAddress:   row[0], // Assuming Mobile number is in the first column
			MessageType:   req.MessageType,
			MessageTxt:    row[1], // Assuming Message is in the second column
			DltTemplateID: req.TemplateID,
			DltEntityID:   ch.c.GetString("sms.dltEntityID"),
		})
	}

	// Create the request
	nicRequest := Request{
		Version: "2.0",
		Request: RequestDetail{
			Username:    NICUsername,
			Pin:         NICPassword,
			MessageList: messageList,
		},
	}

	// Convert the request to XML
	xmlData, err := xml.MarshalIndent(nicRequest, "", "    ")
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert data to XML"})
		return
	}

	// Print the generated XML data for inspection
	fmt.Println("Generated XML:")
	fmt.Println(string(xmlData))

	// Send the XML data to the NIC URL
	// NICBulkURL := ch.c.NICBulkURL()
	NICBulkURL := ch.c.GetString("sms.nic.bulk.url")
	resp, err := http.Post(NICBulkURL, "application/xml", bytes.NewBuffer(xmlData))
	if err != nil || resp.StatusCode != http.StatusOK {
		gctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send data to NIC"})
		return
	}
	defer resp.Body.Close()

	// Read the response body
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		// handleError(ctx, "Failed to read response from NIC")
		apierrors.HandleWithMessage(gctx, "Failed to read response from NIC")
		return
	}

	// Parse the XML response into NICResponse struct
	var nicResponse NICResponse
	if err := xml.Unmarshal(responseData, &nicResponse); err != nil {
		handleError(gctx, "Failed to parse NIC response")
		return
	}

	// Construct the JSON response
	responseJSON := &domain.NicResponse{
		Timestamp: nicResponse.Timestamp,
		RequestID: nicResponse.RequestID,
		Code:      nicResponse.Code,
		Info:      nicResponse.Info,
	}

	// handleSuccess(c, responseJSON)

	rsp := response.NewSendBulkSMSResponseOld(responseJSON)
	apiRsp := response.SendBulkSMSAPIResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data:                 rsp,
	}

	handleSuccess(gctx, apiRsp)
}
*/

// SendBulkSMS godoc.
//
//	@Summary		Send processed Excel file data
//	@Description	This API reads the output Excel file, constructs messages, and sends them to the NIC service in XML format.
//	@Tags			BulkSMS
//	@ID				SendBulkSMSHandler
//	@Accept			json
//	@Produce		json
//	@Param			sendBulkSMSRequest	body		[]sendBulkSMSRequest			true	"Request Body"
//	@Success		200					{object}	response.SendBulkSMSAPIResponse	"Successful operation"
//	@Failure		400					{object}	apierrors.APIErrorResponse		"Bad Request"
//	@Failure		401					{object}	apierrors.APIErrorResponse		"Unauthorized"
//	@Failure		403					{object}	apierrors.APIErrorResponse		"Forbidden"
//	@Failure		404					{object}	apierrors.APIErrorResponse		"Data not found"
//	@Failure		409					{object}	apierrors.APIErrorResponse		"Data conflict errpr"
//	@Failure		422					{object}	apierrors.APIErrorResponse		"Binding or Validation error"
//	@Failure		500					{object}	apierrors.APIErrorResponse		"Internal server error"
//	@Failure		502					{object}	apierrors.APIErrorResponse		"Bad Gateway"
//	@Failure		504					{object}	apierrors.APIErrorResponse		"Gateway Timeout"
//	@Router			/bulk-sms [post]
func (ch *MgApplicationHandler) SendBulkSMSHandler(gctx *gin.Context) {
	var req []models.SendBulkSMSRequest
	if err := gctx.BindJSON(&req); err != nil {
		log.Error(gctx, "Binding failed for sendBulkSMSRequest: %s", err.Error())
		apierrors.HandleBindingError(gctx, err)
		return
	}

	// if !ch.vs.handleValidation(ctx, req) {
	// 	fmt.Println("Validation failed for the request.") // Debugging log
	// 	return
	// }
	for _, sms := range req {
		if err := sms.Validate(); err != nil {
			// fmt.Printf("sms request is %v", sms)
			apierrors.HandleValidationError(gctx, err)
			log.Error(gctx, "Validation failed for sendBulkSMSRequest: %s", err.Error())
			return
		}
	}

	// Ensure req is not empty and
	if len(req) == 0 {
		log.Info(gctx, "Request array is empty")
		gctx.JSON(http.StatusBadRequest, gin.H{"error": "Empty request"})
		return
	}

	//Setting NIC Credentials Based on SenderID
	var NICUsername, NICPassword string
	senderID := req[0].SenderID
	fmt.Println("SenderID:", senderID)

	switch senderID {
	case "INPOST":
		// NICUsername = ch.c.NICBulkUsername()
		NICUsername = ch.c.GetString("sms.nic.bulk.username")
		// NICPassword = ch.c.NICBulkPassword()
		NICPassword = ch.c.GetString("sms.nic.bulk.password")
	case "DOPBNK", "DOPCBS":
		// NICUsername = ch.c.NICDOPBNKUserName()
		NICUsername = ch.c.GetString("nic.DOPBNKusername")
		// NICPassword = ch.c.NICDOPBNKPassword()
		NICPassword = ch.c.GetString("nic.DOPBNKpassword")
	case "DOPPLI":
		// NICUsername = ch.c.NICDOPPLIUserName()
		NICUsername = ch.c.GetString("sms.nic.DOPPLIuserName")
		// NICPassword = ch.c.NICDOPPLIPassword()
		NICPassword = ch.c.GetString("sms.nic.DOPPLIpassword")
	default:
		log.Error(gctx, "Unknown SenderID provided: %s", senderID)
	}

	// Constructing Message List (Skipping First Row)
	var messageList []MessageList
	for _, row := range req {
		messageType := req[0].MessageType // Assuming MessageType is the same for all entries
		if messageType == "UC" {
			row.MessageText = UnicodemsgConvertNIC(row.MessageText)
		} else {
			messageType = "PM"
		}

		messageList = append(messageList, MessageList{
			FromAddress:   senderID,
			DestAddress:   row.MobileNumber, // Assuming MobileNumber is the field for the destination
			MessageType:   messageType,
			MessageTxt:    row.MessageText,   // Assuming MessageText is the field for the message
			DltTemplateID: req[0].TemplateID, // Assuming TemplateID is the same for all entries
			// DltEntityID:   ch.c.DltEntityID(),
			DltEntityID: ch.c.GetString("sms.dltEntityID"),
		})
	}

	// Creating NIC Request and Convert to XML
	nicRequest := Request{
		Version: "2.0",
		Request: RequestDetail{
			Username:    NICUsername,
			Pin:         NICPassword,
			MessageList: messageList,
		},
	}

	xmlData, err := xml.MarshalIndent(nicRequest, "", "    ")
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert data to XML"})
		return
	}
	fmt.Println("Generated XML:", string(xmlData))

	// Sending XML Data to NIC Bulk URL
	// NICBulkURL := ch.c.NICBulkURL()
	NICBulkURL := ch.c.GetString("sms.nic.bulk.url")
	resp, err := http.Post(NICBulkURL, "application/xml", bytes.NewBuffer(xmlData))
	if err != nil {
		log.Error(gctx, "HTTP Post Error: %s", err.Error())
		// ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send data to NIC"})
		apierrors.HandleWithMessage(gctx, "Failed to send data to NIC")
		return
	}
	// fmt.Println("response is", resp)
	defer resp.Body.Close()

	// Read and Parse the NIC Response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		apierrors.HandleWithMessage(gctx, "Failed to read response from NIC")
		return
	}

	var nicResponse domain.NicResponseXml
	if err := xml.Unmarshal(responseData, &nicResponse); err != nil {
		fmt.Println("XML Unmarshal Error:", err)
		apierrors.HandleWithMessage(gctx, "Failed to parse NIC response")
		return
	}
	fmt.Println("Parsed NIC response:", nicResponse)

	// Construct and Send the Final JSON Response
	// responseJSON := domain.NicResponse{
	// 	Timestamp: nicResponse.Timestamp,
	// 	RequestID: nicResponse.RequestID,
	// 	Code:      nicResponse.Code,
	// 	Info:      nicResponse.Info,
	// }

	// // rsp := response.SendBulkResponse(responseJSON)
	// rsp1 := response.SendBulkResponse(&responseJSON)
	// apiRsp1 := response.SendBulkSMSResponse{
	// 	StatusCodeAndMessage: port.CreateSuccess,
	// 	Data:                 rsp1,
	// }

	rsp := response.NewSendBulkSMSResponse(&nicResponse)
	apiRsp := response.SendBulkSMSAPIResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data:                 rsp,
	}

	handleCreateSuccess(gctx, apiRsp)
}

// IsShuttingDown checks if the application is in the process of shutting down
// func IsShuttingDown() bool {
// 	return isShuttingDown.Load().(bool)
// }
