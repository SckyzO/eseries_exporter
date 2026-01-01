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

func TestControllerStatisticsCollector(t *testing.T) {
	analyzedControllerData, err := os.ReadFile("testdata/analysed-controller-statistics.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	controllerData, err := os.ReadFile("testdata/controller-statistics.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	inventoryData, err := os.ReadFile("testdata/controllers.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	expected := `# HELP eseries_controller_average_read_op_size_bytes Controller statistic averageReadOpSize
# TYPE eseries_controller_average_read_op_size_bytes gauge
eseries_controller_average_read_op_size_bytes{controller="070000000000000000000001",controller_label="A"} 39687.27392305163
eseries_controller_average_read_op_size_bytes{controller="070000000000000000000002",controller_label="B"} 73664.54585344449
# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
# TYPE eseries_exporter_collect_error gauge
eseries_exporter_collect_error{collector="controller-statistics"} 0
`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "hardware-inventory") {
			_, _ = rw.Write(inventoryData)
		} else if strings.HasSuffix(req.URL.Path, "analyzed/controller-statistics") {
			_, _ = rw.Write(analyzedControllerData)
		} else {
			_, _ = rw.Write(controllerData)
		}
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
	collector := NewControllerStatisticsExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 56 {
		t.Errorf("Unexpected collection count %d, expected 56", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_controller_average_read_op_size_bytes",
		"eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestControllerStatisticsCollectorError(t *testing.T) {
	expected := `# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
# TYPE eseries_exporter_collect_error gauge
eseries_exporter_collect_error{collector="controller-statistics"} 1
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
	collector := NewControllerStatisticsExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 2 {
		t.Errorf("Unexpected collection count %d, expected 2", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_controller_average_read_op_size_bytes", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}