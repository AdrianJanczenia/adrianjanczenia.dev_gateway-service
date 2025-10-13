package registry

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		HTTPPort string `yaml:"httpPort"`
	}
	Services struct {
		Content struct {
			GRPC struct {
				Addr string `yaml:"addr"`
			} `yaml:"grpc"`
			HTTP struct {
				Addr string `yaml:"addr"`
			} `yaml:"http"`
		} `yaml:"content"`
	}
	RabbitMQ struct {
		URL      string `yaml:"url"`
		Topology struct {
			Exchange            string `yaml:"exchange"`
			CVRequestRoutingKey string `yaml:"cv_request_routing_key"`
		} `yaml:"topology"`
	} `yaml:"rabbitmq"`
}

var Cfg *Config

func LoadConfig() (*Config, error) {
	env := os.Getenv("APP_ENV")
	if env != "production" {
		env = "local"
	}
	configPath := filepath.Join("config", env, "config.yml")
	log.Printf("INFO: loading configuration from %s", configPath)

	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
