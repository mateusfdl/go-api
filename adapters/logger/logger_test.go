package logger_test

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/mateusfdl/go-api/adapters/logger"
)

func TestSugaredLogging(t *testing.T) {
	cfg := logger.Config{Level: "debug", Sugared: true}
	logInstance := logger.New(cfg)

	output := catchSTDOut(func() {
		logInstance.Debug("Debug message", "key1", "value1", "key2", "value2")
		logInstance.Info("Info message", "key1", "value1")
		logInstance.Warn("Warning message", "key", "value")
		logInstance.Error("Error message")
	})

	if !containsAll(output, []string{"DEBUG", "Info message key1=value1", "Warning message key=value"}) {
		t.Errorf("Sugared logging did not format correctly: %s", output)
	}
}

func TestJSONLogging(t *testing.T) {
	cfg := logger.Config{Level: "info", Sugared: false}
	logInstance := logger.New(cfg)

	output := catchSTDOut(func() {
		logInstance.Debug("Debug message")
		logInstance.Info("Info message", "key", "value")
		logInstance.Warn("Warning message", "key1", "value1", "key2", "value2")
		logInstance.Error("Error message", "key", "value")
	})

	if contains(output, "Debug message") {
		t.Error("JSON logging should not include Debug message when level is set to info")
	}

	if !containsAll(output, []string{"\"level\": \"INFO\"", "\"message\": \"Info message\"", "\"key\": \"value\""}) {
		t.Errorf("JSON logging did not format correctly: %s", output)
	}
}

func TestLogLevelFiltering(t *testing.T) {
	cases := []struct {
		level    string
		expected bool
		assert   func(string) bool
	}{
		{"debug", true, func(s string) bool { return containsAll(s, []string{"DEBUG", "INFO", "WARN", "ERROR"}) }},
		{"info", true, func(s string) bool { return containsAll(s, []string{"INFO", "WARN", "ERROR"}) }},
		{"warn", true, func(s string) bool { return containsAll(s, []string{"WARN", "ERROR"}) }},
		{"error", true, func(s string) bool { return contains(s, "ERROR") }},
	}

	for _, c := range cases {
		cfg := logger.Config{Level: c.level, Sugared: true}
		logInstance := logger.New(cfg)

		output := catchSTDOut(func() {
			logInstance.Debug("Debug message")
			logInstance.Info("Info message")
			logInstance.Warn("Warning message")
			logInstance.Error("Error message")
		})

		if c.expected != c.assert(output) {
			t.Errorf("Log level filtering did not work as expected for level %s", c.level)
		}
	}
}

func containsAll(s string, substrings []string) bool {
	for _, substring := range substrings {
		if !contains(s, substring) {
			return false
		}
	}
	return true
}

func contains(s, substring string) bool {
	return strings.Contains(s, substring)
}

func catchSTDOut(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	f()
	return buf.String()
}
