package config_test

import (
	"os"
	"testing"

	"github.com/mateusfdl/go-api/config"
)

func TestNewAppConfig(t *testing.T) {
	os.Setenv("ENV", "test")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_SUGARED", "true")
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("HTTP_PORT", "8080")
	os.Setenv("HTTP_TIMEOUT", "10")

	c, err := config.NewAppConfig()

	if err != nil {
		t.Fatalf("NewAppConfig() failed: %v", err)
	}

	if c.Env != "test" {
		t.Errorf("Expected env to be 'test', but got '%s'", c.Env)
	}

	if c.Logger.Level != "debug" {
		t.Errorf("Expected logger level to be 'debug', but got '%s'", c.Logger.Level)
	}

	if c.Logger.Sugared != true {
		t.Errorf("Expected logger sugared to be true, but got '%t'", c.Logger.Sugared)
	}

	if c.Mongo.URI != "mongodb://localhost:27017" {
		t.Errorf("Expected mongo URI to be 'mongodb://localhost:27017', but got '%s'", c.Mongo.URI)
	}

	if c.HTTP.Port != 8080 {
		t.Errorf("Expected http port to be 8080, but got '%d'", c.HTTP.Port)
	}

	if c.HTTP.Timeout != 10 {

		t.Errorf("Expected http timeout to be 10, but got '%d'", c.HTTP.Timeout)
	}
}

func TestEnvNotSet(t *testing.T) {
	os.Clearenv()

	_, err := config.NewAppConfig()

	if err == nil {
		t.Fatalf("Expected not set env error, but got nil")
	}
}

func TestInvalidEnv(t *testing.T) {
	os.Setenv("ENV", "invalid")

	_, err := config.NewAppConfig()
	if err == nil {
		t.Fatalf("Expected invalid env error , but got nil")
	}
}

func TestInvalidLogLevel(t *testing.T) {
	os.Setenv("ENV", "test")
	os.Setenv("LOG_LEVEL", "invalid")

	_, err := config.NewAppConfig()

	if err == nil {
		t.Fatalf("Expected invalid log level error, but got nil")
	}
}

func TestInvalidLogSugared(t *testing.T) {
	os.Setenv("ENV", "test")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_SUGARED", "invalid")

	_, err := config.NewAppConfig()

	if err == nil {
		t.Fatalf("Expected invalid sugared flag error, but got nil")
	}
}

func TestInvalidHttpPort(t *testing.T) {
	os.Setenv("ENV", "test")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_SUGARED", "true")
	os.Setenv("HTTP_PORT", "invalid")

	_, err := config.NewAppConfig()

	if err == nil {
		t.Fatalf("Expected invalid http port error, but got nil")
	}
}

func TestInvalidHttpTimeout(t *testing.T) {
	os.Setenv("ENV", "test")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_SUGARED", "true")
	os.Setenv("HTTP_PORT", "8080")
	os.Setenv("HTTP_TIMEOUT", "invalid")

	_, err := config.NewAppConfig()

	if err == nil {
		t.Fatalf("Expected invalid http timeout, but got nil")
	}
}

func TestInvalidMongoURI(t *testing.T) {
	os.Setenv("ENV", "test")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_SUGARED", "true")
	os.Setenv("HTTP_PORT", "8080")
	os.Setenv("HTTP_TIMEOUT", "10")
	os.Setenv("MONGO_URI", "invalid")

	_, err := config.NewAppConfig()

	if err == nil {
		t.Fatalf("Expected invalid mongo URI error, but got nil")
	}
}
