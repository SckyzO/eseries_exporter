package collector

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sckyzo/eseries_exporter/internal/config"
)

type StoragePool struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Label            string `json:"label"`
	TotalRaidedSpace string `json:"totalRaidedSpace"`
	UsedSpace        string `json:"usedSpace"`
	FreeSpace        string `json:"freeSpace"`
	RaidLevel        string `json:"raidLevel"`
	RaidStatus       string `json:"raidStatus"`
	State            string `json:"state"`
	DiskPool         bool   `json:"diskPool"`
	Offline          bool   `json:"offline"`
}

type StoragePoolsCollector struct {
	target config.Target
	logger *slog.Logger

	capacityBytes    *prometheus.Desc
	usedBytes        *prometheus.Desc
	freeBytes        *prometheus.Desc
	utilizationRatio *prometheus.Desc
	status           *prometheus.Desc
	state            *prometheus.Desc
	offline          *prometheus.Desc
}

func init() {
	registerCollector("storage-pools", false, NewStoragePoolsExporter)
}

func NewStoragePoolsExporter(target config.Target, logger *slog.Logger) Collector {
	logger = logger.With("collector", "storage-pools")

	return &StoragePoolsCollector{
		target: target,
		logger: logger,

		capacityBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "capacity_bytes"),
			"Total capacity of the storage pool in bytes",
			[]string{"pool", "raid_level", "status", "type"}, nil,
		),
		usedBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "used_bytes"),
			"Used capacity of the storage pool in bytes",
			[]string{"pool", "raid_level", "status"}, nil,
		),
		freeBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "free_bytes"),
			"Free capacity of the storage pool in bytes",
			[]string{"pool", "raid_level", "status"}, nil,
		),
		utilizationRatio: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "utilization_ratio"),
			"Utilization ratio of the storage pool (0-1)",
			[]string{"pool", "raid_level"}, nil,
		),
		status: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "status"),
			"Status of the storage pool (1 for optimal, 0 otherwise)",
			[]string{"pool", "raid_level", "status"}, nil,
		),
		state: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "state"),
			"Current state of the pool (1 for complete, 0 otherwise)",
			[]string{"pool", "raid_level", "state"}, nil,
		),
		offline: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "pool", "offline"),
			"Whether the pool is offline (1) or online (0)",
			[]string{"pool", "raid_level"}, nil,
		),
	}
}

func (c *StoragePoolsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.capacityBytes
	ch <- c.usedBytes
	ch <- c.freeBytes
	ch <- c.utilizationRatio
	ch <- c.status
	ch <- c.state
	ch <- c.offline
}

func (c *StoragePoolsCollector) Collect(ch chan<- prometheus.Metric) {
	pools, err := c.collectStoragePools()
	if err != nil {
		c.logger.Error("Collection failed", "error", err)
		return
	}

	for _, pool := range pools {
		poolType := "volume_group"
		if pool.DiskPool {
			poolType = "disk_pool"
		}

		// Parse capacities
		totalCapacity, _ := strconv.ParseFloat(pool.TotalRaidedSpace, 64)
		usedCapacity, _ := strconv.ParseFloat(pool.UsedSpace, 64)
		freeCapacity, _ := strconv.ParseFloat(pool.FreeSpace, 64)

		// Capacity
		ch <- prometheus.MustNewConstMetric(
			c.capacityBytes,
			prometheus.GaugeValue,
			totalCapacity,
			pool.Label, pool.RaidLevel, pool.RaidStatus, poolType,
		)

		// Used
		ch <- prometheus.MustNewConstMetric(
			c.usedBytes,
			prometheus.GaugeValue,
			usedCapacity,
			pool.Label, pool.RaidLevel, pool.RaidStatus,
		)

		// Free
		ch <- prometheus.MustNewConstMetric(
			c.freeBytes,
			prometheus.GaugeValue,
			freeCapacity,
			pool.Label, pool.RaidLevel, pool.RaidStatus,
		)

		// Utilization ratio
		utilizationRatio := 0.0
		if totalCapacity > 0 {
			utilizationRatio = usedCapacity / totalCapacity
		}
		ch <- prometheus.MustNewConstMetric(
			c.utilizationRatio,
			prometheus.GaugeValue,
			utilizationRatio,
			pool.Label, pool.RaidLevel,
		)

		// Status (1 for optimal, 0 otherwise)
		statusValue := 0.0
		if pool.RaidStatus == "optimal" {
			statusValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.status,
			prometheus.GaugeValue,
			statusValue,
			pool.Label, pool.RaidLevel, pool.RaidStatus,
		)

		// State (1 for complete, 0 otherwise)
		stateValue := 0.0
		if pool.State == "complete" {
			stateValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.state,
			prometheus.GaugeValue,
			stateValue,
			pool.Label, pool.RaidLevel, pool.State,
		)

		// Offline
		offlineValue := 0.0
		if pool.Offline {
			offlineValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.offline,
			prometheus.GaugeValue,
			offlineValue,
			pool.Label, pool.RaidLevel,
		)
	}
}

func (c *StoragePoolsCollector) collectStoragePools() ([]StoragePool, error) {
	poolsBody, err := getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/storage-pools", c.target.Name), c.logger)
	if err != nil {
		return nil, err
	}

	var pools []StoragePool
	if err := json.Unmarshal(poolsBody, &pools); err != nil {
		return nil, fmt.Errorf("failed to unmarshal storage pools: %w", err)
	}

	return pools, nil
}
