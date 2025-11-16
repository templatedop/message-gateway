package handler

import (
	"MgApplication/core/domain"
	repo "MgApplication/repo/postgres"
	"context"
	"regexp"

	v1 "MgApplication/gen/smsrequest/v1"

	config "MgApplication/api-config"
	log "MgApplication/api-log"

	"connectrpc.com/connect"
)

// MgApplication Handler represents the HTTP handler for MgApplication related requests
type MgApplicationHandlergrpc struct {
	ch  *MgApplicationHandler
	svc *repo.MgApplicationRepository
	c   *config.Config
}

// MgApplication Handler creates a new MgApplicatPion Handler instance
func NewMgApplicationHandlergrpc(ch *MgApplicationHandler, svc *repo.MgApplicationRepository, c *config.Config) *MgApplicationHandlergrpc {
	return &MgApplicationHandlergrpc{
		ch,
		svc,
		c,
	}
}

func (mh *MgApplicationHandlergrpc) CreateSMSRequestHandler(ctx context.Context,
	req *connect.Request[v1.CreateSMSRequestHandlerRequest]) (resp *connect.Response[v1.CreateSMSRequestHandlerResponse], err error) {
	msgreq := domain.MsgRequest{
		FacilityID:    req.Msg.FacilityId,
		ApplicationID: req.Msg.ApplicationId,
		Priority:      int(req.Msg.Priority),
		MessageText:   req.Msg.MessageText,
		SenderID:      req.Msg.SenderId,
		MobileNumbers: req.Msg.MobileNumbers,
		EntityId:      req.Msg.EntityId,
		TemplateID:    req.Msg.TemplateId,
		MessageType:   req.Msg.MessageType,
	}

	//Fetch Entity ID from config, if not assigned
	// msgreq.EntityId = ch.c.DltEntityID()
	msgreq.EntityId = mh.c.GetString("sms.dltEntityID")
	log.Debug(ctx, "Entity ID is : %s", msgreq.EntityId)

	var gateway string
	// msgStoreRequest := ch.c.MessageStoreRequest()
	msgStoreRequest := mh.c.GetInt("sms.msgstorerequest")
	log.Debug(ctx, "Message Store Request ID is : %d", msgStoreRequest)

	if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
		//priorites are 1-OTP, 2-Transactional, 3-Promotional, 4-Bulk. If store is true or for Promotional and Bulk info will be saved.
		savedresponse, err := mh.svc.SaveMsgRequest(&ctx, &msgreq)
		if err != nil {
			log.Error(ctx, "DB Error in SaveMsgRequest: %s", err.Error())
			// apierrors.HandleDBError(ctx, err)
			return nil, err
		}
		gateway = savedresponse.Gateway
	} else {
		savedresponse, err := mh.svc.GetGateway(&ctx, &msgreq)
		if err != nil {
			log.Error(ctx, "DB Error in GetGateway: %s", err.Error())
			// apierrors.HandleDBError(ctx, err)
			return nil, err
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

			rsp, err := mh.ch.SendSMSCDAC(SMSParams{
				mh.c.GetString("sms.cdac.username"),
				mh.c.GetString("sms.cdac.password"),
				msgreq.MessageText,
				msgreq.SenderID,
				msgreq.MobileNumbers,
				mh.c.GetString("sms.cdac.securekey"),
				msgreq.TemplateID, msgreq.MessageType})
			if err != nil {
				msgresponse := domain.MsgResponse{
					CommunicationID:  msgreq.CommunicationID,
					CompleteResponse: rsp,
					ResponseCode:     "02",
					ResponseText:     err.Error(),
					ReferenceID:      "",
				}
				_, _ = mh.svc.SaveResponse(&ctx, &msgresponse)
				// ch.vs.handleError(ctx, err)
				// apierrors.HandleError(ctx, err)
				return nil, err
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
					msgStoreRequest := mh.c.GetInt("sms.msgstorerequest")
					if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
						msgresponse := domain.MsgResponse{
							CommunicationID:  msgreq.CommunicationID,
							CompleteResponse: rsp,
							ResponseCode:     "400",
							ResponseText:     "Invalid Response",
							ReferenceID:      "",
						}
						_, _ = mh.svc.SaveResponse(&ctx, &msgresponse)
						// apierrors.HandleWithMessage(ctx, "Invalid Response")
						return nil, err
					}

				} else {
					//if error and format is not good
					errorNumber := matches[1]
					errorMessage := matches[2]
					// customError := CustomError{Message: "401, " + errorMessage}
					msgStoreRequest := mh.c.GetInt("sms.msgstorerequest")
					if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
						msgresponse := domain.MsgResponse{
							CommunicationID:  msgreq.CommunicationID,
							CompleteResponse: rsp,
							ResponseCode:     errorNumber,
							ResponseText:     errorMessage,
							ReferenceID:      "",
						}
						_, _ = mh.svc.SaveResponse(&ctx, &msgresponse)
					}
					// ch.vs.handleError(ctx, customError)
					// apierrors.HandleError(ctx, customError)
					return nil, err
				}
			} else {

				pattern := `^(\d{3}),MsgID = (\d+)`
				re := regexp.MustCompile(pattern)
				matches := re.FindStringSubmatch(rsp)
				if len(matches) >= 3 {
					//if success and format is good
					responseCode := matches[1]
					referenceID := matches[2]
					msgStoreRequest := mh.c.GetInt("sms.msgstorerequest")
					if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
						msgresponse := domain.MsgResponse{
							CommunicationID:  msgreq.CommunicationID,
							CompleteResponse: rsp,
							ResponseCode:     responseCode,
							ResponseText:     "Submitted Successfully",
							ReferenceID:      referenceID,
						}
						_, _ = mh.svc.SaveResponse(&ctx, &msgresponse)
						// handleSuccess(ctx, msgresponse)
						// rsp := response.NewCreateSMSResponse(&msgresponse)
						// apiRsp := response.CreateSMSAPIResponse{
						// 	StatusCodeAndMessage: port.CreateSuccess,
						// 	Data:                 rsp,
						// }
						// handleCreateSuccess(ctx, apiRsp)
						// return nil, err
						return connect.NewResponse(
							&v1.CreateSMSRequestHandlerResponse{}), nil
					}

				} else {
					// msgStoreRequest := mh.c.MessageStoreRequest()
					msgStoreRequest := mh.c.GetInt("sms.msgstorerequest")
					if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
						msgresponse := domain.MsgResponse{
							CommunicationID:  msgreq.CommunicationID,
							CompleteResponse: rsp,
							ResponseCode:     "402",
							ResponseText:     "Submitted Successfully",
							ReferenceID:      "",
						}
						_, _ = mh.svc.SaveResponse(&ctx, &msgresponse)
						// handleSuccess(ctx, msgresponse)
						// rsp := response.NewCreateSMSResponse(&msgresponse)
						// apiRsp := response.CreateSMSAPIResponse{
						// 	StatusCodeAndMessage: port.CreateSuccess,
						// 	Data:                 rsp,
						// }
						// handleCreateSuccess(ctx, apiRsp)
						return connect.NewResponse(
							&v1.CreateSMSRequestHandlerResponse{}), nil
					}

				}

			}
		} else if gateway == "2" {
			var NICUsername, NICPassword string
			if msgreq.SenderID == "INPOST" {
				NICUsername = mh.c.GetString("sms.nic.INPOSTUserName")
				NICPassword = mh.c.GetString("sms.nic.INPOSTPassword")
			} else if (msgreq.SenderID == "DOPBNK") || (msgreq.SenderID == "DOPCBS") {
				NICUsername = mh.c.GetString("sms.nic.DOPBNKUserName")
				NICPassword = mh.c.GetString("sms.nic.DOPBNKPassword")
			} else if msgreq.SenderID == "DOPPLI" {
				NICUsername = mh.c.GetString("sms.nic.DOPPLIUserName")
				NICPassword = mh.c.GetString("sms.nic.DOPPLIPassword")
			}

			// rsp, err := SendSMSNIC(NICUsername, NICPassword, msgreq.MessageText, msgreq.SenderID, msgreq.MobileNumbers, msgreq.EntityId, msgreq.TemplateID, msgreq.MessageType)
			rsp, err := mh.ch.SendSMSNIC(SMSParams{
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
				_, _ = mh.svc.SaveResponse(&ctx, &msgresponse)
				// apierrors.HandleError(ctx, err)
				return nil, err
			}
			pattern := `Request ID=(\d+)~code=([A-Z0-9]+)`
			re := regexp.MustCompile(pattern)
			matches := re.FindStringSubmatch(rsp)
			if len(matches) >= 3 {
				// If success and format is good
				requestID := matches[1]
				responseCode := matches[2]
				// msgStoreRequest := mh.c.MessageStoreRequest()
				msgStoreRequest := mh.c.GetInt("sms.msgstorerequest")
				if msgStoreRequest == 1 || msgreq.Priority == 3 || msgreq.Priority == 4 {
					msgresponse := domain.MsgResponse{
						CommunicationID:  msgreq.CommunicationID,
						CompleteResponse: rsp,
						ResponseCode:     responseCode,
						ResponseText:     "Submitted Successfully",
						ReferenceID:      requestID,
					}
					_, _ = mh.svc.SaveResponse(&ctx, &msgresponse)
					// handleSuccess(ctx, msgresponse)
					// rsp := response.NewCreateSMSResponse(&msgresponse)
					// apiRsp := response.CreateSMSAPIResponse{
					// 	StatusCodeAndMessage: port.CreateSuccess,
					// 	Data:                 rsp,
					// }
					// handleCreateSuccess(ctx, apiRsp)
					// return nil, err
					return connect.NewResponse(
						&v1.CreateSMSRequestHandlerResponse{}), nil

				}
			}

		} else {
			// customError := CustomError{Message: "Invalid Gateway"}
			// ch.vs.handleError(ctx, customError)
			// apierrors.HandleWithMessage(ctx, "Invalid Gateway")
		}
	} else {
		// handleSuccess(ctx, "Stored Successfully")
		// apiRsp := response.CreateSMSAPIResponse{
		// 	StatusCodeAndMessage: port.CreateSuccess,
		// 	// Data:                 rsp,
		// }
		// handleCreateSuccess(ctx, apiRsp)
		return connect.NewResponse(
			&v1.CreateSMSRequestHandlerResponse{}), nil
	}
	return
}
