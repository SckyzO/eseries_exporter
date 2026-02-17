package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"

	collector "github.com/sckyzo/eseries_exporter/internal/collectors"
	"github.com/sckyzo/eseries_exporter/internal/config"
)

var (
	configFile    = kingpin.Flag("config.file", "Path to exporter config file").Default("eseries_exporter.yaml").String()
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9313").String()
	logLevel      = kingpin.Flag("log.level", "Log level (debug, info, warn, error)").Default("info").String()
	logFormat     = kingpin.Flag("log.format", "Log format (text, json)").Default("text").String()
)

func metricsHandler(c *config.Config, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := r.URL.Query().Get("target")
		if t == "" {
			http.Error(w, "'target' parameter must be specified", http.StatusBadRequest)
			return
		}
		m := r.URL.Query().Get("module")
		if m == "" {
			m = "default"
		}
		module, ok := c.Modules[m]
		if !ok {
			http.Error(w, fmt.Sprintf("Unknown module %s", m), http.StatusNotFound)
			return
		}

		proxyURL, err := url.Parse(module.ProxyURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to parse ProxyURL %s", module.ProxyURL), http.StatusNotFound)
			logger.Error("Unable to parse ProxyURL", "url", module.ProxyURL, "error", err)
			return
		}

		target := config.Target{
			Name:       t,
			User:       module.User,
			Password:   module.Password,
			ProxyURL:   module.ProxyURL,
			BaseURL:    proxyURL,
			Collectors: module.Collectors,
		}

		httpClient := &http.Client{
			Timeout: time.Duration(module.Timeout) * time.Second,
		}

		if proxyURL.Scheme == "https" {
			logger.Debug("Setting up SSL transport", "url", module.ProxyURL)
			rootCAs, err := x509.SystemCertPool()
			if err != nil {
				logger.Error("Error loading system cert pool, creating empty cert pool", "error", err)
				rootCAs = x509.NewCertPool()
			}
			if module.RootCA != "" {
				certs, err := os.ReadFile(module.RootCA)
				if err != nil {
					logger.Error("Error loading root CA", "rootCA", module.RootCA, "error", err)
					http.Error(w, "Error loading root CA", http.StatusBadRequest)
					return
				}
				if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
					logger.Error("Error appending root CA to pool", "rootCA", module.RootCA)
				}
			}
			tlsConfig := &tls.Config{
				InsecureSkipVerify: module.InsecureSSL,
				RootCAs:            rootCAs,
			}
			httpClient.Transport = &http.Transport{TLSClientConfig: tlsConfig}
		}
		target.HttpClient = httpClient

		registry := prometheus.NewRegistry()
		eseriesCollector := collector.NewCollector(target, logger)

		// Register all sub-collectors
		for _, col := range eseriesCollector.Collectors {
			if err := registry.Register(col); err != nil {
				logger.Error("Collector registration failed", "collector", col, "error", err)
			}
		}

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

func setupLogger(levelStr, formatStr string) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{}

	switch levelStr {
	case "debug":
		opts.Level = slog.LevelDebug
	case "warn":
		opts.Level = slog.LevelWarn
	case "error":
		opts.Level = slog.LevelError
	default:
		opts.Level = slog.LevelInfo
	}

	if formatStr == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}
	return slog.New(handler)
}

func main() {
	kingpin.Version(version.Print("eseries_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := setupLogger(*logLevel, *logFormat)

	logger.Info("Starting eseries_exporter", "version", version.Info(), "build_context", version.BuildContext())

	sc := &config.SafeConfig{}
	if err := sc.ReloadConfig(*configFile); err != nil {
		logger.Error("Error loading config", "file", *configFile, "error", err)
		os.Exit(1)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/eseries", metricsHandler(sc.C, logger))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>E-Series Exporter</title></head>
			<body>
			<h1>E-Series Exporter</h1>
			<p><a href="/eseries">Run Prometheus Scrape</a></p>
			<p><a href="/metrics">Exporter Metrics</a></p>
			</body>
			</html>`))
	})

	logger.Info("Listening on", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		logger.Error("Error starting server", "error", err)
		os.Exit(1)
	}
}
