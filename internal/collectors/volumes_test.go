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

func TestVolumesCollector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		data, _ := os.ReadFile("testdata/volumes-response.json")
		rw.Write(data)
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	target := config.Target{
		Name:       "test-array",
		BaseURL:    baseURL,
		HttpClient: server.Client(),
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	collector := NewVolumesExporter(target, logger)

	// Test metrics collection
	expectedMetrics := 18 // 3 volumes * 6 metrics each
	count := testutil.CollectAndCount(collector)
	if count != expectedMetrics {
		t.Errorf("Expected %d metrics, got %d", expectedMetrics, count)
	}

	// Test specific metrics
	expected := `
		# HELP eseries_volume_capacity_bytes Total capacity of the volume in bytes
		# TYPE eseries_volume_capacity_bytes gauge
		eseries_volume_capacity_bytes{pool="040000006D039EA000CF32BB000000D868E4C6E2",raid_level="raid6",status="optimal",type="standardVolume",volume="Volume_1"} 5.1316269252608e+13
		eseries_volume_capacity_bytes{pool="040000006D039EA000CF32BB000000D868E4C6E2",raid_level="raid6",status="optimal",type="thinVolume",volume="Volume_2"} 5.1316269252608e+13
		eseries_volume_capacity_bytes{pool="040000006D039EA000CF32BB000000D868E4C6E2",raid_level="raid6",status="failed",type="standardVolume",volume="Volume_3"} 1.073741824e+10
		# HELP eseries_volume_mapped Whether the volume is mapped to a host (1) or not (0)
		# TYPE eseries_volume_mapped gauge
		eseries_volume_mapped{pool="040000006D039EA000CF32BB000000D868E4C6E2",volume="Volume_1"} 1
		eseries_volume_mapped{pool="040000006D039EA000CF32BB000000D868E4C6E2",volume="Volume_2"} 0
		eseries_volume_mapped{pool="040000006D039EA000CF32BB000000D868E4C6E2",volume="Volume_3"} 0
		# HELP eseries_volume_mappings_total Number of host mappings for this volume
		# TYPE eseries_volume_mappings_total gauge
		eseries_volume_mappings_total{pool="040000006D039EA000CF32BB000000D868E4C6E2",volume="Volume_1"} 1
		eseries_volume_mappings_total{pool="040000006D039EA000CF32BB000000D868E4C6E2",volume="Volume_2"} 0
		eseries_volume_mappings_total{pool="040000006D039EA000CF32BB000000D868E4C6E2",volume="Volume_3"} 0
		# HELP eseries_volume_offline Whether the volume is offline (1) or online (0)
		# TYPE eseries_volume_offline gauge
		eseries_volume_offline{pool="040000006D039EA000CF32BB000000D868E4C6E2",volume="Volume_1"} 0
		eseries_volume_offline{pool="040000006D039EA000CF32BB000000D868E4C6E2",volume="Volume_2"} 0
		eseries_volume_offline{pool="040000006D039EA000CF32BB000000D868E4C6E2",volume="Volume_3"} 1
		# HELP eseries_volume_status Status of the volume (1 for optimal, 0 otherwise)
		# TYPE eseries_volume_status gauge
		eseries_volume_status{pool="040000006D039EA000CF32BB000000D868E4C6E2",status="optimal",volume="Volume_1"} 1
		eseries_volume_status{pool="040000006D039EA000CF32BB000000D868E4C6E2",status="optimal",volume="Volume_2"} 1
		eseries_volume_status{pool="040000006D039EA000CF32BB000000D868E4C6E2",status="failed",volume="Volume_3"} 0
		# HELP eseries_volume_thin_provisioned Whether the volume uses thin provisioning (1) or not (0)
		# TYPE eseries_volume_thin_provisioned gauge
		eseries_volume_thin_provisioned{pool="040000006D039EA000CF32BB000000D868E4C6E2",volume="Volume_1"} 0
		eseries_volume_thin_provisioned{pool="040000006D039EA000CF32BB000000D868E4C6E2",volume="Volume_2"} 1
		eseries_volume_thin_provisioned{pool="040000006D039EA000CF32BB000000D868E4C6E2",volume="Volume_3"} 0
	`

	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected)); err != nil {
		t.Errorf("Unexpected metrics:\n%v", err)
	}
}

func TestVolumesCollectorError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("error\n"))
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	target := config.Target{
		Name:       "test-array",
		BaseURL:    baseURL,
		HttpClient: server.Client(),
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	collector := NewVolumesExporter(target, logger)

	// Should return 0 metrics on error
	count := testutil.CollectAndCount(collector)
	if count != 0 {
		t.Errorf("Expected 0 metrics on error, got %d", count)
	}
}
