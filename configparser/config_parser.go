package configparser

import (
	"fmt"
	"io"
	"os"
	"time"
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
	BackendServers []BackendServer `yaml:"backend_servers"`
	Strategy       string          `yaml:"strategy"`
	HealthCheck    HealthCheck     `yaml:"health_check"`
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

func ParseConfig() (Config, error) {
	file, err := os.Open("config.yaml")
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
	fmt.Printf("CONFIG:\n%+v\n\n", config)

	return config, nil
}
