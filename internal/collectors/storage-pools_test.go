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

func TestStoragePoolsCollector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		data, _ := os.ReadFile("testdata/storage-pools-response.json")
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
	collector := NewStoragePoolsExporter(target, logger)

	// Test metrics collection
	expectedMetrics := 14 // 2 pools * 7 metrics each
	count := testutil.CollectAndCount(collector)
	if count != expectedMetrics {
		t.Errorf("Expected %d metrics, got %d", expectedMetrics, count)
	}

	// Test specific metrics
	expected := `
		# HELP eseries_pool_capacity_bytes Total capacity of the storage pool in bytes
		# TYPE eseries_pool_capacity_bytes gauge
		eseries_pool_capacity_bytes{pool="Pool_1",raid_level="raidDiskPool",status="optimal",type="disk_pool"} 1.02632538505216e+14
		eseries_pool_capacity_bytes{pool="Pool_2",raid_level="raid5",status="degraded",type="volume_group"} 5e+13
		# HELP eseries_pool_free_bytes Free capacity of the storage pool in bytes
		# TYPE eseries_pool_free_bytes gauge
		eseries_pool_free_bytes{pool="Pool_1",raid_level="raidDiskPool",status="optimal"} 0
		eseries_pool_free_bytes{pool="Pool_2",raid_level="raid5",status="degraded"} 2e+13
		# HELP eseries_pool_offline Whether the pool is offline (1) or online (0)
		# TYPE eseries_pool_offline gauge
		eseries_pool_offline{pool="Pool_1",raid_level="raidDiskPool"} 0
		eseries_pool_offline{pool="Pool_2",raid_level="raid5"} 0
		# HELP eseries_pool_state Current state of the pool (1 for complete, 0 otherwise)
		# TYPE eseries_pool_state gauge
		eseries_pool_state{pool="Pool_1",raid_level="raidDiskPool",state="complete"} 1
		eseries_pool_state{pool="Pool_2",raid_level="raid5",state="reconstructing"} 0
		# HELP eseries_pool_status Status of the storage pool (1 for optimal, 0 otherwise)
		# TYPE eseries_pool_status gauge
		eseries_pool_status{pool="Pool_1",raid_level="raidDiskPool",status="optimal"} 1
		eseries_pool_status{pool="Pool_2",raid_level="raid5",status="degraded"} 0
		# HELP eseries_pool_used_bytes Used capacity of the storage pool in bytes
		# TYPE eseries_pool_used_bytes gauge
		eseries_pool_used_bytes{pool="Pool_1",raid_level="raidDiskPool",status="optimal"} 1.02632538505216e+14
		eseries_pool_used_bytes{pool="Pool_2",raid_level="raid5",status="degraded"} 3e+13
		# HELP eseries_pool_utilization_ratio Utilization ratio of the storage pool (0-1)
		# TYPE eseries_pool_utilization_ratio gauge
		eseries_pool_utilization_ratio{pool="Pool_1",raid_level="raidDiskPool"} 1
		eseries_pool_utilization_ratio{pool="Pool_2",raid_level="raid5"} 0.6
	`

	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected)); err != nil {
		t.Errorf("Unexpected metrics:\n%v", err)
	}
}

func TestStoragePoolsCollectorError(t *testing.T) {
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
	collector := NewStoragePoolsExporter(target, logger)

	// Should return 0 metrics on error
	count := testutil.CollectAndCount(collector)
	if count != 0 {
		t.Errorf("Expected 0 metrics on error, got %d", count)
	}
}
