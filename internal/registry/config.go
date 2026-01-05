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

	overrideFromEnv("CONTENT_SERVICE_GRPC_ADDR", &cfg.Services.Content.GRPC.Addr)
	overrideFromEnv("CONTENT_SERVICE_HTTP_ADDR", &cfg.Services.Content.HTTP.Addr)
	overrideFromEnv("RABBITMQ_URL", &cfg.RabbitMQ.URL)

	return &cfg, nil
}

func overrideFromEnv(envKey string, configValue *string) {
	if value, exists := os.LookupEnv(envKey); exists && value != "" {
		*configValue = value
	}
}
