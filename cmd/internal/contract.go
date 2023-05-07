package internal

import (
	"crypto/rsa"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	Coreum = "coreum"
)

type AppConfig struct {
	Port            string
	TokenTimeToLive int64
	PrivateKey      *rsa.PrivateKey
	PublicKey       *rsa.PublicKey
	Interval		time.Duration
}

// MustString func returns environment variable value as a string value,
// If variable doesn't exist or is not set, exits from the runtime
func MustString(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("required ENV %q is not set", key)
	}
	if value == "" {
		log.Fatalf("required ENV %q is empty", key)
	}
	return value
}

// GetString func returns environment variable value as a string value,
// If variable doesn't exist or is not set, returns fallback value
func GetString(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}

// MustInt func returns environment variable value as an integer value,
// If variable doesn't exist or is not set, exits from the runtime
func MustInt(key string) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("required ENV %q is not set", key)
	}
	if value == "" {
		log.Fatalf("required ENV %q is empty", key)
	}
	res, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		log.Fatalf("required ENV %q must be a number but it's %q", key, value)
	}
	return int(res)
}

// GetInt func returns environment variable value as a integer value,
// If variable doesn't exist or is not set, returns fallback value
func GetInt(key string, fallback int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	res, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return fallback
	}
	return int(res)
}

// GetFloat func returns environment variable value as a integer value,
// If variable doesn't exist or is not set, returns fallback value
func GetFloat(key string, fallback int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	res, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return int(res)
}
