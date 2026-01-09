package registry

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		HTTPPort string `yaml:"httpPort"`
	}
	Infrastructure struct {
		Retry struct {
			MaxAttempts  int           `yaml:"maxAttempts"`
			DelaySeconds time.Duration `yaml:"delaySeconds"`
		} `yaml:"retry"`
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
		Captcha struct {
			HTTP struct {
				Addr string `yaml:"addr"`
			} `yaml:"http"`
		} `yaml:"captcha"`
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
	type yamlConfig struct {
		Server struct {
			HTTPPort string `yaml:"httpPort"`
		} `yaml:"server"`
		Infrastructure struct {
			Retry struct {
				MaxAttempts  int `yaml:"maxAttempts"`
				DelaySeconds int `yaml:"delaySeconds"`
			} `yaml:"retry"`
		} `yaml:"infrastructure"`
		Services struct {
			Content struct {
				GRPC struct {
					Addr string `yaml:"addr"`
				} `yaml:"grpc"`
				HTTP struct {
					Addr string `yaml:"addr"`
				} `yaml:"http"`
			} `yaml:"content"`
			Captcha struct {
				HTTP struct {
					Addr string `yaml:"addr"`
				} `yaml:"http"`
			} `yaml:"captcha"`
		} `yaml:"services"`
		RabbitMQ struct {
			URL      string `yaml:"url"`
			Topology struct {
				Exchange            string `yaml:"exchange"`
				CVRequestRoutingKey string `yaml:"cv_request_routing_key"`
			} `yaml:"topology"`
		} `yaml:"rabbitmq"`
	}

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

	var yc yamlConfig
	if err := yaml.NewDecoder(f).Decode(&yc); err != nil {
		return nil, err
	}

	cfg := &Config{}
	cfg.Server.HTTPPort = yc.Server.HTTPPort
	cfg.Infrastructure.Retry.MaxAttempts = yc.Infrastructure.Retry.MaxAttempts
	cfg.Infrastructure.Retry.DelaySeconds = time.Duration(yc.Infrastructure.Retry.DelaySeconds) * time.Second
	cfg.Services.Content.GRPC.Addr = yc.Services.Content.GRPC.Addr
	cfg.Services.Content.HTTP.Addr = yc.Services.Content.HTTP.Addr
	cfg.Services.Captcha.HTTP.Addr = yc.Services.Captcha.HTTP.Addr
	cfg.RabbitMQ.URL = yc.RabbitMQ.URL
	cfg.RabbitMQ.Topology = yc.RabbitMQ.Topology

	overrideFromEnv("CONTENT_SERVICE_GRPC_ADDR", &cfg.Services.Content.GRPC.Addr)
	overrideFromEnv("CONTENT_SERVICE_HTTP_ADDR", &cfg.Services.Content.HTTP.Addr)
	overrideFromEnv("CAPTCHA_SERVICE_HTTP_ADDR", &cfg.Services.Captcha.HTTP.Addr)
	overrideFromEnv("RABBITMQ_URL", &cfg.RabbitMQ.URL)

	return cfg, nil
}

func overrideFromEnv(envKey string, configValue *string) {
	if value, exists := os.LookupEnv(envKey); exists && value != "" {
		*configValue = value
	}
}
