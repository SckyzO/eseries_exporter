package collector

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sckyzo/eseries_exporter/internal/config"
)

type Volume struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Label            string          `json:"label"`
	Capacity         string          `json:"capacity"`
	TotalSizeInBytes string          `json:"totalSizeInBytes"`
	Status           string          `json:"status"`
	ThinProvisioned  bool            `json:"thinProvisioned"`
	ListOfMappings   []VolumeMapping `json:"listOfMappings"`
	VolumeGroupRef   string          `json:"volumeGroupRef"`
	DiskPool         bool            `json:"diskPool"`
	Offline          bool            `json:"offline"`
	Mapped           bool            `json:"mapped"`
	RaidLevel        string          `json:"raidLevel"`
	VolumeUse        string          `json:"volumeUse"`
}

type VolumeMapping struct {
	ID string `json:"id"`
}

type VolumesCollector struct {
	target config.Target
	logger *slog.Logger

	capacityBytes   *prometheus.Desc
	status          *prometheus.Desc
	mapped          *prometheus.Desc
	mappingsTotal   *prometheus.Desc
	thinProvisioned *prometheus.Desc
	offline         *prometheus.Desc
}

func init() {
	registerCollector("volumes", false, NewVolumesExporter)
}

func NewVolumesExporter(target config.Target, logger *slog.Logger) Collector {
	logger = logger.With("collector", "volumes")

	return &VolumesCollector{
		target: target,
		logger: logger,

		capacityBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "volume", "capacity_bytes"),
			"Total capacity of the volume in bytes",
			[]string{"volume", "pool", "status", "raid_level", "type"}, nil,
		),
		status: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "volume", "status"),
			"Status of the volume (1 for optimal, 0 otherwise)",
			[]string{"volume", "pool", "status"}, nil,
		),
		mapped: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "volume", "mapped"),
			"Whether the volume is mapped to a host (1) or not (0)",
			[]string{"volume", "pool"}, nil,
		),
		mappingsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "volume", "mappings_total"),
			"Number of host mappings for this volume",
			[]string{"volume", "pool"}, nil,
		),
		thinProvisioned: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "volume", "thin_provisioned"),
			"Whether the volume uses thin provisioning (1) or not (0)",
			[]string{"volume", "pool"}, nil,
		),
		offline: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "volume", "offline"),
			"Whether the volume is offline (1) or online (0)",
			[]string{"volume", "pool"}, nil,
		),
	}
}

func (c *VolumesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.capacityBytes
	ch <- c.status
	ch <- c.mapped
	ch <- c.mappingsTotal
	ch <- c.thinProvisioned
	ch <- c.offline
}

func (c *VolumesCollector) Collect(ch chan<- prometheus.Metric) {
	volumes, err := c.collectVolumes()
	if err != nil {
		c.logger.Error("Collection failed", "error", err)
		return
	}

	for _, volume := range volumes {
		poolLabel := c.getPoolLabel(volume.VolumeGroupRef)

		// Capacity
		capacity, _ := strconv.ParseFloat(volume.TotalSizeInBytes, 64)
		ch <- prometheus.MustNewConstMetric(
			c.capacityBytes,
			prometheus.GaugeValue,
			capacity,
			volume.Label, poolLabel, volume.Status, volume.RaidLevel, volume.VolumeUse,
		)

		// Status (1 for optimal, 0 otherwise)
		statusValue := 0.0
		if volume.Status == "optimal" {
			statusValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.status,
			prometheus.GaugeValue,
			statusValue,
			volume.Label, poolLabel, volume.Status,
		)

		// Mapped
		mappedValue := 0.0
		if volume.Mapped {
			mappedValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.mapped,
			prometheus.GaugeValue,
			mappedValue,
			volume.Label, poolLabel,
		)

		// Mappings total
		ch <- prometheus.MustNewConstMetric(
			c.mappingsTotal,
			prometheus.GaugeValue,
			float64(len(volume.ListOfMappings)),
			volume.Label, poolLabel,
		)

		// Thin provisioned
		thinValue := 0.0
		if volume.ThinProvisioned {
			thinValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.thinProvisioned,
			prometheus.GaugeValue,
			thinValue,
			volume.Label, poolLabel,
		)

		// Offline
		offlineValue := 0.0
		if volume.Offline {
			offlineValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(
			c.offline,
			prometheus.GaugeValue,
			offlineValue,
			volume.Label, poolLabel,
		)
	}
}

func (c *VolumesCollector) collectVolumes() ([]Volume, error) {
	volumesBody, err := getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/volumes", c.target.Name), c.logger)
	if err != nil {
		return nil, err
	}

	var volumes []Volume
	if err := json.Unmarshal(volumesBody, &volumes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal volumes: %w", err)
	}

	return volumes, nil
}

func (c *VolumesCollector) getPoolLabel(poolRef string) string {
	if poolRef == "" {
		return "unknown"
	}
	return poolRef
}
