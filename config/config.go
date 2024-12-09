package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/mateusfdl/go-api/adapters/http"
	"github.com/mateusfdl/go-api/adapters/logger"
	"github.com/mateusfdl/go-api/adapters/mongo"
)

type AppConfig struct {
	Env    string
	Logger logger.Config
	HTTP   http.Config
	Mongo  mongo.Config
}

func NewAppConfig() (AppConfig, error) {
	env, err := getAndValidateEnv("ENV", []string{"development", "production", "test"})
	if err != nil {
		return AppConfig{}, err
	}

	loggerConfig, err := getLoggerConfig()
	if err != nil {
		return AppConfig{}, err
	}
	mongoConfig, err := getMongoConfig()
	if err != nil {
		return AppConfig{}, err
	}
	httpConfig, err := getHttpConfig()
	if err != nil {
		return AppConfig{}, err
	}

	return AppConfig{
		Env:    env,
		Logger: loggerConfig,
		HTTP:   httpConfig,
		Mongo:  mongoConfig,
	}, nil
}

func getLoggerConfig() (logger.Config, error) {
	level, err := getAndValidateEnv("LOG_LEVEL", []string{"debug", "info", "warn", "error"})
	if err != nil {
		return logger.Config{}, err
	}

	isSugared, err := getEnvAsBool("LOG_SUGARED", true)
	if err != nil {
		return logger.Config{}, err
	}

	return logger.Config{
		Level:   level,
		Sugared: isSugared,
	}, nil
}

func getHttpConfig() (http.Config, error) {
	port, err := getEnvAsInt("HTTP_PORT", 8080)
	if err != nil {
		return http.Config{}, err
	}

	timeout, err := getEnvAsInt("HTTP_TIMEOUT", 10)
	if err != nil {
		return http.Config{}, err
	}

	return http.Config{
		Port:    port,
		Timeout: timeout,
	}, nil
}

func getMongoConfig() (mongo.Config, error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		return mongo.Config{}, errors.New("environment variable MONGO_URI is not set")
	}
	if !strings.HasPrefix(uri, "mongodb://") {
		return mongo.Config{}, errors.New("invalid value for environment variable MONGO_URI: " + uri)
	}

	dbNames := os.Getenv("MONGO_DB_NAME")
	if dbNames == "" {
		return mongo.Config{}, errors.New("environment variable MONGO_DB_NAME is not set")
	}

	return mongo.Config{
		URI:    uri,
		DBName: dbNames,
	}, nil
}

func getAndValidateEnv(envName string, expected []string) (string, error) {
	value := os.Getenv(envName)
	if value == "" {
		return "", errors.New("environment variable " + envName + " is not set")
	}

	for _, e := range expected {
		if value == e {
			return value, nil
		}
	}

	return "", errors.New("invalid value for environment variable " + envName + ": " + value)
}

func getEnvAsBool(envName string, defaultValue bool) (bool, error) {
	value := os.Getenv(envName)
	if value == "" {
		return defaultValue, nil
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, errors.New("invalid boolean value for environment variable " + envName + ": " + value)
	}

	return boolValue, nil
}

func getEnvAsInt(envName string, defaultValue int) (int, error) {
	value := os.Getenv(envName)
	if value == "" {
		return defaultValue, nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("invalid integer value for environment variable " + envName + ": " + value)
	}

	return intValue, nil
}
