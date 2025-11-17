// Package handler provides HTTP handlers for managing message requests in the MgApplication.
// It includes functions for creating SMS requests, converting messages to Unicode, and sending messages via different gateways.
//
// This package depends on several other packages for its functionality:
// - "MgApplication/core/domain": Contains domain models for the application.
// - "MgApplication/core/port": Defines ports for interacting with external systems.
// - "MgApplication/handler/response": Handles responses for the handlers.
// - "MgApplication/repo/postgres": Provides repository implementations for PostgreSQL.
// - "github.com/gin-gonic/gin": A web framework for building HTTP servers.
// - "MgApplication/api-config": Configuration management.
// - "MgApplication/api-errors": Error handling.
// - "MgApplication/api-log": Logging.
// - "MgApplication/api-validation": Validation utilities.
//
// The main handler in this package is MgApplicationHandler, which provides methods for creating and managing SMS requests.

package handler

import (
	"MgApplication/core/domain"
	"MgApplication/core/port"
	"MgApplication/handler/response"
	"MgApplication/models"
	repo "MgApplication/repo/postgres"
	"bytes"
	"context"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"crypto/rand"
	"crypto/sha1"
	"crypto/sha512"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	// _ "time"

	config "MgApplication/api-config"
	apierrors "MgApplication/api-errors"
	log "MgApplication/api-log"

	"github.com/gin-gonic/gin"
)

// MgApplication Handler represents the HTTP handler for MgApplication related requests
type MgApplicationHandler struct {
	svc *repo.MgApplicationRepository
	c   *config.Config
}

// MgApplication Handler creates a new MgApplicatPion Handler instance
func NewMgApplicationHandler(svc *repo.MgApplicationRepository, c *config.Config) *MgApplicationHandler {
	return &MgApplicationHandler{
		svc,
		c,
	}
}

// HTML numeric character references
func UnicodemsgConvertCDAC(message string) string {
	var UnicodeMessage strings.Builder
	for _, char := range message {
		UnicodeMessage.WriteString(fmt.Sprintf("&#%d;", char))
	}
	return UnicodeMessage.String()
}

// Hexadecimal Unicode Code Points
func UnicodemsgConvertNIC(message string) string {
	var UnicodeMessage strings.Builder
	for _, char := range message {
		UnicodeMessage.WriteString(fmt.Sprintf("%04X", char))
	}
	return UnicodeMessage.String()
}

// Use models.CreateSMSRequest instead

// CreateMessageRequest godoc
//
//	@Summary		Creates a message request
//	@Description	Creates message requests for application for registered templates
//	@Tags			SMS Request
//	@ID				CreateSMSRequestHandler
//	@Accept			json
//	@Produce		json
//	@Param			createSMSRequest	body		createSMSRequest				true	"Creates Message request"
//	@Success		201					{object}	response.CreateSMSAPIResponse	"Success"
//	@Failure		400					{object}	apierrors.APIErrorResponse		"Bad Request"
//	@Failure		401					{object}	apierrors.APIErrorResponse		"Unauthorized"
//	@Failure		403					{object}	apierrors.APIErrorResponse		"Forbidden"
//	@Failure		404					{object}	apierrors.APIErrorResponse		"Data not found"
//	@Failure		409					{object}	apierrors.APIErrorResponse		"Data conflict errpr"
//	@Failure		422					{object}	apierrors.APIErrorResponse		"Binding or Validation error"
//	@Failure		500					{object}	apierrors.APIErrorResponse		"Internal server error"
//	@Failure		502					{object}	apierrors.APIErrorResponse		"Bad Gateway"
//	@Failure		504					{object}	apierrors.APIErrorResponse		"Gateway Timeout"
//	@Router			/sms-request [post]
func (ch *MgApplicationHandler) CreateSMSRequestHandler(ctx *gin.Context) {
	log.Debug(ctx, "Inside CreateSMSRequestHandler function")
	var req models.CreateSMSRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error(ctx, "Binding failed for CreateSMSRequestHandler: %s", err.Error())
		apierrors.HandleBindingError(ctx, err)
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for CreateSMSRequestHandler: %s", err.Error())
		return
	}

	msgreq := domain.MsgRequest{
		FacilityID:    req.FacilityID,
		ApplicationID: req.ApplicationID,
		Priority:      req.Priority,
		// MessageText:   NormalizeAndClean2(req.MessageText),
		MessageText:   req.MessageText,
		SenderID:      req.SenderID,
		MobileNumbers: req.MobileNumbers,
		EntityId:      req.EntityId,
		TemplateID:    req.TemplateID,
		MessageType:   req.MessageType,
	}

	//Fetch Entity ID from config, if not assigned
	msgreq.EntityId = ch.c.GetString("sms.dltEntityID")
	// log.Debug(ctx, "Entity ID is : %s", msgreq.EntityId)
	gctx := context.Background()

	//**********************************************************************************
	//added by phani for sending msg to kafka topic if Priority is not 1(Other than OTP)
	//**********************************************************************************
	if msgreq.Priority != 1 && msgreq.Priority != 2 {

		log.Debug(ctx, "Pushing Data to Kafka : %s", msgreq)
		resp, err := ch.svc.SendMsgToKafka(&gctx, ch.c.GetString("sms.kafka.url"), ch.c.GetString("sms.kafka.schema"), &msgreq)
		if err != nil {
			log.Error(ctx, "Error in Pushing Message to Kafka: %s", err.Error())
			apierrors.HandleDBError(ctx, err)
			return
		}
		log.Debug(ctx, "Push Data to Kafka : %s", msgreq)
		log.Debug(ctx, "Response from Kafka is : %s", resp)
		apiRsp := response.CreateSMSAPIResponseKafka{
			StatusCodeAndMessage: port.CreateSuccess,
			Data:                 resp,
		}
		handleCreateSuccess(ctx, apiRsp)
		return
	}
	//**********************************************************************************
	//End- added by phani for sending msg to kafka topic if Priority is not 1(Other than OTP)
	//**********************************************************************************

	var gateway string
	msgStoreRequest := ch.c.GetInt("sms.msgstorerequest")
	// log.Debug(ctx, "Message Store Request ID is : %d", msgStoreRequest)
	if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
		//priorites are 1-OTP, 2-Transactional, 3-Promotional, 4-Bulk. If store is true or for Promotional and Bulk info will be saved.
		savedresponse, err := ch.svc.SaveMsgRequestTx(&gctx, &msgreq)
		if err != nil {
			log.Error(ctx, "DB Error in SaveMsgRequestTx: %s", err.Error())
			apierrors.HandleDBError(ctx, err)
			return
		}
		gateway = savedresponse.Gateway
	} else {
		savedresponse, err := ch.svc.GetGateway(&gctx, &msgreq)
		if err != nil {
			log.Error(ctx, "DB Error in GetGateway: %s", err.Error())
			apierrors.HandleDBError(ctx, err)
			return
		}
		gateway = savedresponse.Gateway

	}
	// log.Debug(ctx, "Gateway is : %s", gateway)

	//UC - Unicode message ; PM - Plaintext message
	if msgreq.MessageType == "UC" {
		if msgreq.Gateway == "1" {
			msgreq.MessageText = UnicodemsgConvertCDAC(msgreq.MessageText)
		} else {
			msgreq.MessageText = UnicodemsgConvertNIC(msgreq.MessageText)
		}
	} else {
		msgreq.MessageType = "PM"
	}
	// log.Debug(ctx, "Message Type is : %s", msgreq.MessageType)

	if msgreq.Priority == 1 || msgreq.Priority == 2 {
		if gateway == "1" {
			rsp, err := ch.SendSMSCDAC(SMSParams{
				Username:     ch.c.GetString("sms.cdac.username"),
				Password:     ch.c.GetString("sms.cdac.password"),
				Message:      msgreq.MessageText,
				SenderID:     msgreq.SenderID,
				MobileNumber: msgreq.MobileNumbers,
				SecureKey:    ch.c.GetString("sms.cdac.securekey"),
				TemplateID:   msgreq.TemplateID,
				MessageType:  msgreq.MessageType,
			})
			if err != nil {
				msgresponse := domain.MsgResponse{
					CommunicationID:  msgreq.CommunicationID,
					CompleteResponse: rsp,
					ResponseCode:     "02",
					ResponseText:     err.Error(),
					ReferenceID:      "",
				}
				_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
				apierrors.HandleError(ctx, err)
				return
			}
			log.Debug(ctx, "Response from SendSMSCDAC is : %s", rsp)

			SMSResponse := rsp[:5]

			if SMSResponse == "Error" {
				pattern := `Error (\d+) : (.+)`
				re := regexp.MustCompile(pattern)
				matches := re.FindStringSubmatch(rsp)
				if len(matches) < 3 {
					msgStoreRequest := ch.c.GetInt("sms.msgstorerequest")
					if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
						msgresponse := domain.MsgResponse{
							CommunicationID:  msgreq.CommunicationID,
							CompleteResponse: rsp,
							ResponseCode:     "400",
							ResponseText:     "Invalid Response",
							ReferenceID:      "",
						}
						_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
						apierrors.HandleWithMessage(ctx, "Invalid Response")
						return
					}

				} else {
					errorNumber := matches[1]
					errorMessage := matches[2]
					customError := CustomError{Message: "401, " + errorMessage}
					msgStoreRequest := ch.c.GetInt("sms.msgstorerequest")
					if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
						msgresponse := domain.MsgResponse{
							CommunicationID:  msgreq.CommunicationID,
							CompleteResponse: rsp,
							ResponseCode:     errorNumber,
							ResponseText:     errorMessage,
							ReferenceID:      "",
						}
						_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
					}
					apierrors.HandleError(ctx, customError)
					return
				}
			} else {

				pattern := `^(\d{3}),MsgID = (\d+)`
				re := regexp.MustCompile(pattern)
				matches := re.FindStringSubmatch(rsp)
				if len(matches) >= 3 {
					responseCode := matches[1]
					referenceID := matches[2]
					msgStoreRequest := ch.c.GetInt("sms.msgstorerequest")
					if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
						msgresponse := domain.MsgResponse{
							CommunicationID:  msgreq.CommunicationID,
							CompleteResponse: rsp,
							ResponseCode:     responseCode,
							ResponseText:     "Submitted Successfully",
							ReferenceID:      referenceID,
						}
						_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
						rsp := response.NewCreateSMSResponse(&msgresponse)
						apiRsp := response.CreateSMSAPIResponse{
							StatusCodeAndMessage: port.CreateSuccess,
							Data:                 rsp,
						}
						handleCreateSuccess(ctx, apiRsp)
						return
					}

				} else {
					msgStoreRequest := ch.c.GetInt("sms.msgstorerequest")
					if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
						msgresponse := domain.MsgResponse{
							CommunicationID:  msgreq.CommunicationID,
							CompleteResponse: rsp,
							ResponseCode:     "402",
							ResponseText:     "Submitted Successfully",
							ReferenceID:      "",
						}
						_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
						rsp := response.NewCreateSMSResponse(&msgresponse)
						apiRsp := response.CreateSMSAPIResponse{
							StatusCodeAndMessage: port.CreateSuccess,
							Data:                 rsp,
						}
						handleCreateSuccess(ctx, apiRsp)
						return
					}

				}

			}
		} else if gateway == "2" {
			var NICUsername, NICPassword string
			switch msgreq.SenderID {
			case "INPOST":
				NICUsername = ch.c.GetString("sms.nic.INPOSTUserName")
				NICPassword = ch.c.GetString("sms.nic.INPOSTPassword")
			case "DOPBNK", "DOPCBS":
				NICUsername = ch.c.GetString("sms.nic.DOPBNKUserName")
				NICPassword = ch.c.GetString("sms.nic.DOPBNKPassword")
			case "DOPPLI":
				NICUsername = ch.c.GetString("sms.nic.DOPPLIUserName")
				NICPassword = ch.c.GetString("sms.nic.DOPPLIPassword")
			default:
				log.Error(ctx, "Invalid SenderID: %s", msgreq.SenderID)
				apierrors.HandleWithMessage(ctx, "Invalid SenderID")
				return
			}

			// rsp, err := ch.SendSMSNIC(NICUsername, NICPassword, msgreq.MessageText, msgreq.SenderID, msgreq.MobileNumbers, msgreq.EntityId, msgreq.TemplateID, msgreq.MessageType)
			rsp, err := ch.SendSMSNIC(SMSParams{
				Username:     NICUsername,
				Password:     NICPassword,
				Message:      msgreq.MessageText,
				SenderID:     msgreq.SenderID,
				MobileNumber: msgreq.MobileNumbers,
				TemplateID:   msgreq.TemplateID,
				MessageType:  msgreq.MessageType,
			})

			if err != nil {
				msgresponse := domain.MsgResponse{
					CommunicationID:  msgreq.CommunicationID,
					CompleteResponse: rsp,
					ResponseCode:     "02",
					ResponseText:     err.Error(),
					ReferenceID:      "",
				}
				_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
				// ch.vs.handleError(ctx, err)
				apierrors.HandleError(ctx, err)
				return
			}
			pattern := `Request ID=(\d+)~code=([A-Z0-9]+)`
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(rsp)
			if len(matches) >= 3 {
				// If success and format is good
				requestID := matches[1]
				responseCode := matches[2]
				// msgStoreRequest := ch.c.MessageStoreRequest()
				msgStoreRequest := ch.c.GetInt("sms.msgstorerequest")
				if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
					msgresponse := domain.MsgResponse{
						CommunicationID:  msgreq.CommunicationID,
						CompleteResponse: rsp,
						ResponseCode:     responseCode,
						ResponseText:     "Submitted Successfully",
						ReferenceID:      requestID,
					}
					_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
					// handleSuccess(ctx, msgresponse)
					rsp := response.NewCreateSMSResponse(&msgresponse)
					apiRsp := response.CreateSMSAPIResponse{
						StatusCodeAndMessage: port.CreateSuccess,
						Data:                 rsp,
					}
					handleCreateSuccess(ctx, apiRsp)
					return
				}
			}

		} else {
			// customError := CustomError{Message: "Invalid Gateway"}
			// ch.vs.handleError(ctx, customError)
			log.Error(ctx, "Invalid Gateway: %s", gateway)
			apierrors.HandleWithMessage(ctx, "Invalid Gateway")
		}
	} else {
		// handleSuccess(ctx, "Stored Successfully")
		apiRsp := response.CreateSMSAPIResponse{
			StatusCodeAndMessage: port.CreateSuccess,
			// Data:                 rsp,
		}
		handleCreateSuccess(ctx, apiRsp)
	}
}

func (ch *MgApplicationHandler) CreateSMSRequestHandlerKafka(ctx *gin.Context) {
	log.Debug(ctx, "Inside CreateSMSRequestHandler function")
	var req models.CreateSMSRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Error(ctx, "Binding failed for CreateSMSRequestHandler: %s", err.Error())
		// ch.vs.handleError(ctx, err)
		apierrors.HandleBindingError(ctx, err)
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for CreateSMSRequestHandler: %s", err.Error())
		return
	}

	msgreq := domain.MsgRequest{
		FacilityID:    req.FacilityID,
		ApplicationID: req.ApplicationID,
		Priority:      req.Priority,
		MessageText:   req.MessageText,
		SenderID:      req.SenderID,
		MobileNumbers: req.MobileNumbers,
		EntityId:      req.EntityId,
		TemplateID:    req.TemplateID,
		MessageType:   req.MessageType,
	}

	//Fetch Entity ID from config, if not assigned
	// msgreq.EntityId = ch.c.DltEntityID()
	msgreq.EntityId = ch.c.GetString("sms.dltEntityID")
	log.Debug(ctx, "Entity ID is : %s", msgreq.EntityId)
	gctx := context.Background()

	var gateway string
	// msgStoreRequest := ch.c.MessageStoreRequest()
	msgStoreRequest := ch.c.GetInt("sms.msgstorerequest")
	log.Debug(ctx, "Message Store Request ID is : %d", msgStoreRequest)

	//priorites are 1-OTP, 2-Transactional, 3-Promotional, 4-Bulk. If store is true or for Promotional and Bulk info will be saved.
	savedresponse, err := ch.svc.SaveMsgRequestTx(&gctx, &msgreq)
	if err != nil {
		log.Error(ctx, "DB Error in SaveMsgRequestTx: %s", err.Error())
		// ch.vs.handledbError(ctx, err)
		apierrors.HandleDBError(ctx, err)
		return
	}
	gateway = savedresponse.Gateway

	// log.Debug(ctx, "Gateway is : %s", gateway)

	//UC - Unicode message ; PM - Plaintext message
	if msgreq.MessageType == "UC" {
		if msgreq.Gateway == "1" {
			msgreq.MessageText = UnicodemsgConvertCDAC(msgreq.MessageText)
		} else {
			msgreq.MessageText = UnicodemsgConvertNIC(msgreq.MessageText)
		}
	} else {
		msgreq.MessageType = "PM"
	}
	// log.Debug(ctx, "Message Type is : %s", msgreq.MessageType)

	if gateway == "1" {
		// rsp, err := SendSMSCDAC(ch.c.CDACUserName(), ch.c.CDACPassword(), msgreq.MessageText, msgreq.SenderID, msgreq.MobileNumbers, ch.c.CDACSecureKey(), msgreq.TemplateID, msgreq.MessageType)
		rsp, err := ch.SendSMSCDAC(SMSParams{
			ch.c.GetString("sms.cdac.username"),
			ch.c.GetString("sms.cdac.password"),
			msgreq.MessageText,
			msgreq.SenderID,
			msgreq.MobileNumbers,
			ch.c.GetString("sms.cdac.securekey"),
			msgreq.TemplateID,
			msgreq.MessageType})
		if err != nil {
			msgresponse := domain.MsgResponse{
				CommunicationID:  msgreq.CommunicationID,
				CompleteResponse: rsp,
				ResponseCode:     "02",
				ResponseText:     err.Error(),
				ReferenceID:      "",
			}
			_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
			// ch.vs.handleError(ctx, err)
			apierrors.HandleError(ctx, err)
			return
		}
		log.Debug(ctx, "Response from SendSMSCDAC is : %s", rsp)

		SMSResponse := rsp[:5]

		if SMSResponse == "Error" {
			pattern := `Error (\d+) : (.+)`
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(rsp)
			if len(matches) < 3 {
				//if error and format of the message is good
				// fmt.Println("No matches found.")
				//  customError := CustomError{Message: "Invalid Response"}
				msgStoreRequest := ch.c.GetInt("sms.msgstorerequest")
				if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
					msgresponse := domain.MsgResponse{
						CommunicationID:  msgreq.CommunicationID,
						CompleteResponse: rsp,
						ResponseCode:     "400",
						ResponseText:     "Invalid Response",
						ReferenceID:      "",
					}
					_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
					// ch.vs.handleError(ctx, customError)
					apierrors.HandleWithMessage(ctx, "Invalid Response")
					return
				}

			} else {
				//if error and format is not good
				errorNumber := matches[1]
				errorMessage := matches[2]
				customError := CustomError{Message: "401, " + errorMessage}
				msgStoreRequest := ch.c.GetInt("sms.msgstorerequest")
				if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
					msgresponse := domain.MsgResponse{
						CommunicationID:  msgreq.CommunicationID,
						CompleteResponse: rsp,
						ResponseCode:     errorNumber,
						ResponseText:     errorMessage,
						ReferenceID:      "",
					}
					_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
				}
				// ch.vs.handleError(ctx, customError)
				apierrors.HandleError(ctx, customError)
				return
			}
		} else {

			pattern := `^(\d{3}),MsgID = (\d+)`
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(rsp)
			if len(matches) >= 3 {
				//if success and format is good
				responseCode := matches[1]
				referenceID := matches[2]
				msgStoreRequest := ch.c.GetInt("sms.msgstorerequest")
				if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
					msgresponse := domain.MsgResponse{
						CommunicationID:  msgreq.CommunicationID,
						CompleteResponse: rsp,
						ResponseCode:     responseCode,
						ResponseText:     "Submitted Successfully",
						ReferenceID:      referenceID,
					}
					_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
					// handleSuccess(ctx, msgresponse)
					rsp := response.NewCreateSMSResponse(&msgresponse)
					apiRsp := response.CreateSMSAPIResponse{
						StatusCodeAndMessage: port.CreateSuccess,
						Data:                 rsp,
					}
					handleCreateSuccess(ctx, apiRsp)
					return
				}

			} else {
				// msgStoreRequest := ch.c.MessageStoreRequest()
				msgStoreRequest := ch.c.GetInt("sms.msgstorerequest")
				if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
					msgresponse := domain.MsgResponse{
						CommunicationID:  msgreq.CommunicationID,
						CompleteResponse: rsp,
						ResponseCode:     "402",
						ResponseText:     "Submitted Successfully",
						ReferenceID:      "",
					}
					_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
					// handleSuccess(ctx, msgresponse)
					rsp := response.NewCreateSMSResponse(&msgresponse)
					apiRsp := response.CreateSMSAPIResponse{
						StatusCodeAndMessage: port.CreateSuccess,
						Data:                 rsp,
					}
					handleCreateSuccess(ctx, apiRsp)
					return
				}

			}

		}
	} else if gateway == "2" {
		var NICUsername, NICPassword string
		switch msgreq.SenderID {
		case "INPOST":
			NICUsername = ch.c.GetString("sms.nic.INPOSTUserName")
			NICPassword = ch.c.GetString("sms.nic.INPOSTPassword")
		case "DOPBNK", "DOPCBS":
			NICUsername = ch.c.GetString("sms.nic.DOPBNKUserName")
			NICPassword = ch.c.GetString("sms.nic.DOPBNKPassword")
		case "DOPPLI":
			NICUsername = ch.c.GetString("sms.nic.DOPPLIUserName")
			NICPassword = ch.c.GetString("sms.nic.DOPPLIPassword")
		default:
			log.Error(ctx, "Invalid SenderID: %s", msgreq.SenderID)
			apierrors.HandleWithMessage(ctx, "Invalid SenderID")
			return
		}

		// rsp, err := SendSMSNIC(NICUsername, NICPassword, msgreq.MessageText, msgreq.SenderID, msgreq.MobileNumbers, msgreq.EntityId, msgreq.TemplateID, msgreq.MessageType)
		rsp, err := ch.SendSMSNIC(SMSParams{
			Username:     NICUsername,
			Password:     NICPassword,
			Message:      msgreq.MessageText,
			SenderID:     msgreq.SenderID,
			MobileNumber: msgreq.MobileNumbers,
			TemplateID:   msgreq.TemplateID,
			MessageType:  msgreq.MessageType,
		})

		if err != nil {
			msgresponse := domain.MsgResponse{
				CommunicationID:  msgreq.CommunicationID,
				CompleteResponse: rsp,
				ResponseCode:     "02",
				ResponseText:     err.Error(),
				ReferenceID:      "",
			}
			_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
			// ch.vs.handleError(ctx, err)
			apierrors.HandleError(ctx, err)
			return
		}
		pattern := `Request ID=(\d+)~code=([A-Z0-9]+)`
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(rsp)
		if len(matches) >= 3 {
			// If success and format is good
			requestID := matches[1]
			responseCode := matches[2]
			// msgStoreRequest := ch.c.MessageStoreRequest()
			msgStoreRequest := ch.c.GetInt("sms.msgstorerequest")
			if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
				msgresponse := domain.MsgResponse{
					CommunicationID:  msgreq.CommunicationID,
					CompleteResponse: rsp,
					ResponseCode:     responseCode,
					ResponseText:     "Submitted Successfully",
					ReferenceID:      requestID,
				}
				_, _ = ch.svc.SaveResponseTx(&gctx, &msgresponse)
				// handleSuccess(ctx, msgresponse)
				rsp := response.NewCreateSMSResponse(&msgresponse)
				apiRsp := response.CreateSMSAPIResponse{
					StatusCodeAndMessage: port.CreateSuccess,
					Data:                 rsp,
				}
				handleCreateSuccess(ctx, apiRsp)
				return
			}
		}

	} else {
		// customError := CustomError{Message: "Invalid Gateway"}
		// ch.vs.handleError(ctx, customError)
		apierrors.HandleWithMessage(ctx, "Invalid Gateway")
	}

}

func (ch *MgApplicationHandler) SendTestMessage(ctx *gin.Context, payload map[string]interface{}) (map[string]interface{}, error) {

	url := ch.c.GetString("client.baseurl")

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Error(ctx, "Unable to marshal payload in SendTestMessage function %s", err.Error())
		apierrors.HandleMarshalError(ctx, err)
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
			Renegotiation:      tls.RenegotiateOnceAsClient,
		},
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	SMSResponse, err := client.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Error(ctx, "Error calling SMS Provider URL %s", err.Error())
		// apierrors.HandleErrorWithCustomMessage(ctx, "Error calling SMS Provider URL", err)
		apierrors.HandleErrorWithStatusCodeAndMessage(apierrors.HTTPErrorBadGateway, "Error calling SMS Provider URL: ", err)
		return nil, fmt.Errorf("failed to send request to SMS provider: %w", err)
	}
	defer SMSResponse.Body.Close()

	if SMSResponse.StatusCode != http.StatusCreated {
		apierrors.HandleWithMessage(ctx, "unable to send the message")
		return nil, fmt.Errorf("SMS provider returned status: %s", SMSResponse.Status)
	}

	// Decoding the response JSON into a map for structured access
	var responseData map[string]interface{}
	if err := json.NewDecoder(SMSResponse.Body).Decode(&responseData); err != nil {
		log.Error(ctx, "Failed to decode SMS provider response body %s", err.Error())
		return nil, fmt.Errorf("failed to decode SMS provider response: %w", err)
	}

	log.Info(ctx, "SMS sent successfully: %v", responseData)
	return responseData, nil
}

/*
func (ch *MgApplicationHandler) CreateTestSMSHandlerOld(ctx *gin.Context) {
	payload := make(map[string]interface{})

	payload["application_id"] = "4"
	payload["facility_id"] = "facility1"
	payload["priority"] = 1
	payload["message_text"] = "Dear Customer, OTP for booking is 1234, please do not share it with anyone - INDPOST"
	payload["sender_id"] = "INPOST"
	payload["mobile_numbers"] = "9634294395"
	payload["entity_id"] = "1001081725895192800"
	payload["template_id"] = "1007344609998507114"

	ch.SendTestMessage(ctx, payload)
}
*/

// Use models.CreateTestSMSRequest instead

/*
func (ch *MgApplicationHandler) CreateTestSMSHandlerOld2(ctx *gin.Context) {

	var req createTestSMSRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "JSON Binding failed for createTestSMSRequest: %s", err.Error())
		return
	}

	if err := validation.ValidateStruct(req); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for createTestSMSRequest: %s", err.Error())
		return
	}

	// Prepare the payload
	payload := map[string]interface{}{
		"application_id": "4",
		"facility_id":    "facility1",
		"priority":       1,
		"message_text":   "Dear Customer, OTP for booking is 1234, please do not share it with anyone - INDPOST",
		"sender_id":      "INPOST",
		"mobile_numbers": req.MobileNumber,
		"entity_id":      "1001081725895192800",
		"template_id":    "1007344609998507114",
	}

	// Send the SMS using SendTestMessage and capture the response or error
	rsp, err := ch.SendTestMessage(ctx, payload)
	if err != nil {
		log.Error(ctx, "Failed to send test SMS: %s", err.Error())
		apierrors.HandleError(ctx, err)
		return
	}

	// Return success response with SMS details
	apiRsp := response.TestSMSAPIResponse{
		//	StatusCodeAndMessage: port.CreateSuccess,
		// Message:              "Test SMS sent successfully",
		Data: rsp,
	}

	log.Debug(ctx, "CreateTestSMSHandler response: %v", apiRsp)
	handleSuccess(ctx, apiRsp)
}
*/

// CreateTestMessageHandler godoc
//
//	@Summary		Creates a test message request
//	@Description	Creates a new test message requests
//	@Tags			SMS Request
//	@ID				CreateTestSMSHandler
//	@Accept			json
//	@Produce		json
//	@Param			createSMSRequest	body		createTestSMSRequest		true	"Creates Message request"
//	@Success		201					{object}	response.TestSMSAPIResponse	"Success"
//	@Failure		400					{object}	apierrors.APIErrorResponse	"Bad Request"
//	@Failure		401					{object}	apierrors.APIErrorResponse	"Unauthorized"
//	@Failure		403					{object}	apierrors.APIErrorResponse	"Forbidden"
//	@Failure		404					{object}	apierrors.APIErrorResponse	"Data not found"
//	@Failure		409					{object}	apierrors.APIErrorResponse	"Data conflict errpr"
//	@Failure		422					{object}	apierrors.APIErrorResponse	"Binding or Validation error"
//	@Failure		500					{object}	apierrors.APIErrorResponse	"Internal server error"
//	@Failure		502					{object}	apierrors.APIErrorResponse	"Bad Gateway"
//	@Failure		504					{object}	apierrors.APIErrorResponse	"Gateway Timeout"
//	@Router			/test-sms-request [post]
func (ch *MgApplicationHandler) CreateTestSMSHandler(ctx *gin.Context) {
	var req models.CreateTestSMSRequest

	// Bind and validate JSON payload for mobile number
	if err := ctx.ShouldBindJSON(&req); err != nil {
		apierrors.HandleBindingError(ctx, err)
		log.Error(ctx, "JSON Binding failed for createTestSMSRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(ctx, err)
		log.Error(ctx, "Validation failed for createTestSMSRequest: %s", err.Error())
		return
	}

	// Prepare the payload
	payload := map[string]interface{}{
		"application_id": "4",
		"facility_id":    "facility1",
		"priority":       1,
		"message_text":   "Dear Customer, OTP for booking is 1234, please do not share it with anyone - INDPOST",
		"sender_id":      "INPOST",
		"mobile_numbers": req.MobileNumber,
		"entity_id":      "1001081725895192800",
		"template_id":    "1007344609998507114",
		"gateway":        "1",
		"message_type":   "PM",
	}

	// Send the SMS using SendTestMessage and capture the response or error
	rsp, err := ch.SendTestMessage(ctx, payload)
	if err != nil {
		log.Error(ctx, "Failed to send test SMS: %s", err.Error())
		apierrors.HandleError(ctx, err)
		return
	}

	// apiRsp := response.TestSMSAPIResponse{
	//StatusCodeAndMessage: port.CreateSuccess,
	// Message:              "Test SMS sent successfully",
	// Data: rsp,
	// }

	// log.Debug(ctx, "CreateTestSMSHandler response: %v", apiRsp)
	// handleSuccess(ctx, apiRsp)
	log.Debug(ctx, "CreateTestSMSHandler response: %v", rsp)
	handleSuccess(ctx, rsp)
}

type EditMgApplicationRequest struct {
	ApplicationID   uint64 `json:"application_id"`
	ApplicationName string `json:"application_name"`
	RequestType     string `json:"request_type"`
}

// type listApplicationsRequest struct {
// 	Skip  uint64 `form:"skip" binding:"required,min=0" example:"0"`
// 	Limit uint64 `form:"limit" binding:"required,min=5" example:"5"`
// }

type CustomError struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func GenerateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	randomString := base64.URLEncoding.EncodeToString(randomBytes)[:length]
	return randomString, nil
}

type SMSParams struct {
	Username     string
	Password     string
	Message      string
	SenderID     string
	MobileNumber string
	SecureKey    string
	TemplateID   string
	MessageType  string
}

func (ch *MgApplicationHandler) SendSMSCDAC(req SMSParams) (string, error) {
	log.Debug(nil, "Inside SendSMSCDAC function")
	log.Debug(nil, "req is : %v", req)
	var responseString string

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: false,
			},
		},
	}

	// Encrypt the password using MD5
	encryptedPassword, err := MD5(req.Password)
	if err != nil {
		log.Error(nil, "CDAC password encryption failed: %s", err.Error())
		apierrors.HandleErrorWithCustomMessage(nil, "CDAC password encryption failed", err)
		return "", err
	}
	// log.Debug(nil, "CDAC encryptedPassword is : %s", encryptedPassword)

	// Generate hash key
	hashKey := hashGenerator(req.Username, req.SenderID, req.Message, req.SecureKey)
	// log.Debug(nil, "CDAC hashKey is : %s", hashKey)

	// Prepare the request parameters
	data := url.Values{}
	data.Set("username", req.Username)
	data.Set("password", encryptedPassword)
	data.Set("mobileno", req.MobileNumber)
	data.Set("senderid", req.SenderID)
	data.Set("content", req.Message)
	if req.MessageType == "UC" {
		data.Set("smsservicetype", "unicodemsg")
	} else if strings.Contains(req.Message, "otp") || strings.Contains(req.Message, "OTP") {
		data.Set("smsservicetype", "otpmsg")
	} else {
		data.Set("smsservicetype", "singlemsg")
	}
	data.Set("key", hashKey)
	data.Set("templateid", req.TemplateID)

	// Make the HTTP POST request
	url := ch.c.GetString("sms.cdac.url")
	log.Debug(nil, "CDAC URL is : %s", url)

	resp, err := client.PostForm(url, data)
	if err != nil {
		log.Error(nil, "CDAC API Call failed: %s", err.Error())
		apierrors.HandleErrorWithCustomMessage(nil, "CDAC sendSMS API Call failed", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(nil, "Error reading response body: %s", err.Error())
		apierrors.HandleErrorWithCustomMessage(nil, "Error reading CDAC sendSMS response body", err)
		return "", err
	}

	// Check the HTTP response status
	//sample response: 402,MsgID = 060320251741252969158appostsms
	if resp.StatusCode != http.StatusOK {
		log.Error(nil, "CDAC sendSMS API returned non-OK status: %s", resp.Status)
		apierrors.HandleErrorWithCustomMessage(nil, "CDAC sendSMS API call failed", err)
		return "", fmt.Errorf("CDAC SMS Gateway returned non-OK status: %s", resp.Status)
	} else {
		log.Debug(nil, "CDAC sendSMS API call success: %s", resp.Status)
	}

	// Convert the response body to a string
	responseString = string(body)
	log.Debug(nil, "CDAC responseString is : %s", responseString)
	return responseString, nil
}

// func SendSMSNIC(username string, password string, message string, senderId string, mobileNumber string, entityId string, templateId string, messageType string) (string, error) {
func (ch *MgApplicationHandler) SendSMSNIC(smsreq SMSParams) (string, error) {

	log.Debug(nil, "Inside SendSMSNIC function")
	// log.Debug(nil, "smsreq is : %+v", smsreq)

	// baseURL := "https://smsgw.sms.gov.in/failsafe/HttpLink"

	baseURL := ch.c.GetString("sms.nic.url")
	// log.Debug(nil, "NIC Base URL is : %s", baseURL)
	entityId := ch.c.GetString("sms.dltEntityID")

	queryString := fmt.Sprintf("?username=%s&pin=%s&message=%s&mnumber=%s&signature=%s&dlt_entity_id=%s&dlt_template_id=%s&msgType=%s",
		smsreq.Username, smsreq.Password, smsreq.Message, smsreq.MobileNumber, smsreq.SenderID, entityId, smsreq.TemplateID, smsreq.MessageType)
	// log.Debug(nil, "NIC Query String is : %s", queryString)

	fullURL := baseURL + queryString
	// log.Debug(nil, "NIC Full URL is : %s", fullURL)

	// req, err := http.NewRequest("POST", fullURL, nil)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		log.Error(nil, "Failed to create NIC HTTP request: %s", err.Error())
		apierrors.HandleErrorWithCustomMessage(nil, "Failed to create HTTP request", err)
		return "", err
	}
	log.Debug(nil, "NIC HTTP request is : %+v", req)

	// Set the Content-Type header to application/x-www-form-urlencoded

	// Execute the HTTP request
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: false,
			},
			// Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(nil, "NIC sendSMS API call failed: %s", err.Error())
		// apierrors.HandleErrorWithCustomMessage(nil, "Failed to execute HTTP request", err)
		return "", err
	}
	log.Debug(nil, "NIC HTTP response is : %+v", resp)

	defer resp.Body.Close()

	// Check the HTTP response status
	if resp.StatusCode != http.StatusOK {
		log.Info(nil, "NIC sendSMS API call failed: %s", resp.Status)
		return "", fmt.Errorf("SMS Gateway returned non-OK status: %d %s", resp.StatusCode, resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	log.Debug(nil, "NIC response body is : %s", string(body))

	// Convert the body to a string for further processing
	responseString := string(body)

	if strings.Contains(responseString, "Message Accepted") {
		return responseString, nil
	} else {
		return "", fmt.Errorf("unexpected response from sms gateway: %s", responseString)
	}
}

func hashGenerator(userName string, senderId string, content string, secureKey string) string {
	finalString := userName + senderId + content + secureKey

	hashGen := finalString
	md := sha512.New()
	md.Write([]byte(hashGen))
	byteData := md.Sum(nil)

	sb := ""
	for _, b := range byteData {
		sb += fmt.Sprintf("%02x", b)
	}

	return sb
}

func MD5(text string) (string, error) {
	// Create a new SHA-1 hash instance
	hash := sha1.New()

	// Write the text to the hash
	_, err := io.WriteString(hash, text)
	if err != nil {
		return "", err
	}

	// Get the hash sum as a byte slice
	hashInBytes := hash.Sum(nil)

	// Convert the byte slice to a hexadecimal string
	md5String := convertedToHex(hashInBytes)

	return md5String, nil
}

func convertedToHex(data []byte) string {
	var buf []rune

	for i := 0; i < len(data); i++ {
		halfOfByte := (data[i] >> 4) & 0x0F
		twoHalfBytes := 0

		for twoHalfBytes < 2 {
			if 0 <= halfOfByte && halfOfByte <= 9 {
				buf = append(buf, rune('0'+halfOfByte))
			} else {
				buf = append(buf, rune('a'+(halfOfByte-10)))
			}

			halfOfByte = data[i] & 0x0F
			twoHalfBytes++
		}
	}

	return string(buf)
}

/*
func convertMap(mapData map[string]interface{}) map[string]string {
	result := make(map[string]string)

	for key, value := range mapData {
		// Convert the value to a string
		strValue, ok := value.(string)
		if !ok {
			// Handle the case where the value cannot be converted to a string
			strValue = fmt.Sprintf("%v", value)
		}

		// Add the key-value pair to the result map
		result[key] = strValue
	}

	return result
}
*/

func (ce CustomError) Error() string {
	return fmt.Sprintf("{Message: %s}", ce.Message)
}

// type FetchSMSRequestStatusHandlerRequest struct {
// 	MessageID uint64 `json:"message_id" validate:"required" example:"250220251740500435482appostsms"`
// }

// func (ch *MgApplicationHandler) FetchSMSRequestStatusHandler (gctx *gin.Context){
// 	var req FetchSMSRequestStatusHandlerRequest
// 	if err := gctx.ShouldBindJSON(&req); err != nil {
// 		apierrors.HandleBindingError(gctx, err)
// 		log.Error(gctx, "JSON Binding failed for FetchSMSRequestStatusHandlerRequest: %s", err.Error())
// 		return
// 	}

// 	if err := validation.ValidateStruct(req); err != nil {
// 		apierrors.HandleValidationError(gctx, err)
// 		log.Error(gctx, "Validation failed for FetchSMSRequestStatusHandlerRequest: %s", err.Error())
// 		return
// 	}

// 	// Fetch the SMS request status
// 	status, err := ch.svc.FetchSMSRequestStatusRepo(gctx, req.MessageID)
// 	if err != nil {
// 		apierrors.HandleDBError(gctx, err)
// 		log.Error(gctx, "Failed to fetch SMS request status: %s", err.Error())
// 		return
// 	}

// 	// Return the status in the response
// 	apiRsp := response.FetchSMSRequestStatusAPIResponse{
// 		StatusCodeAndMessage: port.FetchSuccess,
// 		Data:                 status,
// 	}

// 	log.Debug(gctx, "FetchSMSRequestStatusHandler response: %v", apiRsp)
// 	handleSuccess(gctx, apiRsp)

// }

// Use models.FetchCDACSMSDeliveryStatusRequest instead

func (ch *MgApplicationHandler) FetchCDACSMSDeliveryStatusHandler(gctx *gin.Context) {

	log.Debug(gctx, "Inside FetchCDACSMSDeliveryStatusHandler")

	var req models.FetchCDACSMSDeliveryStatusRequest
	if err := gctx.ShouldBindQuery(&req); err != nil {
		apierrors.HandleBindingError(gctx, err)
		log.Error(gctx, "Query Binding failed for FetchCDACSMSDeliveryStatusRequest: %s", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		apierrors.HandleValidationError(gctx, err)
		log.Error(gctx, "Validation failed for FetchCDACSMSDeliveryStatusRequest: %s", err.Error())
		return
	}

	cdacUserName := ch.c.GetString("sms.cdac.username")
	cdacPwd := ch.c.GetString("sms.cdac.password")
	var IsPwdEncrypted bool

	//Encrypting the password
	cdacPassword, err := MD5(cdacPwd)
	if err != nil {
		log.Error(gctx, "Failed to encrypt password: %s", err.Error())
		apierrors.HandleError(gctx, err)
		IsPwdEncrypted = false
		return
	} else {
		IsPwdEncrypted = true
	}

	smsDeliveryStatus := domain.CDACSMSDeliveryStatusRequest{
		UserName:       cdacUserName,
		Password:       cdacPassword,
		MessageID:      req.ReferenceID + cdacUserName,
		IsPwdEncrypted: IsPwdEncrypted,
	}
	log.Debug(gctx, "FetchCDACSMSDeliveryStatusHandler request: %v", smsDeliveryStatus)

	//API call to fetch the SMS delivery status

	baseURL := ch.c.GetString("sms.cdac.deliverystatusurl")
	params := url.Values{}
	params.Add("userid", smsDeliveryStatus.UserName)
	params.Add("password", smsDeliveryStatus.Password)
	params.Add("msgid", smsDeliveryStatus.MessageID)
	params.Add("pwd_encrypted", strconv.FormatBool(smsDeliveryStatus.IsPwdEncrypted))

	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	fmt.Println("delivery status url is:", url) // url := "https://msdgweb.mgov.gov.in/ReportAPI/csvreport
	method := "GET"

	client := &http.Client{}
	apireq, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Error(gctx, "Failed to build API Request: %s", err.Error())
		apierrors.HandleError(gctx, err)
		return
	}

	apiresponse, err := client.Do(apireq)
	if err != nil {
		log.Error(gctx, "CDAC Delivery status API call failed: %s", err.Error())
		apierrors.HandleError(gctx, err)
		return
	}
	defer apiresponse.Body.Close()

	if apiresponse.StatusCode != http.StatusOK {
		log.Error(gctx, "CDAC Delivery status API returned non-OK status: %d %s", apiresponse.StatusCode, apiresponse.Status)
		apierrors.HandleWithMessage(gctx, "CDAC Delivery status API returned non-OK status")
		return
	}

	body, err := io.ReadAll(apiresponse.Body)
	if err != nil {
		log.Error(gctx, "Failed to read response body: %s", err.Error())
		apierrors.HandleError(gctx, err)
		return
	}
	log.Debug(gctx, "CDAC Delivery status API Raw response: %v", string(body))

	// store the SMS request status
	// status, err := ch.svc.FetchCDACSMSDeliveryStatusRepo(gctx, smsDeliveryStatus)
	// if err != nil {
	// 	apierrors.HandleDBError(gctx, err)
	// 	log.Error(gctx, "Failed to call FetchCDACSMSDeliveryStatusRepo : %s", err.Error())
	// 	return
	// }

	// Return the status in the response
	statusLines := strings.Split(string(body), "\n")
	var statusResponses []*response.FetchCDACSMSDeliveryStatusResponse

	for _, line := range statusLines {
		status := strings.Split(line, ",")
		if len(status) < 3 {
			log.Error(gctx, "Invalid status response: %v", status)
			apierrors.HandleWithMessage(gctx, "Invalid status response")
			return
		}

		statusResponse := &response.FetchCDACSMSDeliveryStatusResponse{
			MobileNumber: status[0],
			SMSStatus:    status[1],
			TimeStamp:    status[2],
		}
		statusResponses = append(statusResponses, statusResponse)
	}

	apiRsp := response.FetchCDACSMSDeliveryStatusAPIResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data:                 statusResponses,
	}

	log.Debug(gctx, "FetchCDACSMSDeliveryStatusHandler response: %v", apiRsp)
	handleSuccess(gctx, apiRsp)
}
