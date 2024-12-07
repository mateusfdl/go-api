package logger

import (
	"fmt"
	"log"
	"strings"
)

type Logger struct {
	Level   string
	Sugared bool
}

func New(cfg *Config) *Logger {
	return &Logger{
		Level:   strings.ToLower(cfg.Level),
		Sugared: cfg.Sugared,
	}
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	if l.shouldLog("debug") {
		l.log("DEBUG", "\x1b[36m", msg, fields...)
	}
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	if l.shouldLog("info") {
		l.log("INFO", "\x1b[32m", msg, fields...)
	}
}

func (l *Logger) Warn(msg string, fields ...interface{}) {
	if l.shouldLog("warn") {
		l.log("WARN", "\x1b[33m", msg, fields...)
	}
}

func (l *Logger) Error(msg string, fields ...interface{}) {
	if l.shouldLog("error") {
		l.log("ERROR", "\x1b[31m", msg, fields...)
	}
}

// log handles the actual logging, applying colors if Sugared is enabled
func (l *Logger) log(level, color, msg string, fields ...interface{}) {
	if l.Sugared {
		log.Printf("%s[%s]\x1b[0m %s %s", color, level, msg, formatFields(fields...))
	} else {
		log.Printf("[%s] %s %s", level, msg, formatFields(fields...))
	}
}

// shouldLog checks if the message level should be logged based on the configured log level
func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{"debug": 1, "info": 2, "warn": 3, "error": 4}
	configLevel, ok := levels[l.Level]
	if !ok {
		configLevel = 1 // Default to the highest verbosity if an invalid level is configured
	}
	messageLevel := levels[level]
	return messageLevel >= configLevel
}

// Helper function to format fields as key=value pairs
func formatFields(fields ...interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	var builder strings.Builder
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			builder.WriteString(fmt.Sprintf("%v=%v ", fields[i], fields[i+1]))
		} else {
			builder.WriteString(fmt.Sprintf("%v ", fields[i]))
		}
	}
	return strings.TrimSpace(builder.String())
}
