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

type BackendServer struct {
	Address           string        `yaml:"address"`
	Weight            int           `yaml:"weight"`
	MaxConn           int           `yaml:"max_connections"`
	HC_Endpoint       string        `yaml:"hc_endpoint"`
	HC_Interval       time.Duration `yaml:"hc_interval"`
	CurrentWeight     int
	ActiveConnections int64
}

type HealthCheck struct {
	Enabled            bool          `yaml:"enabled"`
	Interval           time.Duration `yaml:"interval"`
	Timeout            time.Duration `yaml:"timeout"`
	UnhealthyThreshold int           `yaml:"unhealthy_threshold"`
	HealthyThreshold   int           `yaml:"healthy_threshold"`
}

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
		return errors.New("Bind address wasn't provided.")
	}

	if len(cfg.BackendServers) == 0 {
		return errors.New("Backend servers wasn't provided.")
	}

	for i, s := range cfg.BackendServers {
		if s.Address == "" {
			return errors.New(fmt.Sprintf("Miss address for backend server #%d.", i))
		}
	}

	if !isValidStrategy(cfg.Strategy) {
		validStrategies := fmt.Sprintf("%s, %s, %s", loadbalancer.RoundRobin, loadbalancer.WeightedRoundRobin, loadbalancer.LeastConnections)
		return errors.New(fmt.Sprintf("Provided strategy %s is invalid. Valid strategies: %s.", cfg.Strategy, validStrategies))
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
