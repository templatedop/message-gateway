package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	apierrors "MgApplication/api-errors"
	l "MgApplication/api-log"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

// Default values for Resty client initialization
const (
	DefaultTimeout      = 10 * time.Second
	DefaultRetryWait    = 1 * time.Second
	DefaultMaxRetryWait = 3 * time.Second
	DefaultMaxRetries   = 3
)

// ClientConfig encapsulates configuration parameters for Resty client initialization.
type ClientConfig struct {
	Timeout      time.Duration
	RetryWait    time.Duration
	MaxRetryWait time.Duration
	MaxRetries   int
	AppName      string
}

var (
	once          sync.Once
	urlInitOnce   sync.Once
	restyClient   *resty.Client
	globalAppName string
	globalMethod  string
	globalURL     string
	globalTimeout time.Duration
)

// AuthAPIResponse represents the structure of the response received from the authorization API.
type AuthAPIResponse struct {
	Success bool     `json:"success"`
	Message bool     `json:"message"`
	ErrorNo []string `json:"errorno,omitempty"`
	Data    *Data    `json:"data,omitempty"`
}

type Data struct {
	Message string `json:"message"`
}

// Payload represents the request payload sent to the authorization API.
type Payload struct {
	UserID         string `json:"user_id"`
	Endpoint       string `json:"endpoint"`
	ResourceMethod string `json:"resource_method"`
}

// AuthResult represents the structured response of the authorization check.
type AuthResult struct {
	Authorization bool `json:"authorization"`
}

// Init initializes the Resty client and global variables with the provided configuration.
func Init(config ClientConfig) error {
	once.Do(func() {
		// Set default values if not provided
		if config.Timeout == 0 {
			config.Timeout = DefaultTimeout
		}
		if config.RetryWait == 0 {
			config.RetryWait = DefaultRetryWait
		}
		if config.MaxRetryWait == 0 {
			config.MaxRetryWait = DefaultMaxRetryWait
		}
		if config.MaxRetries == 0 {
			config.MaxRetries = DefaultMaxRetries
		}

		restyClient = NewRestyClient(config.Timeout, config.RetryWait, config.MaxRetryWait, config.MaxRetries)
		globalAppName = config.AppName
		globalTimeout = config.Timeout

	})

	if restyClient == nil {
		appError := apierrors.NewAppError("Resty client Uninitialized", "500", nil)
		l.Error(nil, &appError)
		return &appError
	}
	return nil // No error, successful initialization
}

// Authorize method performs an authorization request using the internal Resty client.
func Authorize(ctx *gin.Context) (*AuthResult, error) {

	urlInitOnce.Do(func() {
		globalURL, globalMethod = getBaseURLAndMethod(ctx)
	})

	payload := Payload{
		UserID:         ctx.GetHeader("X-User-ID"),
		Endpoint:       globalAppName + ctx.Request.URL.Path,
		ResourceMethod: ctx.Request.Method,
	}

	// Validate the payload
	if validationError := validatePayload(payload); validationError != nil {
		// appError := apierrors.NewAppError(validationError.Error(), "401", validationError)
		l.Error(nil, "validation Error: %s", validationError.Error())
		return &AuthResult{
			Authorization: true,
		}, nil
	}

	// Making the actual HTTP request using the internal Resty client.
	resp, err := CallAuthorizationAPI(ctx, payload)
	if err != nil {
		appError := apierrors.NewAppError(err.Error(), "500", err)
		l.Error(nil, appError.Pretty)
		return nil, &appError
	}

	var responseParsed AuthAPIResponse
	if err := json.Unmarshal(resp.Body(), &responseParsed); err != nil {
		appError := apierrors.NewAppError(err.Error(), "500", err)
		l.Error(nil, appError.Pretty)
		return nil, &appError
	}

	if responseParsed.Message {
		l.Info(ctx, "Authorization for user "+payload.UserID+" is authorized to access "+payload.Endpoint)
	} else {
		l.Warn(ctx, "Authorization for user "+payload.UserID+" is not authorized to access "+payload.Endpoint)
	}

	return &AuthResult{
		Authorization: responseParsed.Message,
	}, nil
}

// validatePayload checks that all required fields in the Payload struct are populated.
func validatePayload(payload Payload) error {
	var errMsg string

	// Check for missing fields and append to error message if found
	if payload.UserID == "" {
		errMsg = "X-User-ID is missing in request header"
	}

	// Return all accumulated errors as a single error
	if len(errMsg) > 0 {
		l.Error(nil, errors.New(errMsg))
		return errors.New(errMsg)
	}

	return nil // No errors
}

func getBaseURLAndMethod(c *gin.Context) (string, string) {
	scheme := "https://"
	authApiPath := "beitrolemgmt/rolemanagement/v1/resource/check-access"
	if strings.Contains(c.Request.Host, "localhost") || strings.Contains(c.Request.Host, "127.0.0.1") {
		return fmt.Sprintf("%sdev.cept.gov.in/%s", scheme, authApiPath), "POST"
	}

	return fmt.Sprintf("%s%s/%s", scheme, c.Request.Host, authApiPath), "POST"
}
