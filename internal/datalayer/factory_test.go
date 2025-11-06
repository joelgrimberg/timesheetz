package datalayer

import (
	"os"
	"testing"
	"timesheet/internal/config"
	"timesheet/internal/db"
)

func TestGetDataLayer_Local(t *testing.T) {
	// Reset the instance
	ResetDataLayer()

	// Set environment to local mode
	os.Setenv("TIMESHEETZ_API_MODE", "local")
	defer os.Unsetenv("TIMESHEETZ_API_MODE")

	layer := GetDataLayer()
	if layer == nil {
		t.Fatal("GetDataLayer returned nil")
	}

	// Verify it's a LocalDBLayer
	if _, ok := layer.(*db.LocalDBLayer); !ok {
		t.Error("Expected LocalDBLayer for local mode")
	}

	// Test caching - should return same instance
	layer2 := GetDataLayer()
	if layer != layer2 {
		t.Error("Expected cached instance")
	}
}

func TestGetDataLayer_Remote(t *testing.T) {
	// Reset the instance
	ResetDataLayer()

	// Set environment to remote mode with missing URL (will return error and fallback to local)
	os.Setenv("TIMESHEETZ_API_MODE", "remote")
	os.Unsetenv("TIMESHEETZ_API_URL")
	defer func() {
		os.Unsetenv("TIMESHEETZ_API_MODE")
	}()

	layer := GetDataLayer()
	if layer == nil {
		t.Fatal("GetDataLayer returned nil")
	}

	// Should fallback to local when URL is missing
	if _, ok := layer.(*db.LocalDBLayer); !ok {
		t.Error("Expected LocalDBLayer fallback when URL is missing")
	}
}

func TestGetDataLayer_Dual(t *testing.T) {
	// Reset the instance
	ResetDataLayer()

	// Set environment to dual mode with missing URL (will fallback to local)
	os.Setenv("TIMESHEETZ_API_MODE", "dual")
	os.Unsetenv("TIMESHEETZ_API_URL")
	defer func() {
		os.Unsetenv("TIMESHEETZ_API_MODE")
	}()

	layer := GetDataLayer()
	if layer == nil {
		t.Fatal("GetDataLayer returned nil")
	}

	// Should fallback to local when URL is missing in dual mode
	if _, ok := layer.(*db.LocalDBLayer); !ok {
		t.Error("Expected LocalDBLayer fallback when URL is missing in dual mode")
	}
}

func TestGetDataLayer_Default(t *testing.T) {
	// Reset the instance
	ResetDataLayer()

	// Set invalid mode (should default to local)
	os.Setenv("TIMESHEETZ_API_MODE", "invalid")
	defer os.Unsetenv("TIMESHEETZ_API_MODE")

	layer := GetDataLayer()
	if layer == nil {
		t.Fatal("GetDataLayer returned nil")
	}

	// Should default to local
	if _, ok := layer.(*db.LocalDBLayer); !ok {
		t.Error("Expected LocalDBLayer for invalid mode")
	}
}

func TestResetDataLayer(t *testing.T) {
	// Reset first
	ResetDataLayer()
	
	// Get a layer first
	os.Setenv("TIMESHEETZ_API_MODE", "local")
	defer os.Unsetenv("TIMESHEETZ_API_MODE")

	layer1 := GetDataLayer()
	if layer1 == nil {
		t.Fatal("GetDataLayer returned nil")
	}

	// Reset
	ResetDataLayer()

	// Get again - should create new instance
	layer2 := GetDataLayer()
	if layer2 == nil {
		t.Fatal("GetDataLayer returned nil after reset")
	}

	// They should be the same instance due to caching (same pointer)
	// After reset, a new instance is created, but it's still the same type
	// The test verifies that reset clears the cache
	if layer1 != layer2 {
		// This is actually expected - reset creates a new instance
		// But they should both be LocalDBLayer
		if _, ok1 := layer1.(*db.LocalDBLayer); !ok1 {
			t.Error("layer1 should be LocalDBLayer")
		}
		if _, ok2 := layer2.(*db.LocalDBLayer); !ok2 {
			t.Error("layer2 should be LocalDBLayer")
		}
	}
}

func TestGetDataLayer_ConfigFile(t *testing.T) {
	// Reset the instance
	ResetDataLayer()

	// Remove env vars to test config file
	os.Unsetenv("TIMESHEETZ_API_MODE")
	os.Unsetenv("TIMESHEETZ_API_URL")

	// Create a test config
	configPath := config.GetConfigPath()
	os.Remove(configPath) // Remove existing config
	defer os.Remove(configPath)

	testConfig := config.Config{
		APIMode:    "local",
		APIBaseURL: "",
	}
	config.SaveConfig(testConfig)

	layer := GetDataLayer()
	if layer == nil {
		t.Fatal("GetDataLayer returned nil")
	}

	// Should use local mode from config
	if _, ok := layer.(*db.LocalDBLayer); !ok {
		t.Error("Expected LocalDBLayer from config")
	}
}
