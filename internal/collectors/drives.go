package collector

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sckyzo/eseries_exporter/internal/config"
)

var (
	driveStatuses = []string{
		"optimal",
		"failed",
		"replaced",
		"bypassed",
		"unresponsive",
		"removed",
		"incompatible",
		"dataRelocation",
		"preFailCopy",
		"preFailCopyPending",
		"__UNDEFINED",
	}
)

type DrivesInventory struct {
	Drives []Drive `json:"drives"`
	Trays  []Tray  `json:"trays"`
}

type Drive struct {
	ID               string                `json:"id"`
	Status           string                `json:"status"`
	PhysicalLocation DrivePhysicalLocation `json:"physicalLocation"`
	TrayID           string
	Slot             string
}

type DrivePhysicalLocation struct {
	Slot    int    `json:"slot"`
	TrayRef string `json:"trayRef"`
}

type DrivesCollector struct {
	Status *prometheus.Desc
	target config.Target
	logger *slog.Logger
}

func init() {
	registerCollector("drives", true, NewDrivesExporter)
}

func NewDrivesExporter(target config.Target, logger *slog.Logger) Collector {
	return &DrivesCollector{
		Status: prometheus.NewDesc(prometheus.BuildFQName(namespace, "drive", "status"),
			"Drive status", []string{"tray", "slot", "status"}, nil),
		target: target,
		logger: logger,
	}
}

func (c *DrivesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Status
}

func (c *DrivesCollector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Debug("Collecting drives metrics")
	collectTime := time.Now()
	var errorMetric int
	metrics, err := c.collect()
	if err != nil {
		c.logger.Error("Collection failed", "error", err)
		errorMetric = 1
	}

	trays := make(map[string]int)
	for _, t := range metrics.Trays {
		trays[t.TrayRef] = t.ID
	}
	var ids []string
	for _, d := range metrics.Drives {
		if trayId, ok := trays[d.PhysicalLocation.TrayRef]; ok {
			d.TrayID = strconv.Itoa(trayId)
		}
		d.Slot = strconv.Itoa(d.PhysicalLocation.Slot)
		id := fmt.Sprintf("%s-%s", d.TrayID, d.Slot)
		if sliceContains(ids, id) {
			c.logger.Error("Duplicate drive entry detected, skipping", "tray", d.TrayID, "slot", d.Slot, "status", d.Status)
			errorMetric = 1
			continue
		}
		ids = append(ids, id)
		for _, driveStatus := range driveStatuses {
			var value float64
			if driveStatus == d.Status {
				value = 1
			}
			ch <- prometheus.MustNewConstMetric(c.Status, prometheus.GaugeValue, value, d.TrayID, d.Slot, driveStatus)
		}
		var unknown float64
		if !sliceContains(driveStatuses, d.Status) {
			unknown = 1
		}
		ch <- prometheus.MustNewConstMetric(c.Status, prometheus.GaugeValue, unknown, d.TrayID, d.Slot, "unknown")
	}

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, float64(errorMetric), "drives")
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "drives")
}

func (c *DrivesCollector) collect() (DrivesInventory, error) {
	var metrics DrivesInventory
	body, err := getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/hardware-inventory", c.target.Name), c.logger)
	if err != nil {
		return metrics, err
	}
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		return metrics, err
	}
	if len(metrics.Drives) == 0 {
		return metrics, fmt.Errorf("No drives returned")
	}
	return metrics, nil
}
