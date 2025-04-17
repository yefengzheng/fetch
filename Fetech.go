package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Endpoint struct {
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
}

type DomainStats struct {
	Success int
	Total   int
}

var stats = make(map[string]*DomainStats)

func CheckHealth(endpoint Endpoint) {
	client := &http.Client{}
	// Default to GET if method not provided
	method := strings.ToUpper(endpoint.Method)
	if method == "" {
		method = "GET"
	}

	// Prepare request body
	reqBody := bytes.NewReader([]byte{})
	if method == "POST" || method == "PUT" || method == "PATCH" {
		reqBody = bytes.NewReader([]byte(endpoint.Body))
	}

	// Create request
	req, err := http.NewRequest(method, endpoint.URL, reqBody)
	if err != nil {
		log.Printf("Error creating request for %s: %v\n", endpoint.URL, err)
		return
	}

	// Set headers
	for key, value := range endpoint.Headers {
		req.Header.Set(key, value)
	}

	// Measure response time
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)

	domain, err := extractDomain(endpoint.URL)
	if err != nil {
		log.Printf("Error extracting domain from %s: %v\n", endpoint.URL, err)
	}
	if stats[domain] == nil {
		stats[domain] = &DomainStats{}
	}
	stats[domain].Total++

	if err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 && duration <= 500*time.Millisecond {
		stats[domain].Success++
	} else if err != nil {
		log.Printf("Request failed for %s: %v\n", endpoint.URL, err)
	}

	if resp != nil {
		resp.Body.Close()
	}
}

func extractDomain(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	host := u.Host
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx] // strip port
	}
	return host, nil
}

func monitorEndpoints(endpoints []Endpoint) {
	for _, endpoint := range endpoints {
		domain, err := extractDomain(endpoint.URL)
		if err != nil {
			log.Printf("Error extracting domain from %s: %v\n", endpoint.URL, err)
			return
		}
		if stats[domain] == nil {
			stats[domain] = &DomainStats{}
		}
	}

	for {
		for _, endpoint := range endpoints {
			CheckHealth(endpoint)
		}
		logResults()
		time.Sleep(15 * time.Second)
	}
}

func logResults() {
	for domain, stat := range stats {
		// Drop decimal part, no rounding
		percentage := 0
		if stat.Total > 0 {
			percentage = int(100 * stat.Success / stat.Total)
		}
		fmt.Printf("%s has %d%% availability\n", domain, percentage)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <config_file>")
	}

	filePath := os.Args[1]
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	var endpoints []Endpoint
	if err := yaml.Unmarshal(data, &endpoints); err != nil {
		log.Fatal("Error parsing YAML:", err)
	}

	monitorEndpoints(endpoints)
}
