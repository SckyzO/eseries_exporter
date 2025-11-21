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

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sckyzo/eseries_exporter/internal/config"
)

type VolumeStatistics struct {
	VolumeID     string `json:"volumeId"`
	VolumeName   string `json:"volumeName"`
	Label        string
	TotalIops    float64 `json:"totalIops"`
	ReadIops     float64 `json:"readIops"`
	WriteIops    float64 `json:"writeIops"`
	TotalBytes   float64 `json:"totalBytes"`
	ReadBytes    float64 `json:"readBytes"`
	WriteBytes   float64 `json:"writeBytes"`
	ResponseTime float64 `json:"responseTime"`
	CacheHits    int     `json:"cacheHits"`
	CacheMisses  int     `json:"cacheMisses"`
	LatencyMs    float64 `json:"latencyMs"`
}

type VolumeStatisticsCollector struct {
	TotalIops     *prometheus.Desc
	ReadIops      *prometheus.Desc
	WriteIops     *prometheus.Desc
	TotalBytes    *prometheus.Desc
	ReadBytes     *prometheus.Desc
	WriteBytes    *prometheus.Desc
	ResponseTime  *prometheus.Desc
	CacheHits     *prometheus.Desc
	CacheMisses   *prometheus.Desc
	LatencyMs     *prometheus.Desc
	CacheHitRatio *prometheus.Desc
	target        config.Target
	logger        log.Logger
}

func init() {
	registerCollector("volume-statistics", true, NewVolumeStatisticsExporter)
}

func NewVolumeStatisticsExporter(target config.Target, logger log.Logger) Collector {
	labels := []string{"volume", "volume_name"}
	return &VolumeStatisticsCollector{
		TotalIops: prometheus.NewDesc(prometheus.BuildFQName(namespace, "volume", "iops_total"),
			"Volume statistic totalIops", labels, nil),
		ReadIops: prometheus.NewDesc(prometheus.BuildFQName(namespace, "volume", "read_iops_total"),
			"Volume statistic readIops", labels, nil),
		WriteIops: prometheus.NewDesc(prometheus.BuildFQName(namespace, "volume", "write_iops_total"),
			"Volume statistic writeIops", labels, nil),
		TotalBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "volume", "bytes_total"),
			"Volume statistic totalBytes", labels, nil),
		ReadBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "volume", "read_bytes_total"),
			"Volume statistic readBytes", labels, nil),
		WriteBytes: prometheus.NewDesc(prometheus.BuildFQName(namespace, "volume", "write_bytes_total"),
			"Volume statistic writeBytes", labels, nil),
		ResponseTime: prometheus.NewDesc(prometheus.BuildFQName(namespace, "volume", "response_time_seconds"),
			"Volume statistic responseTime", labels, nil),
		CacheHits: prometheus.NewDesc(prometheus.BuildFQName(namespace, "volume", "cache_hits_total"),
			"Volume statistic cacheHits", labels, nil),
		CacheMisses: prometheus.NewDesc(prometheus.BuildFQName(namespace, "volume", "cache_misses_total"),
			"Volume statistic cacheMisses", labels, nil),
		LatencyMs: prometheus.NewDesc(prometheus.BuildFQName(namespace, "volume", "latency_milliseconds"),
			"Volume statistic latencyMs", labels, nil),
		CacheHitRatio: prometheus.NewDesc(prometheus.BuildFQName(namespace, "volume", "cache_hit_ratio"),
			"Volume cache hit ratio (0.0-1.0)", labels, nil),
		target: target,
		logger: logger,
	}
}

func (c *VolumeStatisticsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.TotalIops
	ch <- c.ReadIops
	ch <- c.WriteIops
	ch <- c.TotalBytes
	ch <- c.ReadBytes
	ch <- c.WriteBytes
	ch <- c.ResponseTime
	ch <- c.CacheHits
	ch <- c.CacheMisses
	ch <- c.LatencyMs
	ch <- c.CacheHitRatio
}

func (c *VolumeStatisticsCollector) Collect(ch chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "Collecting volume-statistics metrics")
	collectTime := time.Now()
	var errorMetric int
	statistics, err := c.collect()
	if err != nil {
		level.Error(c.logger).Log("msg", err)
		errorMetric = 1
	}

	for _, s := range statistics {
		// Calculate cache hit ratio
		totalRequests := s.CacheHits + s.CacheMisses
		cacheHitRatio := 0.0
		if totalRequests > 0 {
			cacheHitRatio = float64(s.CacheHits) / float64(totalRequests)
		}

		ch <- prometheus.MustNewConstMetric(c.TotalIops, prometheus.CounterValue, s.TotalIops, s.VolumeID, s.VolumeName)
		ch <- prometheus.MustNewConstMetric(c.ReadIops, prometheus.CounterValue, s.ReadIops, s.VolumeID, s.VolumeName)
		ch <- prometheus.MustNewConstMetric(c.WriteIops, prometheus.CounterValue, s.WriteIops, s.VolumeID, s.VolumeName)
		ch <- prometheus.MustNewConstMetric(c.TotalBytes, prometheus.CounterValue, s.TotalBytes, s.VolumeID, s.VolumeName)
		ch <- prometheus.MustNewConstMetric(c.ReadBytes, prometheus.CounterValue, s.ReadBytes, s.VolumeID, s.VolumeName)
		ch <- prometheus.MustNewConstMetric(c.WriteBytes, prometheus.CounterValue, s.WriteBytes, s.VolumeID, s.VolumeName)
		ch <- prometheus.MustNewConstMetric(c.ResponseTime, prometheus.GaugeValue, s.ResponseTime, s.VolumeID, s.VolumeName)
		ch <- prometheus.MustNewConstMetric(c.CacheHits, prometheus.CounterValue, float64(s.CacheHits), s.VolumeID, s.VolumeName)
		ch <- prometheus.MustNewConstMetric(c.CacheMisses, prometheus.CounterValue, float64(s.CacheMisses), s.VolumeID, s.VolumeName)
		ch <- prometheus.MustNewConstMetric(c.LatencyMs, prometheus.GaugeValue, s.LatencyMs, s.VolumeID, s.VolumeName)
		ch <- prometheus.MustNewConstMetric(c.CacheHitRatio, prometheus.GaugeValue, cacheHitRatio, s.VolumeID, s.VolumeName)
	}

	ch <- prometheus.MustNewConstMetric(collectError, prometheus.GaugeValue, float64(errorMetric), "volume-statistics")
	ch <- prometheus.MustNewConstMetric(collectDuration, prometheus.GaugeValue, time.Since(collectTime).Seconds(), "volume-statistics")
}

func (c *VolumeStatisticsCollector) collect() ([]VolumeStatistics, error) {
	var statistics []VolumeStatistics
	var body []byte
	var err error

	// Get analyzed volume statistics - these include calculated metrics
	body, err = getRequest(c.target, fmt.Sprintf("/devmgr/v2/storage-systems/%s/analyzed/volume-statistics?statisticsFetchTime=60", c.target.Name), c.logger)
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

	// Set labels
	for i := range statistics {
		s := &statistics[i]
		s.Label = s.VolumeName
	}

	return statistics, nil
}
