package response

import (
	"MgApplication/core/domain"
	"MgApplication/core/port"
	"encoding/xml"
)

type bulkSMSInitiateResponse struct {
	MsgResponse string
	ReferenceID string
}

func NewBulkSMSInitiateResponse(rsp string, referenceId string) bulkSMSInitiateResponse {
	response := bulkSMSInitiateResponse{
		MsgResponse: rsp,
		ReferenceID: referenceId,
	}
	return response

}

type BulkSMSInitiateAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      bulkSMSInitiateResponse `json:"data"`
}

type ValidateBulkSMSOTPAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      bool `json:"data"`
}

/*
type transformedDataRsp struct {
	MobileNumber string `json:"mobile_number"`
	Message      string `json:"message"`
}

type buildTargetFileResponse struct {
	Rows     []transformedDataRsp `json:"rows"`
	UniqueID string               `json:"unique_id"`
}

func NewBuildTargetFileResponse(transRows []domain.TransformedData, uniqueID string) buildTargetFileResponse {
	var transformedRows []transformedDataRsp
	for _, row := range transRows {
		transformedRow := transformedDataRsp{
			MobileNumber: row.MobileNumber,
			Message:      row.Message,
		}
		transformedRows = append(transformedRows, transformedRow)
	}

	response := buildTargetFileResponse{
		Rows:     transformedRows,
		UniqueID: uniqueID,
	}
	return response
}

type BuildTargetFileAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`        // Common metadata like status code and message
	Data                      buildTargetFileResponse `json:"data"` // Embedded BuildTargetFileRsp struct
}
*/

type sendBulkSMSResponse struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id"`
	Code      string `json:"code"`
	Info      string `json:"info"`
}

// func NewSendBulkSMSResponseOld(bulk *domain.NicResponse) *sendBulkSMSResponse {
// 	response := sendBulkSMSResponse{
// 		Timestamp: bulk.Timestamp,
// 		RequestID: bulk.RequestID,
// 		Code:      bulk.Code,
// 		Info:      bulk.Info,
// 	}
// 	return &response

// }

type NicXmlResponse struct {
	XMLName   xml.Name `xml:"a2wml"`
	Version   string   `xml:"response>version"`
	Timestamp string   `xml:"response>timestamp"`
	RequestID string   `xml:"response>request ID"`
	Code      string   `xml:"response>code"`
	Info      string   `xml:"response>info"`
}

func NewSendBulkSMSResponse(bulk *domain.NicResponseXml) *sendBulkSMSResponse {
	response := sendBulkSMSResponse{
		Timestamp: bulk.Timestamp,
		RequestID: bulk.RequestID,
		Code:      bulk.Code,
		Info:      bulk.Info,
	}
	return &response

}

type SendBulkSMSAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      *sendBulkSMSResponse `json:"data"`
}
