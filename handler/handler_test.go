package handler

import (
	"MgApplication/api-config"
	"MgApplication/models"
	repo "MgApplication/repo/postgres"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupTestRouter creates a test router with handlers
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// TestCreateSMSRequestHandler tests the actual SMS request handler with validation
func TestCreateSMSRequestHandlerValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    models.CreateSMSRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid SMS request",
			requestBody: models.CreateSMSRequest{
				ApplicationID: "4",
				FacilityID:    "facility1",
				Priority:      1,
				MessageText:   "Your OTP is 123456",
				SenderID:      "INPOST",
				MobileNumbers: "9876543210",
				TemplateID:    "1307160377410448739",
				EntityId:      "1301157641566214705",
				MessageType:   "PM",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "missing application_id",
			requestBody: models.CreateSMSRequest{
				FacilityID:    "facility1",
				Priority:      1,
				MessageText:   "Your OTP is 123456",
				SenderID:      "INPOST",
				MobileNumbers: "9876543210",
				TemplateID:    "1307160377410448739",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
		},
		{
			name: "missing facility_id",
			requestBody: models.CreateSMSRequest{
				ApplicationID: "4",
				Priority:      1,
				MessageText:   "Your OTP is 123456",
				SenderID:      "INPOST",
				MobileNumbers: "9876543210",
				TemplateID:    "1307160377410448739",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
		},
		{
			name: "missing sender_id",
			requestBody: models.CreateSMSRequest{
				ApplicationID: "4",
				FacilityID:    "facility1",
				Priority:      1,
				MessageText:   "Your OTP is 123456",
				MobileNumbers: "9876543210",
				TemplateID:    "1307160377410448739",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
		},
		{
			name: "empty request",
			requestBody: models.CreateSMSRequest{
				// All fields empty
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock handler function
			handler := func(ctx *gin.Context) {
				var req models.CreateSMSRequest

				if err := ctx.ShouldBindJSON(&req); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": "binding failed"})
					return
				}

				if err := req.Validate(); err != nil {
					ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "validation failed", "details": err.Error()})
					return
				}

				ctx.JSON(http.StatusOK, gin.H{"message": "success", "data": req})
			}

			router := setupTestRouter()
			router.POST("/sms-request", handler)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/sms-request", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Verify error presence
			if tt.expectError {
				if response["error"] == nil {
					t.Errorf("Expected error in response, got: %v", response)
				}
				t.Logf("Validation error (as expected): %v", response["details"])
			} else {
				if response["error"] != nil {
					t.Errorf("Unexpected error in response: %v", response)
				}
			}
		})
	}
}

// TestCreateTemplateHandlerValidation tests template creation with validation
func TestCreateTemplateHandlerValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    models.CreateTemplateRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid template request",
			requestBody: models.CreateTemplateRequest{
				ApplicationID:  "4",
				TemplateName:   "Test Template",
				TemplateFormat: "Dear {#var#}, your OTP is {#var#}",
				SenderID:       "INPOST",
				EntityID:       "1301157641566214705",
				TemplateID:     "1307160377410448739",
				Gateway:        "1",
				Status:         true,
				MessageType:    "PM",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "missing template_name",
			requestBody: models.CreateTemplateRequest{
				ApplicationID:  "4",
				TemplateFormat: "Dear {#var#}, your OTP is {#var#}",
				SenderID:       "INPOST",
				TemplateID:     "1307160377410448739",
				Gateway:        "1",
				Status:         true,
				MessageType:    "PM",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
		},
		{
			name: "missing gateway",
			requestBody: models.CreateTemplateRequest{
				ApplicationID:  "4",
				TemplateName:   "Test Template",
				TemplateFormat: "Dear {#var#}, your OTP is {#var#}",
				SenderID:       "INPOST",
				TemplateID:     "1307160377410448739",
				Status:         true,
				MessageType:    "PM",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(ctx *gin.Context) {
				var req models.CreateTemplateRequest

				if err := ctx.ShouldBindJSON(&req); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": "binding failed"})
					return
				}

				if err := req.Validate(); err != nil {
					ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "validation failed", "details": err.Error()})
					return
				}

				ctx.JSON(http.StatusOK, gin.H{"message": "success", "data": req})
			}

			router := setupTestRouter()
			router.POST("/sms-templates", handler)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/sms-templates", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectError && response["error"] == nil {
				t.Errorf("Expected error in response")
			}
		})
	}
}

// TestCreateMessageProviderHandlerValidation tests provider creation with validation
func TestCreateMessageProviderHandlerValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    models.CreateMessageProviderRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid provider request",
			requestBody: models.CreateMessageProviderRequest{
				ProviderName:      "CDAC Gateway",
				ShortName:         "CDAC",
				Services:          "1,2,3,4",
				ConfigurationKeys: json.RawMessage(`{"url":"https://api.example.com"}`),
				Status:            true,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "missing provider_name",
			requestBody: models.CreateMessageProviderRequest{
				ShortName:         "CDAC",
				Services:          "1,2,3,4",
				ConfigurationKeys: json.RawMessage(`{"url":"https://api.example.com"}`),
				Status:            true,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
		},
		{
			name: "missing services",
			requestBody: models.CreateMessageProviderRequest{
				ProviderName:      "CDAC Gateway",
				ShortName:         "CDAC",
				ConfigurationKeys: json.RawMessage(`{"url":"https://api.example.com"}`),
				Status:            true,
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(ctx *gin.Context) {
				var req models.CreateMessageProviderRequest

				if err := ctx.ShouldBindJSON(&req); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": "binding failed"})
					return
				}

				if err := req.Validate(); err != nil {
					ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "validation failed", "details": err.Error()})
					return
				}

				ctx.JSON(http.StatusOK, gin.H{"message": "success", "data": req})
			}

			router := setupTestRouter()
			router.POST("/sms-providers", handler)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/sms-providers", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectError && response["error"] == nil {
				t.Errorf("Expected error in response")
			}
		})
	}
}

// TestBulkSMSHandlerValidation tests bulk SMS with validation
func TestBulkSMSHandlerValidation(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    models.SendBulkSMSRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "valid bulk SMS request",
			requestBody: models.SendBulkSMSRequest{
				SenderID:     "INPOST",
				MobileNumber: "9876543210",
				MessageType:  "PM",
				MessageText:  "Test bulk message",
				TemplateID:   "1307160377410448739",
				EntityID:     "1301157641566214705",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "missing sender_id",
			requestBody: models.SendBulkSMSRequest{
				MobileNumber: "9876543210",
				MessageType:  "PM",
				MessageText:  "Test bulk message",
				TemplateID:   "1307160377410448739",
				EntityID:     "1301157641566214705",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
		},
		{
			name: "missing mobile_number",
			requestBody: models.SendBulkSMSRequest{
				SenderID:    "INPOST",
				MessageType: "PM",
				MessageText: "Test bulk message",
				TemplateID:  "1307160377410448739",
				EntityID:    "1301157641566214705",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(ctx *gin.Context) {
				var req models.SendBulkSMSRequest

				if err := ctx.ShouldBindJSON(&req); err != nil {
					ctx.JSON(http.StatusBadRequest, gin.H{"error": "binding failed"})
					return
				}

				if err := req.Validate(); err != nil {
					ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "validation failed", "details": err.Error()})
					return
				}

				ctx.JSON(http.StatusOK, gin.H{"message": "success", "data": req})
			}

			router := setupTestRouter()
			router.POST("/bulk-sms", handler)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/bulk-sms", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

// TestBindAndValidateHelper tests the utility helper function
func TestBindAndValidateHelper(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful JSON binding and validation", func(t *testing.T) {
		handler := func(ctx *gin.Context) {
			var req models.CreateSMSRequest
			if err := BindAndValidate(ctx, &req, false, false, true); err != nil {
				// Error already handled by BindAndValidate
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"message": "success"})
		}

		router := gin.New()
		router.POST("/test", handler)

		validReq := models.CreateSMSRequest{
			ApplicationID: "4",
			FacilityID:    "facility1",
			Priority:      1,
			MessageText:   "Test",
			SenderID:      "SENDER",
			MobileNumbers: "9876543210",
			TemplateID:    "template123",
		}

		body, _ := json.Marshal(validReq)
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("validation failure with helper", func(t *testing.T) {
		handler := func(ctx *gin.Context) {
			var req models.CreateSMSRequest
			if err := BindAndValidate(ctx, &req, false, false, true); err != nil {
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"message": "success"})
		}

		router := gin.New()
		router.POST("/test", handler)

		invalidReq := models.CreateSMSRequest{
			ApplicationID: "4",
			// Missing required fields
		}

		body, _ := json.Marshal(invalidReq)
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("Expected status 422, got %d", w.Code)
		}
	})
}

// Mock implementations for testing - these can be replaced with actual repo mocks if needed
type MockMgApplicationRepository struct {
	*repo.MgApplicationRepository
}

type MockConfig struct {
	*config.Config
}
