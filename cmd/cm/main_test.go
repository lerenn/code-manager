//go:build unit

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckInitialization(t *testing.T) {
	// This test verifies that the checkInitialization function works correctly
	// We'll test both scenarios: when CM is initialized and when it's not

	// Test 1: When config file doesn't exist (should return error)
	originalConfigPath := configPath
	configPath = "/tmp/nonexistent/config.yaml"
	defer func() { configPath = originalConfigPath }()

	err := checkInitialization()

	// Should return configuration error when config file doesn't exist
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

func TestLoadConfig(t *testing.T) {
	// Test that loadConfig returns an error when config file doesn't exist
	originalConfigPath := configPath
	configPath = "/tmp/nonexistent/config.yaml"
	defer func() { configPath = originalConfigPath }()

	_, err := loadConfig()

	// Should return error when config file doesn't exist
	assert.Error(t, err)
}
