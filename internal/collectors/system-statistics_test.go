package collector

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/sckyzo/eseries_exporter/internal/config"
)

func TestSystemStatisticsCollector(t *testing.T) {
	fixtureData, err := os.ReadFile("testdata/system-statistics.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	expected := `# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
# TYPE eseries_exporter_collect_error gauge
eseries_exporter_collect_error{collector="system-statistics"} 0
# HELP eseries_system_average_read_op_size_bytes System statistic averageReadOpSize
# TYPE eseries_system_average_read_op_size_bytes gauge
eseries_system_average_read_op_size_bytes 17357.11013434037
`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write(fixtureData)
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
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	collector := NewSystemStatisticsExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 14 {
		t.Errorf("Unexpected collection count %d, expected 14", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_system_average_read_op_size_bytes",
		"eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestSystemStatisticsCollectorError(t *testing.T) {
	expected := `# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
# TYPE eseries_exporter_collect_error gauge
eseries_exporter_collect_error{collector="system-statistics"} 1
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
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	collector := NewSystemStatisticsExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 2 {
		t.Errorf("Unexpected collection count %d, expected 2", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_system_average_read_op_size_bytes", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}
