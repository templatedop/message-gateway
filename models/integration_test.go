package models

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// Mock handler to test validation in HTTP context
func createSMSHandler(c *gin.Context) {
	var req CreateSMSRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "binding failed", "details": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "validation failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success", "data": req})
}

func TestValidationInHTTPContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name: "valid SMS request",
			requestBody: CreateSMSRequest{
				ApplicationID: "123",
				FacilityID:    "facility1",
				Priority:      1,
				MessageText:   "Test message",
				SenderID:      "SENDER",
				MobileNumbers: "9876543210",
				TemplateID:    "template123",
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name: "missing required fields",
			requestBody: CreateSMSRequest{
				ApplicationID: "123",
				FacilityID:    "facility1",
				// Missing other required fields
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectSuccess:  false,
		},
		{
			name: "completely empty request",
			requestBody: CreateSMSRequest{
				// All fields empty
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectSuccess:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router
			router := gin.New()
			router.POST("/sms", createSMSHandler)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/sms", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Verify success/failure
			if tt.expectSuccess {
				if response["message"] != "success" {
					t.Errorf("Expected success response, got: %v", response)
				}
			} else {
				if response["error"] == "" {
					t.Errorf("Expected error response, got: %v", response)
				}
				t.Logf("Validation error (as expected): %v", response["details"])
			}
		})
	}
}

func TestMultipleRequestValidations(t *testing.T) {
	// Test validating multiple different request types
	tests := []struct {
		name      string
		validator func() error
		wantErr   bool
	}{
		{
			name: "valid CreateSMSRequest",
			validator: func() error {
				req := CreateSMSRequest{
					ApplicationID: "123",
					FacilityID:    "facility1",
					Priority:      1,
					MessageText:   "Test",
					SenderID:      "SENDER",
					MobileNumbers: "9876543210",
					TemplateID:    "template123",
				}
				return req.Validate()
			},
			wantErr: false,
		},
		{
			name: "valid CreateTemplateRequest",
			validator: func() error {
				req := CreateTemplateRequest{
					ApplicationID:  "123",
					TemplateName:   "Test Template",
					TemplateFormat: "Hello {#var#}",
					SenderID:       "SENDER",
					TemplateID:     "template123",
					Gateway:        "1",
					Status:         true,
					MessageType:    "PM",
				}
				return req.Validate()
			},
			wantErr: false,
		},
		{
			name: "valid SendBulkSMSRequest",
			validator: func() error {
				req := SendBulkSMSRequest{
					SenderID:     "SENDER",
					MobileNumber: "9876543210",
					MessageType:  "PM",
					MessageText:  "Test",
					TemplateID:   "template123",
					EntityID:     "entity123",
				}
				return req.Validate()
			},
			wantErr: false,
		},
		{
			name: "invalid CreateSMSRequest",
			validator: func() error {
				req := CreateSMSRequest{
					ApplicationID: "123",
					// Missing required fields
				}
				return req.Validate()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConcurrentValidation(t *testing.T) {
	// Test that validation is thread-safe
	const goroutines = 100
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			req := CreateSMSRequest{
				ApplicationID: "123",
				FacilityID:    "facility1",
				Priority:      id,
				MessageText:   "Test message",
				SenderID:      "SENDER",
				MobileNumbers: "9876543210",
				TemplateID:    "template123",
			}

			// Validate multiple times
			for j := 0; j < 100; j++ {
				_ = req.Validate()
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

func BenchmarkHTTPValidation(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/sms", createSMSHandler)

	req := CreateSMSRequest{
		ApplicationID: "123",
		FacilityID:    "facility1",
		Priority:      1,
		MessageText:   "Test message",
		SenderID:      "SENDER",
		MobileNumbers: "9876543210",
		TemplateID:    "template123",
	}

	body, _ := json.Marshal(req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpReq := httptest.NewRequest(http.MethodPost, "/sms", bytes.NewBuffer(body))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httpReq)
	}
}
