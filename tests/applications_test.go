package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)

// CreateMessageApplicationHandler
func TestCreateMessageApplicationHandlerSuccess(t *testing.T) {
	input := `{
			"application_name": "New Application Jathin2456890",
			"request_type": "4",
			"status": true
		}`
	req := httptest.NewRequest("POST", "/v1/applications", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreateMessageApplicationHandlerBindingError(t *testing.T) {
	input := `{
			"application_name": "New Application Jathin2456890",
			"request_type": "4",
			"status23": true,
		}`
	req := httptest.NewRequest("POST", "/v1/applications", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateMessageApplicationHandlerValidationError(t *testing.T) {
	input := `{
			"application_name": "New Application Jathin2456890",
			"request_type": "",
			"status23": true
		}`
	req := httptest.NewRequest("POST", "/v1/applications", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

/*
func Test_Create_Message_Application_MissingParam(t *testing.T) {
	input := `{}`
	req := httptest.NewRequest("POST", "/v1/applications", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
*/

// ListMessageApplicationsHandler
func TestListMessageApplicationsHandlerSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/applications", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestListMessageApplicationsHandlerBindingError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/applications?status=false", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListMessageApplicationsHandlerValidationError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/applications", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestListMessageApplicationsHandlerInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/applications", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListMessageApplicationsHandlerActiveSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/applications?status=true", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// FetchApplicationHandler
func TestFetchApplicationHandlerSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/applications/5", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFetchApplicationHandlerBindingError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/applications/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFetchApplicationHandlerValidationError(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/applications/4a", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestFetchApplicationHandlerInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/applications/5y", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// UpdateMessageApplicationHandler
func TestUpdateMessageApplicationHandlerSuccess(t *testing.T) {
	input := `{
		"application_name": "CEPT",
		"request_type": "3,4",
		"status": true
	}`

	req := httptest.NewRequest("PUT", "/v1/applications/3", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdateMessageApplicationHandlerBindingError(t *testing.T) {
	input := `{
		"application_name": "CEPT",
		"request_type": "3,4",
		"status23": true
	}`
	req := httptest.NewRequest("PUT", "/v1/applications/2", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateMessageApplicationHandlerValidationError(t *testing.T) {
	input := `{
		"applicationname": "CEPT",
		"request_type": "3,4",
		"status": true
	}`
	req := httptest.NewRequest("PUT", "/v1/applications/3", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

// ToggleApplicationStatusHandler
func TestToggleApplicationStatusHandlerSuccess(t *testing.T) {
	req := httptest.NewRequest("PUT", "/v1/applications/4/status", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestToggleApplicationStatusHandlerBindingError(t *testing.T) {
	req := httptest.NewRequest("PUT", "/v1/applications//status", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestToggleApplicationStatusHandlerValidationError(t *testing.T) {
	req := httptest.NewRequest("PUT", "/v1/applications/a/status", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestToggleApplicationStatusHandlerInvalidParam(t *testing.T) {
	req := httptest.NewRequest("PUT", "/v1/applications/4/status", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
