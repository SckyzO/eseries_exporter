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

func TestStorageSystemsCollector(t *testing.T) {
	fixtureData, err := os.ReadFile("testdata/storage-systems.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	expected := `# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
# TYPE eseries_exporter_collect_error gauge
eseries_exporter_collect_error{collector="storage-systems"} 0
# HELP eseries_storage_system_status Storage System status, 1=optimal 0=all other states
# TYPE eseries_storage_system_status gauge
eseries_storage_system_status{status="lockDown"} 0
eseries_storage_system_status{status="needsAttn"} 0
eseries_storage_system_status{status="neverContacted"} 0
eseries_storage_system_status{status="newDevice"} 0
eseries_storage_system_status{status="offline"} 0
eseries_storage_system_status{status="optimal"} 1
eseries_storage_system_status{status="removed"} 0
eseries_storage_system_status{status="unknown"} 0
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
	collector := NewStorageSystemsExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 10 {
		t.Errorf("Unexpected collection count %d, expected 10", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_storage_system_status", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestStorageSystemsCollectorError(t *testing.T) {
	expected := `# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
# TYPE eseries_exporter_collect_error gauge
eseries_exporter_collect_error{collector="storage-systems"} 1
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
	collector := NewStorageSystemsExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 2 {
		t.Errorf("Unexpected collection count %d, expected 2", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_storage_system_status", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}