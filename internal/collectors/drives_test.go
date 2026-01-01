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

func TestDrivesCollector(t *testing.T) {
	fixtureData, err := os.ReadFile("testdata/drives.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	expected := `# HELP eseries_drive_status Drive status
# TYPE eseries_drive_status gauge
eseries_drive_status{slot="53",status="bypassed",tray="0"} 0
eseries_drive_status{slot="53",status="dataRelocation",tray="0"} 0
eseries_drive_status{slot="53",status="failed",tray="0"} 1
eseries_drive_status{slot="53",status="incompatible",tray="0"} 0
eseries_drive_status{slot="53",status="optimal",tray="0"} 0
eseries_drive_status{slot="53",status="preFailCopy",tray="0"} 0
eseries_drive_status{slot="53",status="preFailCopyPending",tray="0"} 0
eseries_drive_status{slot="53",status="removed",tray="0"} 0
eseries_drive_status{slot="53",status="replaced",tray="0"} 0
eseries_drive_status{slot="53",status="unknown",tray="0"} 0
eseries_drive_status{slot="53",status="unresponsive",tray="0"} 0
eseries_drive_status{slot="53",status="__UNDEFINED",tray="0"} 0
eseries_drive_status{slot="58",status="bypassed",tray="0"} 0
eseries_drive_status{slot="58",status="dataRelocation",tray="0"} 0
eseries_drive_status{slot="58",status="failed",tray="0"} 0
eseries_drive_status{slot="58",status="incompatible",tray="0"} 0
eseries_drive_status{slot="58",status="optimal",tray="0"} 1
eseries_drive_status{slot="58",status="preFailCopy",tray="0"} 0
eseries_drive_status{slot="58",status="preFailCopyPending",tray="0"} 0
eseries_drive_status{slot="58",status="removed",tray="0"} 0
eseries_drive_status{slot="58",status="replaced",tray="0"} 0
eseries_drive_status{slot="58",status="unknown",tray="0"} 0
eseries_drive_status{slot="58",status="unresponsive",tray="0"} 0
eseries_drive_status{slot="58",status="__UNDEFINED",tray="0"} 0
# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
# TYPE eseries_exporter_collect_error gauge
eseries_exporter_collect_error{collector="drives"} 0
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
	collector := NewDrivesExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 26 {
		t.Errorf("Unexpected collection count %d, expected 26", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_drive_status", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestDrivesCollectorDuplicates(t *testing.T) {
	fixtureData, err := os.ReadFile("testdata/drives-duplicate.json")
	if err != nil {
		t.Fatalf("Error loading fixture data: %s", err.Error())
	}
	expected := `# HELP eseries_drive_status Drive status
# TYPE eseries_drive_status gauge
eseries_drive_status{slot="53",status="bypassed",tray="0"} 0
eseries_drive_status{slot="53",status="dataRelocation",tray="0"} 0
eseries_drive_status{slot="53",status="failed",tray="0"} 1
eseries_drive_status{slot="53",status="incompatible",tray="0"} 0
eseries_drive_status{slot="53",status="optimal",tray="0"} 0
eseries_drive_status{slot="53",status="preFailCopy",tray="0"} 0
eseries_drive_status{slot="53",status="preFailCopyPending",tray="0"} 0
eseries_drive_status{slot="53",status="removed",tray="0"} 0
eseries_drive_status{slot="53",status="replaced",tray="0"} 0
eseries_drive_status{slot="53",status="unknown",tray="0"} 0
eseries_drive_status{slot="53",status="unresponsive",tray="0"} 0
eseries_drive_status{slot="53",status="__UNDEFINED",tray="0"} 0
eseries_drive_status{slot="58",status="bypassed",tray="0"} 0
eseries_drive_status{slot="58",status="dataRelocation",tray="0"} 0
eseries_drive_status{slot="58",status="failed",tray="0"} 0
eseries_drive_status{slot="58",status="incompatible",tray="0"} 0
eseries_drive_status{slot="58",status="optimal",tray="0"} 1
eseries_drive_status{slot="58",status="preFailCopy",tray="0"} 0
eseries_drive_status{slot="58",status="preFailCopyPending",tray="0"} 0
eseries_drive_status{slot="58",status="removed",tray="0"} 0
eseries_drive_status{slot="58",status="replaced",tray="0"} 0
eseries_drive_status{slot="58",status="unknown",tray="0"} 0
eseries_drive_status{slot="58",status="unresponsive",tray="0"} 0
eseries_drive_status{slot="58",status="__UNDEFINED",tray="0"} 0
# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
# TYPE eseries_exporter_collect_error gauge
eseries_exporter_collect_error{collector="drives"} 1
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
	collector := NewDrivesExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 26 {
		t.Errorf("Unexpected collection count %d, expected 26", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_drive_status", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}

func TestDrivesCollectorError(t *testing.T) {
	expected := `# HELP eseries_exporter_collect_error Indicates if error has occurred during collection
# TYPE eseries_exporter_collect_error gauge
eseries_exporter_collect_error{collector="drives"} 1
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
	collector := NewDrivesExporter(target, logger)
	gatherers := setupGatherer(collector)
	if val, err := testutil.GatherAndCount(gatherers); err != nil {
		t.Errorf("Unexpected error: %v", err)
	} else if val != 2 {
		t.Errorf("Unexpected collection count %d, expected 2", val)
	}
	if err := testutil.GatherAndCompare(gatherers, strings.NewReader(expected),
		"eseries_drive_status", "eseries_exporter_collect_error"); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}
}