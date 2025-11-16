package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)

//SMSDashboardHandler
func TestGetDashboardDataSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-dashboard", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetDashboardDataInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-dashboard", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

//SentSMSStatusReportHandler
func TestSmsSentStatusSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-sent-status-report?from-date=01-01-2024&to-date=02-08-2024", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSmsSentStatusInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-sent-status-report?from-date=01-01-2024&to-date=02-08-2024", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

//AggregateSMSUsageReportHandler
func TestAggregateSmsReportAppwiseSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/aggregate-sms-report?from-date=01-01-2024&to-date=02-09-2024&report-type=1", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAggregateSmsReportInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/aggregate-sms-report?from-date=01-01-2024&to-date=02-09-2024&report-type=1", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAggregateSmsReportTemplatewiseSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/aggregate-sms-report?from-date=01-01-2024&to-date=02-09-2024&report-type=2", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAggregateSmsReportProviderwiseSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/aggregate-sms-report?from-date=01-01-2024&to-date=02-09-2024&report-type=3", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}


