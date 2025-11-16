package response

import (
	"MgApplication/core/domain"
	"MgApplication/core/port"
	"time"
)

type smsDashboardResponse struct {
	TotalSMSSent       int64 `json:"total_sms_sent" db:"total_sms_sent"`
	TotalOTPMessages   int64 `json:"total_otp_messages" db:"total_otps"`
	TotalTransMessages int64 `json:"total_trans_messages" db:"total_transactions"`
	TotalBulkMessages  int64 `json:"total_bulk_messages" db:"total_bulk_sms"`
	TotalPromMessages  int64 `json:"total_prom_messages" db:"total_promotional_sms"`
	TotalTemplates     int64 `json:"total_templates" db:"total_templates"`
	TotalProviders     int64 `json:"total_providers" db:"total_providers"`
	TotalApplications  int64 `json:"total_applications" db:"total_applications"`
}

func NewSMSDashboardResponse(dashboard *domain.SMSDashboard) *smsDashboardResponse {
	response := smsDashboardResponse{
		TotalSMSSent:       dashboard.TotalSMSSent,
		TotalOTPMessages:   dashboard.TotalOTPMessages,
		TotalTransMessages: dashboard.TotalTransMessages,
		TotalBulkMessages:  dashboard.TotalBulkMessages,
		TotalPromMessages:  dashboard.TotalPromMessages,
		TotalTemplates:     dashboard.TotalTemplates,
		TotalProviders:     dashboard.TotalProviders,
		TotalApplications:  dashboard.TotalApplications,
	}
	return &response
}

type SMSDashboardAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      *smsDashboardResponse `json:"data"`
}

type smsSentStatusReportResponse struct {
	SerialNo        uint64    `json:"serial_no" db:"serial_number"`
	CreatedDate     time.Time `json:"created_date" db:"created_date"`
	CommunicationID *string   `json:"comm_id" db:"communication_id"`
	ApplicationID   *string   `json:"application_id" db:"application_id"`
	FacilityID      *string   `json:"facility_id" db:"facility_id"`
	MessagePriority *int64    `json:"message_priority" db:"priority"`
	MessageText     *string   `json:"message_text" db:"message_text"`
	MobileNumber    *int64    `json:"mobile_number" db:"mobile_number"`
	GatewayID       *string   `json:"gateway_id" db:"gateway"`
	Status          string    `json:"status" db:"status"`
}

func NewSMSSentStatusReportResponse(reports []domain.SMSReport) []smsSentStatusReportResponse {
	var response []smsSentStatusReportResponse
	for _, report := range reports {
		ReportResponse := smsSentStatusReportResponse{
			SerialNo:        report.SerialNo,
			CreatedDate:     report.CreatedDate,
			CommunicationID: report.CommunicationID,
			ApplicationID:   report.ApplicationID,
			FacilityID:      report.FacilityID,
			MessagePriority: report.MessagePriority,
			MessageText:     report.MessageText,
			MobileNumber:    report.MobileNumber,
			GatewayID:       report.GatewayID,
			Status:          report.Status,
		}
		response = append(response, ReportResponse)
	}
	return response
}

type SMSSentStatusReportAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse     `json:",inline"`
	Data                      []smsSentStatusReportResponse `json:"data"`
}

type aggregateSMSReportResponse struct {
	SerialNo        uint64    `json:"serial_no" db:"serial_number"`
	ApplicationName string    `json:"application_name" db:"application_name"`
	TemplateName    string    `json:"template_name" db:"template_name"`
	ProviderName    string    `json:"provider_name" db:"provider_name"`
	CreatedDate     time.Time `json:"created_date" db:"created_date"`
	TotalSMS        int64     `json:"total_sms" db:"total_sms"`
	Success         int64     `json:"success" db:"success"`
	Failed          int64     `json:"failed" db:"failed"`
}

func NewAggregateSMSReportResponse(reports []domain.SMSAggregateReport) []aggregateSMSReportResponse {
	var response []aggregateSMSReportResponse
	for _, report := range reports {
		ReportResponse := aggregateSMSReportResponse{
			SerialNo:        report.SerialNo,
			ApplicationName: report.ApplicationName,
			TemplateName:    report.TemplateName,
			ProviderName:    report.ProviderName,
			CreatedDate:     report.CreatedDate,
			TotalSMS:        report.TotalSMS,
			Success:         report.Success,
			Failed:          report.Failed,
		}
		response = append(response, ReportResponse)
	}
	return response
}

type AggregateSMSReportAPIResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse     `json:",inline"`
	Data                      []aggregateSMSReportResponse `json:"data"`
}
