package client

import (
	"context"
	"time"

	apierrors "MgApplication/api-errors"

	"github.com/go-resty/resty/v2"
)

// NewRestyClient creates and returns a new Resty client with custom configuration.
// The client is configured with timeout, retry count, and retry wait times. It also
// adds a retry condition that triggers retries when a 5xx server error or any network
// error occurs.
func NewRestyClient(Timeout, RetryWait, MaxRetryWait time.Duration, MaxRetries int) *resty.Client {
	client := resty.New().
		SetTimeout(Timeout).
		SetRetryCount(MaxRetries).
		SetRetryWaitTime(RetryWait).
		SetRetryMaxWaitTime(MaxRetryWait).
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
		}).
		SetDebug(true).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() >= 500 || err != nil
		})

	return client
}

// CallAuthorizationAPI sends an HTTP request using the Resty client.
// It takes the context, client, payload, HTTP method, and URL as inputs. The function
// configures the request with the payload and executes it with the specified method and URL.
func CallAuthorizationAPI(ctx context.Context, payload Payload) (*resty.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, globalTimeout)
	defer cancel()

	if restyClient == nil {
		appError := apierrors.NewAppError("Resty Client initialization error", 500, nil)
		return nil, &appError
	}
	request := restyClient.R().
		SetBody(payload).
		SetHeader("Content-Type", "application/json").
		SetContext(ctx)

	return request.Execute(globalMethod, globalURL)
}
