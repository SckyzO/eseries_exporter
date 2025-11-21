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
	"time"

	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sckyzo/eseries_exporter/internal/config"
)

type InterfaceStatistics struct {
	InterfaceID     string `json:"interfaceId"`
	Label           string
	TotalIops       float64 `json:"totalIops"`
	ReadIops        float64 `json:"readIops"`
	WriteIops       float64 `json:"writeIops"`
	TotalBytes      float64 `json:"totalBytes"`
	ReadBytes       float64 `json:"readBytes"`
	WriteBytes      float64 `json:"writeBytes"`
	ResponseTime    float64 `json:"responseTime"`
	ErrorsTotal     int     `json:"errorsTotal"`
	LinkUtilization float64 `json:"linkUtilization"`
}

type InterfaceStatisticsCollector struct {
	TotalIops       *prometheus.Desc
	ReadIops        *prometheus.Desc
	WriteIops       *prometheus.Desc
	TotalBytes      *prometheus.Desc
	ReadBytes       *prometheus.Desc
	WriteBytes      *prometheus.Desc
	ResponseTime    *prometheus.Desc
	ErrorsTotal     *prometheus.Desc
	LinkUtilization *prometheus.Desc
	target          config.Target
	logger          *slog.Logger
}

func init() {
	registerCollector("interface-statistics", true, NewInterfaceStatisticsExporter)
}

func NewInterfaceStatisticsExporter(target config.Target, logger *slog.Logger) Collector {
	labels := []string{"interface", "interface_label"}
	return &InterfaceStatisticsCollector{
		TotalIops: prometheus.NewDesc(prometheus.BuildFQName(namespace, "interface", "iops_total"),
			"Interface statistic totalIops", labels, nil),
		ReadIops: prometheus.NewDesc(prometheus.BuildFQName(namespace, "interface", "read_iops_total"),
			"Interface statistic readIops", labels, nil),
		WriteIops: prometheus.NewDesc(prometheus.BuildFQName(namespace, "interface", "write_iops_total"),
			"Interface statistic writeIops", labels, nil),
		TotalBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "interface", "bytes_total"),
			"Interface statistic totalBytes", labels, nil),
		ReadBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "interface", "read_bytes_total"),
			"Interface statistic readBytes", labels, nil),
		WriteBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "interface", "write_bytes_total"),
			"Interface statistic writeBytes", labels, nil),
		ResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "interface", "response_time_seconds"),
			"Interface statistic responseTime", labels, nil),
		ErrorsTotal: prometheus.NewDesc(prometheus.BuildFQName(namespace, "interface", "errors_total"),
			"Interface statistic errorsTotal", labels, nil),
		LinkUtilization: prometheus.NewDesc(prometheus.BuildFQName(namespace, "interface", "link_utilization_ratio"),
			"Interface statistic linkUtilization (0.0-1.0 ratio)", labels, nil),
		target: target,
		logger: logger,
	}
}

func (c *InterfaceStatisticsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.TotalIops
	ch <- c.ReadIops
	ch <- c.WriteIops
	ch <- c.TotalBytes
	ch <- c.ReadBytes
	ch <- c.WriteBytes
	ch <- c.ResponseTime
	ch <- c.ErrorsTotal
	ch <- c.LinkUtilization
}

func (c *InterfaceStatisticsCollector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Debug("Collecting interface statistics metrics")
	collectTime := time.Now()
	var errorMetric int
	statistics, err := c.collect()
	if err != nil {
		c.logger.Error("interface statistics collection error", "err", err)
		errorMetric = 1
	}

	for _, s := range statistics {
		ch <- prometheus.MustNewConstMetric(c.TotalIops, prometheus.CounterValue, s.TotalIops, s.InterfaceID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.ReadIops, prometheus.CounterValue, s.ReadIops, s.InterfaceID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.WriteIops, prometheus.CounterValue, s.WriteIops, s.InterfaceID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.TotalBytes, prometheus.CounterValue, s.TotalBytes, s.InterfaceID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.ReadBytes, prometheus.CounterValue, s.ReadBytes, s.InterfaceID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.WriteBytes, prometheus.CounterValue, s.WriteBytes, s.InterfaceID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.ResponseTime, prometheus.GaugeValue, s.ResponseTime, s.InterfaceID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.ErrorsTotal, prometheus.CounterValue, float64(s.ErrorsTotal), s.InterfaceID, s.Label)
		ch <- prometheus.MustNewConstMetric(c.LinkUtilization, prometheus.GaugeValue, s.LinkUtilization, s.InterfaceID, s.Label)
	}

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, float64(errorMetric), "interface-statistics")
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "interface-statistics")
}

func (c *InterfaceStatisticsCollector) collect() ([]InterfaceStatistics, error) {
	var statistics []InterfaceStatistics
	var body []byte
	var err error

	// Get analyzed interface statistics - these include calculated metrics
	body, err = getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/analyzed/interface-statistics?statisticsFetchTime=60", c.target.Name), c.logger)
	if err != nil {
		return nil, err
	}

	var objmap map[string]json.RawMessage
	err = json.Unmarshal(body, &objmap)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(objmap["statistics"], &statistics)
	if err != nil {
		return nil, err
	}

	// Set labels (interface ID is used as label for now)
	for i := range statistics {
		s := &statistics[i]
		s.Label = s.InterfaceID
	}

	return statistics, nil
}
