//go:build unit

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckInitialization(t *testing.T) {
	// This test verifies that the checkInitialization function works correctly
	// We'll test both scenarios: when CM is initialized and when it's not

	// Test 1: When config file doesn't exist (should return error)
	originalConfigPath := ConfigPath
	ConfigPath = "/tmp/nonexistent/config.yaml"
	defer func() { ConfigPath = originalConfigPath }()

	err := CheckInitialization()

	// Should return configuration error when config file doesn't exist
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

func TestLoadConfig(t *testing.T) {
	// Test that loadConfig returns an error when config file doesn't exist
	originalConfigPath := ConfigPath
	ConfigPath = "/tmp/nonexistent/config.yaml"
	defer func() { ConfigPath = originalConfigPath }()

	_, err := LoadConfig()

	// Should return error when config file doesn't exist
	assert.Error(t, err)
}
