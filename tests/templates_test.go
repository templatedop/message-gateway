package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)

// CreateTemplateHandler
func TestCreateTemplateHandlerSuccess(t *testing.T) {
	input := `{
	"template_local_id":"569",
	"application_id":"69",
	"template_name":"Test Template safron",
	"template_format":"Your OTP is {#val} for {#val} for",
	"sender_id":"INPOST",
	"entity_id":"16507160377410448739",
	"template_id":"165071603777774104478739",
	"message_type":"PM",
	"gateway":"1",
	"status":true
	}`
	req := httptest.NewRequest("POST", "/v1/sms-templates", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreateTemplateHandlerBindingError(t *testing.T) {
	input := `{
		"template_local_id":"569",
		"application_id":"69",
		"template_name":"Test Template 1233666",
		"template_format":"Your OTP is {#val} for {#val} for",
		"sender_id":"INPOST",
		"entityid":"16507160377410448739",
		"template_id":"165071603777774104478739",
		"message_type":"PM",
		"gateway":"1",
		"status":true
		}`
	req := httptest.NewRequest("POST", "/v1/sms-templates", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateTemplateHandlerValidationError(t *testing.T) {
	input := `{
		"template_local_id":"569",
		"application_id":"69abc",
		"template_name":"Test Template 1233666",
		"template_format":"Your OTP is {#val} for {#val} for",
		"sender_id":"INPOST",
		"entity_id":"16507160377410448739",
		"template_id":"165071603777774104478739",
		"message_type":"PM",
		"gateway":"1",
		"status":true
		}`
	req := httptest.NewRequest("POST", "/v1/sms-templates", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateTemplateHandlerMissingParam(t *testing.T) {
	input := `{}`
	req := httptest.NewRequest("POST", "/v1/sms-templates", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// ListTemplatesHandler
func TestListTemplatesHandlerSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates?skip=0&limit=0", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestListTemplatesHandlerBindingError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates?sk=0", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListTemplatesHandlerValidationError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates?skip=200abc", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestListTemplatesHandlerInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates?skip=0&limit=0", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// FetchTemplateHandler
func TestFetchTemplateHandlerSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/312", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFetchTemplateHandlerBindingError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/312", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFetchTemplateHandlerValidationError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/312", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// FetchTemplateByApplicationHandler
func TestFetchTemplateByApplicationHandlerSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/name?application-id=10", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFetchTemplateByApplicationHandlerBindingError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/name?applicationid=10", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFetchTemplateByApplicationHandlerValidationError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/name?application-id=abc", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFetchTemplateByApplicationHandlerInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/name?application-id=10", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// FetchTemplateDetailsHandler
func TestFetchTemplateDetailsHandlerSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/details?application-id=10", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFetchTemplateDetailsHandlerBindingError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/details?applicationid=10", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFetchTemplateDetailsHandlerValidationError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/details?application-id=10a", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFetchTemplateDetailsHandlerInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/details?application-id=10", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFetchTemplateDetailsHandlerByTemplateLocalIdSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/details?template-local-id=10", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFetchTemplateDetailsHandlerByTemplateLocalIdInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-templates/details?template-local-id=10", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFetchTemplateDetailsHandlerByTemplateFormatSuccess(t *testing.T) {
	// req := httptest.NewRequest("GET", "/v1/sms-templates/details?template-format=Article No. {%23var%23} Out for delivery through {%23var%23} on {%23var%23}. Use OTP {%23var%23} at the time of Delivery - INDIA POST", nil)
	req := httptest.NewRequest("GET",
		"/v1/sms-templates/details?template-format=Article%20No.%20{%23var%23}%20Out%20for%20delivery%20through%20{%23var%23}%20on%20{%23var%23}.%20Use%20OTP%20{%23var%23}%20at%20the%20time%20of%20Delivery%20-%20INDIA%20POST",
		nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFetchTemplateDetailsHandlerByTemplateFormatInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET",
		"/v1/sms-templates/details?templateformat=Article%20No.%20{%23var%23}%20Out%20for%20delivery%20through%20{%23var%23}%20on%20{%23var%23}.%20Use%20OTP%20{%23var%23}%20at%20the%20time%20of%20Delivery%20-%20INDIA%20POST",
		nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ToggleTemplateStatusHandler
func TestToggleTemplateStatusHandlerSuccess(t *testing.T) {
	req := httptest.NewRequest("PUT", "/v1/sms-templates/312/status", bytes.NewBuffer([]byte(`{"status":true}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestToggleTemplateStatusHandlerBindingError(t *testing.T) {
	req := httptest.NewRequest("PUT", "/v1/sms-templates//status", bytes.NewBuffer([]byte(`{"status":true}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestToggleTemplateStatusHandlerValidationError(t *testing.T) {
	req := httptest.NewRequest("PUT", "/v1/sms-templates/312abc/status", bytes.NewBuffer([]byte(`{"status":true}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestToggleTemplateStatusHandlerInvalidParam(t *testing.T) {
	req := httptest.NewRequest("PUT", "/v1/sms-templates/312/status", bytes.NewBuffer([]byte(`{"status":true}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// UpdateTemplateHandler
func TestUpdateTemplateHandlerSuccess(t *testing.T) {
	input := `{
	"template_local_id":"569",
	"application_id":"69",
	"template_name":"Test Template safron",
	"template_format":"Your OTP is {#val} for {#val} for",
	"sender_id":"INPOST",
	"entity_id":"16507160377410448739",
	"template_id":"165071603777774104478739",
	"message_type":"PM",
	"gateway":"1",
	"status":true
	}`
	req := httptest.NewRequest("PUT", "/v1/sms-templates/312", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdateTemplateHandlerBindingError(t *testing.T) {
	input := `{
		"template_local_id":"569",
		"application_id":"69",
		"template_name":"Test Template 1233666",
		"template_format":"Your OTP is {#val} for {#val} for",
		"sender_id":"INPOST",
		"entityid":"16507160377410448739",
		"template_id":"165071603777774104478739",
		"message_type":"PM",
		"gateway":"1",
		"status":true
		}`
	req := httptest.NewRequest("PUT", "/v1/sms-templates/312", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateTemplateHandlerValidationError(t *testing.T) {
	input := `{
		"template_local_id":"569",
		"application_id":"69abc",
		"template_name":"Test Template 1233666",
		"template_format":"Your OTP is {#val} for {#val} for",
		"sender_id":"INPOST",
		"entity_id":"16507160377410448739",
		"template_id":"165071603777774104478739",
		"message_type":"PM",
		"gateway":"1",
		"status":true
		}`
	req := httptest.NewRequest("PUT", "/v1/sms-templates/312", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}