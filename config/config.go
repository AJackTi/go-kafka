package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

const DefaultPath = "config/config.yml"

type (
	Config struct {
		App    App                  `yaml:"app"`
		HTTP   HTTP                 `yaml:"http"`
		Log    Log                  `yaml:"logger"`
		MySQL  MySQL                `yaml:"mysql"`
		Kafka  Kafka                `yaml:"kafka"`
		Events KafkaPublisherConfig `yaml:"kafkaPublisherConfig"`
	}

	App struct {
		Name    string `yaml:"name" env:"APP_NAME"`
		Env     string `yaml:"env" env:"APP_ENV"`
		Version string `yaml:"version" env:"APP_VERSION"`
	}

	HTTP struct {
		Port string `yaml:"port" env:"HTTP_PORT"`
		Cors bool   `yaml:"cors" env:"HTTP_CORS"`
	}

	Log struct {
		Level string `yaml:"logLevel" env:"LOG_LEVEL"`
	}

	MySQL struct {
		URL string `yaml:"url" env:"MYSQL_URL"`
	}

	Kafka struct {
		Brokers    []string `yaml:"brokers" env:"BROKERS" env-separator:","`
		GroupID    string   `yaml:"groupID" env:"GROUP_ID"`
		InitTopics bool     `yaml:"initTopics" env:"INIT_TOPICS"`
	}

	KafkaPublisherConfig struct {
		Topic             string `yaml:"topic" env:"TOPIC"`
		TopicPrefix       string `yaml:"topicPrefix" env:"TOPIC_PREFIX"`
		Partitions        int    `yaml:"partitions" env:"PARTITIONS"`
		ReplicationFactor int    `yaml:"replicationFactor" env:"REPLICATION_FACTOR"`
	}
)

func NewConfig() (*Config, error) {
	path := strings.TrimSpace(os.Getenv("CONFIG_FILE"))
	if path == "" {
		path = DefaultPath
	}

	return Load(path)
}

func Load(path string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read config environment: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg Config) Validate() error {
	switch {
	case strings.TrimSpace(cfg.App.Name) == "":
		return fmt.Errorf("validate config: app name is required")
	case strings.TrimSpace(cfg.App.Env) == "":
		return fmt.Errorf("validate config: app environment is required")
	case strings.TrimSpace(cfg.App.Version) == "":
		return fmt.Errorf("validate config: app version is required")
	case !validPort(cfg.HTTP.Port):
		return fmt.Errorf("validate config: http port must be between 1 and 65535")
	case !validLogLevel(cfg.Log.Level):
		return fmt.Errorf("validate config: log level must be debug, info, warn, or error")
	case strings.TrimSpace(cfg.MySQL.URL) == "":
		return fmt.Errorf("validate config: mysql url is required")
	case len(cfg.Kafka.Brokers) == 0:
		return fmt.Errorf("validate config: kafka brokers are required")
	case hasBlank(cfg.Kafka.Brokers):
		return fmt.Errorf("validate config: kafka brokers cannot contain blanks")
	case strings.TrimSpace(cfg.Kafka.GroupID) == "":
		return fmt.Errorf("validate config: kafka group id is required")
	case strings.TrimSpace(cfg.Events.Topic) == "":
		return fmt.Errorf("validate config: kafka topic is required")
	case strings.TrimSpace(cfg.Events.TopicPrefix) == "":
		return fmt.Errorf("validate config: kafka topic prefix is required")
	case cfg.Events.Partitions < 1:
		return fmt.Errorf("validate config: kafka partitions must be at least 1")
	case cfg.Events.ReplicationFactor < 1:
		return fmt.Errorf("validate config: kafka replication factor must be at least 1")
	default:
		return nil
	}
}

func validPort(value string) bool {
	port, err := strconv.Atoi(strings.TrimSpace(value))
	return err == nil && port > 0 && port <= 65535
}

func validLogLevel(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

func hasBlank(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}
	return false
}
