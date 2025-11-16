package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)

func TestInitiateBulkSmsSuccess(t *testing.T) {
	input := `{
	"application_id":"4",
	"template_name":"490",
	"mobile_no":"1111111111",
	"message_type":"PM",
	"test_msg":"Dear priyatham, OTP for testing is 12345. Don't share it with anyone - INDIAPOST"
	}`
	req := httptest.NewRequest("POST", "/v1/bulk-sms-initiate", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestInitiateBulkSmsMissingParam(t *testing.T) {
	input := `{}`
	req := httptest.NewRequest("POST", "/v1/bulk-sms-initiate", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestValidateTestSmsSuccess(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/bulk-sms-validate-otp?reference-id=4iq3zb8hzfaod8srftch&test-string=12345", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestValidateTestSmsInvalidParam(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/bulk-sms-validate-otp?reference-id=4iq3zb8hzfaod8srftch&test-string=12345", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSendFileSuccess(t *testing.T) {
	input := `[{
	"mobile_number":"90000000",
	"message_text":"Dear Candidate, please login to testvar1 to download Your Admit Card for testvar5 - INDIAPOST",
	"sender_id":"INPOST",
	"template_id":"1007441579539105995",
	"entity_id":"1001081725895192800",
	"message_type":"PM"
	},
	{
	"mobile_number":"9000000001",
	"message_text":"Dear Candidate, please login to testvar2 to download Your Admit Card for testvar6 - INDIAPOST",
	"sender_id":"INPOST",
	"template_id":"1007441579539105995",
	"entity_id":"1001081725895192800",
	"message_type":"PM"
	},
	{
	"mobile_number":"9000000002",
	"message_text":"Dear Candidate, please login to testvar3 to download Your Admit Card for testvar7 - INDIAPOST",
	"sender_id":"INPOST",
	"template_id":"1007441579539105995",
	"entity_id":"1001081725895192800",
	"message_type":"PM"
	}]`
	req := httptest.NewRequest("POST", "/v1/bulk-sms", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestSendFileMissingParam(t *testing.T) {
	input := `{}`
	req := httptest.NewRequest("POST", "/v1/bulk-sms", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}
