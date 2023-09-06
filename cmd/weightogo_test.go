package main

import (
	"net/http"
	"strings"
	"sync"
	"testing"
	"weightogo/configparser"
	"weightogo/loadbalancer"
)

func TestIntegration(t *testing.T) {
	config, err := configparser.ParseConfig()
	if err != nil {
		t.Fatalf("Cannot parse config %v", err)
	}

	servers := parseServers(config.BackendServers)
	var wg sync.WaitGroup
	go startServers(servers, &wg)
	wg.Wait()

	go main()

	testData := "test_data"
	for i := 0; i <= 10; i++ {
		testLoadBalancer(t, config.General.BindAddress, testData)
	}
}

func TestHealthcheckIntegration(t *testing.T) {
	config, err := configparser.ParseConfig()
	if err != nil {
		t.Fatalf("Cannot parse config %v", err)
	}

	servers := parseServers(config.BackendServers)
	// Not running third server
	var wg sync.WaitGroup
	go startServers(servers[:2], &wg)
	wg.Wait()

	go main()

	healthyServers := getHealthyServers(servers)
	if len(healthyServers) != 2 {
		t.Fatalf("Expected 2 healthy servers, got %d", len(healthyServers))
	}
	for _, server := range healthyServers {
		if server.Address == "http://localhost:8002" {
			t.Fatal("localhost:8002 should not be marked as healthy")
		}
	}
}

func startServers(servers []loadbalancer.Server, wg *sync.WaitGroup) {
	for _, s := range servers {
		wg.Add(1)
		go func(address string) {
			defer wg.Done()
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				var body = make([]byte, 1000)
				r.Body.Read(body)
				w.Write([]byte(body))
			})

			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
			})

			http.ListenAndServe(address, mux)
		}(s.Address)
	}
}

func testLoadBalancer(t *testing.T, address string, testData string) {
	r := strings.NewReader(testData)
	req, err := http.NewRequest(http.MethodPost, "http://"+address, r)
	if err != nil {
		t.Fatalf("Cannot create new request: %v", err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error sending request %v", err)
	}

	respBody := make([]byte, len(testData))
	res.Body.Read(respBody)
	if string(respBody) != string(testData) {
		t.Fatalf("Response is not equal to testData. expected=%s, got=%s.", string(testData), string(respBody))
	}
}
