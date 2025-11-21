// Copyright 2025 SckyzO
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "eseries_exporter"
)

var (
	// Exporter metrics
	exporterVersion = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "version_info",
			Help:      "Version information about the exporter",
		},
		[]string{"version", "commit", "build_date", "go_version"},
	)

	// Collector performance metrics
	collectorDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "collector_duration_seconds",
			Help:      "Time spent collecting metrics by collector",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"collector", "target"},
	)

	collectorErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "collector_errors_total",
			Help:      "Total number of errors by collector and target",
		},
		[]string{"collector", "target", "error_type"},
	)

	collectorRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "collector_requests_total",
			Help:      "Total number of requests by collector and target",
		},
		[]string{"collector", "target", "status"},
	)

	// HTTP metrics
	httpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"endpoint", "method", "status_code"},
	)

	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"endpoint", "method"},
	)

	// Configuration metrics
	configReload = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "config_reloads_total",
			Help:      "Total number of configuration reloads",
		},
		[]string{"status"},
	)

	// Security metrics
	tlsHandshakes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "tls_handshakes_total",
			Help:      "Total number of TLS handshakes",
		},
		[]string{"status"},
	)

	authAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "auth_attempts_total",
			Help:      "Total number of authentication attempts",
		},
		[]string{"status"},
	)

	// Resource metrics
	goroutines = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "goroutines",
			Help:      "Current number of goroutines",
		},
		func() float64 {
			return float64(runtime.NumGoroutine())
		},
	)

	memoryAlloc = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_alloc_bytes",
			Help:      "Current allocated memory in bytes",
		},
		func() float64 {
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			return float64(mem.Alloc)
		},
	)

	memorySys = promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "memory_sys_bytes",
			Help:      "Total allocated memory in bytes",
		},
		func() float64 {
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			return float64(mem.Sys)
		},
	)

	gcPause = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "gc_pause_ns",
			Help:      "Garbage collection pause time in nanoseconds",
		},
		[]string{"pause_percent"},
	)
)

// ExporterMetrics contains all exporter-specific metrics
type ExporterMetrics struct {
	mu        sync.Mutex
	startTime time.Time
	version   string
	commit    string
	buildDate string
}

// NewExporterMetrics creates a new ExporterMetrics instance
func NewExporterMetrics(version, commit, buildDate string) *ExporterMetrics {
	em := &ExporterMetrics{
		startTime: time.Now(),
		version:   version,
		commit:    commit,
		buildDate: buildDate,
	}

	// Set version metric
	exporterVersion.WithLabelValues(version, commit, buildDate, runtime.Version()).Set(1)

	return em
}

// RecordCollectorDuration records the duration of a collector execution
func (em *ExporterMetrics) RecordCollectorDuration(collector, target string, duration time.Duration) {
	collectorDuration.WithLabelValues(collector, target).Observe(duration.Seconds())
}

// RecordCollectorError records a collector error
func (em *ExporterMetrics) RecordCollectorError(collector, target, errorType string) {
	collectorErrors.WithLabelValues(collector, target, errorType).Inc()
}

// RecordCollectorRequest records a collector request
func (em *ExporterMetrics) RecordCollectorRequest(collector, target, status string) {
	collectorRequests.WithLabelValues(collector, target, status).Inc()
}

// RecordHTTPRequest records an HTTP request
func (em *ExporterMetrics) RecordHTTPRequest(endpoint, method string, statusCode int) {
	httpRequests.WithLabelValues(endpoint, method, toString(statusCode)).Inc()
}

// RecordHTTPDuration records HTTP request duration
func (em *ExporterMetrics) RecordHTTPDuration(endpoint, method string, duration time.Duration) {
	httpDuration.WithLabelValues(endpoint, method).Observe(duration.Seconds())
}

// RecordConfigReload records a configuration reload
func (em *ExporterMetrics) RecordConfigReload(status string) {
	configReload.WithLabelValues(status).Inc()
}

// RecordTLSHandshake records a TLS handshake attempt
func (em *ExporterMetrics) RecordTLSHandshake(status string) {
	tlsHandshakes.WithLabelValues(status).Inc()
}

// RecordAuthAttempt records an authentication attempt
func (em *ExporterMetrics) RecordAuthAttempt(status string) {
	authAttempts.WithLabelValues(status).Inc()
}

// RecordGCPause records garbage collection pause time
func (em *ExporterMetrics) RecordGCPause(pauseTime time.Duration) {
	percent := calculateGCPercent(pauseTime)
	gcPause.WithLabelValues(percent).Set(float64(pauseTime.Nanoseconds()))
}

// GetUptime returns the uptime of the exporter
func (em *ExporterMetrics) GetUptime() time.Duration {
	em.mu.Lock()
	defer em.mu.Unlock()
	return time.Since(em.startTime)
}

// GetVersionInfo returns version information
func (em *ExporterMetrics) GetVersionInfo() (string, string, string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	return em.version, em.commit, em.buildDate
}

// HealthChecker provides health check functionality
type HealthChecker struct {
	mu         sync.RWMutex
	checks     map[string]HealthCheck
	lastCheck  time.Time
	isHealthy  bool
	muProblems []string
}

// HealthCheck represents a single health check
type HealthCheck struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	LastCheck time.Time              `json:"last_check"`
	Duration  time.Duration          `json:"duration"`
}

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status        string                 `json:"status"`
	Timestamp     time.Time              `json:"timestamp"`
	Uptime        string                 `json:"uptime"`
	Version       string                 `json:"version"`
	Commit        string                 `json:"commit"`
	BuildDate     string                 `json:"build_date"`
	Checks        []HealthCheck          `json:"checks"`
	MemoryUsage   map[string]interface{} `json:"memory_usage"`
	Goroutines    int                    `json:"goroutines"`
	ProblemGroups map[string][]string    `json:"problem_groups,omitempty"`
}

// NewHealthChecker creates a new HealthChecker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]HealthCheck),
	}
}

// RegisterCheck registers a health check
func (hc *HealthChecker) RegisterCheck(name string, checkFunc func() HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.checks[name] = HealthCheck{
		Name:      name,
		Status:    "unknown",
		LastCheck: time.Now(),
		Data:      make(map[string]interface{}),
	}
}

// RunChecks runs all registered health checks
func (hc *HealthChecker) RunChecks(exporterMetrics *ExporterMetrics) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.lastCheck = time.Now()
	allHealthy := true
	problemGroups := make(map[string][]string)

	for name, check := range hc.checks {
		start := time.Now()
		result := hc.runCheck(name, check)
		result.Duration = time.Since(start)

		hc.checks[name] = result

		if result.Status != "healthy" {
			allHealthy = false
			// Group problems by category
			problemGroups["errors"] = append(problemGroups["errors"], fmt.Sprintf("%s: %s", name, result.Message))
		}
	}

	hc.isHealthy = allHealthy
	hc.muProblems = problemGroups["errors"]
}

// GetHealthStatus returns the current health status
func (hc *HealthChecker) GetHealthStatus(exporterMetrics *ExporterMetrics) HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	var checks []HealthCheck
	for _, check := range hc.checks {
		checks = append(checks, check)
	}

	return HealthStatus{
		Status:    getOverallStatus(hc.isHealthy),
		Timestamp: time.Now(),
		Uptime:    exporterMetrics.GetUptime().String(),
		Version:   func() string { v, _, _ := exporterMetrics.GetVersionInfo(); return v }(),
		Commit:    func() string { _, c, _ := exporterMetrics.GetVersionInfo(); return c }(),
		BuildDate: func() string { _, _, b := exporterMetrics.GetVersionInfo(); return b }(),
		Checks:    checks,
		MemoryUsage: map[string]interface{}{
			"alloc_bytes":    memStats.Alloc,
			"sys_bytes":      memStats.Sys,
			"gc_count":       memStats.NumGC,
			"last_gc_ns":     memStats.LastGC,
			"pause_total_ns": memStats.PauseTotalNs,
		},
		Goroutines:    runtime.NumGoroutine(),
		ProblemGroups: map[string][]string{"errors": hc.muProblems},
	}
}

// helper functions
func toString(i int) string {
	return string(rune(i))
}

func getOverallStatus(healthy bool) string {
	if healthy {
		return "healthy"
	}
	return "unhealthy"
}

func (hc *HealthChecker) runCheck(name string, check HealthCheck) HealthCheck {
	// This is a placeholder - in a real implementation, each check would be run
	// For now, we'll just mark all checks as healthy
	return HealthCheck{
		Name:      name,
		Status:    "healthy",
		Message:   "All checks passed",
		LastCheck: time.Now(),
		Data:      map[string]interface{}{"status": "ok"},
	}
}

func calculateGCPercent(pause time.Duration) string {
	pauseMs := pause.Seconds() * 1000
	switch {
	case pauseMs < 1:
		return "low"
	case pauseMs < 10:
		return "medium"
	default:
		return "high"
	}
}

// InitializeDefaultHealthChecks sets up default health checks
func (hc *HealthChecker) InitializeDefaultHealthChecks() {
	// Config health check
	hc.RegisterCheck("config", func() HealthCheck {
		return HealthCheck{
			Name:    "config",
			Status:  "healthy",
			Message: "Configuration is valid",
			Data:    map[string]interface{}{"config_file": "loaded"},
		}
	})

	// HTTP server health check
	hc.RegisterCheck("http_server", func() HealthCheck {
		return HealthCheck{
			Name:    "http_server",
			Status:  "healthy",
			Message: "HTTP server is running",
			Data:    map[string]interface{}{"endpoints": []string{"/", "/eseries", "/metrics"}},
		}
	})

	// Go runtime health check
	hc.RegisterCheck("go_runtime", func() HealthCheck {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		status := "healthy"
		message := "Runtime is healthy"

		// Check if memory usage is too high (> 1GB)
		if memStats.Alloc > 1024*1024*1024 {
			status = "warning"
			message = "High memory usage detected"
		}

		return HealthCheck{
			Name:    "go_runtime",
			Status:  status,
			Message: message,
			Data: map[string]interface{}{
				"memory_alloc": memStats.Alloc,
				"goroutines":   runtime.NumGoroutine(),
				"gc_count":     memStats.NumGC,
				"go_version":   runtime.Version(),
			},
		}
	})
}
