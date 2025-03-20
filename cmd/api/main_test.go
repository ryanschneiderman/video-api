package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ryanschneiderman/video-api/internal/app"
	"github.com/ryanschneiderman/video-api/internal/metrics"
)

func TestMetricsEndpointWithTraffic(t *testing.T) {
	// Set Gin to test mode.
	gin.SetMode(gin.TestMode)

	// Create a fresh custom registry.
	reg := prometheus.NewRegistry()
	// Create a new API metrics instance and register it.
	apiMetrics := metrics.NewAPIMetrics()
	apiMetrics.Register(reg)

	// Create a fake app instance.
	fApp := &app.App{} // Use a minimal stub or fake as needed.

	// Setup router using our helper with the custom registry and metrics instance.
	router := setupRouter(fApp, reg, apiMetrics)

	// Simulate an HTTP GET request to trigger the middleware.
	req1, err := http.NewRequest("GET", "/videos/test-id", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	rr1 := httptest.NewRecorder()
	router.ServeHTTP(rr1, req1)
	// We do not assert the response code for this endpoint since our focus is on metric updates.

	// Now simulate a GET request to /metrics.
	req2, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatalf("could not create /metrics request: %v", err)
	}
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)

	// Check for HTTP 200 OK.
	if rr2.Code != http.StatusOK {
		t.Errorf("expected status code 200 on /metrics, got %d", rr2.Code)
	}

	// Verify that the /metrics output contains our custom metric.
	metricsOutput := rr2.Body.String()
	if !strings.Contains(metricsOutput, "myapp_http_requests_total") {
		t.Errorf("expected /metrics output to contain 'myapp_http_requests_total', got: %s", metricsOutput)
	}

	// Optionally, check that the metric counter is non-zero by looking for label output.
	if !strings.Contains(metricsOutput, "myapp_http_requests_total{") {
		t.Errorf("expected /metrics output to include our counter labels, got: %s", metricsOutput)
	}
}

func TestVideoEndpointsExist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reg := prometheus.NewRegistry()
	apiMetrics := metrics.NewAPIMetrics()
	apiMetrics.Register(reg)

	fApp := &app.App{}
	router := setupRouter(fApp, reg, apiMetrics)

	// Test that POST /videos route is registered.
	req, err := http.NewRequest("POST", "/videos", nil)
	if err != nil {
		t.Fatalf("could not create POST /videos request: %v", err)
	}
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code == http.StatusNotFound {
		t.Errorf("POST /videos route not found, got %d", rr.Code)
	}

	// Test that GET /videos/:id route is registered.
	req, err = http.NewRequest("GET", "/videos/test-id", nil)
	if err != nil {
		t.Fatalf("could not create GET /videos/:id request: %v", err)
	}
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code == http.StatusNotFound {
		t.Errorf("GET /videos/:id route not found, got %d", rr.Code)
	}
}
