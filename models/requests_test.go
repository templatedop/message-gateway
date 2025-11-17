package models

import (
	"testing"

	govaliderrors "github.com/templatedop/govalid/validation/errors"
)

func TestCreateSMSRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateSMSRequest
		wantErr bool
		errType string
	}{
		{
			name: "valid request with all required fields",
			req: CreateSMSRequest{
				ApplicationID: "123",
				FacilityID:    "facility1",
				Priority:      1,
				MessageText:   "Test SMS message",
				SenderID:      "SENDER",
				MobileNumbers: "9876543210",
				TemplateID:    "template123",
			},
			wantErr: false,
		},
		{
			name: "missing application id",
			req: CreateSMSRequest{
				FacilityID:    "facility1",
				Priority:      1,
				MessageText:   "Test SMS message",
				SenderID:      "SENDER",
				MobileNumbers: "9876543210",
				TemplateID:    "template123",
			},
			wantErr: true,
			errType: "required",
		},
		{
			name: "missing facility id",
			req: CreateSMSRequest{
				ApplicationID: "123",
				Priority:      1,
				MessageText:   "Test SMS message",
				SenderID:      "SENDER",
				MobileNumbers: "9876543210",
				TemplateID:    "template123",
			},
			wantErr: true,
			errType: "required",
		},
		{
			name: "missing multiple required fields",
			req: CreateSMSRequest{
				ApplicationID: "123",
			},
			wantErr: true,
			errType: "required",
		},
		{
			name: "zero priority (required validation)",
			req: CreateSMSRequest{
				ApplicationID: "123",
				FacilityID:    "facility1",
				Priority:      0,
				MessageText:   "Test SMS message",
				SenderID:      "SENDER",
				MobileNumbers: "9876543210",
				TemplateID:    "template123",
			},
			wantErr: true,
			errType: "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSMSRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErr {
				validationErrors, ok := err.(govaliderrors.ValidationErrors)
				if !ok {
					t.Errorf("Expected ValidationErrors type, got %T", err)
					return
				}
				if len(validationErrors) == 0 {
					t.Error("Expected validation errors, got none")
				}
			}
		})
	}
}

func TestCreateTemplateRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateTemplateRequest
		wantErr bool
	}{
		{
			name: "valid template request",
			req: CreateTemplateRequest{
				ApplicationID:  "123",
				TemplateName:   "Test Template",
				TemplateFormat: "Hello {#var#}",
				SenderID:       "SENDER",
				TemplateID:     "template123",
				Gateway:        "1",
				Status:         true,
				MessageType:    "PM",
			},
			wantErr: false,
		},
		{
			name: "missing required fields",
			req: CreateTemplateRequest{
				ApplicationID: "123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTemplateRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateMessageProviderRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateMessageProviderRequest
		wantErr bool
	}{
		{
			name: "valid provider request",
			req: CreateMessageProviderRequest{
				ProviderName:      "Test Provider",
				ShortName:         "TP",
				Services:          "1,2,3",
				ConfigurationKeys: []byte(`{"key1":"value1"}`),
				Status:            true,
			},
			wantErr: false,
		},
		{
			name: "missing provider name",
			req: CreateMessageProviderRequest{
				ShortName:         "TP",
				Services:          "1,2,3",
				ConfigurationKeys: []byte(`{"key1":"value1"}`),
				Status:            true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateMessageProviderRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationErrorDetails(t *testing.T) {
	req := CreateSMSRequest{
		// Missing all required fields
	}

	err := req.Validate()
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	validationErrors, ok := err.(govaliderrors.ValidationErrors)
	if !ok {
		t.Fatalf("Expected ValidationErrors type, got %T", err)
	}

	// Check that we got multiple validation errors
	if len(validationErrors) == 0 {
		t.Error("Expected at least one validation error")
	}

	// Verify error structure
	for _, ve := range validationErrors {
		if ve.Path == "" {
			t.Error("Expected non-empty Path in validation error")
		}
		if ve.Reason == "" {
			t.Error("Expected non-empty Reason in validation error")
		}
		if ve.Type == "" {
			t.Error("Expected non-empty Type in validation error")
		}
	}

	t.Logf("Validation errors: %v", validationErrors)
}

func TestBulkSMSRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     SendBulkSMSRequest
		wantErr bool
	}{
		{
			name: "valid bulk SMS request",
			req: SendBulkSMSRequest{
				SenderID:     "SENDER",
				MobileNumber: "9876543210",
				MessageType:  "PM",
				MessageText:  "Test message",
				TemplateID:   "template123",
				EntityID:     "entity123",
			},
			wantErr: false,
		},
		{
			name: "missing sender id",
			req: SendBulkSMSRequest{
				MobileNumber: "9876543210",
				MessageType:  "PM",
				MessageText:  "Test message",
				TemplateID:   "template123",
				EntityID:     "entity123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SendBulkSMSRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNilRequestValidation(t *testing.T) {
	// Test that nil requests are handled correctly
	var req *CreateSMSRequest = nil

	err := ValidateCreateSMSRequest(req)
	if err == nil {
		t.Error("Expected error for nil request, got nil")
	}

	if err != ErrNilCreateSMSRequest {
		t.Errorf("Expected ErrNilCreateSMSRequest, got %v", err)
	}
}

func BenchmarkValidateCreateSMSRequest(b *testing.B) {
	req := CreateSMSRequest{
		ApplicationID: "123",
		FacilityID:    "facility1",
		Priority:      1,
		MessageText:   "Test SMS message",
		SenderID:      "SENDER",
		MobileNumbers: "9876543210",
		TemplateID:    "template123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Validate()
	}
}

func BenchmarkValidateCreateSMSRequestWithErrors(b *testing.B) {
	req := CreateSMSRequest{
		// Missing required fields to trigger validation errors
		ApplicationID: "123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Validate()
	}
}
