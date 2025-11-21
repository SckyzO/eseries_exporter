package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	collector "github.com/sckyzo/eseries_exporter/internal/collectors"
	"github.com/sckyzo/eseries_exporter/internal/config"
)

const (
	version = "2.0.0"
)

var (
	listenAddress = flag.String("web.listen-address", ":9313", "Address to listen on for web interface and telemetry.")
	metricsPath   = flag.String("web.metrics-path", "/metrics", "Path under which to expose metrics.")
	webPath       = flag.String("web.web-path", "/eseries", "Path under which to expose metrics.")
	configFile    = flag.String("config.file", "eseries_exporter.yaml", "Path to configuration file.")
	eseriesHost   = flag.String("eseries.host", "", "The host of the E-Series storage system.")
	eseriesUser   = flag.String("eseries.user", "admin", "The user of the E-Series storage system.")
	eseriesPass   = flag.String("eseries.pass", "admin", "The password of the E-Series storage system.")
	eseriesPort   = flag.String("eseries.port", "8443", "The port of the E-Series storage system.")
	logLevel      = flag.String("log.level", "info", "The logging level (debug, info, warn, error).")

	collectorsLog = make(map[string]*slog.Logger)
)

func main() {
	flag.Parse()

	// Setup logging
	logger := setupLogging(*logLevel)

	// Create metrics registry
	registry := prometheus.NewRegistry()

	// Initialize exporter metrics (optional for now)
	// exporterMetrics := metrics.NewExporterMetrics(version, "", "")

	// Load configuration
	var cfg config.SafeConfig
	if err := cfg.ReloadConfig(*configFile); err != nil {
		logger.Error("Failed to load config file", "err", err)
		os.Exit(1)
	}

	// Get target configuration
	var target config.Target
	if *eseriesHost != "" {
		// Use command line arguments
		target = createTargetFromFlags()
	} else {
		// Use configuration file - simple approach for now
		logger.Warn("Config file approach not fully implemented, using command line flags instead")
		target = createTargetFromFlags()
	}

	// Create HTTP client
	target.HttpClient = createHTTPClient(target)

	// Create logger with target context
	targetLogger := logger.With("target", target.Name)

	// Create the E-Series collector
	esc := collector.NewCollector(target, targetLogger)

	// Wrap EseriesCollector to implement prometheus.Collector
	collector := &prometheusCollector{esc: esc}

	// Register the collector
	if err := registry.Register(collector); err != nil {
		logger.Error("Failed to register collector", "err", err)
		os.Exit(1)
	}

	// Create metrics handler
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	// Setup HTTP handlers
	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html>
<head><title>E-Series Exporter</title></head>
<body>
<h1>E-Series Exporter</h1>
<p><a href='%s'>E-Series Metrics</a></p>
</body>
</html>`, *metricsPath)
	})
	http.Handle(*webPath, handler)

	// Health endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "OK\n")
	})

	// Start server
	logger.Info("Starting E-Series exporter", "version", version, "address", *listenAddress, "path", *metricsPath)

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		logger.Error("Failed to start server", "err", err)
		os.Exit(1)
	}
}

// prometheusCollector wraps EseriesCollector to implement prometheus.Collector interface
type prometheusCollector struct {
	esc *collector.EseriesCollector
}

func (c *prometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	// Send all metric descriptions
	for _, collector := range c.esc.Collectors {
		collector.Describe(ch)
	}
}

func (c *prometheusCollector) Collect(ch chan<- prometheus.Metric) {
	// Collect metrics from all sub-collectors
	for name, collector := range c.esc.Collectors {
		collectorLogger, exists := collectorsLog[name]
		if !exists {
			continue // Skip if no logger for this collector
		}
		collectorLogger.Debug("Collecting metrics", "collector", name)
		collector.Collect(ch)
	}
}

func setupLogging(level string) *slog.Logger {
	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}

func createTargetFromFlags() config.Target {
	var target config.Target
	target.Name = *eseriesHost
	target.User = *eseriesUser
	target.Password = *eseriesPass
	target.Collectors = nil // Use default collectors
	return target
}

func createHTTPClient(target config.Target) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // This should be configurable
		},
		Proxy: http.ProxyFromEnvironment,
	}
	return &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}
}
