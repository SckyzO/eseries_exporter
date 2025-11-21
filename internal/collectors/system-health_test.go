// MIT License
//
// Copyright (c) 2025 Contributors to the E-Series Exporter project
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package collector

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sckyzo/eseries_exporter/internal/config"
)

func TestSystemHealthCollector(t *testing.T) {
	healthData, err := os.ReadFile("testdata/system-health.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	expected := `
	# HELP eseries_system_status System health status (1=optimal, 2=degraded, 3=failed)
	# TYPE eseries_system_status gauge
	eseries_system_status{system_id="e5660-01",system_name="e5660-01",model="5600"} 1
	# HELP eseries_system_security_key_enabled Security key management enabled
	# TYPE eseries_system_security_key_enabled gauge
	eseries_system_security_key_enabled{system_id="e5660-01",system_name="e5660-01",model="5600"} 1
	# HELP eseries_system_drive_count Total number of drives in system
	# TYPE eseries_system_drive_count gauge
	eseries_system_drive_count{system_id="e5660-01",system_name="e5660-01",model="5600"} 180
	# HELP eseries_system_tray_count Number of drive trays
	# TYPE eseries_system_tray_count gauge
	eseries_system_tray_count{system_id="e5660-01",system_name="e5660-01",model="5600"} 3
	# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
	# TYPE eseries_exporter_collect_error gauge
	eseries_exporter_collect_error{collector="system-health"} 0
	`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write(healthData)
	}))
	defer server.Close()
	baseURL, _ := url.Parse(server.URL)
	target := config.Target{
		Name:       "test",
		User:       "test",
		Password:   "test",
		BaseURL:    baseURL,
		HttpClient: &http.Client{},
	}
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	collector := NewSystemHealthExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 22 { // 20 metrics + collect_error + collect_duration
		t.Errorf("Unexpected collection count %d, expected 22", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_system_status",
		"eseries_system_security_key_enabled",
		"eseries_system_drive_count",
		"eseries_system_tray_count",
		"eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSystemHealthCollectorError(t *testing.T) {
	expected := `
	# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
	# TYPE eseries_exporter_collect_error gauge
	eseries_exporter_collect_error{collector="system-health"} 1
	`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		http.Error(rw, "error", http.StatusNotFound)
	}))
	defer server.Close()
	baseURL, _ := url.Parse(server.URL)
	target := config.Target{
		Name:       "test",
		User:       "test",
		Password:   "test",
		BaseURL:    baseURL,
		HttpClient: &http.Client{},
	}
	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	collector := NewSystemHealthExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 2 {
		t.Errorf("Unexpected collection count %d, expected 2", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_system_status", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSystemHealthParseFunctions(t *testing.T) {
	// Test parseSizeToBytes
	testCases := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"", 0, false},
		{"0", 0, false},
		{"12345", 12345, false},
		{"invalid", 0, true},
	}

	for _, tc := range testCases {
		result, err := parseSizeToBytes(tc.input)
		if tc.hasError && err == nil {
			t.Errorf("Expected error for input %s, but got none", tc.input)
		}
		if !tc.hasError && result != tc.expected {
			t.Errorf("parseSizeToBytes(%s) = %d, expected %d", tc.input, result, tc.expected)
		}
	}

	// Test boolToFloat64
	if boolToFloat64(true) != 1.0 {
		t.Error("boolToFloat64(true) should return 1.0")
	}
	if boolToFloat64(false) != 0.0 {
		t.Error("boolToFloat64(false) should return 0.0")
	}
}
