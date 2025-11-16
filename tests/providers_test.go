package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)

// CreateMessageProviderHandler
func TestCreateMessageProviderHandlerrSuccess(t *testing.T) {
	input := `{
		"provider_name": "New Provider Jathin2445",
		"short_name": "NPJ2",
		"services": "1,3,4",
		"status": true,
		"configuration_keys": "[
		{\"keyname\":\"key1\",\"keyvalue\":\"key_value1\"},
		{\"keyname\":\"key2\",\"keyvalue\":\"key_value2\"},
		{\"keyname\":\"key3\",\"keyvalue\":\"key_value3\"},
		{\"keyname\":\"key4\",\"keyvalue\":\"key_value4\"}
		]"
	}`
	req := httptest.NewRequest("POST", "/v1/sms-providers", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreateMessageProviderHandlerBindingError(t *testing.T) {
	input := `{
		"provider_name": "New Provider Jathin2445",
		"short_name": "NPJ2",
		"services": "1,3,4",
		"status": true,
		"configuration_keys": "[
		{\"keyname\":\"key1\",\"keyvalue\":\"key_value1\"},
		{\"keyname\":\"key2\",\"keyvalue\":\"key_value2\"},
		{\"keyname\":\"key3\",\"keyvalue\":\"key_value3\"},
		{\"keyname\":\"key4\",\"keyvalue\":\"key_value4\"}
		]"
	}`
	req := httptest.NewRequest("POST", "/v1/sms-providers", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateMessageProviderHandlerValidationError(t *testing.T) {
	input := `{
		"provider_name": "New Provider Jathin2445",
		"short_name": "NPJ2",
		"services": "1,3,4",
		"status": true,
		"configuration_keys": "[
		{\"keyname\":\"key1\",\"keyvalue\":\"key_value1\"},
		{\"keyname\":\"key2\",\"keyvalue\":\"key_value2\"},
		{\"keyname\":\"key3\",\"keyvalue\":\"key_value3\"},
		{\"keyname\":\"key4\",\"keyvalue\":\"key_value4\"}
		]"
	}`
	req := httptest.NewRequest("POST", "/v1/sms-providers", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

/*
func Test_CreateMessageProviderHandler_MissingParam(t *testing.T) {
	input := `{}`
	req := httptest.NewRequest("POST", "/v1/sms-providers", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
*/

// ListMessageProvidersHandler
func TestListMessageProvidersHandlerSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-providers", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestListMessageProvidersHandlerBindingError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-providers?status=false", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListMessageProvidersHandlerValidationError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-providers", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestListMessageProvidersHandlerInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-providers", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// FetchMessageProviderHandler
func TestFetchMessageProviderHandlerSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-providers/5", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFetchMessageProviderHandlerBindingError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-providers/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFetchMessageProviderHandlerValidationError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-providers/4a", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestFetchMessageProviderHandlerInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/sms-providers/5y", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// UpdateMessageProviderHandler
func TestUpdateMessageProviderHandlerSuccess(t *testing.T) {
	input := `{
		"provider_name": "testcasen",
		"services": "3,4",
		"status": true
	}`

	req := httptest.NewRequest("PUT", "/v1/sms-providers/4", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdateMessageProviderHandlerBindingError(t *testing.T) {
	input := `{
		"providername": "testcasen",
		"services": "3,4",
		"status": true
	}`
	req := httptest.NewRequest("PUT", "/v1/sms-providers/4", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateMessageProviderHandlerValidationError(t *testing.T) {
	input := `{
		"provider_name": "testcasen",
		"services": "3,4",
		"status": b
	}`
	req := httptest.NewRequest("PUT", "/v1/sms-providers/4", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// ToggleMessageProviderStatusHandler
func TestToggleMessageProviderStatusHandlerSuccess(t *testing.T) {
	req := httptest.NewRequest("PUT", "/v1/sms-providers/4/status", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestToggleMessageProviderStatusHandlerBindingError(t *testing.T) {
	req := httptest.NewRequest("PUT", "/v1/sms-providers//status", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestToggleMessageProviderStatusHandlerValidationError(t *testing.T) {
	req := httptest.NewRequest("PUT", "/v1/sms-providers/a/status", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestToggleMessageProviderStatusHandlerInvalidParam(t *testing.T) {
	req := httptest.NewRequest("PUT", "/v1/sms-providers/4/status", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
