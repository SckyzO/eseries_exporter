package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sckyzo/eseries_exporter/internal/config"
)

const (
	address = "localhost:19313"
)

func SetupServer() *config.Config {
	fixtureData, err := os.ReadFile("../../internal/collectors/testdata/drives.json")
	if err != nil {
		fmt.Printf("Error loading fixture data: %s", err.Error())
		os.Exit(1)
	}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write(fixtureData)
	}))
	sslServer := httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write(fixtureData)
	}))
	module := &config.Module{
		User:       "test",
		Password:   "test",
		Collectors: []string{"drives"},
		ProxyURL:   server.URL,
	}
	sslModule := &config.Module{
		User:        "test",
		Password:    "test",
		Collectors:  []string{"drives"},
		ProxyURL:    sslServer.URL,
		RootCA:      "../../internal/collectors/testdata/rootCA.crt",
		InsecureSSL: true,
	}
	sslBadModule := &config.Module{
		User:        "test",
		Password:    "test",
		Collectors:  []string{"drives"},
		ProxyURL:    sslServer.URL,
		RootCA:      "/dne",
		InsecureSSL: true,
	}
	c := &config.Config{}
	c.Modules = make(map[string]*config.Module)
	c.Modules["default"] = module
	c.Modules["ssl"] = sslModule
	c.Modules["ssl-error"] = sslBadModule
	return c
}

func TestMetricsHandler(t *testing.T) {
	c := SetupServer()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Channel to signal server is ready
	serverReady := make(chan bool)

	go func() {
		http.Handle("/eseries", metricsHandler(c, logger))
		// We can't easily signal readiness with ListenAndServe, so we'll just wait a bit
		// In a real refactor, main/server setup would be decoupled
		close(serverReady)
		err := http.ListenAndServe(address, nil)
		if err != nil && err != http.ErrServerClosed {
			// This might fail if port is taken, but for test logic we hope it's free
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	<-serverReady
	// Give it a tiny bit of time to actually bind
	time.Sleep(100 * time.Millisecond)

	body, err := queryExporter("target=test1", http.StatusOK)
	if err != nil {
		t.Fatalf("Unexpected error GET /eseries: %s", err.Error())
	}
	if !strings.Contains(body, "eseries_exporter_collect_error{collector=\"drives\"} 0") {
		t.Errorf("Unexpected value for eseries_exporter_collect_error")
	}

	body, err = queryExporter("target=test1&module=ssl", http.StatusOK)
	if err != nil {
		t.Fatalf("Unexpected error GET /eseries: %s", err.Error())
	}
	if !strings.Contains(body, "eseries_exporter_collect_error{collector=\"drives\"} 0") {
		t.Errorf("Unexpected value for eseries_exporter_collect_error")
	}

	_, _ = queryExporter("target=test1&module=ssl-error", http.StatusBadRequest)

	_, _ = queryExporter("", http.StatusBadRequest)

	_, _ = queryExporter("module=dne", http.StatusNotFound)
}

func queryExporter(param string, want int) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/eseries?%s", address, param))
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if err := resp.Body.Close(); err != nil {
		return "", err
	}
	if have := resp.StatusCode; want != have {
		return "", fmt.Errorf("want /eseries status code %d, have %d. Body:\n%s", want, have, b)
	}
	return string(b), nil
}
