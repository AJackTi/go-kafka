package logger_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"

	"github.com/AJackTi/go-kafka/pkg/logger"
)

func TestLoggerPreservesSeverity(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	log := logger.NewWithWriter("debug", &output)

	log.Debug("debug message")
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	var levels []string
	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		var entry map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("decode log entry: %v", err)
		}
		level, ok := entry["level"].(string)
		if !ok {
			t.Fatalf("log entry has no level: %v", entry)
		}
		levels = append(levels, level)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan log output: %v", err)
	}

	want := []string{"debug", "info", "warn", "error"}
	if len(levels) != len(want) {
		t.Fatalf("levels = %v, want %v", levels, want)
	}
	for i := range want {
		if levels[i] != want[i] {
			t.Fatalf("levels[%d] = %q, want %q (all levels: %v)", i, levels[i], want[i], levels)
		}
	}
}
