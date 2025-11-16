package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHealthCheckNoRoute(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHealthCheckNoMethod(t *testing.T) {
	req := httptest.NewRequest("PUT", "/healthz", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestSwagger(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1/swagger/docs/index.html#/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	Router.Engine.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
