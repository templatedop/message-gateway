package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)


func TestCreateMessageRequestSuccess(t *testing.T) {
	input := `{
		"application_id":"4",
		"facility_id":"facility1",
		"priority":1,
		"message_text":"Dear Customer, OTP for booking is 1234, please do not share it with anyone - INDPOST",
		"sender_id":"INPOST",
		"mobile_numbers":"1111111111",
		"entity_id":"1001081725895192800",
		"template_id":"1007344609998507114"
	}`
	req := httptest.NewRequest("POST", "/v1/sms-request", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreateMessageRequestDbErrorAppTemplateMapping(t *testing.T) {
	input := `{
		"application_id":"5",
		"facility_id":"8587808287",
		"message_text":"Dear Customer, Your contract with ID 123 has been approved.- INDIAPOST",
		"mobile_numbers":"1111111111",
		"priority":1,
		"sender_id":"INPOST",
		"entity_id":"1001081725895192800",
		"template_id":"1007781334111659021"
		}`
	req := httptest.NewRequest("POST", "/v1/sms-request", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreateMessageRequestMissingParam(t *testing.T) {
	input := `{}`
	req := httptest.NewRequest("POST", "/v1/sms-request", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestCreateMessageRequestUnicodeCdacSuccess(t *testing.T) {
	input := `{
		"application_id":"23",
		"facility_id":"facility1",
		"priority":1,
		"message_text": "प्रिय Phani, abc ऐप के लिए ओटीपी (OTP) 123456 है। कृपया इसे किसी के साथ साझा न करें। यह 5 मिनट के लिए वैध है। - भारतीय डाक",
		"sender_id":"INPOST",
		"mobile_numbers":"1111111111",
		"entity_id":"1001081725895192800",
		"template_id":"1007667344967470975",
		"message_type":"UC"
	}`
	req := httptest.NewRequest("POST", "/v1/sms-request", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}


func TestCreateMessageRequestUnicodeNicSuccess(t *testing.T) {
	input := `{
		"application_id":"2",
		"facility_id":"facility1",
		"priority":1,
		"message_text": "प्रिय ग्राहक सुकन्या समृद्धि योजना खाता खोलकर अपनी प्यारी बिटिया का भविष्य सुरक्षित करने के लिए डाक विभाग आपको बधाई देता है।",
		"sender_id":"DOPBNK",
		"mobile_numbers":"1111111111",
		"entity_id":"1001081725895192800",
		"template_id":"1007340152087500154",
		"message_type":"UC"
	}`
	req := httptest.NewRequest("POST", "/v1/sms-request", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}
func TestTestMessageSuccess(t *testing.T) {
	input := `{"mobile_number":"1111111111"}`
	req := httptest.NewRequest("POST", "/v1/test-sms-request", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestTestMessageMissingParam(t *testing.T) {
	input := `{}`
	req := httptest.NewRequest("POST", "/v1/test-sms-request", bytes.NewBuffer([]byte(input)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}