package collector

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sckyzo/eseries_exporter/internal/config"
)

const (
	namespace = "eseries"
)

var (
	collectorState  = make(map[string]bool)
	factories       = make(map[string]func(target config.Target, logger *slog.Logger) Collector)
	collectDuration = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "exporter", "collector_duration_seconds"),
		"Collector time duration.",
		[]string{"collector"}, nil)
	collectError = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "exporter", "collect_error"),
		"Indicates if error has occurred during collection",
		[]string{"collector"}, nil)
)

type Collector interface {
	Describe(ch chan<- *prometheus.Desc)
	Collect(ch chan<- prometheus.Metric)
}

type EseriesCollector struct {
	Collectors map[string]Collector
}

func registerCollector(collector string, isDefaultEnabled bool, factory func(target config.Target, logger *slog.Logger) Collector) {
	collectorState[collector] = isDefaultEnabled
	factories[collector] = factory
}

func NewCollector(target config.Target, logger *slog.Logger) *EseriesCollector {
	collectors := make(map[string]Collector)
	for key, enabled := range collectorState {
		enable := false
		if target.Collectors == nil && enabled {
			enable = true
		} else if sliceContains(target.Collectors, key) {
			enable = true
		}
		
		if enable {
			// Create a child logger with collector context
			collectorLogger := logger.With("collector", key, "target", target.Name)
			collectors[key] = factories[key](target, collectorLogger)
		}
	}
	return &EseriesCollector{Collectors: collectors}
}

func sliceContains(slice []string, str string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, str) {
			return true
		}
	}
	return false
}

func getRequest(target config.Target, path string, logger *slog.Logger) ([]byte, error) {
	rel := &url.URL{Path: path}
	u := target.BaseURL.ResolveReference(rel)
	// We handle potential unescaping errors implicitly via URL parsing
	unescaped, err := url.PathUnescape(u.String())
	if err != nil {
		logger.Error("Failed to unescape URL", "url", u.String(), "error", err)
		// Fallback to original URL if unescape fails
		unescaped = u.String()
	}
	
	req, err := http.NewRequest("GET", unescaped, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(target.User, target.Password)

	logger.Debug("Performing GET request", "url", u.String())
	
	resp, err := target.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode != http.StatusOK {
		logger.Error("Response error", "code", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("%s", body)
	}
	return body, nil
}