// Package configparser responsible for parsing yaml config
// into Config struct which contains load balancer settings
package configparser

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"weightogo/loadbalancer"
	"weightogo/logger"

	"gopkg.in/yaml.v3"
)

// Config contains loadbalancer settings parsed from yaml config file
type Config struct {
	General struct {
		BindAddress string        `yaml:"bind_address"`
		LogLevel    string        `yaml:"log_level"`
		MaxConn     int           `yaml:"max_connections"`
		ConnTimeout time.Duration `yaml:"connection_timeout"`
	}
	BackendServers []BackendServer       `yaml:"backend_servers"`
	Strategy       loadbalancer.Strategy `yaml:"strategy"`
	HealthCheck    HealthCheck           `yaml:"health_check"`
}

// BackendServer contains loadbalancer settings for each individual server
type BackendServer struct {
	Address           string        `yaml:"address"`
	Weight            int           `yaml:"weight"`
	MaxConn           int           `yaml:"max_connections"`
	HCEndpoint        string        `yaml:"hc_endpoint"`
	HCInterval        time.Duration `yaml:"hc_interval"`
	CurrentWeight     int
	ActiveConnections int64
}

// HealthCheck contains healthcheck settings
type HealthCheck struct {
	Enabled            bool          `yaml:"enabled"`
	Interval           time.Duration `yaml:"interval"`
	Timeout            time.Duration `yaml:"timeout"`
	UnhealthyThreshold int           `yaml:"unhealthy_threshold"`
	HealthyThreshold   int           `yaml:"healthy_threshold"`
}

// ParseConfig parses yaml load balancer config into Config struct
// Returns Config struct and error
func ParseConfig(p string) (Config, error) {
	path := filepath.Join("/", "etc", "weightogo", "config.yaml")
	if p != "" {
		path = p
	}
	isTest := os.Getenv("WEIGHTOGO_TEST")

	if isTest == "true" {
		path = "mock_config.yaml"
	}

	file, err := os.Open(path)
	if err != nil {
		logger.Logger.Error("Cannot open config file.")
		return Config{}, err
	}
	content, err := io.ReadAll(file)
	if err != nil {
		logger.Logger.Error("Cannot read config file.")
		return Config{}, err
	}

	config := Config{}

	err = yaml.Unmarshal(content, &config)
	if err != nil {
		logger.Logger.Error("Cannot unmarshal config file.")
		return Config{}, err
	}

	if err := validateConfig(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func validateConfig(cfg *Config) error {
	if cfg.General.BindAddress == "" {
		return errors.New("bind address wasn't provided")
	}

	if len(cfg.BackendServers) == 0 {
		return errors.New("backend servers wasn't provided")
	}

	for i, s := range cfg.BackendServers {
		if s.Address == "" {
			return fmt.Errorf("miss address for backend server #%d", i)
		}
	}

	if !isValidStrategy(cfg.Strategy) {
		validStrategies := fmt.Sprintf("%s, %s, %s", loadbalancer.RoundRobin, loadbalancer.WeightedRoundRobin, loadbalancer.LeastConnections)
		return fmt.Errorf("provided strategy %s is invalid. Valid strategies: %s", cfg.Strategy, validStrategies)
	}

	return nil
}

func isValidStrategy(strategy loadbalancer.Strategy) bool {
	switch strategy {
	case loadbalancer.RoundRobin, loadbalancer.WeightedRoundRobin, loadbalancer.LeastConnections:
		return true
	default:
		return false
	}
}
