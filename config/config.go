package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		App                  `yaml:"app"`
		HTTP                 `yaml:"http"`
		Log                  `yaml:"logger"`
		PG                   `yaml:"postgres"`
		Kafka                `yaml:"kafka"`
		KafkaPublisherConfig `yaml:"kafkaPublisherConfig"`
	}

	// App -.
	App struct {
		Name    string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}

	// HTTP -.
	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" yaml:"logLevel"   env:"LOG_LEVEL"`
	}

	// PG -.
	PG struct {
		PoolMax int    `env-required:"true" yaml:"poolMax" env:"PG_POOL_MAX"`
		URL     string `env-required:"true"                 env:"PG_URL"`
	}

	Kafka struct {
		Brokers    []string `env-required:"true" yaml:"brokers" env:"BROKERS"`
		GroupID    string   `env-required:"true" yaml:"groupID" env:"BROKERS"`
		InitTopics bool     `env-required:"true" yaml:"initTopics" env:"INIT_TOPICS"`
	}

	KafkaPublisherConfig struct {
		Topic             string `env-required:"true" yaml:"topic" 				env:"TOPIC"`
		TopicPrefix       string `env-required:"true" yaml:"topicPrefix" 		env:"TOPIC_PREFIX"`
		Partitions        int    `env-required:"true" yaml:"partitions" 		env:"PARTITIONS"`
		ReplicationFactor int    `env-required:"true" yaml:"replicationFactor" env:"REPLICATION_FACTOR"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("../../config/config.yml", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
