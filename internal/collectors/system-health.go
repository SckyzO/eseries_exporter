// Copyright 2025 Contributors to the E-Series Exporter project
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sckyzo/eseries_exporter/internal/config"
)

type SystemHealth struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Model                    string `json:"model"`
	ChassisSerialNumber      string `json:"chassisSerialNumber"`
	WWN                      string `json:"wwn"`
	FwVersion                string `json:"fwVersion"`
	NvsramVersion            string `json:"nvsramVersion"`
	BootVersion              string `json:"bootVersion"`
	AppVersion               string `json:"appVersion"`
	Status                   string `json:"status"`
	AsupEnabled              bool   `json:"asupEnabled"`
	PasswordSet              bool   `json:"passwordSet"`
	CertificateStatus        string `json:"certificateStatus"`
	DriveChannelPortDisabled bool   `json:"driveChannelPortDisabled"`
	RecoveryModeEnabled      bool   `json:"recoveryModeEnabled"`
	SecurityKeyEnabled       bool   `json:"securityKeyEnabled"`
	SimplexModeEnabled       bool   `json:"simplexModeEnabled"`
	DriveCount               int    `json:"driveCount"`
	TrayCount                int    `json:"trayCount"`
	UsedPoolSpace            string `json:"usedPoolSpace"`
	FreePoolSpace            string `json:"freePoolSpace"`
	HotSpareCount            int    `json:"hotSpareCount"`
	HostSpareCountInStandby  int    `json:"hostSpareCountInStandby"`
	HostSparesUsed           int    `json:"hostSparesUsed"`
	LastContacted            string `json:"lastContacted"`
	BootTime                 string `json:"bootTime"`
}

type SystemHealthCollector struct {
	SystemStatus        *prometheus.Desc
	FirmwareVersion     *prometheus.Desc
	NvsramVersion       *prometheus.Desc
	BootVersion         *prometheus.Desc
	AppVersion          *prometheus.Desc
	CertificateStatus   *prometheus.Desc
	SecurityKeyEnabled  *prometheus.Desc
	DriveCount          *prometheus.Desc
	TrayCount           *prometheus.Desc
	HotSpareCount       *prometheus.Desc
	HostSpareCount      *prometheus.Desc
	UsedPoolSpace       *prometheus.Desc
	FreePoolSpace       *prometheus.Desc
	RecoveryModeEnabled *prometheus.Desc
	SimplexModeEnabled  *prometheus.Desc
	AsupEnabled         *prometheus.Desc
	PasswordSet         *prometheus.Desc
	Uptime              *prometheus.Desc
	target              config.Target
	logger              log.Logger
}

func init() {
	registerCollector("system-health", true, NewSystemHealthExporter)
}

func NewSystemHealthExporter(target config.Target, logger log.Logger) Collector {
	labels := []string{"system_id", "system_name", "model"}
	return &SystemHealthCollector{
		SystemStatus: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "status"),
			"System health status (1=optimal, 2=degraded, 3=failed)", labels, nil),
		FirmwareVersion: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "firmware_version"),
			"System firmware version", labels, nil),
		NvsramVersion: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "nvsram_version"),
			"System NVSRAM version", labels, nil),
		BootVersion: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "boot_version"),
			"System boot version", labels, nil),
		AppVersion: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "app_version"),
			"System application version", labels, nil),
		CertificateStatus: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "certificate_status"),
			"SSL certificate status", labels, nil),
		SecurityKeyEnabled: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "security_key_enabled"),
			"Security key management enabled", labels, nil),
		DriveCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "drive_count"),
			"Total number of drives in system", labels, nil),
		TrayCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "tray_count"),
			"Number of drive trays", labels, nil),
		HotSpareCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "hot_spare_count"),
			"Number of hot spare drives", labels, nil),
		HostSpareCount: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "host_spare_count"),
			"Number of host spare drives in standby", labels, nil),
		UsedPoolSpace: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "used_pool_space_bytes"),
			"Used pool space in bytes", labels, nil),
		FreePoolSpace: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "free_pool_space_bytes"),
			"Free pool space in bytes", labels, nil),
		RecoveryModeEnabled: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "recovery_mode_enabled"),
			"System in recovery mode", labels, nil),
		SimplexModeEnabled: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "simplex_mode_enabled"),
			"System in simplex mode (single controller)", labels, nil),
		AsupEnabled: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "asup_enabled"),
			"AutoSupport feature enabled", labels, nil),
		PasswordSet: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "password_set"),
			"System password configured", labels, nil),
		Uptime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "system", "uptime_seconds"),
			"System uptime in seconds", labels, nil),
		target: target,
		logger: logger,
	}
}

func (c *SystemHealthCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.SystemStatus
	ch <- c.FirmwareVersion
	ch <- c.NvsramVersion
	ch <- c.BootVersion
	ch <- c.AppVersion
	ch <- c.CertificateStatus
	ch <- c.SecurityKeyEnabled
	ch <- c.DriveCount
	ch <- c.TrayCount
	ch <- c.HotSpareCount
	ch <- c.HostSpareCount
	ch <- c.UsedPoolSpace
	ch <- c.FreePoolSpace
	ch <- c.RecoveryModeEnabled
	ch <- c.SimplexModeEnabled
	ch <- c.AsupEnabled
	ch <- c.PasswordSet
	ch <- c.Uptime
}

func (c *SystemHealthCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "Collecting system-health metrics")
	collectTime := time.Now()
	var errorMetric int
	health, err := c.collect()
	if err != nil {
		level.Error(c.logger).Log("msg", err)
		errorMetric = 1
	}

	// Convert string space values to bytes
	usedSpace, _ := parseSizeToBytes(health.UsedPoolSpace)
	freeSpace, _ := parseSizeToBytes(health.FreePoolSpace)

	// Calculate uptime
	uptime := 0.0
	if health.BootTime != "" {
		if bootTime, err := time.Parse(time.RFC3339, health.BootTime); err == nil {
			uptime = time.Since(bootTime).Seconds()
		}
	}

	// Convert status to numeric value
	statusValue := 0.0
	switch health.Status {
	case "optimal":
		statusValue = 1.0
	case "degraded":
		statusValue = 2.0
	case "failed":
		statusValue = 3.0
	}

	ch <- prometheus.MustNewConstMetric(c.SystemStatus, prometheus.GaugeValue, statusValue, health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.FirmwareVersion, prometheus.GaugeValue, 1, health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.NvsramVersion, prometheus.GaugeValue, 1, health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.BootVersion, prometheus.GaugeValue, 1, health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.AppVersion, prometheus.GaugeValue, 1, health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.CertificateStatus, prometheus.GaugeValue, 1, health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.SecurityKeyEnabled, prometheus.GaugeValue, boolToFloat64(health.SecurityKeyEnabled), health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.DriveCount, prometheus.GaugeValue, float64(health.DriveCount), health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.TrayCount, prometheus.GaugeValue, float64(health.TrayCount), health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.HotSpareCount, prometheus.GaugeValue, float64(health.HotSpareCount), health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.HostSpareCount, prometheus.GaugeValue, float64(health.HostSpareCountInStandby), health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.UsedPoolSpace, prometheus.GaugeValue, float64(usedSpace), health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.FreePoolSpace, prometheus.GaugeValue, float64(freeSpace), health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.RecoveryModeEnabled, prometheus.GaugeValue, boolToFloat64(health.RecoveryModeEnabled), health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.SimplexModeEnabled, prometheus.GaugeValue, boolToFloat64(health.SimplexModeEnabled), health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.AsupEnabled, prometheus.GaugeValue, boolToFloat64(health.AsupEnabled), health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.PasswordSet, prometheus.GaugeValue, boolToFloat64(health.PasswordSet), health.ID, health.Name, health.Model)
	ch <- prometheus.MustNewConstMetric(c.Uptime, prometheus.GaugeValue, uptime, health.ID, health.Name, health.Model)

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, float64(errorMetric), "system-health")
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "system-health")
}

func (c *SystemHealthCollector) collect() (SystemHealth, error) {
	var health SystemHealth
	var body []byte
	var err error

	// Get system health information
	body, err = getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s", c.target.Name), c.logger)
	if err != nil {
		return health, err
	}

	err = json.Unmarshal(body, &health)
	if err != nil {
		return health, err
	}

	return health, nil
}

// Helper functions
func parseSizeToBytes(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, nil
	}
	// NetApp E-Series returns space as string, try to parse as int64
	return strconv.ParseInt(sizeStr, 10, 64)
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}
