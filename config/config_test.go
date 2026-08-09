package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AJackTi/go-kafka/config"
)

func TestLoadReadsFileAndEnvironmentOverrides(t *testing.T) {
	path := writeConfig(t, validConfig)
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("GROUP_ID", "task-projector-test")
	t.Setenv("BROKERS", "kafka-a:9092,kafka-b:9092")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Env != "test" {
		t.Fatalf("App.Env = %q, want test", cfg.App.Env)
	}
	if cfg.HTTP.Port != "9090" {
		t.Fatalf("HTTP.Port = %q, want 9090", cfg.HTTP.Port)
	}
	if cfg.Kafka.GroupID != "task-projector-test" {
		t.Fatalf("Kafka.GroupID = %q, want task-projector-test", cfg.Kafka.GroupID)
	}

	wantBrokers := []string{"kafka-a:9092", "kafka-b:9092"}
	if strings.Join(cfg.Kafka.Brokers, ",") != strings.Join(wantBrokers, ",") {
		t.Fatalf("Kafka.Brokers = %v, want %v", cfg.Kafka.Brokers, wantBrokers)
	}
}

func TestNewConfigUsesConfigFileEnvironment(t *testing.T) {
	path := writeConfig(t, validConfig)
	t.Setenv("CONFIG_FILE", path)

	cfg, err := config.NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if cfg.App.Name != "go-kafka" {
		t.Fatalf("App.Name = %q, want go-kafka", cfg.App.Name)
	}
}

func TestLoadRejectsMissingKafkaBrokers(t *testing.T) {
	t.Parallel()

	path := writeConfig(t, strings.Replace(validConfig, "brokers: [\"localhost:9092\"]", "brokers: []", 1))

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "kafka brokers") {
		t.Fatalf("Load() error = %q, want kafka brokers validation", err)
	}
}

func TestLoadRejectsInvalidRuntimeValues(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"invalid port":      strings.Replace(validConfig, `port: "8000"`, `port: "not-a-port"`, 1),
		"invalid log level": strings.Replace(validConfig, `logLevel: "debug"`, `logLevel: "trace"`, 1),
		"blank broker":      strings.Replace(validConfig, `brokers: ["localhost:9092"]`, `brokers: [""]`, 1),
	}

	for name, contents := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := config.Load(writeConfig(t, contents))
			if err == nil {
				t.Fatal("Load() error = nil, want validation error")
			}
		})
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

const validConfig = "" +
	"app:\n" +
	"  name: \"go-kafka\"\n" +
	"  env: \"localhost\"\n" +
	"  version: \"0.1.0\"\n" +
	"http:\n" +
	"  port: \"8000\"\n" +
	"  cors: true\n" +
	"logger:\n" +
	"  logLevel: \"debug\"\n" +
	"mysql:\n" +
	"  url: \"mysql:password@tcp(localhost:3306)/go_kafka?parseTime=true\"\n" +
	"kafka:\n" +
	"  brokers: [\"localhost:9092\"]\n" +
	"  groupID: \"go-kafka\"\n" +
	"  initTopics: true\n" +
	"kafkaPublisherConfig:\n" +
	"  topic: \"event_created\"\n" +
	"  topicPrefix: \"eventStore\"\n" +
	"  partitions: 1\n" +
	"  replicationFactor: 1\n"
