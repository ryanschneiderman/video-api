package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ryanschneiderman/video-api/internal/metrics"
)

// fakeProcessor simulates the worker.Processor behavior.
type fakeProcessor struct {
	shouldError bool
}

func (f *fakeProcessor) ProcessMessages(ctx context.Context) error {
	// Simulate processing delay.
	time.Sleep(10 * time.Millisecond)
	if f.shouldError {
		return errors.New("simulated error")
	}
	return nil
}

// TestWorkerProcessingMetricsSuccess verifies that when processing succeeds,
// the processing duration metric is updated.
func TestWorkerProcessingMetricsSuccess(t *testing.T) {
	// Create a fresh registry and register worker metrics.
	reg := prometheus.NewRegistry()
	workerMetrics := metrics.NewWorkerMetrics()
	workerMetrics.Register(reg)

	// Create a fake processor that simulates successful processing.
	fp := &fakeProcessor{shouldError: false}
	ctx := context.Background()

	// Simulate one iteration of processing.
	start := time.Now()
	err := fp.ProcessMessages(ctx)
	duration := time.Since(start)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	workerMetrics.ProcessingDuration.Observe(duration.Seconds())

	// Now, simulate a /metrics request to verify that the metric is present.
	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatalf("could not create /metrics request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "worker_processing_duration_seconds") {
		t.Errorf("expected metrics output to contain 'worker_processing_duration_seconds', got: %s", body)
	}
}

// TestWorkerProcessingMetricsError verifies that when processing errors occur,
// the error counter metric is incremented.
func TestWorkerProcessingMetricsError(t *testing.T) {
	// Create a fresh registry and register worker metrics.
	reg := prometheus.NewRegistry()
	workerMetrics := metrics.NewWorkerMetrics()
	workerMetrics.Register(reg)

	// Create a fake processor that simulates an error.
	fp := &fakeProcessor{shouldError: true}
	ctx := context.Background()

	err := fp.ProcessMessages(ctx)
	if err == nil {
		t.Error("expected an error, got nil")
	}
	workerMetrics.Errors.Inc()

	// Simulate a /metrics request.
	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatalf("could not create /metrics request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "worker_errors_total") {
		t.Errorf("expected metrics output to contain 'worker_errors_total', got: %s", body)
	}
}
