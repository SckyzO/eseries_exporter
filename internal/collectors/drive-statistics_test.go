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

func TestDriveStatisticsCollector(t *testing.T) {
	analyzedDriveData, err := os.ReadFile("testdata/analysed-drive-statistics.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	driveData, err := os.ReadFile("testdata/drive-statistics.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	inventoryData, err := os.ReadFile("testdata/drives.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	expected := `# HELP eseries_drive_average_read_op_size_bytes Drive statistic averageReadOpSize
# TYPE eseries_drive_average_read_op_size_bytes gauge
eseries_drive_average_read_op_size_bytes{slot="58",tray="0"} 39620.99569760295
eseries_drive_average_read_op_size_bytes{slot="53",tray="0"} 21312.646464646463
# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
# TYPE eseries_exporter_collect_error gauge
eseries_exporter_collect_error{collector="drive-statistics"} 0
`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "hardware-inventory") {
			_, _ = rw.Write(inventoryData)
		} else if strings.HasSuffix(req.URL.Path, "analysed-drive-statistics") {
			_, _ = rw.Write(analyzedDriveData)
		} else {
			_, _ = rw.Write(driveData)
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
	collector := NewDriveStatisticsExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 48 {
		t.Errorf("Unexpected collection count %d, expected 48", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_drive_average_read_op_size_bytes",
		"eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestDriveStatisticsCollectorError(t *testing.T) {
	expected := `# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
# TYPE eseries_exporter_collect_error gauge
eseries_exporter_collect_error{collector="drive-statistics"} 1
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
	collector := NewDriveStatisticsExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 2 {
		t.Errorf("Unexpected collection count %d, expected 2", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_drive_average_read_op_size_bytes", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}