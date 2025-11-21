// Copyright 2020 Trey Dockendorf
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

package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
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

// LoggerAdapter implements go-kit/log.Logger interface but uses slog internally
type LoggerAdapter struct {
	*slog.Logger
	prefix string
}

func (a *LoggerAdapter) Log(keyvals ...interface{}) error {
	if len(keyvals)%2 != 0 {
		return fmt.Errorf("keyvals must have even length")
	}

	// Extract level and message from keyvals
	level := slog.LevelInfo
	var msg string

	attrs := make([]any, 0)
	for i := 0; i < len(keyvals); i += 2 {
		key := keyvals[i].(string)
		val := keyvals[i+1]

		switch key {
		case "level":
			if levelStr, ok := val.(string); ok {
				switch levelStr {
				case "debug":
					level = slog.LevelDebug
				case "info":
					level = slog.LevelInfo
				case "warn":
					level = slog.LevelWarn
				case "error":
					level = slog.LevelError
				}
			}
		case "msg":
			if msgStr, ok := val.(string); ok {
				msg = msgStr
			}
		default:
			attrs = append(attrs, key, val)
		}
	}

	attrs = append(attrs, "prefix", a.prefix)
	a.Logger.Log(context.Background(), level, msg, attrs...)
	return nil
}

func NewCompatibleLogger(logger *slog.Logger, prefix string) log.Logger {
	return &LoggerAdapter{Logger: logger, prefix: prefix}
}

func metricsHandler(c *config.Config, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registry := prometheus.NewRegistry()

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
			http.Error(w, fmt.Sprintf("Unknown module %s", t), http.StatusNotFound)
			return
		}
		target := config.Target{
			Name:       t,
			User:       module.User,
			Password:   module.Password,
			ProxyURL:   module.ProxyURL,
			Collectors: module.Collectors,
		}
		proxyURL, err := url.Parse(module.ProxyURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to parse ProxyURL %s", module.ProxyURL), http.StatusNotFound)
			return
		} else {
			target.BaseURL = proxyURL
		}
		httpClient := &http.Client{
			Timeout: time.Duration(module.Timeout) * time.Second,
		}
		if proxyURL.Scheme == "https" {
			compatibleLogger := NewCompatibleLogger(getSlogLogger(), "https")
			compatibleLogger.Log("msg", "Setting up SSL transport", "url", module.ProxyURL)
			rootCAs, err := x509.SystemCertPool()
			if err != nil {
				compatibleLogger.Log("msg", "Error loading system cert pool, creating empty cert pool", "err", err)
				rootCAs = x509.NewCertPool()
			}
			if module.RootCA != "" {
				certs, err := os.ReadFile(module.RootCA)
				if err != nil {
					compatibleLogger.Log("msg", "Error loading root CA", "rootCA", module.RootCA, "err", err)
					http.Error(w, "Error loading root CA", http.StatusBadRequest)
					return
				}
				if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
					compatibleLogger.Log("msg", "Error appending root CA to pool", "rootCA", module.RootCA, "err", err)
				}
			}
			tlsConfig := &tls.Config{
				InsecureSkipVerify: module.InsecureSSL,
				RootCAs:            rootCAs,
			}
			tlsTransport := &http.Transport{TLSClientConfig: tlsConfig}
			httpClient.Transport = tlsTransport
		}
		target.HttpClient = httpClient
		eseriesCollector := collector.NewCollector(target, logger)
		for key, collector := range eseriesCollector.Collectors {
			logger.Log("msg", fmt.Sprintf("Enabled collector %s", key))
			registry.MustRegister(collector)
		}

		gatherers := prometheus.Gatherers{registry}

		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

// Global slog logger instance
var globalSlogLogger *slog.Logger

func getSlogLogger() *slog.Logger {
	return globalSlogLogger
}

func main() {
	metricsEndpoint := "/eseries"
	kingpin.Version(version.Print("eseries_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	// Setup structured logging
	var handler slog.Handler
	if *logFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLogLevel(*logLevel),
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLogLevel(*logLevel),
		})
	}

	logger := slog.New(handler)

	// Set global slog logger
	globalSlogLogger = logger

	compatibleLogger := NewCompatibleLogger(logger, "main")

	compatibleLogger.Log("msg", "Starting eseries_exporter", "version", version.Info())
	compatibleLogger.Log("msg", "Build context", "build_context", version.BuildContext())
	compatibleLogger.Log("msg", "Starting Server", "address", *listenAddress)

	sc := &config.SafeConfig{}

	if err := sc.ReloadConfig(*configFile); err != nil {
		compatibleLogger.Log("msg", "Error loading config", "err", err)
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//nolint:errcheck
		w.Write([]byte(`<html>
             <head><title>E-Series Exporter</title></head>
             <body>
             <h1>E-Series Exporter</h1>
             <p><a href='` + metricsEndpoint + `'>E-Series Metrics</a></p>
             <p><a href='/metrics'>Exporter Metrics</a></p>
             </body>
             </html>`))
	})
	http.Handle(metricsEndpoint, metricsHandler(sc.C, compatibleLogger))
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		compatibleLogger.Log("err", err)
		os.Exit(1)
	}
}

func getLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
